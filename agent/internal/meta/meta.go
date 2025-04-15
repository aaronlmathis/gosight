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
	"github.com/shirou/gopsutil/v4/host"
)

func BuildHostMeta(cfg *config.Config, addTags map[string]string, agentID, agentVersion string) *model.Meta {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		utils.Warn("Failed to get hostname: %v", err)
	}

	ip := utils.GetLocalIP()
	if ip == "" {
		ip = "unknown"
		utils.Warn("Failed to get local IP address")
	}
	hostInfo, err := host.Info()
	if err != nil {
		utils.Warn("Failed to get host info: %v", err)
		hostInfo = &host.InfoStat{}
	}

	tags := utils.MergeMaps(cfg.Agent.CustomTags, addTags)

	// Add rich system info as tags
	tags["host_id"] = hostInfo.HostID
	tags["platform"] = hostInfo.Platform
	tags["platform_family"] = hostInfo.PlatformFamily
	tags["platform_version"] = hostInfo.PlatformVersion
	tags["kernel_version"] = hostInfo.KernelVersion
	tags["virtualization_system"] = hostInfo.VirtualizationSystem
	tags["virtualization_role"] = hostInfo.VirtualizationRole

	return &model.Meta{
		Hostname:             hostname,
		IPAddress:            ip,
		OS:                   hostInfo.OS,
		OSVersion:            hostInfo.PlatformVersion,
		Platform:             hostInfo.Platform,
		PlatformFamily:       hostInfo.PlatformFamily,
		PlatformVersion:      hostInfo.PlatformVersion,
		KernelArchitecture:   hostInfo.KernelArch,
		VirtualizationSystem: hostInfo.VirtualizationSystem,
		VirtualizationRole:   hostInfo.VirtualizationRole,
		HostID:               hostInfo.HostID,
		KernelVersion:        hostInfo.KernelVersion,
		Architecture:         runtime.GOARCH,
		Version:              agentVersion,
		AgentID:              agentID,
		AgentVersion:         agentVersion,
		Tags:                 tags,
	}
}

func BuildContainerMeta(cfg *config.Config, addTags map[string]string, agentID, agentVersion string) *model.Meta {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		utils.Warn("Failed to get hostname: %v", err)
	}

	ip := utils.GetLocalIP()
	if ip == "" {
		ip = "unknown"
		utils.Warn("Failed to get local IP address")
	}

	tags := utils.MergeMaps(cfg.Agent.CustomTags, addTags)

	return &model.Meta{
		Hostname:     hostname,
		IPAddress:    ip,
		AgentID:      agentID,
		AgentVersion: agentVersion,
		Tags:         tags,
	}
}
