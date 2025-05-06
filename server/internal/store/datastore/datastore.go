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

package datastore

import (
	"context"

	"github.com/aaronlmathis/gosight/shared/model"
)

type DataStore interface {
	// Satisfy Buffered Data Store interface
	Write(ctx context.Context, batches []*model.ProcessPayload) error

	// Agent methods
	UpsertAgent(ctx context.Context, agent *model.Agent) error
	GetAgentByHostname(ctx context.Context, hostname string) (*model.Agent, error)
	GetAgentByID(ctx context.Context, id string) (*model.Agent, error)
	ListAgents(ctx context.Context) ([]*model.Agent, error)

	// Container methods
	UpsertContainer(ctx context.Context, container *model.Container) error
	GetContainerByID(ctx context.Context, id string) (*model.Container, error)
	ListContainers(ctx context.Context) ([]*model.Container, error)

	// Tag methods
	GetTags(ctx context.Context, endpointID string) (map[string]string, error)
	SetTags(ctx context.Context, endpointID string, tags map[string]string) error
	DeleteTag(ctx context.Context, endpointID, key string) error
	ListTags(ctx context.Context, endpointID string) (map[string]string, error)
	GetAllTags(ctx context.Context) ([]model.Tag, error)

	ListKeys(ctx context.Context) ([]string, error)
	ListValues(ctx context.Context, key string) ([]string, error)

	// Process methods
	InsertFullProcessPayload(ctx context.Context, payload *model.ProcessPayload) error
	InsertProcessSnapshot(ctx context.Context, snap *model.ProcessPayload) (int64, error)
	InsertProcessInfos(ctx context.Context, snapshotID int64, payload *model.ProcessPayload) error
	QueryProcessInfos(ctx context.Context, filter *model.ProcessQueryFilter) ([]model.ProcessInfo, error)

	// Lifecycle
	Close() error
}
