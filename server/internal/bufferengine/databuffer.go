package bufferengine

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/datastore"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type DataStore interface {
	Write(ctx context.Context, batches []*model.ProcessPayload) error
}

type BufferedDataStore struct {
	name          string
	underlying    datastore.DataStore
	buffer        []*model.ProcessPayload
	mu            sync.Mutex
	maxSize       int
	flushInterval time.Duration
	ctx           context.Context
}

func NewBufferedDataStore(ctx context.Context, name string, store datastore.DataStore, maxSize int, flushInterval time.Duration) *BufferedDataStore {
	return &BufferedDataStore{
		name:          name,
		underlying:    store,
		buffer:        make([]*model.ProcessPayload, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
		ctx:           ctx,
	}
}

func (b *BufferedDataStore) Name() string {
	return b.name
}

func (b *BufferedDataStore) Interval() time.Duration {
	return b.flushInterval
}

func (b *BufferedDataStore) WriteAny(payload interface{}) error {
	p, ok := payload.(*model.ProcessPayload)
	if !ok {
		return errors.New("invalid payload type for process data")
	}
	return b.Write(p)
}

func (b *BufferedDataStore) Write(payload *model.ProcessPayload) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buffer = append(b.buffer, payload)
	if len(b.buffer) >= b.maxSize {
		return b.flushLocked()
	}
	return nil
}

func (b *BufferedDataStore) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

func (b *BufferedDataStore) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}
	toFlush := b.buffer
	b.buffer = make([]*model.ProcessPayload, 0, b.maxSize)
	utils.Debug("Flushing %d process payloads from buffer", len(toFlush))
	return b.underlying.Write(b.ctx, toFlush)
}

func (b *BufferedDataStore) Close() error {
	return b.Flush()
}
