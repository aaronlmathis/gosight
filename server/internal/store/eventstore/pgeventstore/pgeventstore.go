package pgeventstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/google/uuid"
)

type PGEventStore struct {
	db *sql.DB
}

func NewPGEventStore(db *sql.DB) *PGEventStore {
	return &PGEventStore{db: db}
}

// AddEvent adds an event to the event store.
// If the event ID is empty, a new UUID is generated.
// The event is stored in the database with the provided details.
func (s *PGEventStore) AddEvent(ctx context.Context, e model.EventEntry) error {
	if e.ID == "" {
		e.ID = uuid.NewString()
	}
	meta, _ := json.Marshal(e.Meta)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO events (
			id, timestamp, level, type, category, message,
			source, scope, target, endpoint_id, meta
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11
		)
	`,
		e.ID, e.Timestamp, e.Level, e.Type, e.Category, e.Message,
		e.Source, e.Scope, e.Target, e.EndpointID, meta)
	return err
}

// QueryEvents retrieves events from the event store based on the provided filter.
// The filter can include various criteria such as level, type, category,
// source, scope, target, and time range.
// The results are returned as a slice of EventEntry structs.
func (s *PGEventStore) QueryEvents(filter model.EventFilter) ([]model.EventEntry, error) {
	q := `SELECT id, timestamp, level, type, category, message,
	          source, scope, target, endpoint_id, meta
	      FROM events WHERE 1=1`
	args := []interface{}{}
	arg := func(v interface{}) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if filter.Level != "" {
		q += " AND level = " + arg(filter.Level)
	}
	if filter.Type != "" {
		q += " AND type = " + arg(filter.Type)
	}
	if filter.Category != "" {
		q += " AND category = " + arg(filter.Category)
	}
	if filter.Scope != "" {
		q += " AND scope = " + arg(filter.Scope)
	}
	if filter.Target != "" {
		q += " AND target = " + arg(filter.Target)
	}
	if filter.Source != "" {
		q += " AND source ILIKE " + arg("%"+filter.Source+"%")
	}
	if filter.Contains != "" {
		q += " AND message ILIKE " + arg("%"+filter.Contains+"%")
	}
	if filter.Start != nil {
		q += " AND timestamp >= " + arg(*filter.Start)
	}
	if filter.End != nil {
		q += " AND timestamp <= " + arg(*filter.End)
	}
	if filter.EndpointID != "" {
		q += " AND endpoint_id = " + arg(filter.EndpointID)
	}
	if filter.HostID != "" {
		q += " AND meta->>'host_id' = " + arg(filter.HostID)
	}

	q += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		q += " LIMIT " + arg(filter.Limit)
	}

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.EventEntry
	for rows.Next() {
		var e model.EventEntry
		var meta []byte
		err := rows.Scan(&e.ID, &e.Timestamp, &e.Level, &e.Type, &e.Category, &e.Message,
			&e.Source, &e.Scope, &e.Target, &e.EndpointID, &meta)
		if err != nil {
			continue
		}
		_ = json.Unmarshal(meta, &e.Meta)
		results = append(results, e)
	}
	return results, rows.Err()
}
