package metastore

import (
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

type MetaTracker struct {
	mu    sync.RWMutex
	store map[string]model.Meta
}

func NewMetaTracker() *MetaTracker {
	return &MetaTracker{
		store: make(map[string]model.Meta),
	}
}

func (m *MetaTracker) Set(endpointID string, meta model.Meta) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[endpointID] = meta
}

func (m *MetaTracker) Get(endpointID string) (model.Meta, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.store[endpointID]
	return meta, ok
}
