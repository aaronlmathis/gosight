package cache

import (
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

const snapshotRetention = 30 * time.Minute

// ProcessCache stores recent process snapshots per endpoint.
type ProcessCache interface {
	Add(snapshot model.ProcessSnapshot)
	Get(endpointID string) []model.ProcessSnapshot
	Prune()
}

type processCache struct {
	mu        sync.RWMutex
	snapshots map[string][]model.ProcessSnapshot // key: endpointID
}

func NewProcessCache() ProcessCache {
	return &processCache{
		snapshots: make(map[string][]model.ProcessSnapshot),
	}
}

func (c *processCache) Add(snapshot model.ProcessSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Append snapshot to ring buffer per endpoint
	ep := snapshot.EndpointID
	c.snapshots[ep] = append(c.snapshots[ep], snapshot)

	// Trim old snapshots right after append
	cutoff := time.Now().Add(-snapshotRetention)
	buf := c.snapshots[ep]
	for i := 0; i < len(buf); i++ {
		if buf[i].Timestamp.After(cutoff) {
			c.snapshots[ep] = buf[i:]
			return
		}
	}
	c.snapshots[ep] = nil // all entries too old
}

func (c *processCache) Get(endpointID string) []model.ProcessSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshots[endpointID]
}

func (c *processCache) Prune() {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := time.Now().Add(-snapshotRetention)
	for ep, buf := range c.snapshots {
		for i := 0; i < len(buf); i++ {
			if buf[i].Timestamp.After(cutoff) {
				c.snapshots[ep] = buf[i:]
				break
			}
		}
	}
}
