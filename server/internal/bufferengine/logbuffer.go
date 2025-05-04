// bufferengine/log_buffer.go
package bufferengine

import (
	"fmt"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type LogStore interface {
	Write(entries []model.LogPayload) error
}

type BufferedLogStore struct {
	name          string
	underlying    LogStore
	buffer        []model.LogPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
}

func NewBufferedLogStore(name string, store LogStore, maxSize int, flushInterval time.Duration) *BufferedLogStore {
	return &BufferedLogStore{
		name:          name,
		underlying:    store,
		buffer:        make([]model.LogPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
	}
}

func (b *BufferedLogStore) Name() string {
	return b.name
}

func (b *BufferedLogStore) Interval() time.Duration {
	return b.flushInterval
}

func (b *BufferedLogStore) WriteAny(payload interface{}) error {
	p, ok := payload.(model.LogPayload)
	if !ok {
		return fmt.Errorf("BufferedLogStore: invalid payload type %T", payload)
	}
	return b.Write(p)
}

func (b *BufferedLogStore) Write(payload model.LogPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

func (b *BufferedLogStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

func (b *BufferedLogStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]model.LogPayload, 0, b.maxSize)
	utils.Debug("Flushing %d log payloads from buffer", len(toFlush))
	return b.underlying.Write(toFlush)
}

func (b *BufferedLogStore) Close() error {
	return b.Flush()
}
