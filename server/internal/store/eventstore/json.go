package eventstore

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
)

// JSONEventStore is a file-backed implementation of the EventStore interface.
// It stores events in a JSON file and loads them into memory on startup.
// The events are stored in a slice, and the maximum number of events is limited
// to a specified value. When the limit is reached, the oldest event is dropped
// to make room for the new one. The store is thread-safe and uses a read-write
// mutex to ensure that multiple goroutines can read from the store concurrently,
// while only one goroutine can write to the store at a time.

type JSONEventStore struct {
	path   string
	lock   sync.RWMutex
	events []model.EventEntry
	max    int
}

// NewJSONEventStore creates a file-backed event store
func NewJSONEventStore(path string, max int) (*JSONEventStore, error) {
	store := &JSONEventStore{
		path:   path,
		max:    max,
		events: []model.EventEntry{},
	}

	// Try to load existing events from file
	if data, err := os.ReadFile(path); err == nil {
		json.Unmarshal(data, &store.events)
	}

	return store, nil
}

// AddEvent adds a new event entry to the store.
// If the store is full, it drops the oldest event to make room for the new one.
func (j *JSONEventStore) AddEvent(e model.EventEntry) {
	j.lock.Lock()
	defer j.lock.Unlock()

	// Trim if necessary
	if len(j.events) >= j.max {
		j.events = j.events[1:]
	}
	j.events = append(j.events, e)

	// Save to disk (non-blocking alternative could be queue+flush)
	go j.saveToFile()
}

// GetRecent retrieves the most recent event entries from the store.
// It returns a slice of EventEntry objects, limited to the specified number.
// If the limit exceeds the number of stored events, it returns all available events.
func (j *JSONEventStore) GetRecent(limit int) []model.EventEntry {
	j.lock.RLock()
	defer j.lock.RUnlock()

	if limit > len(j.events) {
		limit = len(j.events)
	}
	return append([]model.EventEntry(nil), j.events[len(j.events)-limit:]...)
}

// saveToFile saves the current events to the JSON file.
// This is called in a goroutine after adding an event to avoid blocking the main thread.

func (j *JSONEventStore) saveToFile() {
	j.lock.RLock()
	defer j.lock.RUnlock()

	data, err := json.MarshalIndent(j.events, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(j.path, data, 0644)
}
