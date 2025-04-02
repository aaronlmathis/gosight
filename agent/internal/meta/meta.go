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

// // gosight/agent/internal/meta/meta.go

package meta

import (
	"os"
	"runtime"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func BuildMeta(cfg *config.AgentConfig, addTags map[string]string) *model.Meta {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		utils.Warn("⚠️ Failed to get hostname: %v", err)
	}

	ip := utils.GetLocalIP()
	if ip == "" {
		ip = "unknown"
		utils.Warn("⚠️ Failed to get local IP address")
	}

	tags := utils.MergeMaps(cfg.CustomTags, addTags)

	return &model.Meta{
		Hostname:     hostname,
		IPAddress:    ip,
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		Version:      "0.1", // you could inject this via build flags
		Tags:         tags,
	}
}
