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
	"fmt"
	"runtime/debug"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type PGDataStore struct {
	db *sql.DB
}

func New(db *sql.DB) *PGDataStore {
	return &PGDataStore{db: db}
}

func (s *PGDataStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Agent related functions

// UpsertAgent inserts or updates an agent in the database

func (s *PGDataStore) UpsertAgent(ctx context.Context, agent *model.Agent) error {
	tags, _ := json.Marshal(agent.Labels)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO agents (
			agent_id, host_id, hostname, ip, os, arch, version, labels, endpoint_id, last_seen, status, since
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, now(), $10, $11
		)
		ON CONFLICT (agent_id) DO UPDATE SET
			host_id = EXCLUDED.host_id,
			hostname = EXCLUDED.hostname,
			ip = EXCLUDED.ip,
			os = EXCLUDED.os,
			arch = EXCLUDED.arch,
			version = EXCLUDED.version,
			labels = EXCLUDED.labels,
			endpoint_id = EXCLUDED.endpoint_id,
			last_seen = now(),
			status = EXCLUDED.status,
			since = EXCLUDED.since
	`,
		agent.AgentID,
		agent.HostID,
		agent.Hostname,
		agent.IP,
		agent.OS,
		agent.Arch,
		agent.Version,
		tags,
		agent.EndpointID,
		agent.Status,
		agent.Since,
	)

	return err
}

// GetAgentByAgentID retrieves an agent by its agent_id
// This is used for the agent to check in with the server
func (s *PGDataStore) GetAgentByID(ctx context.Context, id string) (*model.Agent, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT agent_id, host_id, hostname, ip, os, arch, version, labels, last_seen, endpoint_id, status, since
		FROM agents
		WHERE agent_id = $1
	`, id)

	var agent model.Agent
	var tagsRaw []byte
	var since sql.NullString
	var status sql.NullString

	err := row.Scan(
		&agent.AgentID,
		&agent.HostID,
		&agent.Hostname,
		&agent.IP,
		&agent.OS,
		&agent.Arch,
		&agent.Version,
		&tagsRaw,
		&agent.LastSeen,
		&agent.EndpointID,
		&status,
		&since,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(tagsRaw) > 0 {
		if err := json.Unmarshal(tagsRaw, &agent.Labels); err != nil {
			agent.Labels = map[string]string{}
		}
	}

	if status.Valid {
		agent.Status = status.String
	}
	if since.Valid {
		agent.Since = since.String
	}

	return &agent, nil
}

// GetAgentByHostname retrieves an agent by its hostname
func (s *PGDataStore) GetAgentByHostname(ctx context.Context, hostname string) (*model.Agent, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT agent_id, host_id, hostname, ip, os, arch, version, labels, last_seen, endpoint_id, status, since
		FROM agents
		WHERE hostname = $1
	`, hostname)

	var agent model.Agent
	var tagsRaw []byte
	var status sql.NullString
	var since sql.NullString

	err := row.Scan(
		&agent.AgentID,
		&agent.HostID,
		&agent.Hostname,
		&agent.IP,
		&agent.OS,
		&agent.Arch,
		&agent.Version,
		&tagsRaw,
		&agent.LastSeen,
		&agent.EndpointID,
		&status,
		&since,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if len(tagsRaw) > 0 {
		if err := json.Unmarshal(tagsRaw, &agent.Labels); err != nil {
			agent.Labels = map[string]string{}
		}
	}

	if status.Valid {
		agent.Status = status.String
	}
	if since.Valid {
		agent.Since = since.String
	}

	return &agent, nil
}

func (s *PGDataStore) ListAgents(ctx context.Context) ([]*model.Agent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT agent_id, host_id, hostname, ip, os, arch, version, labels, last_seen, endpoint_id, status, since
		FROM agents
	`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var agents []*model.Agent

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("🔥 PANIC in ListAgents loop:", r)
			debug.PrintStack()
		}
	}()

	for rows.Next() {
		agent := &model.Agent{}
		var tagsRaw []byte

		var since sql.NullString

		err := rows.Scan(
			&agent.AgentID,
			&agent.HostID,
			&agent.Hostname,
			&agent.IP,
			&agent.OS,
			&agent.Arch,
			&agent.Version,
			&tagsRaw,
			&agent.LastSeen,
			&agent.EndpointID,
			&agent.Status,
			&since, // 🟢 changed from &agent.Since
		)
		if err != nil {
			utils.Warn("Scan error: %v", err)
			continue
		}

		if since.Valid {
			agent.Since = since.String
		} else {
			agent.Since = ""
		}
		utils.Debug("🧲 Loaded agent: %s | Status: %s", agent.Hostname, agent.Status)

		if len(tagsRaw) > 0 {
			if err := json.Unmarshal(tagsRaw, &agent.Labels); err != nil {
				fmt.Printf("JSON decode failed: %v\nRaw: %s\n", err, string(tagsRaw))
				agent.Labels = map[string]string{}
			}
		} else {
			agent.Labels = map[string]string{}
		}

		agents = append(agents, agent)
	}

	return agents, rows.Err()
}
