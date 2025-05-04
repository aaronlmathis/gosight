// bufferengine/engine.go
package bufferengine

import (
	"context"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/utils"
)

type BufferedStore interface {
	WriteAny(payload interface{}) error
	Flush() error
	Close() error
	Name() string
	Interval() time.Duration
}

type BufferEngine struct {
	stores        []BufferedStore
	flushInterval time.Duration
	maxWorkers    int
	ctx           context.Context
	wg            sync.WaitGroup
}

func NewBufferEngine(ctx context.Context, flushInterval time.Duration, maxWorkers int) *BufferEngine {
	return &BufferEngine{
		flushInterval: flushInterval,
		maxWorkers:    maxWorkers,
		ctx:           ctx,
	}
}

func (e *BufferEngine) RegisterStore(store BufferedStore) {
	e.stores = append(e.stores, store)
	utils.Info("BufferEngine registered store: %s", store.Name())
}

func (e *BufferEngine) Start() {
	utils.Info("BufferEngine starting with %d stores (per-store intervals)", len(e.stores))

	for _, store := range e.stores {
		e.wg.Add(1)
		go func(s BufferedStore) {
			defer e.wg.Done()
			interval := s.Interval()
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			utils.Info("Buffer [%s] started with flush interval: %s", s.Name(), interval)

			for {
				select {
				case <-e.ctx.Done():
					utils.Info("Buffer [%s] shutting down...", s.Name())
					_ = s.Flush() // final flush on shutdown
					return
				case <-ticker.C:
					if err := s.Flush(); err != nil {
						utils.Warn("Flush failed for [%s]: %v", s.Name(), err)
					}
				}
			}
		}(store)
	}
}

func (e *BufferEngine) Stop() {
	utils.Info("BufferEngine waiting for background flush routines to stop...")
	e.wg.Wait()

	for _, store := range e.stores {
		if err := store.Close(); err != nil {
			utils.Warn("Error closing store [%s]: %v", store.Name(), err)
		}
	}
	utils.Info("BufferEngine stopped cleanly")
}
