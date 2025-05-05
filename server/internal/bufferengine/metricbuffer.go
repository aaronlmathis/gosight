// bufferengine/metricbuffer.go
package bufferengine

import (
	"errors"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

type BufferedMetricStore struct {
	name          string
	underlying    MetricStore
	buffer        []model.MetricPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
}

type MetricStore interface {
	Write(payloads []model.MetricPayload) error
}

func NewBufferedMetricStore(name string, store MetricStore, maxSize int, flushInterval time.Duration) *BufferedMetricStore {
	return &BufferedMetricStore{
		name:          name,
		underlying:    store,
		buffer:        make([]model.MetricPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
	}
}

func (b *BufferedMetricStore) Name() string {
	return b.name
}

func (b *BufferedMetricStore) Interval() time.Duration {
	return b.flushInterval
}

func (b *BufferedMetricStore) WriteAny(payload interface{}) error {
	p, ok := payload.(model.MetricPayload)
	if !ok {
		return errors.New("invalid payload type for metrics")
	}
	return b.Write(p)
}

func (b *BufferedMetricStore) Write(payload model.MetricPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

func (b *BufferedMetricStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

func (b *BufferedMetricStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]model.MetricPayload, 0, b.maxSize)
	//utils.Debug("Flushing %d metric payloads from buffer", len(toFlush))
	return b.underlying.Write(toFlush)
}

func (b *BufferedMetricStore) Close() error {
	return b.Flush()
}
