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

type StatisticValues struct {
	Minimum     float64 `json:"min"`
	Maximum     float64 `json:"max"`
	SampleCount int     `json:"count"`
	Sum         float64 `json:"sum"`
}

type Point struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

type Metric struct {
	Namespace         string            `json:"namespace,omitempty"`
	SubNamespace      string            `json:"subnamespace,omitempty"`
	Name              string            `json:"name"`
	Timestamp         time.Time         `json:"timestamp,omitempty"`
	Value             float64           `json:"value,omitempty"`
	StatisticValues   *StatisticValues  `json:"stats,omitempty"`
	Unit              string            `json:"unit,omitempty"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
	StorageResolution int               `json:"resolution,omitempty"`
	Type              string            `json:"type,omitempty"`
}

type Meta struct {
	// General Host Information
	Hostname             string `json:"hostname,omitempty"`
	IPAddress            string `json:"ip_address,omitempty"`
	OS                   string `json:"os,omitempty"`
	OSVersion            string `json:"os_version,omitempty"`
	Platform             string `json:"platform,omitempty"`
	PlatformFamily       string `json:"platform_family,omitempty"`
	PlatformVersion      string `json:"platform_version,omitempty"`
	KernelArchitecture   string `json:"kernel_architecture,omitempty"`
	VirtualizationSystem string `json:"virtualization_system,omitempty"`
	VirtualizationRole   string `json:"virtualization_role,omitempty"`
	HostID               string `json:"host_id,omitempty"`
	KernelVersion        string `json:"kernel_version,omitempty"`
	Architecture         string `json:"architecture,omitempty"`
	EndpointID           string `json:"endpoint_id,omitempty"` // Unique ID for the endpoint

	// Cloud Provider Specific
	CloudProvider    string `json:"cloud_provider,omitempty"` // AWS, Azure, GCP
	Region           string `json:"region,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"` // or Zone
	InstanceID       string `json:"instance_id,omitempty"`
	InstanceType     string `json:"instance_type,omitempty"`
	AccountID        string `json:"account_id,omitempty"`
	ProjectID        string `json:"project_id,omitempty"`     // GCP
	ResourceGroup    string `json:"resource_group,omitempty"` //Azure
	VPCID            string `json:"vpc_id,omitempty"`         // AWS, GCP
	SubnetID         string `json:"subnet_id,omitempty"`      // AWS, GCP, Azure
	ImageID          string `json:"image_id,omitempty"`       // AMI, Image, etc.
	ServiceID        string `json:"service_id,omitempty"`     // if a managed service is the source

	// Containerization/Orchestration
	ContainerID   string `json:"container_id,omitempty"`
	ContainerName string `json:"container_name,omitempty"`
	PodName       string `json:"pod_name,omitempty"`
	Namespace     string `json:"namespace,omitempty"` // K8s namespace
	ClusterName   string `json:"cluster_name,omitempty"`
	NodeName      string `json:"node_name,omitempty"`

	// Application Specific
	Application  string `json:"application,omitempty"`
	Environment  string `json:"environment,omitempty"` // dev, staging, prod
	Service      string `json:"service,omitempty"`     // if a microservice
	Version      string `json:"version,omitempty"`
	DeploymentID string `json:"deployment_id,omitempty"`

	// Network Information
	PublicIP         string `json:"public_ip,omitempty"`
	PrivateIP        string `json:"private_ip,omitempty"`
	MACAddress       string `json:"mac_address,omitempty"`
	NetworkInterface string `json:"network_interface,omitempty"`
	AgentVersion     string `json:"agent_version,omitempty"` // Version of the agent running
	AgentID          string `json:"agent_id,omitempty"`      // Unique ID for the agent

	// Custom Metadata
	Tags map[string]string `json:"tags,omitempty"` // Allow for arbitrary key-value pairs
}

type MetricPayload struct {
	EndpointID string    `json:"endpoint_id"`
	Timestamp  time.Time `json:"timestamp"`
	Metrics    []Metric  `json:"metrics"`
	Meta       *Meta     `json:"meta,omitempty"`
}

type MetricRow struct {
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
	Timestamp int64             `json:"timestamp"` // Unix ms
}

type MetricPoint struct {
	Timestamp int64   `json:"timestamp"` // Unix ms
	Value     float64 `json:"value"`     // Metric value
}

type MetricSelector struct {
	Name         string
	Namespace    string
	SubNamespace string
	Instant      bool // true = stat card, false = chart or long term
}
