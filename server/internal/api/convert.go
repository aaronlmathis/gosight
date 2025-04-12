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

package api

import (
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
)

func ConvertToModelPayload(pbPayload *proto.MetricPayload) model.MetricPayload {
	metrics := make([]model.Metric, 0, len(pbPayload.Metrics))
	var modelMeta *model.Meta
	for _, m := range pbPayload.Metrics {
		metric := model.Metric{
			Namespace:         m.Namespace,
			SubNamespace:      m.Subnamespace,
			Name:              m.Name,
			Timestamp:         m.Timestamp.AsTime(),
			Value:             m.Value,
			Unit:              m.Unit,
			Dimensions:        m.Dimensions,
			StorageResolution: int(m.StorageResolution),
			Type:              m.Type,
		}
		if m.StatisticValues != nil {
			metric.StatisticValues = &model.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		//utils.Debug("Received metric: %v", metric)
		if pbPayload.Meta != nil {
			modelMeta = convertProtoMetaToModelMeta(pbPayload.Meta)
		}

		metrics = append(metrics, metric)
	}
	return model.MetricPayload{
		Host:      pbPayload.Host,
		Timestamp: pbPayload.Timestamp.AsTime(),
		Metrics:   metrics,
		Meta:      modelMeta,
	}
}

func convertProtoMetaToModelMeta(pbMeta *proto.Meta) *model.Meta {
	if pbMeta == nil {
		return nil
	}
	return &model.Meta{
		EndpointID:       pbMeta.EndpointId,
		Hostname:         pbMeta.Hostname,
		IPAddress:        pbMeta.IpAddress,
		OS:               pbMeta.Os,
		OSVersion:        pbMeta.OsVersion,
		KernelVersion:    pbMeta.KernelVersion,
		Architecture:     pbMeta.Architecture,
		CloudProvider:    pbMeta.CloudProvider,
		Region:           pbMeta.Region,
		AvailabilityZone: pbMeta.AvailabilityZone,
		InstanceID:       pbMeta.InstanceId,
		InstanceType:     pbMeta.InstanceType,
		AccountID:        pbMeta.AccountId,
		ProjectID:        pbMeta.ProjectId,
		ResourceGroup:    pbMeta.ResourceGroup,
		VPCID:            pbMeta.VpcId,
		SubnetID:         pbMeta.SubnetId,
		ImageID:          pbMeta.ImageId,
		ServiceID:        pbMeta.ServiceId,
		ContainerID:      pbMeta.ContainerId,
		ContainerName:    pbMeta.ContainerName,
		PodName:          pbMeta.PodName,
		Namespace:        pbMeta.Namespace,
		ClusterName:      pbMeta.ClusterName,
		NodeName:         pbMeta.NodeName,
		Application:      pbMeta.Application,
		Environment:      pbMeta.Environment,
		Service:          pbMeta.Service,
		Version:          pbMeta.Version,
		DeploymentID:     pbMeta.DeploymentId,
		PublicIP:         pbMeta.PublicIp,
		PrivateIP:        pbMeta.PrivateIp,
		MACAddress:       pbMeta.MacAddress,
		NetworkInterface: pbMeta.NetworkInterface,
		Tags:             pbMeta.Tags,
	}
}
