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

package model

import "time"

type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`      // When the log was emitted
	Level     string            `json:"level"`          // info, warn, error, debug, etc.
	Message   string            `json:"message"`        // The actual log content
	Source    string            `json:"source"`         // Collector name or service name (e.g., journald, nginx)
	Category  string            `json:"category"`       // Optional: auth, network, system, app, etc.
	PID       int               `json:"pid,omitempty"`  // Process ID if available
	Fields    map[string]string `json:"fields"`         // Structured fields (JSON logs, key/values)
	Tags      map[string]string `json:"tags"`           // Custom labels/tags (user-defined or enriched)
	Meta      *LogMeta          `json:"meta,omitempty"` // Optional platform/service-specific metadata
}

type LogMeta struct {
	Platform      string            `json:"platform,omitempty"`       // journald, eventlog, syslog, etc.
	AppName       string            `json:"app_name,omitempty"`       // e.g., nginx, sshd
	AppVersion    string            `json:"app_version,omitempty"`    // if known
	ContainerID   string            `json:"container_id,omitempty"`   // if inside a container
	ContainerName string            `json:"container_name,omitempty"` // optional
	Unit          string            `json:"unit,omitempty"`           // For journald: systemd unit name
	Service       string            `json:"service,omitempty"`        // For syslog/Windows Event Log
	EventID       string            `json:"event_id,omitempty"`       // Windows event ID, etc.
	User          string            `json:"user,omitempty"`           // User associated with log entry
	Executable    string            `json:"exe,omitempty"`            // Path to binary if available
	Path          string            `json:"path,omitempty"`           // Original source log path
	Extra         map[string]string `json:"extra,omitempty"`          // For collector-specific fields
}

type LogPayload struct {
	AgentID    string
	HostID     string
	Hostname   string
	EndpointID string
	Timestamp  time.Time
	Logs       []LogEntry
	Meta       *Meta
}

type StoredLog struct {
	LogID string   `json:"log_id"`
	Log   LogEntry `json:"log"`
	Meta  *Meta    `json:"meta,omitempty"` // from LogPayload
}

// LogFilter is used to filter logs based on various criteria.
// It includes the limit of logs to return, the log levels to filter by,
// the unit of the logs, the source of the logs, a string to search for in the logs,
// and the start and end times for the logs.
type LogFilter struct {
	// Time range filter
	Start time.Time // Filter logs from this time onward
	End   time.Time // Filter logs until this time

	// Filtering by log properties
	EndpointID    string // Filter by endpoint ID (e.g., "host-123", "container-xyz")
	Target        string // Filter by target (e.g., "gosight-core", "host-123")
	Level         string // Filter by log level (e.g., "info", "warning", "error")
	Category      string // Filter by category (e.g., "system", "metric", "security")
	Source        string // Filter by source (e.g., "docker", "podman", "system")
	Contains      string // Filter by a substring match in the message
	Unit          string // Filter by systemd unit name (e.g., "nginx.service")
	AppName       string // Filter by application name (e.g., "nginx", "sshd")
	Service       string // For syslog/Windows Event Log
	EventID       string
	User          string
	ContainerID   string // if inside a container
	Platform      string
	ContainerName string // Windows event ID, etc.
	Meta          map[string]string
	Tags          map[string]string // Additional tags to filter by
	Extra         map[string]string
	Fields        map[string]string // Additional fields to filter by
	// Limit and sorting
	Limit  int    // Max number of logs to return
	Order  string // Order direction: "asc" or "desc"
	Cursor time.Time
	Offset int
}

// Standardized Log Categories
/*

system				OS-level logs: kernel, systemd, startup, shutdown
auth				Login attempts, sudo, PAM, user sessions, MFA
security			Firewall logs, intrusion detection, policy enforcement
network				DHCP, DNS, IP allocation, connection errors, interface events
app					Application-specific logs (e.g., nginx, postgres, your Go apps)
container			Podman/Docker lifecycle, logs from inside containers
metric				Logs related to metric collection or processing
gosight				Internal logs from the GoSight agent/server
scheduler			Cron-like jobs, backups, etc.
config				Configuration changes, reloads, validation errors
audit				User actions, RBAC events, config edits, role changes (great for IAM auditing)
alert				Alert evaluation or dispatch activity
*/
