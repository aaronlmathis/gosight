package eventstore

import "github.com/aaronlmathis/gosight/shared/model"

// EventStore is an interface for storing and retrieving events.
// It provides methods to add events and get recent events.
// The implementation of this interface should handle the actual storage
// and retrieval of events, whether it's in memory, a database, or any other
// storage mechanism.

type EventStore interface {
	AddEvent(e model.EventEntry)
	GetRecent(limit int) []model.EventEntry
}
