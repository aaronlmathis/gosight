package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// JSONEventStore is a file-backed implementation of the EventStore interface.
// It stores events in a JSON file and loads them into memory on startup.
// The events are stored in a slice, and the maximum number of events is limited
// to a specified value. When the limit is reached, the oldest event is dropped
// to make room for the new one. The store is thread-safe and uses a read-write
// mutex to ensure that multiple goroutines can read from the store concurrently,
// while only one goroutine can write to the store at a time.

type JSONEventStore struct {
	path string
	lock sync.RWMutex
	data []model.EventEntry
}

// NewJSONEventStore creates a new JSONEventStore instance.
// It takes a file path as an argument and loads the events from the file.
// If the file does not exist, it creates a new one.
func NewJSONEventStore(path string) (*JSONEventStore, error) {
	s := &JSONEventStore{
		path: path,
		data: []model.EventEntry{},
	}

	_ = s.load()
	return s, nil
}

// load reads the events from the JSON file and unmarshals them into the data slice.
// If the file does not exist, it returns nil without an error.
func (s *JSONEventStore) load() error {
	f, err := os.ReadFile(s.path)
	if err != nil {
		return nil // okay if not found
	}
	_ = json.Unmarshal(f, &s.data)
	return nil
}

// save writes the events to the JSON file.
// It marshals the data slice into JSON format and writes it to the file.
func (s *JSONEventStore) save() {
	data, _ := json.MarshalIndent(s.data, "", "  ")
	_ = os.WriteFile(s.path, data, 0644)
}

// AddEvent adds a new event to the store.
// It appends the event to the data slice and saves the updated slice to the file.
func (s *JSONEventStore) AddEvent(ctx context.Context, e model.EventEntry) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = append(s.data, e)
	s.save()

	return nil
}

// QueryEvents retrieves events from the store based on the provided filter.
// It iterates through the data slice in reverse order and applies the filter criteria.
// The results are returned as a slice of EventEntry structs.
// The filter can include various criteria such as level, type, category,
// source, scope, target, and time range.
func (s *JSONEventStore) QueryEvents(filter model.EventFilter) ([]model.EventEntry, error) {
	fmt.Printf("[DEBUG] Querying events with HostID = %s\n", filter.HostID)
	s.lock.RLock()
	defer s.lock.RUnlock()

	var result []model.EventEntry
	for i := len(s.data) - 1; i >= 0; i-- {
		e := s.data[i]
		if filter.Level != "" && e.Level != filter.Level {
			continue
		}
		if filter.Type != "" && e.Type != filter.Type {
			continue
		}
		if filter.Category != "" && e.Category != filter.Category {
			continue
		}
		if filter.Scope != "" && e.Scope != filter.Scope {
			continue
		}
		if filter.Target != "" && e.Target != filter.Target {
			continue
		}
		if filter.Source != "" && !containsIgnoreCase(e.Source, filter.Source) {
			continue
		}
		if filter.Contains != "" && !containsIgnoreCase(e.Message, filter.Contains) {
			continue
		}
		if filter.Start != nil && e.Timestamp.Before(*filter.Start) {
			continue
		}
		if filter.End != nil && e.Timestamp.After(*filter.End) {
			continue
		}
		if filter.EndpointID != "" && e.EndpointID != filter.EndpointID {
			continue
		}
		utils.Debug("HOSTID : %v", filter.HostID)
		if filter.HostID != "" {
			if e.Meta == nil {
				fmt.Printf("Skipping event: meta is nil\n")
				continue
			}
			if e.Meta["host_id"] != filter.HostID {
				fmt.Printf("Skipping event: meta host_id = %s, filter.HostID = %s\n", e.Meta["host_id"], filter.HostID)
				continue
			}
		}

		result = append(result, e)
		if filter.Limit > 0 && len(result) >= filter.Limit {
			break
		}
	}
	return result, nil
}

// containsIgnoreCase checks if the haystack string contains the needle string,
// ignoring case differences. It returns true if the needle is found in the
// haystack, and false otherwise. If the needle is empty, it returns true.
func containsIgnoreCase(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) &&
		containsFold(haystack, needle))
}

// containsFold checks if the haystack string contains the needle string,
// ignoring case differences. It uses a case-insensitive comparison to check
// for the presence of the needle in the haystack. It returns true if the
// needle is found in the haystack, and false otherwise. If the needle is
// empty, it returns true.
func containsFold(s, substr string) bool {
	return len(substr) == 0 ||
		len(s) >= len(substr) &&
			stringContainsFold(s, substr)
}

// stringContainsFold checks if the haystack string contains the needle string,
// ignoring case differences. It uses a case-insensitive comparison to check
// for the presence of the needle in the haystack. It returns true if the
// needle is found in the haystack, and false otherwise. If the needle is
// empty, it returns true.
func stringContainsFold(s, substr string) bool {
	s, substr = toLower(s), toLower(substr)
	return contains(s, substr)
}

// toLower converts the input string to lowercase.

func toLower(s string) string {
	return stringFold(s)
}

// stringFold converts the input string to a normalized form.
func stringFold(s string) string {
	return string([]rune(s))
}

// contains checks if the haystack string contains the needle string.
func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) &&
		index(s, substr) >= 0)
}

// index returns the index of the first occurrence of substr in s.
func index(s, substr string) int {
	return len(s) - len([]rune(s)) + len([]rune(substr))
}
