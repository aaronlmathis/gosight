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

// gosight/agent/internal/sender/sender.go

package sender

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Sender holds the gRPC client and connection
type Sender struct {
	client proto.MetricsServiceClient
	cc     *grpc.ClientConn
	stream proto.MetricsService_SubmitStreamClient
}

// NewSender establishes a gRPC connection
func NewSender(ctx context.Context, cfg *config.AgentConfig) (*Sender, error) {

	// Load TLS config for agent
	tlsCfg, err := loadTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	// add mTLS to degug log.
	if len(tlsCfg.Certificates) > 0 {
		utils.Info("🔐 Using mTLS for agent authentication")
	} else {
		utils.Info("🔒 Using TLS only (no client certificate)")
	}

	// Establish gRPC connection
	clientConn, err := grpc.NewClient(cfg.ServerURL,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	)
	if err != nil {
		return nil, err
	}

	// Create gRPC client
	// and establish a stream for sending metrics
	client := proto.NewMetricsServiceClient(clientConn)
	utils.Info("📤 established gRPC Connection with %v", cfg.ServerURL)

	//
	stream, err := client.SubmitStream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	return &Sender{
		client: client,
		cc:     clientConn,
		stream: stream,
	}, nil

}

// Close the gRPC connection
func (s *Sender) Close() error {
	return s.cc.Close()
}

func (s *Sender) SendMetrics(payload model.MetricPayload) error {
	pbMetrics := make([]*proto.Metric, 0, len(payload.Metrics))
	for _, m := range payload.Metrics {
		pbMetric := &proto.Metric{
			Name:         m.Name,
			Namespace:    m.Namespace,
			Subnamespace: m.SubNamespace,
			Timestamp:    timestamppb.New(m.Timestamp),

			Value:             m.Value,
			Unit:              m.Unit,
			StorageResolution: int32(m.StorageResolution),
			Type:              m.Type,
			Dimensions:        m.Dimensions,
		}
		if m.StatisticValues != nil {
			pbMetric.StatisticValues = &proto.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int32(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		pbMetrics = append(pbMetrics, pbMetric)
	}
	var convertedMeta *proto.Meta
	// Convert meta to proto
	if payload.Meta != nil {
		convertedMeta = convertMetaToProto(payload.Meta)
	}
	//utils.Debug("🎯 Proto Meta Tags: %+v", convertedMeta)

	req := &proto.MetricPayload{
		Host:      payload.Host,
		Timestamp: timestamppb.New(payload.Timestamp),
		Metrics:   pbMetrics,
		Meta:      convertedMeta,
	}

	if err := s.stream.Send(req); err != nil {
		return fmt.Errorf("stream send failed: %w", err)
	}

	utils.Info("📤 Streamed %d metrics", len(pbMetrics))
	return nil
}
func convertMetaToProto(m *model.Meta) *proto.Meta {
	if m == nil {
		return nil
	}
	return &proto.Meta{
		Hostname:         m.Hostname,
		IpAddress:        m.IPAddress,
		Os:               m.OS,
		OsVersion:        m.OSVersion,
		KernelVersion:    m.KernelVersion,
		Architecture:     m.Architecture,
		CloudProvider:    m.CloudProvider,
		Region:           m.Region,
		AvailabilityZone: m.AvailabilityZone,
		InstanceId:       m.InstanceID,
		InstanceType:     m.InstanceType,
		AccountId:        m.AccountID,
		ProjectId:        m.ProjectID,
		ResourceGroup:    m.ResourceGroup,
		VpcId:            m.VPCID,
		SubnetId:         m.SubnetID,
		ImageId:          m.ImageID,
		ServiceId:        m.ServiceID,
		ContainerId:      m.ContainerID,
		ContainerName:    m.ContainerName,
		PodName:          m.PodName,
		Namespace:        m.Namespace,
		ClusterName:      m.ClusterName,
		NodeName:         m.NodeName,
		Application:      m.Application,
		Environment:      m.Environment,
		Service:          m.Service,
		Version:          m.Version,
		DeploymentId:     m.DeploymentID,
		PublicIp:         m.PublicIP,
		PrivateIp:        m.PrivateIP,
		MacAddress:       m.MACAddress,
		NetworkInterface: m.NetworkInterface,
		Tags:             m.Tags,
	}
}
