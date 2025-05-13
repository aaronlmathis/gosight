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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/logstore/victoriametricsvictoriametrics.go

package victorialogstore

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

// buildPrometheusFormatFromWrapped converts a slice of StoredLog into a Prometheus-compatible string format.

func buildPrometheusFormatFromWrapped(logs []*model.StoredLog) string {
	var sb strings.Builder
	maxLabels := 38
	for _, wrapped := range logs {
		log := wrapped.Log
		ts := log.Timestamp.UnixNano() / 1e6
		if ts == 0 {
			fmt.Printf("WARNING: log with log_id=%s has zero timestamp\n", wrapped.LogID)
		}

		labels := BuildPromLabels(wrapped.Meta) // base labels from model.Meta
		labels["log_id"] = wrapped.LogID
		labels["level"] = log.Level
		labels["source"] = log.Source

		if log.Category != "" {
			labels["category"] = log.Category
		}

		if log.PID != 0 {
			labels["pid"] = strconv.Itoa(log.PID)
		}
		add := func(k, v string) {
			if len(labels) < maxLabels {
				labels[k] = v
			}
		}
		for k, v := range log.Tags {
			add(sanitizeLabelKey(k), v)
		}

		if log.Meta != nil {
			if log.Meta.Platform != "" {
				labels["platform"] = log.Meta.Platform
			}
			if log.Meta.Unit != "" {
				labels["unit"] = log.Meta.Unit
			}
			if log.Meta.ContainerID != "" {
				labels["container_id"] = log.Meta.ContainerID
			}
			if log.Meta.ContainerName != "" {
				labels["container_name"] = log.Meta.ContainerName
			}
			if log.Meta.Service != "" {
				labels["service"] = log.Meta.Service
			}
			if log.Meta.AppName != "" {
				labels["app_name"] = log.Meta.AppName
			}
			if log.Meta.User != "" {
				labels["user"] = log.Meta.User
			}
			if log.Meta.EventID != "" {
				labels["event_id"] = log.Meta.EventID
			}
		}

		// Add fields (structured logs)
		for k, v := range log.Fields {
			add(sanitizeLabelKey("field_"+k), v)
		}

		sb.WriteString(fmt.Sprintf("gosight.logs.entry{%s} 1 %d\n",
			formatLabelMap(labels), ts))
	}

	return sb.String()
}

// formatLabelMap prepares potential labels for Prometheus scraping.
// It combines payload.Meta tags and log fields/tags into a single map.
// It converts the labels map to a string in the format: key1="value1",key2="value2",...
// It escapes values properly to ensure they are valid Prometheus label values.

func formatLabelMap(m map[string]string) string {
	var parts []string
	for k, v := range m {
		// strconv.Quote ensures proper escaping of quotes, slashes, etc.
		escaped := strconv.Quote(v)           // yields `"escaped string"`
		escaped = escaped[1 : len(escaped)-1] // remove outer quotes since we wrap it manually
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, escaped))
	}
	return strings.Join(parts, ",")
}

// BuildPromLabels constructs Prometheus-compatible labels from the given Meta object.
// It filters out any labels that are already present in the Meta object to avoid duplication.
// The resulting labels are returned as a map of key-value pairs.

func BuildPromLabels(meta *model.Meta) map[string]string {
	if meta == nil {
		return map[string]string{}
	}

	labels := map[string]string{}

	// Identity and system labels from Meta

	if meta.HostID != "" {
		labels["host_id"] = meta.HostID
	}
	if meta.EndpointID != "" {
		labels["endpoint_id"] = meta.EndpointID
	}
	if meta.Hostname != "" {
		labels["hostname"] = meta.Hostname
	}
	if meta.IPAddress != "" {
		labels["ip_address"] = meta.IPAddress
	}
	if meta.OS != "" {
		labels["os"] = meta.OS
	}
	if meta.OSVersion != "" {
		labels["os_version"] = meta.OSVersion
	}

	if meta.Environment != "" {
		labels["environment"] = meta.Environment
	}
	if meta.Region != "" {
		labels["region"] = meta.Region
	}
	if meta.AvailabilityZone != "" {
		labels["availability_zone"] = meta.AvailabilityZone
	}
	if meta.InstanceID != "" {
		labels["instance_id"] = meta.InstanceID
	}
	if meta.InstanceType != "" {
		labels["instance_type"] = meta.InstanceType
	}
	if meta.AccountID != "" {
		labels["account_id"] = meta.AccountID
	}
	if meta.ProjectID != "" {
		labels["project_id"] = meta.ProjectID
	}
	if meta.ResourceGroup != "" {
		labels["resource_group"] = meta.ResourceGroup
	}
	if meta.VPCID != "" {
		labels["vpc_id"] = meta.VPCID
	}
	if meta.SubnetID != "" {
		labels["subnet_id"] = meta.SubnetID
	}
	if meta.ImageID != "" {
		labels["image_id"] = meta.ImageID
	}
	if meta.ServiceID != "" {
		labels["service_id"] = meta.ServiceID
	}
	if meta.ContainerID != "" {
		labels["container_id"] = meta.ContainerID
	}
	if meta.ContainerName != "" {
		labels["container_name"] = meta.ContainerName
	}
	if meta.PodName != "" {
		labels["pod_name"] = meta.PodName
	}
	if meta.ClusterName != "" {
		labels["cluster_name"] = meta.ClusterName
	}
	if meta.NodeName != "" {
		labels["node_name"] = meta.NodeName
	}
	if meta.ContainerImageID != "" {
		labels["container_image_id"] = meta.ContainerImageID
	}
	if meta.ContainerImageName != "" {
		labels["container_image_name"] = meta.ContainerImageName
	}

	for k, v := range meta.Tags {
		labels[k] = v
	}

	return labels
}

func sanitizeLabelKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

// BuildPromQL constructs a Prometheus Query Language (PromQL) query string
func BuildPromQL(metric string, filters map[string]string) string {
	if len(filters) == 0 {
		return metric
	}
	var parts []string
	for k, v := range filters {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(parts)
	return fmt.Sprintf(`%s{%s}`, metric, strings.Join(parts, ","))
}

// wrapLogs converts a batch of LogPayloads into a slice of StoredLog.
func wrapLogs(batch []model.LogPayload) []*model.StoredLog {
	var result []*model.StoredLog
	for _, payload := range batch {
		meta := payload.Meta
		for _, log := range payload.Logs {
			logID := hash(log.Timestamp.String() + log.Message)
			result = append(result, &model.StoredLog{
				LogID: logID,
				Log:   log,
				Meta:  meta,
			})
		}
	}
	return result
}

// hash returns a short SHA1-based hash (first 10 characters).
func hash(input string) string {
	h := sha1.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))[:10]
}
