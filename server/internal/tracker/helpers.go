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

// Store agent details/heartbeats
// server/internal/store/helpers.go

package tracker

import (
	"fmt"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

// BuildAgentEventMeta builds a map of agent metadata for event tracking.
// This is used to track agent events such as heartbeats and status changes.
// The metadata includes various attributes of the agent such as agent ID,
// host ID, hostname, IP address, OS, architecture, version, endpoint ID,
// status, and uptime in seconds. The last seen time is formatted in RFC3339.
// The function takes an agent object as input and returns a map of strings
// representing the metadata.

func BuildAgentEventMeta(agent *model.Agent) map[string]string {
	meta := map[string]string{
		"agent_id":    agent.AgentID,
		"host_id":     agent.HostID,
		"hostname":    agent.Hostname,
		"ip":          agent.IP,
		"os":          agent.OS,
		"arch":        agent.Arch,
		"version":     agent.Version,
		"endpoint_id": agent.EndpointID,
		"status":      agent.Status,
		"since":       agent.Since,
		"last_seen":   agent.LastSeen.Format(time.RFC3339),
		"uptime_secs": fmt.Sprintf("%.0f", agent.UptimeSeconds),
	}
	return meta
}

// BuildContainerEventMeta builds a map of container metadata for event tracking.
// This is used to track container events such as creation, updates, and deletions.
// The metadata includes various attributes of the container such as container ID,
// name, image, runtime, status, endpoint ID, and host ID. The function takes a
// container object as input and returns a map of strings representing the metadata.
func BuildContainerEventMeta(c *model.Container) map[string]string {
	return map[string]string{
		"container_id": c.ContainerID,
		"name":         c.Name,
		"image":        c.ImageName,
		"image_id":     c.ImageID,
		"runtime":      c.Runtime,
		"status":       c.Status,
		"endpoint_id":  c.EndpointID,
		"host_id":      c.HostID,
	}
}

func NormalizeContainerStatus(raw string) string {
	switch strings.ToLower(raw) {
	case "created":
		return "Created"
	case "running":
		return "Running"
	case "paused":
		return "Paused"
	case "restarting":
		return "Restarting"
	case "removing":
		return "Removing"
	case "exited":
		return "Exited"
	case "dead":
		return "Dead"
	case "stopped":
		return "Stopped"
	default:
		return "Unknown"
	}
}
