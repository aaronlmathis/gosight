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

// gosight/agent/api/convert.go
// convert.go - converts internal metric payloads to protobuf format for gRPC.

package api

import (
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToProtoPayload(payload model.MetricPayload) *proto.MetricPayload {
	metrics := make([]*proto.Metric, 0, len(payload.Metrics))
	for _, m := range payload.Metrics {
		pm := &proto.Metric{
			Namespace:         m.Namespace,
			Name:              m.Name,
			Subnamespace:      m.SubNamespace,
			Timestamp:         timestamppb.New(m.Timestamp),
			Value:             m.Value,
			Unit:              m.Unit,
			Dimensions:        m.Dimensions,
			StorageResolution: int32(m.StorageResolution),
			Type:              m.Type,
		}
		if m.StatisticValues != nil {
			pm.StatisticValues = &proto.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int32(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		metrics = append(metrics, pm)
	}
	// Convert proto meta into model meta
	pbMeta := ConvertMetaToProtoMeta(payload.Meta)
	if pbMeta == nil {
		pbMeta = &proto.Meta{}
	}

	return &proto.MetricPayload{
		Host:      payload.Host,
		Timestamp: timestamppb.New(payload.Timestamp),
		Metrics:   metrics,
		Meta:      pbMeta,
	}
}

func ConvertLogMetaToProto(m *model.LogMeta) *proto.LogMeta {
	if m == nil {
		return nil
	}
	return &proto.LogMeta{
		Os:            m.OS,
		Platform:      m.Platform,
		AppName:       m.AppName,
		AppVersion:    m.AppVersion,
		ContainerId:   m.ContainerID,
		ContainerName: m.ContainerName,
		Unit:          m.Unit,
		Service:       m.Service,
		EventId:       m.EventID,
		User:          m.User,
		Executable:    m.Executable,
		Path:          m.Path,
		Extra:         m.Extra,
	}
}

func ConvertMetaToProtoMeta(m *model.Meta) *proto.Meta {
	if m == nil {
		return nil
	}
	//utils.Debug("📦 Converting meta to proto: %v", m)
	return &proto.Meta{
		EndpointId:           m.EndpointID,
		Hostname:             m.Hostname,
		IpAddress:            m.IPAddress,
		Os:                   m.OS,
		OsVersion:            m.OSVersion,
		Platform:             m.Platform,
		PlatformFamily:       m.PlatformFamily,
		PlatformVersion:      m.PlatformVersion,
		KernelArchitecture:   m.KernelArchitecture,
		VirtualizationSystem: m.VirtualizationSystem,
		VirtualizationRole:   m.VirtualizationRole,
		KernelVersion:        m.KernelVersion,
		Architecture:         m.Architecture,
		CloudProvider:        m.CloudProvider,
		Region:               m.Region,
		AvailabilityZone:     m.AvailabilityZone,
		InstanceId:           m.InstanceID,
		InstanceType:         m.InstanceType,
		AccountId:            m.AccountID,
		ProjectId:            m.ProjectID,
		ResourceGroup:        m.ResourceGroup,
		VpcId:                m.VPCID,
		SubnetId:             m.SubnetID,
		ImageId:              m.ImageID,
		ServiceId:            m.ServiceID,
		ContainerId:          m.ContainerID,
		ContainerName:        m.ContainerName,
		PodName:              m.PodName,
		Namespace:            m.Namespace,
		ClusterName:          m.ClusterName,
		NodeName:             m.NodeName,
		Application:          m.Application,
		Environment:          m.Environment,
		Service:              m.Service,
		Version:              m.Version,
		DeploymentId:         m.DeploymentID,
		PublicIp:             m.PublicIP,
		PrivateIp:            m.PrivateIP,
		MacAddress:           m.MACAddress,
		NetworkInterface:     m.NetworkInterface,
		Tags:                 m.Tags,
		AgentId:              m.AgentID,
		AgentVersion:         m.AgentVersion,
	}
}

func ConvertLogToProtoPayload(payload model.LogPayload) *proto.LogPayload {
	logs := make([]*proto.LogEntry, 0, len(payload.Logs))

	for _, l := range payload.Logs {
		entry := &proto.LogEntry{
			Timestamp: timestamppb.New(l.Timestamp),
			Level:     l.Level,
			Message:   l.Message,
			Source:    l.Source,
			Category:  l.Category,
			Host:      l.Host,
			Pid:       int32(l.PID),
			Fields:    l.Fields,
			Tags:      l.Tags,
		}

		if l.Meta != nil {
			entry.Meta = &proto.LogMeta{
				Os:            l.Meta.OS,
				Platform:      l.Meta.Platform,
				AppName:       l.Meta.AppName,
				AppVersion:    l.Meta.AppVersion,
				ContainerId:   l.Meta.ContainerID,
				ContainerName: l.Meta.ContainerName,
				Unit:          l.Meta.Unit,
				Service:       l.Meta.Service,
				EventId:       l.Meta.EventID,
				User:          l.Meta.User,
				Executable:    l.Meta.Executable,
				Path:          l.Meta.Path,
				Extra:         l.Meta.Extra,
			}
		}

		logs = append(logs, entry)
	}

	meta := ConvertLogMetaToProto(payload.Meta)
	if meta == nil {
		meta = &proto.LogMeta{}
	}
	// TODO fix meta
	return &proto.LogPayload{
		EndpointId: payload.EndpointID,
		Timestamp:  timestamppb.New(payload.Timestamp),
		Logs:       logs,
	}
}
