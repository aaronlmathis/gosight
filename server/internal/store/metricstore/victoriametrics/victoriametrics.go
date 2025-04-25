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

// server/internal/store/victoriametrics.go

package victoriametricstore

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store/metricindex"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type VictoriaStore struct {
	url      string
	queue    chan []model.MetricPayload
	incoming chan []model.MetricPayload
	wg       sync.WaitGroup
	client   *http.Client
	ctx      context.Context

	// batching config
	batchSize     int
	batchTimeout  time.Duration
	batchRetry    int
	batchInterval time.Duration
	MetricIndex   *metricindex.MetricIndex
}

func NewVictoriaStore(ctx context.Context, url string, workers, queueSize, batchSize, timeoutMS, retry, retryIntervalMS int, metricIndex *metricindex.MetricIndex) *VictoriaStore {
	utils.Info("NewVictoriaStore received workers=%d", workers)
	store := &VictoriaStore{
		url:           url,
		queue:         make(chan []model.MetricPayload, queueSize),
		incoming:      make(chan []model.MetricPayload, queueSize),
		client:        &http.Client{Timeout: 10 * time.Second},
		ctx:           ctx,
		batchSize:     batchSize,
		batchTimeout:  time.Duration(timeoutMS) * time.Millisecond,
		batchRetry:    retry,
		batchInterval: time.Duration(retryIntervalMS) * time.Millisecond,
		MetricIndex:   metricIndex,
	}
	if workers == 0 {
		utils.Warn("VictoriaStore called with 0 workers!")
	} else {
		utils.Debug("Spawning %d workers now...", workers)
	}

	for i := 0; i < workers; i++ {
		store.wg.Add(1)

		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					utils.Error("Worker #%d panicked: %v", id, r)
				}
			}()
			utils.Info("🧵 Started worker #%d", id)
			store.worker()
		}(i + 1)
	}

	go store.collectorLoop()

	utils.Info("VictoriaStore initialized with %d workers", workers)
	utils.Debug("NewVictoriaStore created at address: %p", store)

	return store
}

func (v *VictoriaStore) Write(metrics []model.MetricPayload) error {
	//utils.Debug(" store.Write received: %d metrics (store addr: %p)", totalMetricCount(metrics), v)

	select {
	case v.incoming <- metrics:
		//utils.Debug("Write enqueued %d metrics", totalMetricCount(metrics))
		return nil
	default:
		utils.Warn("Incoming buffer full: dropping metrics")
		return fmt.Errorf("incoming buffer full")
	}
}

func (v *VictoriaStore) collectorLoop() {
	utils.Info("collectorLoop started")
	ticker := time.NewTicker(v.batchTimeout)
	defer ticker.Stop()

	//utils.Info("atchTimeout raw = %v\n", v.batchTimeout)
	//utils.Debug("collectorLoop started with timeout: %s", v.batchTimeout)

	var pending []model.MetricPayload

	for {
		select {
		case <-v.ctx.Done():
			utils.Debug("VictoriaStore collector loop exiting")
			if len(pending) > 0 {
				v.enqueue(pending)
			}
			return

		case batch := <-v.incoming:
			//total := totalMetricCount(batch)
			//utils.Debug(" Received payload with %d metrics", total)
			pending = append(pending, batch...)
			currentTotal := totalMetricCount(pending)
			//utils.Debug(" Total metrics pending: %d", currentTotal)

			if currentTotal >= v.batchSize {
				//utils.Info(" Batch size reached: %d metrics, flushing now", currentTotal)
				v.enqueue(pending)
				pending = nil
			}

		case <-ticker.C:
			currentTotal := totalMetricCount(pending)
			//utils.Debug(" Timeout ticked. Pending payloads: %d, metrics: %d", len(pending), currentTotal)

			if currentTotal > 0 {
				//utils.Info(" Timeout flush triggered for %d metrics", currentTotal)
				v.enqueue(pending)
				pending = nil
			}
		}
	}
}

func (v *VictoriaStore) enqueue(batch []model.MetricPayload) {
	//utils.Debug("Enqueue called with %d payloads / %d metrics",		len(batch), totalMetricCount(batch))
	select {
	case v.queue <- batch:
	default:
		utils.Warn("Worker queue full: dropping batch of %d metrics", len(batch))
	}
}

func (v *VictoriaStore) worker() {
	defer v.wg.Done()
	for {
		//utils.Debug(" Worker waiting for batch...")

		select {

		case batch := <-v.queue:
			//utils.Debug(" Worker received batch with %d payloads / %d metrics", len(batch), totalMetricCount(batch))
			v.flush(batch)
		case <-v.ctx.Done():
			utils.Debug("VictoriaStore collector loop exiting")

			return
		}
	}
}

func (v *VictoriaStore) flush(batch []model.MetricPayload) {

	payload := buildPrometheusFormat(batch)

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, _ = gz.Write([]byte(payload))
	_ = gz.Close()

	//utils.Debug(" Flushing batch of %d metrics", len(batch))

	req, err := http.NewRequest("POST", v.url+"/api/v1/import/prometheus", &buf)
	if err != nil {
		utils.Error("Request build failed: %v", err)
		return
	}
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "text/plain")

	for attempt := 0; attempt < v.batchRetry; attempt++ {
		resp, err := v.client.Do(req)
		if err == nil && resp.StatusCode < 300 {
			//utils.Debug("Batch sent successfully to VictoriaMetrics")
			return
		}
		utils.Warn("Retrying batch write... attempt %d", attempt+1)
		time.Sleep(v.batchInterval)
	}
	utils.Error("Failed to write batch after %d retries", v.batchRetry)
}

func (v *VictoriaStore) Close() error {
	utils.Info("Waiting for VictoriaStore workers to finish...")
	v.wg.Wait()
	utils.Info("VictoriaStore shutdown complete")
	return nil
}

func buildPrometheusFormat(batch []model.MetricPayload) string {
	var sb strings.Builder

	for _, payload := range batch {
		ts := payload.Timestamp.UnixNano() / 1e6

		// Core labels from Meta + Tags
		baseLabels := BuildPromLabels(payload.Meta)

		for _, m := range payload.Metrics {
			fullName := normalizeMetricName(m.Namespace, m.SubNamespace, m.Name)

			// Start with base Meta + Tags
			labels := make(map[string]string, len(baseLabels)+len(m.Dimensions))
			for k, v := range baseLabels {
				labels[k] = v
			}

			// Apply metric-specific dimensions (override any key)
			for k, v := range m.Dimensions {
				labels[k] = v
			}

			sb.WriteString(fmt.Sprintf("%s{%s} %f %d\n",
				fullName,
				formatLabelMap(labels),
				m.Value,
				ts,
			))
		}
	}

	return sb.String()
}

func normalizeMetricName(ns, sub, name string) string {
	var parts []string
	if ns != "" {
		parts = append(parts, strings.ToLower(strings.ReplaceAll(ns, "/", ".")))
	}
	if sub != "" {
		parts = append(parts, strings.ToLower(strings.ReplaceAll(sub, "/", ".")))
	}
	parts = append(parts, name)

	return strings.Join(parts, ".")
}

// formatLabelMap prepares potential labels for Prometheus scraping.
// It combines payload.Meta tags and metric dimensions into a single map.
// It converts the labels map to a string in the format: key1="value1",key2="value2",...
// It allows Dimensions to override Meta tags if two keys are the same.

func formatLabelMap(labels map[string]string) string {
	var parts []string
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(parts)
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
	if meta.AgentID != "" {
		labels["agent_id"] = meta.AgentID
	}
	if meta.AgentVersion != "" {
		labels["agent_version"] = meta.AgentVersion
	}
	if meta.HostID != "" {
		labels["host_id"] = meta.HostID
	}
	if meta.EndpointID != "" {
		labels["endpoint_id"] = meta.EndpointID
	}
	if meta.Hostname != "" {
		labels["hostname"] = meta.Hostname
		utils.Debug("HOSTNAME: %s", meta.Hostname)
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
	if meta.Platform != "" {
		labels["platform"] = meta.Platform
	}
	if meta.PlatformFamily != "" {
		labels["platform_family"] = meta.PlatformFamily
	}
	if meta.PlatformVersion != "" {
		labels["platform_version"] = meta.PlatformVersion
	}
	if meta.KernelArchitecture != "" {
		labels["kernel_architecture"] = meta.KernelArchitecture
	}
	if meta.KernelVersion != "" {
		labels["kernel_version"] = meta.KernelVersion
	}
	if meta.VirtualizationSystem != "" {
		labels["virtualization_system"] = meta.VirtualizationSystem
	}
	if meta.VirtualizationRole != "" {
		labels["virtualization_role"] = meta.VirtualizationRole
	}
	if meta.Architecture != "" {
		labels["architecture"] = meta.Architecture
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
	if meta.Application != "" {
		labels["application"] = meta.Application
	}
	if meta.Service != "" {
		labels["service"] = meta.Service
	}
	if meta.Version != "" {
		labels["version"] = meta.Version
	}
	if meta.DeploymentID != "" {
		labels["deployment_id"] = meta.DeploymentID
	}
	if meta.PublicIP != "" {
		labels["public_ip"] = meta.PublicIP
	}
	if meta.PrivateIP != "" {
		labels["private_ip"] = meta.PrivateIP
	}
	if meta.MACAddress != "" {
		labels["mac_address"] = meta.MACAddress
	}
	if meta.NetworkInterface != "" {
		labels["network_interface"] = meta.NetworkInterface
	}

	// Tags (pre-filtered to avoid duplication)
	for k, v := range meta.Tags {
		if _, exists := labels[k]; !exists {
			labels[k] = v
		}
	}

	return labels
}

// formatLabels formats the labels for Prometheus scraping.
// It converts the labels map to a string in the format: key1="value1",key2="value2",...
// This is used for building the Prometheus-compatible metric format.

func formatLabels(meta *model.Meta) string {
	labels := BuildPromLabels(meta)
	parts := make([]string, 0, len(labels))
	for k, v := range labels {
		parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
	}
	sort.Strings(parts) // for deterministic output
	return strings.Join(parts, ",")
}

func totalMetricCount(payloads []model.MetricPayload) int {
	count := 0
	for _, p := range payloads {
		count += len(p.Metrics)
	}
	return count
}

// QueryInstant fetches the latest data points for a given metric with optional label filters.
func (v *VictoriaStore) QueryInstant(metric string, filters map[string]string) ([]model.MetricRow, error) {
	query := BuildPromQL(metric, filters)

	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", v.url, url.QueryEscape(query))

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("VM instant query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]interface{}    `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", parsed.Status)
	}

	var rows []model.MetricRow
	for _, item := range parsed.Data.Result {
		strVal, ok := item.Value[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			continue
		}
		rows = append(rows, model.MetricRow{
			Tags:  item.Metric,
			Value: val,
		})
	}
	return rows, nil
}

// QueryRange fetches time series data for a metric over a time range with optional label filters.
func (v *VictoriaStore) QueryRange(metric string, start, end time.Time, filters map[string]string) ([]model.Point, error) {
	query := BuildPromQL(metric, filters)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", "15s")

	fullURL := fmt.Sprintf("%s/api/v1/query_range?%s", v.url, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("VM range query failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", parsed.Status)
	}

	var points []model.Point
	for _, series := range parsed.Data.Result {
		for _, val := range series.Values {
			tsRaw, ok1 := val[0].(float64)
			valStr, ok2 := val[1].(string)
			if !ok1 || !ok2 {
				continue
			}
			valFloat, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				continue
			}
			points = append(points, model.Point{
				Timestamp: time.Unix(int64(tsRaw), 0).UTC().Format(time.RFC3339),
				Value:     valFloat,
			})
		}
	}
	return points, nil
}

func (v *VictoriaStore) QueryMultiInstant(metricNames []string, filters map[string]string) ([]model.MetricRow, error) {
	//utils.Debug("Executing VictoriaStore.QueryMultiInstant")
	if len(metricNames) == 0 {
		// Default to all known metric names if available
		metricNames = v.GetAllKnownMetricNames()
		//utils.Debug("No metric names provided, using all known metric names: %v", metricNames)
	}

	// Build metric regex selector
	nameSelector := fmt.Sprintf(`__name__=~"%s"`, strings.Join(metricNames, "|"))

	// Combine all label filters
	var labelSelectors []string
	labelSelectors = append(labelSelectors, nameSelector)
	for k, val := range filters {
		// URL-encode the label value to handle special characters
		escapedVal := url.QueryEscape(val)
		labelSelectors = append(labelSelectors, fmt.Sprintf(`%s="%s"`, k, escapedVal))
	}

	// Construct final query string
	query := fmt.Sprintf("{%s}", strings.Join(labelSelectors, ", "))

	// Build full URL
	reqURL := fmt.Sprintf("%s/api/v1/query?query=%s", v.url, url.QueryEscape(query))
	//utils.Debug("QueryMultiInstant URL: %s", reqURL)
	// Make HTTP request
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("victoriametrics error: %s", string(body))
	}

	// Parse response
	var vmResp struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric map[string]string `json:"metric"`
				Value  [2]interface{}    `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&vmResp); err != nil {
		return nil, err
	}

	// Convert to MetricRow
	var rows []model.MetricRow
	for _, item := range vmResp.Data.Result {
		// Extract timestamp
		tsFloat, ok := item.Value[0].(float64)
		if !ok {
			continue
		}
		timestamp := int64(tsFloat * 1000) // seconds → milliseconds

		// Extract metric value
		valStr, ok := item.Value[1].(string)
		if !ok {
			continue
		}
		val, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			continue
		}

		rows = append(rows, model.MetricRow{
			Value:     val,
			Tags:      item.Metric,
			Timestamp: timestamp,
		})
	}

	return rows, nil
}

func (v *VictoriaStore) QueryMultiRange(metrics []string, start, end time.Time, filters map[string]string) ([]model.MetricRow, error) {
	if len(metrics) == 0 {
		return nil, nil
	}

	// Build PromQL selector
	nameSelector := fmt.Sprintf(`__name__=~"%s"`, strings.Join(metrics, "|"))

	var labelSelectors []string
	labelSelectors = append(labelSelectors, nameSelector)
	for k, val := range filters {
		labelSelectors = append(labelSelectors, fmt.Sprintf(`%s="%s"`, k, val))
	}
	query := fmt.Sprintf("{%s}", strings.Join(labelSelectors, ", "))

	// Build URL
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", "15s") // TODO: make configurable?

	fullURL := fmt.Sprintf("%s/api/v1/query_range?%s", v.url, params.Encode())
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("VM QueryMultiRange failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Values [][]interface{}   `json:"values"` // [ [timestamp, value], ... ]
			} `json:"result"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}
	if parsed.Status != "success" {
		return nil, fmt.Errorf("query failed: %s", parsed.Status)
	}

	var rows []model.MetricRow
	for _, series := range parsed.Data.Result {
		for _, val := range series.Values {
			tsRaw, ok1 := val[0].(float64)
			valStr, ok2 := val[1].(string)
			if !ok1 || !ok2 {
				continue
			}
			valFloat, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				continue
			}
			rows = append(rows, model.MetricRow{
				Timestamp: int64(tsRaw * 1000), // convert seconds → ms
				Value:     valFloat,
				Tags:      series.Metric,
			})
		}
	}

	return rows, nil
}

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

func (v *VictoriaStore) GetAllKnownMetricNames() []string {
	return v.MetricIndex.GetAllMetricNames()
}
