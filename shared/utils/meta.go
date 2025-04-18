package utils

import (
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

// ExtractStandardLabels builds a consistent map of labels from model.Meta.
func ExtractStandardLabels(meta *model.Meta) map[string]string {
	if meta == nil {
		return map[string]string{}
	}

	labels := map[string]string{
		"hostname":      meta.Hostname,
		"ip_address":    meta.IPAddress,
		"os":            meta.OS,
		"arch":          meta.Architecture,
		"endpoint_id":   meta.EndpointID,
		"agent_id":      meta.AgentID,
		"agent_version": meta.AgentVersion,
		"job":           "gosight-agent",
	}

	if meta.ContainerID != "" {
		labels["job"] = "gosight-container"
		labels["container_id"] = meta.ContainerID
	}

	if meta.ContainerName != "" {
		labels["instance"] = meta.ContainerName
	} else if meta.PodName != "" {
		labels["instance"] = meta.PodName
	} else {
		labels["instance"] = meta.Hostname
	}

	if ns, ok := meta.Tags["namespace"]; ok {
		labels["namespace"] = strings.ToLower(ns)
	}
	if sub, ok := meta.Tags["subnamespace"]; ok {
		labels["subnamespace"] = strings.ToLower(sub)
	}

	for k, v := range meta.Tags {
		if _, exists := labels[k]; !exists {
			labels[k] = v
		}
	}

	return labels
}

// ExtractLogLabels builds labels from model.LogMeta.
func ExtractLogLabels(meta *model.LogMeta) map[string]string {
	if meta == nil {
		return map[string]string{}
	}

	labels := map[string]string{
		"platform":       meta.Platform,
		"app_name":       meta.AppName,
		"container_id":   meta.ContainerID,
		"container_name": meta.ContainerName,
		"service":        meta.Service,
		"event_id":       meta.EventID,
	}

	for k, v := range meta.Extra {
		if _, exists := labels[k]; !exists {
			labels[k] = v
		}
	}

	return labels
}
