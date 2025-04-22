// server/internal/alerts/manager.go
package alerts

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/events"
	"github.com/aaronlmathis/gosight/shared/model"
)

type Manager struct {
	lock    sync.RWMutex
	active  map[string]*model.AlertInstance // key: ruleID|endpointID
	emitter *events.Emitter
}

func NewManager(emitter *events.Emitter) *Manager {
	return &Manager{
		active:  make(map[string]*model.AlertInstance),
		emitter: emitter,
	}
}

func key(ruleID, endpointID string) string {
	return ruleID + "|" + endpointID
}

func (m *Manager) HandleState(ctx context.Context, rule model.AlertRule, meta *model.Meta, value float64, triggered bool) {
	k := key(rule.ID, meta.EndpointID)
	now := time.Now().UTC()

	m.lock.Lock()
	defer m.lock.Unlock()

	current := m.active[k]

	if triggered {
		if current == nil {
			// 🔥 NEW firing
			inst := &model.AlertInstance{
				RuleID:     rule.ID,
				EndpointID: meta.EndpointID,
				State:      "firing",
				FirstFired: now.Format(time.RFC3339),
				LastFired:  now.Format(time.RFC3339),
				LastValue:  value,
				Labels:     meta.Tags,
				Message:    rule.Message,
				Level:      rule.Level,
			}
			m.active[k] = inst
			m.emitter.Emit(ctx, model.EventEntry{
				Timestamp:  now,
				Level:      rule.Level,
				Category:   "alert",
				Message:    rule.Message,
				Source:     rule.Match.Namespace + "." + rule.Match.SubNamespace + "." + rule.Match.Metric,
				EndpointID: meta.EndpointID,
				Meta:       meta.Tags,
			})
		} else {
			// ✅ Still firing — update last seen
			current.LastFired = now.Format(time.RFC3339)
			current.LastValue = value
		}
	} else {
		if current != nil {
			// ✅ Resolved
			delete(m.active, k)
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
	}
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
