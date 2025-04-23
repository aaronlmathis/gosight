// server/internal/alerts/manager.go
package alerts

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/dispatcher"
	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/server/internal/store/alertstore"
	"github.com/aaronlmathis/gosight/server/internal/websocket"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/google/uuid"
)

type Manager struct {
	lock       sync.RWMutex
	active     map[string]*model.AlertInstance // key: ruleID|endpointID
	emitter    *events.Emitter
	dispatcher *dispatcher.Dispatcher
	store      alertstore.AlertStore
	hub        *websocket.Hub
}

func NewManager(emitter *events.Emitter, dispatcher *dispatcher.Dispatcher, store alertstore.AlertStore, hub *websocket.Hub) *Manager {
	return &Manager{
		active:     make(map[string]*model.AlertInstance),
		emitter:    emitter,
		dispatcher: dispatcher,
		store:      store,
		hub:        hub,
	}
}

func key(ruleID, endpointID string) string {
	return ruleID + "|" + endpointID
}

func (m *Manager) HandleState(
	ctx context.Context,
	rule model.AlertRule,
	meta *model.Meta,
	value float64,
	triggered bool,
) {
	k := key(rule.ID, meta.EndpointID)
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	current := m.active[k]

	if triggered {
		// If already active, apply cooldown logic
		if current != nil {
			if rule.RepeatInterval == 0 || now.Sub(current.LastFired) < rule.RepeatInterval {
				return
			}
			current.LastFired = now
			current.LastValue = value
			// add to alert store
			_ = m.store.UpsertAlert(ctx, current)
			// broadcast to websocket clients
			utils.Debug("ReBroadcasting from alertmgr: %s", current.ID)
			m.hub.BroadcastAlert(*current)
			// emit alert firing event
			m.emitAlertFiringEvent(ctx, rule, meta, now)

			// dispatch Action
			m.dispatcher.Dispatch(ctx, model.EventEntry{
				Timestamp: now,
				Level:     rule.Level,
				Category:  "alert",
				Message:   rule.Message,
				Source:    rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
				Scope:     current.Scope,
				Target:    current.Target,
				Meta:      meta.Tags,
			})
			return
		}
		// Assume its either an alert on an endpoint or global TODO: Add user related alerts logic.
		scope := "global"
		target := "gosight-core"

		if meta.EndpointID != "" {
			scope = "endpoint"
			target = meta.EndpointID
		}

		// New alert firing
		inst := &model.AlertInstance{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			State:      "firing",
			Previous:   "ok",
			Scope:      scope,  // or infer from meta/job/tag
			Target:     target, // the actual target
			FirstFired: now,
			LastFired:  now,
			LastOK:     now, // can set to now on first fire
			LastValue:  value,
			Level:      rule.Level,
			Message:    rule.Message,
			Labels:     meta.Tags, // contains env, team, agent_id, etc.
		}

		m.active[k] = inst

		// Add to alert store
		_ = m.store.UpsertAlert(ctx, inst)

		// Broadcast to websocket clients
		m.hub.BroadcastAlert(*inst)

		// emit alert firing event
		m.emitAlertFiringEvent(ctx, rule, meta, now)
	} else {
		// Transition from firing to resolved
		if current != nil {
			delete(m.active, k)
			m.emitAlertResolvedEvent(ctx, rule, meta, now)
			if rule.NotifyOnResolve {
				m.dispatcher.Dispatch(ctx, model.EventEntry{
					Timestamp: now,
					Level:     rule.Level,
					Category:  "alert",
					Message:   "Resolved: " + rule.Message,
					Source:    rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
					Target:    current.Target, // TODO: -

					Meta: meta.Tags,
				})
			}
			// add to alert store
			_ = m.store.ResolveAlert(ctx, rule.ID, current.Target, now)

			// broadcast to websocket clients
			m.hub.BroadcastAlert(*current)
		}
	}
}

func (m *Manager) emitAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	// Assume its either an alert on an endpoint or global TODO: Add user related alerts logic.
	scope := "global"
	target := "gosight-core"

	if meta.EndpointID != "" {
		scope = "endpoint"
		target = meta.EndpointID
	}
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   rule.Message,
		Source:    rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      meta.Tags,
	}
	event.Meta["rule_id"] = rule.ID

	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

func (m *Manager) emitAlertResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	// Assume its either an alert on an endpoint or global TODO: Add user related alerts logic.
	scope := "global"
	target := "gosight-core"
	m.emitter.Emit(ctx, model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   "Resolved: " + rule.Message,
		Source:    rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      meta.Tags,
	})
}

func (m *Manager) ListActive() []model.AlertInstance {
	m.lock.RLock()
	defer m.lock.RUnlock()

	var list []model.AlertInstance
	for _, v := range m.active {
		list = append(list, *v)
	}
	return list
}
