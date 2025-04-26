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

// gosight/agent/internal/store/datastore/pgsql/pgdatastore.go
// datastore.go - defines the general relational db interface and types

// Package pgstore implements the userstore.Store interface using PostgreSQL
package pgstore

import (
	"database/sql"
)

// Package pgstore implements the userstore.Store interface using PostgreSQL
// PGDataStore is a struct that represents a PostgreSQL data store
// It contains a pointer to the sql.DB object for database operations
type PGDataStore struct {
	db *sql.DB
}

// New creates a new PGDataStore instance
func NewPGDataStore(db *sql.DB) *PGDataStore {
	return &PGDataStore{db: db}
}

// Close closes the database connection
// It checks if the db is not nil before attempting to close it
func (s *PGDataStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
