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
	hub        *websocket.AlertsHub
}

func NewManager(emitter *events.Emitter, dispatcher *dispatcher.Dispatcher, store alertstore.AlertStore, hub *websocket.AlertsHub) *Manager {
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

func (m *Manager) HandleState(ctx context.Context, rule model.AlertRule, meta *model.Meta, value float64, triggered bool) {
	utils.Debug("✅  Handle State... Rule ID: %s, Endpoint ID: %s, Value: %f, Triggered: %v", rule.ID, meta.EndpointID, value, triggered)
	k := key(rule.ID, meta.EndpointID)
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	current := m.active[k]

	if triggered {
		if current != nil {
			if rule.Options.RepeatInterval != "" {
				repeatDur, _ := time.ParseDuration(rule.Options.RepeatInterval)
				if now.Sub(current.LastFired) < repeatDur {
					return
				}
			}
			current.LastFired = now
			current.LastValue = value
			_ = m.store.UpsertAlert(ctx, current)
			m.hub.Broadcast(*current)
			m.emitAlertFiringEvent(ctx, rule, meta, now)
			m.dispatchFiringEvent(ctx, rule, meta, current.Target, now, current.Message)
			return
		}

		scope, target := inferScopeAndTarget(meta)

		inst := &model.AlertInstance{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			State:      "firing",
			Previous:   "ok",
			Scope:      scope,
			Target:     target,
			FirstFired: now,
			LastFired:  now,
			LastOK:     now,
			LastValue:  value,
			Level:      rule.Level,
			Message:    rule.Message,
			Labels:     utils.SafeCopyTags(meta),
		}

		event := model.EventEntry{
			Timestamp: now,
			Level:     rule.Level,
			Category:  "alert",
			Message:   rule.Message,
			Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
			Scope:     scope,
			Target:    target,
			Meta:      utils.SafeCopyTags(meta),
		}
		event.Meta["rule_id"] = rule.ID

		if len(rule.Actions) > 0 {
			for _, actionID := range rule.Actions {
				utils.Debug("Triggering action ID: %s for alert rule: %s", actionID, rule.ID)
				m.dispatcher.TriggerActionByID(ctx, actionID, event)
			}
			m.active[k] = inst
			_ = m.store.UpsertAlert(ctx, inst)
			m.hub.Broadcast(*inst)
			m.emitter.Emit(ctx, event)
			return
		}

		m.active[k] = inst
		_ = m.store.UpsertAlert(ctx, inst)
		m.hub.Broadcast(*inst)
		m.emitAlertFiringEvent(ctx, rule, meta, now)
		m.dispatchFiringEvent(ctx, rule, meta, target, now, rule.Message)
	} else {
		if current != nil {
			delete(m.active, k)
			m.emitAlertResolvedEvent(ctx, rule, meta, now)
			if rule.Options.NotifyOnResolve {
				m.dispatchResolvedEvent(ctx, rule, meta, current.Target, now, "Resolved: "+rule.Message)
			}
			_ = m.store.ResolveAlert(ctx, rule.ID, current.Target, now)
			m.hub.Broadcast(*current)
		}
	}
}

func (m *Manager) HandleLogState(ctx context.Context, rule model.AlertRule, meta *model.Meta, log model.LogEntry, triggered bool) {
	k := key(rule.ID, meta.EndpointID+"|"+log.Timestamp.Format(time.RFC3339Nano))
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	if triggered {
		inst := &model.AlertInstance{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			State:      "firing",
			Previous:   "ok",
			Scope:      "endpoint",
			Target:     meta.EndpointID,
			FirstFired: now,
			LastFired:  now,
			LastOK:     now,
			LastValue:  0,
			Level:      rule.Level,
			Message:    rule.Message + ": " + log.Message,
			Labels:     utils.SafeCopyTags(meta),
		}

		event := model.EventEntry{
			Timestamp: now,
			Level:     rule.Level,
			Category:  "log_alert",
			Message:   rule.Message + ": " + log.Message,
			Source:    log.Source,
			Scope:     "endpoint",
			Target:    meta.EndpointID,
			Meta:      utils.SafeCopyTags(meta),
		}
		event.Meta["rule_id"] = rule.ID

		if len(rule.Actions) > 0 {
			for _, actionID := range rule.Actions {
				m.dispatcher.TriggerActionByID(ctx, actionID, event)
			}
			m.active[k] = inst
			_ = m.store.UpsertAlert(ctx, inst)
			m.hub.Broadcast(*inst)
			m.emitter.Emit(ctx, event)
			return
		}

		m.active[k] = inst
		_ = m.store.UpsertAlert(ctx, inst)
		m.hub.Broadcast(*inst)
		m.emitLogAlertFiringEvent(ctx, rule, meta, log, now)
	}
}

func (m *Manager) emitAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	scope, target := inferScopeAndTarget(meta)
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   rule.Message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyTags(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

func (m *Manager) emitAlertResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, now time.Time) {
	scope, target := inferScopeAndTarget(meta)
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   "Resolved: " + rule.Message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyTags(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
}

func (m *Manager) emitLogAlertFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, log model.LogEntry, now time.Time) {
	event := model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "log_alert",
		Message:   rule.Message + ": " + log.Message,
		Source:    log.Source,
		Scope:     "endpoint",
		Target:    meta.EndpointID,
		Meta:      utils.SafeCopyTags(meta),
	}
	event.Meta["rule_id"] = rule.ID
	m.emitter.Emit(ctx, event)
	m.dispatcher.Dispatch(ctx, event)
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

func inferScopeAndTarget(meta *model.Meta) (string, string) {
	if meta.EndpointID != "" {
		return "endpoint", meta.EndpointID
	}
	return "global", "gosight-core"
}

func (m *Manager) dispatchFiringEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, target string, now time.Time, message string) {
	scope, target := inferScopeAndTarget(meta)
	m.dispatcher.Dispatch(ctx, model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyTags(meta),
	})
}

func (m *Manager) dispatchResolvedEvent(ctx context.Context, rule model.AlertRule, meta *model.Meta, target string, now time.Time, message string) {
	scope, target := inferScopeAndTarget(meta)
	m.dispatcher.Dispatch(ctx, model.EventEntry{
		Timestamp: now,
		Level:     rule.Level,
		Category:  "alert",
		Message:   message,
		Source:    rule.Scope.Namespace + "." + rule.Scope.SubNamespace + "." + rule.Scope.Metric,
		Scope:     scope,
		Target:    target,
		Meta:      utils.SafeCopyTags(meta),
	})
}
