/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/alertstore/pgalertstore/pgalertstore.go
// Description: This file contains the implementation of the AlertStore interface
// using PostgreSQL as the backend. It provides methods to upsert, resolve,
// and list active and historical alerts. The implementation uses the
// database/sql package for database interactions and the encoding/json
// package for JSON encoding and decoding.

package pgalertstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/google/uuid"
)

// PGAlertStore is a PostgreSQL implementation of the AlertStore interface.
type PGAlertStore struct {
	db *sql.DB
}

// NewPGAlertStore creates a new PGAlertStore instance with the given database connection.
func NewPGAlertStore(db *sql.DB) *PGAlertStore {
	return &PGAlertStore{db: db}
}

// GetByID retrieves an alert by its ID from the PostgreSQL database.
// GetByID retrieves an alert instance by ID from the PostgreSQL database.
// GetByID retrieves an alert by its ID from the PostgreSQL database.
func (s *PGAlertStore) GetByID(ctx context.Context, id string) (model.AlertInstance, error) {
	// Ensure ID is not empty
	if id == "" {
		return model.AlertInstance{}, errors.New("invalid alert ID")
	}

	uuidID, err := uuid.Parse(id)
	if err != nil {
		return model.AlertInstance{}, fmt.Errorf("invalid UUID format: %v", err)
	}

	// Define the SQL query to fetch the alert
	query := `
		SELECT 
			id, rule_id, state, previous, scope, target, first_fired, last_fired, 
			last_ok, last_value, level, message, labels, resolved_at
		FROM alerts WHERE id = $1::uuid
	`

	var alert model.AlertInstance
	var resolvedAt sql.NullTime
	var labelsBytes []byte

	// Query the database using QueryRowContext for a single result
	err = s.db.QueryRowContext(ctx, query, uuidID).Scan(
		&alert.ID,
		&alert.RuleID,
		&alert.State,
		&alert.Previous,
		&alert.Scope,
		&alert.Target,
		&alert.FirstFired,
		&alert.LastFired,
		&alert.LastOK,
		&alert.LastValue,
		&alert.Level,
		&alert.Message,
		&labelsBytes,
		&resolvedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.AlertInstance{}, fmt.Errorf("alert with ID %s not found", id)
		}
		return model.AlertInstance{}, fmt.Errorf("failed to retrieve alert: %v", err)
	}

	// Unmarshal the labels if they exist
	if len(labelsBytes) > 0 {
		if err := json.Unmarshal(labelsBytes, &alert.Labels); err != nil {
			return model.AlertInstance{}, fmt.Errorf("failed to unmarshal labels: %v", err)
		}
	}

	// Set resolved_at if the field is not null
	if resolvedAt.Valid {
		alert.ResolvedAt = &resolvedAt.Time
	}

	return alert, nil
}
func (s *PGAlertStore) ListAlerts(ctx context.Context) ([]model.AlertInstance, error) {
	query := `
        SELECT 
            id,
            rule_id,
            state,
            previous,
            scope,
            target,
            first_fired,
            last_fired,
            last_ok,
            resolved_at,
            last_value,
            level,
            message,
            labels
        FROM alerts
        ORDER BY last_fired DESC
    `
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.AlertInstance

	for rows.Next() {
		var a model.AlertInstance
		var resolvedAt sql.NullTime
		var labelsBytes []byte

		err := rows.Scan(
			&a.ID,
			&a.RuleID,
			&a.State,
			&a.Previous,
			&a.Scope,
			&a.Target,
			&a.FirstFired,
			&a.LastFired,
			&a.LastOK,
			&resolvedAt,
			&a.LastValue,
			&a.Level,
			&a.Message,
			&labelsBytes,
		)
		if err != nil {
			return nil, err
		}

		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}

		// Unmarshal JSON labels into map
		if len(labelsBytes) > 0 {
			if err := json.Unmarshal(labelsBytes, &a.Labels); err != nil {
				return nil, err
			}
		}

		alerts = append(alerts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return alerts, nil
}

// UpsertAlert inserts or updates an alert instance in the database.
func (s *PGAlertStore) UpsertAlert(ctx context.Context, a *model.AlertInstance) error {
	if a.ID == "" {
		a.ID = uuid.NewString()
	}
	labels, _ := json.Marshal(a.Labels)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO alerts (
			id, rule_id, state, previous, scope, target, first_fired, last_fired,
			last_ok, resolved_at, last_value, level, message, labels
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (id) DO UPDATE SET
			state = EXCLUDED.state,
			previous = alerts.state,
			last_fired = EXCLUDED.last_fired,
			last_ok = EXCLUDED.last_ok,
			last_value = EXCLUDED.last_value,
			resolved_at = EXCLUDED.resolved_at,
			labels = EXCLUDED.labels;
	`,
		a.ID, a.RuleID, a.State, a.Previous, a.Scope, a.Target, a.FirstFired, a.LastFired,
		a.LastOK, a.ResolvedAt, a.LastValue, a.Level, a.Message, labels)
	return err
}

// ResolveAlert updates the state of an alert instance to "resolved" in the database.

func (s *PGAlertStore) ResolveAlert(ctx context.Context, ruleID, target string, resolvedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE alerts SET
			state = 'resolved',
			resolved_at = $1
		WHERE rule_id = $2 AND target = $3 AND state = 'firing'
	`, resolvedAt, ruleID, target)
	return err
}

// ListActiveAlerts retrieves all active alerts from the database.
// It returns a slice of AlertInstance structs representing the active alerts.

func (s *PGAlertStore) ListActiveAlerts(ctx context.Context) ([]model.AlertInstance, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, rule_id, state, previous, scope, target, first_fired, last_fired,
			last_ok, resolved_at, last_value, level, message, labels
		FROM alerts WHERE state = 'firing'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// ListAlertHistory retrieves all alert instances that have fired since the given time.
// It returns a slice of AlertInstance structs representing the alert history.
// The results are ordered by the last_fired timestamp in descending order.
// This method is useful for retrieving historical alert data for analysis or reporting.

func (s *PGAlertStore) ListAlertHistory(ctx context.Context, since time.Time) ([]model.AlertInstance, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, rule_id, state, previous, scope, target, first_fired, last_fired,
			last_ok, resolved_at, last_value, level, message, labels
		FROM alerts WHERE last_fired >= $1
		ORDER BY last_fired DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// scanAlerts scans the rows returned by a query and converts them into a slice of AlertInstance structs.
// It handles the conversion of JSON-encoded labels and the resolved_at field, which may be null.
// The function returns an error if any issues occur during scanning or conversion.
// This function is used internally by the ListActiveAlerts and ListAlertHistory methods
// to process the results of the database queries.

func scanAlerts(rows *sql.Rows) ([]model.AlertInstance, error) {
	var out []model.AlertInstance
	for rows.Next() {
		var a model.AlertInstance
		var labels []byte
		var resolvedAt sql.NullTime

		err := rows.Scan(&a.ID, &a.RuleID, &a.State, &a.Previous, &a.Scope, &a.Target,
			&a.FirstFired, &a.LastFired, &a.LastOK, &resolvedAt,
			&a.LastValue, &a.Level, &a.Message, &labels)
		if err != nil {
			return nil, err
		}
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		_ = json.Unmarshal(labels, &a.Labels)
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *PGAlertStore) ListAlertsFiltered(ctx context.Context, q model.AlertQuery) ([]model.AlertInstance, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if q.RuleID != "" {
		where = append(where, "rule_id = $"+strconv.Itoa(argIdx))
		args = append(args, q.RuleID)
		argIdx++
	}
	if q.State != "" {
		where = append(where, "state = $"+strconv.Itoa(argIdx))
		args = append(args, q.State)
		argIdx++
	}
	if q.Level != "" {
		where = append(where, "level = $"+strconv.Itoa(argIdx))
		args = append(args, q.Level)
		argIdx++
	}
	if q.Target != "" {
		where = append(where, "target = $"+strconv.Itoa(argIdx))
		args = append(args, q.Target)
		argIdx++
	}
	if q.Scope != "" {
		where = append(where, "scope = $"+strconv.Itoa(argIdx))
		args = append(args, q.Scope)
		argIdx++
	}
	for k, v := range q.Tags {
		where = append(where, "labels->>$"+strconv.Itoa(argIdx)+" = $"+strconv.Itoa(argIdx+1))
		args = append(args, k, v)
		argIdx += 2
	}
	query := `
        SELECT id, rule_id, state, previous, scope, target, first_fired, last_fired,
               last_ok, resolved_at, last_value, level, message, labels
        FROM alerts
    `

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	// Sorting
	sortField := "last_fired"
	if q.Sort != "" {
		switch q.Sort {
		case "first_fired", "last_fired", "last_ok", "rule_id", "target", "state", "level":
			sortField = q.Sort
		}
	}
	order := "DESC"
	if strings.ToLower(q.Order) == "asc" {
		order = "ASC"
	}
	query += " ORDER BY " + sortField + " " + order

	// Pagination
	if q.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", q.Limit)
	}
	if q.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", q.Offset)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

func (s *PGAlertStore) CountAlertsFiltered(ctx context.Context, q model.AlertQuery) (int, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if q.RuleID != "" {
		where = append(where, "rule_id = $"+strconv.Itoa(argIdx))
		args = append(args, q.RuleID)
		argIdx++
	}
	if q.State != "" {
		where = append(where, "state = $"+strconv.Itoa(argIdx))
		args = append(args, q.State)
		argIdx++
	}
	if q.Level != "" {
		where = append(where, "level = $"+strconv.Itoa(argIdx))
		args = append(args, q.Level)
		argIdx++
	}
	if q.Target != "" {
		where = append(where, "target = $"+strconv.Itoa(argIdx))
		args = append(args, q.Target)
		argIdx++
	}
	if q.Scope != "" {
		where = append(where, "scope = $"+strconv.Itoa(argIdx))
		args = append(args, q.Scope)
		argIdx++
	}

	query := `SELECT COUNT(*) FROM alerts`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	var count int
	err := s.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}
