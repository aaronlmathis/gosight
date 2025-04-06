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

// // gosight/agent/internal/meta/tags.go
// // Sets up standard tags for metrics.

package meta

import (
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// BuildStandardTags sets required labels for consistent metric identity and filtering.
func BuildStandardTags(meta *model.Meta, m model.Metric, isContainer bool) {
	if meta.Tags == nil {
		meta.Tags = make(map[string]string)
	}
	// Contextual source of the metric
	meta.Tags["namespace"] = strings.ToLower(m.Namespace)
	meta.Tags["subnamespace"] = strings.ToLower(m.SubNamespace)

	// Producer of metric becomes the "job"
	if isContainer {
		meta.Tags["job"] = "gosight-container"

		if meta.ContainerName != "" {
			meta.Tags["instance"] = meta.ContainerName
		} else if meta.ContainerID != "" {
			meta.Tags["container_id"] = meta.ContainerID
		} else if meta.ImageID != "" {
			meta.Tags["image"] = meta.ImageID
		} else {
			meta.Tags["instance"] = "unknown-container"
		}
	} else {
		meta.Tags["instance"] = meta.Hostname
		meta.Tags["job"] = "gosight-agent"
	}

	// Final identity key
	meta.Tags["endpoint_id"] = utils.GenerateEndpointID(meta)
}
