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

// gosight/agent/internal/utils/id.go
// Package utils provides utility functions for the GoSight agent.
// It includes functions to generate unique IDs for endpoints based on their metadata.

package utils

import (
	"fmt"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

func GenerateEndpointID(meta *model.Meta) string {
	if meta == nil {
		return "unknown"
	}

	provider := strings.ToLower(meta.CloudProvider)

	// Prefer cloud-specific identifiers if present
	switch provider {
	case "aws":
		if meta.AccountID != "" && meta.InstanceID != "" {
			return fmt.Sprintf("aws-%s-%s", sanitize(meta.AccountID), trim(meta.InstanceID))
		}
	case "gcp":
		if meta.ProjectID != "" && meta.InstanceID != "" {
			return fmt.Sprintf("gcp-%s-%s", sanitize(meta.ProjectID), trim(meta.InstanceID))
		}
	case "azure":
		if meta.ResourceGroup != "" && meta.InstanceID != "" {
			return fmt.Sprintf("azure-%s-%s", sanitize(meta.ResourceGroup), trim(meta.InstanceID))
		}
	}

	// Containers take precedence over host fallback
	if meta.ContainerID != "" {
		return "ctr-" + trim(meta.ContainerID)
	}

	// Fallback to host-based identity
	if meta.HostID != "" {
		return "host-" + trim(meta.HostID)
	}

	return "unknown"
}

func GetNamespace(meta *model.Meta) string {
	switch strings.ToLower(meta.CloudProvider) {
	case "aws":
		if meta.ServiceID == "ecs" {
			return "AWS/ECS"
		}
		return "AWS/EC2"
	case "gcp":
		return "GCP/Compute"
	case "azure":
		return "Azure/VM"
	}

	if meta.ContainerID != "" && meta.ClusterName != "" {
		return "K8s/Pod"
	}
	if meta.ContainerID != "" {
		return "Podman"
	}

	return "System"
}

func trim(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func sanitize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "-"))
}
