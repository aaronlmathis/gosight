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
// Model for endpoint (host / container)
// shared/model/endpoint.go

package model

import "time"

type Endpoint struct {
	EndpointID    string            `json:"endpoint_id"` // Computed from HostID
	HostID        string            `json:"host_id"`
	Hostname      string            `json:"hostname"`
	ContainerName string            `json:"container_name,omitempty"` // Only for containers
	IP            string            `json:"ip"`
	OS            string            `json:"os,omitempty"`
	Arch          string            `json:"arch,omitempty"`
	Version       string            `json:"version,omitempty"`
	Labels        map[string]string `json:"labels"`
	ContainerID   string            `json:"container_id,omitempty"` // Only for containers
	Name          string            `json:"name,omitempty"`         // Only for containers
	ImageName     string            `json:"image_name,omitempty"`   // Only for containers
	ImageID       string            `json:"image_id,omitempty"`     // Only for containers
	Runtime       string            `json:"runtime,omitempty"`      // Only for containers
	LastSeen      time.Time         `json:"last_seen"`
	Status        string            `json:"status"` // online/offline
	UptimeSeconds float64           `json:"uptime_seconds"`
	Updated       bool              `json:"-"` // For in-memory use only
}
