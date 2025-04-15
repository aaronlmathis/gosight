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

// gosight/agent/internal/store/datastore/datastore.go
// datastore.go - defines the general relational db interface and types

// Package pgstore implements the userstore.Store interface using PostgreSQL
package pgstore

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/aaronlmathis/gosight/shared/model"
)

type PGDataStore struct {
	db *sql.DB
}

func New(db *sql.DB) *PGDataStore {
	return &PGDataStore{db: db}
}

// Agent related functions

// UpsertAgent inserts or updates an agent in the database

func (s *PGDataStore) UpsertAgent(ctx context.Context, agent *model.AgentStatus) error {
	tags, _ := json.Marshal(agent.Labels)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (
			agent_id, hostname, ip, os, arch, version, labels, last_seen
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, now()
		)
		ON CONFLICT (agent_id) DO UPDATE SET
			ip = EXCLUDED.ip,
			os = EXCLUDED.os,
			arch = EXCLUDED.arch,
			version = EXCLUDED.version,
			labels = EXCLUDED.labels,
			last_seen = now()
	`,
		agent.AgentID,
		agent.Hostname,
		agent.IP,
		agent.OS,
		agent.Arch,
		agent.Version,
		tags,
	)

	return err
}

func (s *PGDataStore) GetAgentByAgentID(ctx context.Context, agentID string) (*model.AgentStatus, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT agent_id, hostname, ip, os, arch, version, labels, last_seen, updated_at 
		FROM agents
		WHERE agent_id = $1
	`, agentID)

	var agent model.AgentStatus

	var tagsRaw []byte

	err := row.Scan(
		&agent.AgentID,
		&agent.Hostname,
		&agent.IP,
		&agent.OS,
		&agent.Arch,
		&agent.Version,
		&tagsRaw,
		&agent.LastSeen,
		&agent.Updated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tagsRaw, &agent.Labels); err != nil {
		return nil, err
	}

	return &agent, nil
}
func (s *PGDataStore) GetAgentByHostname(ctx context.Context, hostname string) (*model.AgentStatus, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT agent_id, hostname, ip, os, arch, version, labels, last_seen, updated_at
		FROM agents
		WHERE hostname = $1
	`, hostname)

	var agent model.AgentStatus

	var tagsRaw []byte

	err := row.Scan(
		&agent.AgentID,
		&agent.Hostname,
		&agent.IP,
		&agent.OS,
		&agent.Arch,
		&agent.Version,
		&tagsRaw,
		&agent.LastSeen,
		&agent.Updated,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(tagsRaw, &agent.Labels); err != nil {
		return nil, err
	}

	return &agent, nil
}

func (s *PGDataStore) ListAgents(ctx context.Context) ([]*model.AgentStatus, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT agent_id, hostname, ip, os, arch, version, labels, last_seen, updated_at FROM agents
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*model.AgentStatus

	for rows.Next() {
		var a model.AgentStatus
		var tagsRaw []byte

		err := rows.Scan(
			&a.AgentID,
			&a.Hostname,
			&a.IP,
			&a.OS,
			&a.Arch,
			&a.Version,
			&tagsRaw,
			&a.LastSeen,
			&a.Updated,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(tagsRaw, &a.Labels); err != nil {
			return nil, err
		}

		agents = append(agents, &a)
	}

	return agents, rows.Err()
}
