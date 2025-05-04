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
// Package model contains the data structures used in GoSight.
package model

import "time"

// ProcessInfo represents information about a single process.
type ProcessInfo struct {
	PID        int               `json:"pid"`
	PPID       int               `json:"ppid"`
	User       string            `json:"user"`
	Executable string            `json:"exe"`
	Cmdline    string            `json:"cmdline"`
	CPUPercent float64           `json:"cpu_percent"`
	MemPercent float64           `json:"mem_percent"`
	Threads    int               `json:"threads"`
	StartTime  time.Time         `json:"start_time"`
	Tags       map[string]string `json:"tags,omitempty"` // container_id, container_name, etc
}

// ProcessSnapshot represents a snapshot of processes running on a host at a specific time.
type ProcessSnapshot struct {
	Timestamp  time.Time     `json:"timestamp"`
	HostID     string        `json:"host_id"`
	EndpointID string        `json:"endpoint_id"`
	Processes  []ProcessInfo `json:"processes"`
	Meta       *Meta         `json:"meta"` // Same as metric/log meta
}

type ProcessPayload struct {
	AgentID    string        `json:"agent_id"`
	HostID     string        `json:"host_id"`
	Hostname   string        `json:"hostname"`
	EndpointID string        `json:"endpoint_id"`
	Processes  []ProcessInfo `json:"processes"`
	Meta       *Meta         `json:"meta,omitempty"`
	Timestamp  time.Time     `json:"timestamp"`
}

type ProcessQueryFilter struct {
	EndpointID      string
	Start           time.Time
	End             time.Time
	MinCPU          float64
	MaxCPU          float64
	MinMemory       float64
	MaxMemory       float64
	User            string
	PID             int
	PPID            int
	ExeContains     string
	CmdlineContains string
	Tags            map[string]string
	SortBy          string
	SortDesc        bool
	Limit           int
	Offset          int
}
