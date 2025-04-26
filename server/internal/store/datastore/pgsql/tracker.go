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

// gosight/agent/internal/store/datastore/pgdatastore/tracker.go
// tracker.go - defines the db fucntions for tracking endpoints

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

// ListAgents retrieves all agents from the database
// This is used for the web UI to display all agents
// and for the agent to check in with the server
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

// Container related functions

// UpsertContainer inserts or updates a container in the database
// This is used for the agent to check in with the server
// It also updates the last_seen field to the current time
// and the status field to "Running" if the container is running
// or "Stopped" if the container is stopped
// UpsertContainer inserts or updates a container in the database
func (s *PGDataStore) UpsertContainer(ctx context.Context, c *model.Container) error {
	labels, _ := json.Marshal(c.Labels) // You already use this

	_, err := s.db.ExecContext(ctx, `
        INSERT INTO containers (
            container_id,
            endpoint_id,
            host_id,
            name,
            image,
            image_id,
            runtime,
            status,
            last_seen,
            labels
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )
        ON CONFLICT (container_id) DO UPDATE SET
            endpoint_id = EXCLUDED.endpoint_id,
            host_id = EXCLUDED.host_id,
            name = EXCLUDED.name,
            image = EXCLUDED.image,
            image_id = EXCLUDED.image_id,
            runtime = EXCLUDED.runtime,
            status = EXCLUDED.status,
            last_seen = EXCLUDED.last_seen,
            labels = EXCLUDED.labels;
    `,
		c.ContainerID,
		c.EndpointID,
		c.HostID,
		c.Name,
		c.ImageName,
		c.ImageID,
		c.Runtime,
		c.Status,
		c.LastSeen,
		labels,
	)

	return err
}

// ListContainers retrieves all containers from the database
// This is used for the web UI to display all containers
// and for the agent to check in with the server
func (s *PGDataStore) ListContainers(ctx context.Context) ([]*model.Container, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT container_id, endpoint_id, host_id, name, image, image_id, runtime, status, last_seen, labels
		FROM containers
	`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var containers []*model.Container
	for rows.Next() {
		var c model.Container
		var labelsJSON []byte

		err := rows.Scan(
			&c.ContainerID,
			&c.EndpointID,
			&c.HostID,
			&c.Name,
			&c.ImageName,
			&c.ImageID,
			&c.Runtime,
			&c.Status,
			&c.LastSeen,
			&labelsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %v", err)
		}

		_ = json.Unmarshal(labelsJSON, &c.Labels)
		containers = append(containers, &c)
	}

	return containers, nil
}

// GetContainerByID retrieves a container by its container_id
func (s *PGDataStore) GetContainerByID(ctx context.Context, id string) (*model.Container, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT container_id, endpoint_id, host_id, name, image, image_id, runtime, status, last_seen, labels
		FROM containers WHERE container_id = $1
	`, id)

	var c model.Container
	var labelsJSON []byte

	err := row.Scan(
		&c.ContainerID,
		&c.EndpointID,
		&c.HostID,
		&c.Name,
		&c.ImageName,
		&c.ImageID,
		&c.Runtime,
		&c.Status,
		&c.LastSeen,
		&labelsJSON,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(labelsJSON, &c.Labels)
	return &c, nil
}
