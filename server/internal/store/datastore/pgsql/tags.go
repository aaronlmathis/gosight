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

// gosight/agent/internal/store/datastore/pgdatastore/tags.go
// tracker.go - defines the db fucntions for managing resource tags.

package pgstore

import (
	"context"

	"github.com/aaronlmathis/gosight/shared/model"
)

// GetTags retrieves all tags for a given endpoint ID.
func (s *PGDataStore) GetTags(ctx context.Context, endpointID string) (map[string]string, error) {

	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM tags WHERE endpoint_id = $1`, endpointID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tags := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		tags[key] = value
	}
	return tags, nil
}

func (s *PGDataStore) GetAllTags(ctx context.Context) ([]model.Tag, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT endpoint_id, key, value
		FROM tags
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var t model.Tag
		if err := rows.Scan(&t.EndpointID, &t.Key, &t.Value); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// SetTags replaces all tags for a given endpoint ID with the provided tags.
func (s *PGDataStore) SetTags(ctx context.Context, endpointID string, tags map[string]string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First delete existing tags
	_, err = tx.ExecContext(ctx, `DELETE FROM tags WHERE endpoint_id = $1`, endpointID)
	if err != nil {
		return err
	}

	// Then insert new tags
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO tags (endpoint_id, key, value) VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range tags {
		if _, err := stmt.Exec(endpointID, k, v); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// DeleteTag removes a specific tag for a given endpoint ID.
func (s *PGDataStore) DeleteTag(ctx context.Context, endpointID, key string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tags WHERE endpoint_id = $1 AND key = $2`, endpointID, key)
	return err
}

// ListTags retrieves all tags for a given endpoint ID.
func (s *PGDataStore) ListTags(ctx context.Context, endpointID string) (map[string]string, error) {
	return s.GetTags(ctx, endpointID)
}

// ListKeys returns all unique tag keys
func (s *PGDataStore) ListKeys(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT key FROM tags ORDER BY key ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// ListValues returns all values for a given key
func (s *PGDataStore) ListValues(ctx context.Context, key string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT value FROM tags WHERE key = $1 ORDER BY value ASC`, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}
