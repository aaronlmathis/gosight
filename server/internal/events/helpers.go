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

package events

import (
	"github.com/aaronlmathis/gosight/shared/model"
)

func BuildLogEventMeta(log *model.LogEntry, payload *model.LogPayload) map[string]string {
	meta := make(map[string]string)

	// Core identity from payload (not payload.Meta)
	if payload.AgentID != "" {
		meta["agent_id"] = payload.AgentID
	}
	if payload.HostID != "" {
		meta["host_id"] = payload.HostID
	}
	if payload.Hostname != "" {
		meta["hostname"] = payload.Hostname
	}
	if payload.EndpointID != "" {
		meta["endpoint_id"] = payload.EndpointID
	}

	// LogMeta (from log.Meta)
	if log.Meta != nil {
		if log.Meta.Platform != "" {
			meta["log_platform"] = log.Meta.Platform
		}
		if log.Meta.AppName != "" {
			meta["app_name"] = log.Meta.AppName
		}
		if log.Meta.AppVersion != "" {
			meta["app_version"] = log.Meta.AppVersion
		}
		if log.Meta.ContainerID != "" {
			meta["container_id"] = log.Meta.ContainerID
		}
		if log.Meta.ContainerName != "" {
			meta["container_name"] = log.Meta.ContainerName
		}
		if log.Meta.Unit != "" {
			meta["unit"] = log.Meta.Unit
		}
		if log.Meta.Service != "" {
			meta["service"] = log.Meta.Service
		}
		if log.Meta.EventID != "" {
			meta["event_id"] = log.Meta.EventID
		}
		if log.Meta.User != "" {
			meta["user"] = log.Meta.User
		}
		if log.Meta.Executable != "" {
			meta["exe"] = log.Meta.Executable
		}
		if log.Meta.Path != "" {
			meta["path"] = log.Meta.Path
		}
		for k, v := range log.Meta.Extra {
			if v != "" {
				meta[k] = v
			}
		}
	}

	// System + cloud metadata (from payload.Meta)
	if payload.Meta != nil {
		add := func(k, v string) {
			if v != "" {
				meta[k] = v
			}
		}

		// System
		add("ip", payload.Meta.IPAddress)
		add("os", payload.Meta.OS)
		add("os_version", payload.Meta.OSVersion)
		add("platform", payload.Meta.Platform)
		add("platform_family", payload.Meta.PlatformFamily)
		add("platform_version", payload.Meta.PlatformVersion)
		add("kernel_architecture", payload.Meta.KernelArchitecture)
		add("kernel_version", payload.Meta.KernelVersion)
		add("virtualization_system", payload.Meta.VirtualizationSystem)
		add("virtualization_role", payload.Meta.VirtualizationRole)
		add("architecture", payload.Meta.Architecture)

		// Cloud
		add("cloud_provider", payload.Meta.CloudProvider)
		add("region", payload.Meta.Region)
		add("availability_zone", payload.Meta.AvailabilityZone)
		add("instance_id", payload.Meta.InstanceID)
		add("instance_type", payload.Meta.InstanceType)
		add("account_id", payload.Meta.AccountID)
		add("project_id", payload.Meta.ProjectID)
		add("resource_group", payload.Meta.ResourceGroup)
		add("vpc_id", payload.Meta.VPCID)
		add("subnet_id", payload.Meta.SubnetID)
		add("image_id", payload.Meta.ImageID)
		add("service_id", payload.Meta.ServiceID)

		// Container/K8s
		add("container_id", payload.Meta.ContainerID)
		add("container_name", payload.Meta.ContainerName)
		add("container_image_id", payload.Meta.ContainerImageID)
		add("container_image_name", payload.Meta.ContainerImageName)
		add("pod_name", payload.Meta.PodName)
		add("namespace", payload.Meta.Namespace)
		add("cluster_name", payload.Meta.ClusterName)
		add("node_name", payload.Meta.NodeName)

		// App
		add("application", payload.Meta.Application)
		add("environment", payload.Meta.Environment)
		add("service", payload.Meta.Service)
		add("version", payload.Meta.Version)
		add("deployment_id", payload.Meta.DeploymentID)

		// Network
		add("public_ip", payload.Meta.PublicIP)
		add("private_ip", payload.Meta.PrivateIP)
		add("mac_address", payload.Meta.MACAddress)
		add("network_interface", payload.Meta.NetworkInterface)

		// Tags (custom)
		for k, v := range payload.Meta.Tags {
			if v != "" {
				meta[k] = v
			}
		}
	}

	return meta
}