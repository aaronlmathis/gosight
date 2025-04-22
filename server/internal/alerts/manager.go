// server/internal/alerts/manager.go
package alerts

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/dispatcher"
	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/shared/model"
)

type Manager struct {
	lock       sync.RWMutex
	active     map[string]*model.AlertInstance // key: ruleID|endpointID
	emitter    *events.Emitter
	dispatcher *dispatcher.Dispatcher
}

func NewManager(emitter *events.Emitter, dispatcher *dispatcher.Dispatcher) *Manager {
	return &Manager{
		active:     make(map[string]*model.AlertInstance),
		emitter:    emitter,
		dispatcher: dispatcher,
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

			if rule.Cooldown > 0 && now.Sub(current.LastFired) < rule.Cooldown {
				return // Suppress re-alert within cooldown
			}
			current.LastFired = now
			current.LastValue = value
			return
		}

		// New alert firing
		inst := &model.AlertInstance{
			RuleID:     rule.ID,
			EndpointID: meta.EndpointID,
			State:      "firing",
			Previous:   "ok",
			FirstFired: now,
			LastFired:  now,
			LastValue:  value,
			Labels:     meta.Tags,
			Message:    rule.Message,
			Level:      rule.Level,
		}
		m.active[k] = inst

		m.emitAlertFiringEvent(ctx, rule, meta, now)
	} else {
		// Transition from firing to resolved
		if current != nil {
			delete(m.active, k)
			m.emitAlertResolvedEvent(ctx, rule, meta, now)
		}
	}
}

func (m *Manager) emitAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	event := model.EventEntry{
		Timestamp:  now,
		Level:      rule.Level,
		Category:   "alert",
		Message:    rule.Message,
		Source:     rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
		EndpointID: meta.EndpointID,
		Meta:       meta.Tags,
	}
	event.Meta["rule_id"] = rule.ID

	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

func (m *Manager) emitAlertResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	m.emitter.Emit(ctx, model.EventEntry{
		Timestamp:  now,
		Level:      rule.Level,
		Category:   "alert",
		Message:    "Resolved: " + rule.Message,
		Source:     rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
		EndpointID: meta.EndpointID,
		Meta:       meta.Tags,
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
