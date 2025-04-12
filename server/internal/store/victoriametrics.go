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

package store

import (
	"bytes"
	"compress/gzip"
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

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type VictoriaStore struct {
	url      string
	queue    chan []model.MetricPayload
	incoming chan []model.MetricPayload
	wg       sync.WaitGroup
	client   *http.Client
	stopChan chan struct{}

	// batching config
	batchSize     int
	batchTimeout  time.Duration
	batchRetry    int
	batchInterval time.Duration
}

func NewVictoriaStore(url string, workers, queueSize, batchSize, timeoutMS, retry, retryIntervalMS int) *VictoriaStore {
	utils.Info("📊 NewVictoriaStore received workers=%d", workers)
	store := &VictoriaStore{
		url:           url,
		queue:         make(chan []model.MetricPayload, queueSize),
		incoming:      make(chan []model.MetricPayload, queueSize),
		client:        &http.Client{Timeout: 10 * time.Second},
		stopChan:      make(chan struct{}),
		batchSize:     batchSize,
		batchTimeout:  time.Duration(timeoutMS) * time.Millisecond,
		batchRetry:    retry,
		batchInterval: time.Duration(retryIntervalMS) * time.Millisecond,
	}
	if workers == 0 {
		utils.Warn("⚠️ VictoriaStore called with 0 workers!")
	} else {
		utils.Debug("🧵 Spawning %d workers now...", workers)
	}

	for i := 0; i < workers; i++ {
		store.wg.Add(1)

		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					utils.Error("💥 Worker #%d panicked: %v", id, r)
				}
			}()
			utils.Info("🧵 Started worker #%d", id)
			store.worker()
		}(i + 1)
	}

	go store.collectorLoop()

	utils.Info("VictoriaStore initialized with %d workers", workers)
	utils.Debug("🏗️ NewVictoriaStore created at address: %p", store)

	return store
}

func (v *VictoriaStore) Write(metrics []model.MetricPayload) error {
	utils.Debug("✉️ store.Write received: %d metrics (store addr: %p)", totalMetricCount(metrics), v)

	select {
	case v.incoming <- metrics:
		utils.Debug("✅ Write enqueued %d metrics", totalMetricCount(metrics))
		return nil
	default:
		utils.Warn("❌ Incoming buffer full: dropping metrics")
		return fmt.Errorf("incoming buffer full")
	}
}

func (v *VictoriaStore) collectorLoop() {
	utils.Info("🌀 collectorLoop started")
	ticker := time.NewTicker(v.batchTimeout)
	defer ticker.Stop()

	utils.Info("⏱️ batchTimeout raw = %v\n", v.batchTimeout)
	utils.Debug("🕰️ collectorLoop started with timeout: %s", v.batchTimeout)

	var pending []model.MetricPayload

	for {
		select {
		case <-v.stopChan:
			utils.Debug("🛑 collectorLoop received stop signal")
			if len(pending) > 0 {
				utils.Debug("🛑 Flushing %d pending payloads on shutdown", len(pending))
				v.enqueue(pending)
			}
			return

		case batch := <-v.incoming:
			total := totalMetricCount(batch)
			utils.Debug("📥 Received payload with %d metrics", total)
			pending = append(pending, batch...)
			currentTotal := totalMetricCount(pending)
			utils.Debug("📊 Total metrics pending: %d", currentTotal)

			if currentTotal >= v.batchSize {
				utils.Info("📦 Batch size reached: %d metrics, flushing now", currentTotal)
				v.enqueue(pending)
				pending = nil
			}

		case <-ticker.C:
			currentTotal := totalMetricCount(pending)
			//utils.Debug("⏰ Timeout ticked. Pending payloads: %d, metrics: %d", len(pending), currentTotal)

			if currentTotal > 0 {
				utils.Info("⏳ Timeout flush triggered for %d metrics", currentTotal)
				v.enqueue(pending)
				pending = nil
			}
		}
	}
}

func (v *VictoriaStore) enqueue(batch []model.MetricPayload) {
	utils.Debug("📦 Enqueue called with %d payloads / %d metrics",
		len(batch), totalMetricCount(batch))
	select {
	case v.queue <- batch:
	default:
		utils.Warn("Worker queue full: dropping batch of %d metrics", len(batch))
	}
}

func (v *VictoriaStore) worker() {
	defer v.wg.Done()
	for {
		utils.Debug("👷 Worker waiting for batch...")

		select {

		case batch := <-v.queue:
			utils.Debug("👷 Worker received batch with %d payloads / %d metrics", len(batch), totalMetricCount(batch))
			v.flush(batch)
		case <-v.stopChan:
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

	utils.Debug("🚀 Flushing batch of %d metrics", len(batch))

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
			utils.Debug("Batch sent successfully to VictoriaMetrics")
			return
		}
		utils.Warn("Retrying batch write... attempt %d", attempt+1)
		time.Sleep(v.batchInterval)
	}
	utils.Error("Failed to write batch after %d retries", v.batchRetry)
}

func (v *VictoriaStore) Close() error {
	close(v.stopChan)
	v.wg.Wait()
	utils.Info("VictoriaStore shutdown complete")
	return nil
}

func buildPrometheusFormat(batch []model.MetricPayload) string {
	var sb strings.Builder
	for _, payload := range batch {
		ts := payload.Timestamp.UnixNano() / 1e6
		for _, m := range payload.Metrics {
			fullName := normalizeMetricName(m.Namespace, m.SubNamespace, m.Name) // ← Right here
			sb.WriteString(fmt.Sprintf("%s{%s} %f %d\n",
				fullName,
				formatLabels(payload.Meta),
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

func formatLabels(meta *model.Meta) string {
	if meta == nil {
		return "" // Return empty string if meta is nil
	}

	var out []string

	// Add all meta fields as labels, skipping empty strings
	if meta.Hostname != "" {
		out = append(out, fmt.Sprintf(`hostname="%s"`, meta.Hostname))
	}
	if meta.IPAddress != "" {
		out = append(out, fmt.Sprintf(`ip_address="%s"`, meta.IPAddress))
	}
	if meta.OS != "" {
		out = append(out, fmt.Sprintf(`os="%s"`, meta.OS))
	}
	if meta.OSVersion != "" {
		out = append(out, fmt.Sprintf(`os_version="%s"`, meta.OSVersion))
	}
	if meta.KernelVersion != "" {
		out = append(out, fmt.Sprintf(`kernel_version="%s"`, meta.KernelVersion))
	}
	if meta.Architecture != "" {
		out = append(out, fmt.Sprintf(`architecture="%s"`, meta.Architecture))
	}
	if meta.CloudProvider != "" {
		out = append(out, fmt.Sprintf(`cloud_provider="%s"`, meta.CloudProvider))
	}
	if meta.Region != "" {
		out = append(out, fmt.Sprintf(`region="%s"`, meta.Region))
	}
	if meta.AvailabilityZone != "" {
		out = append(out, fmt.Sprintf(`availability_zone="%s"`, meta.AvailabilityZone))
	}
	if meta.InstanceID != "" {
		out = append(out, fmt.Sprintf(`instance_id="%s"`, meta.InstanceID))
	}
	if meta.InstanceType != "" {
		out = append(out, fmt.Sprintf(`instance_type="%s"`, meta.InstanceType))
	}
	if meta.AccountID != "" {
		out = append(out, fmt.Sprintf(`account_id="%s"`, meta.AccountID))
	}
	if meta.ProjectID != "" {
		out = append(out, fmt.Sprintf(`project_id="%s"`, meta.ProjectID))
	}
	if meta.ResourceGroup != "" {
		out = append(out, fmt.Sprintf(`resource_group="%s"`, meta.ResourceGroup))
	}
	if meta.VPCID != "" {
		out = append(out, fmt.Sprintf(`vpc_id="%s"`, meta.VPCID))
	}
	if meta.SubnetID != "" {
		out = append(out, fmt.Sprintf(`subnet_id="%s"`, meta.SubnetID))
	}
	if meta.ImageID != "" {
		out = append(out, fmt.Sprintf(`image_id="%s"`, meta.ImageID))
	}
	if meta.ServiceID != "" {
		out = append(out, fmt.Sprintf(`service_id="%s"`, meta.ServiceID))
	}
	if meta.ContainerID != "" {
		out = append(out, fmt.Sprintf(`container_id="%s"`, meta.ContainerID))
	}
	if meta.ContainerName != "" {
		out = append(out, fmt.Sprintf(`container_name="%s"`, meta.ContainerName))
	}
	if meta.PodName != "" {
		out = append(out, fmt.Sprintf(`pod_name="%s"`, meta.PodName))
	}
	if meta.Namespace != "" {
		out = append(out, fmt.Sprintf(`namespace="%s"`, meta.Namespace))
	}
	if meta.ClusterName != "" {
		out = append(out, fmt.Sprintf(`cluster_name="%s"`, meta.ClusterName))
	}
	if meta.NodeName != "" {
		out = append(out, fmt.Sprintf(`node_name="%s"`, meta.NodeName))
	}
	if meta.Application != "" {
		out = append(out, fmt.Sprintf(`application="%s"`, meta.Application))
	}
	if meta.Environment != "" {
		out = append(out, fmt.Sprintf(`environment="%s"`, meta.Environment))
	}
	if meta.Service != "" {
		out = append(out, fmt.Sprintf(`service="%s"`, meta.Service))
	}
	if meta.Version != "" {
		out = append(out, fmt.Sprintf(`version="%s"`, meta.Version))
	}
	if meta.DeploymentID != "" {
		out = append(out, fmt.Sprintf(`deployment_id="%s"`, meta.DeploymentID))
	}
	if meta.PublicIP != "" {
		out = append(out, fmt.Sprintf(`public_ip="%s"`, meta.PublicIP))
	}
	if meta.PrivateIP != "" {
		out = append(out, fmt.Sprintf(`private_ip="%s"`, meta.PrivateIP))
	}
	if meta.MACAddress != "" {
		out = append(out, fmt.Sprintf(`mac_address="%s"`, meta.MACAddress))
	}
	if meta.NetworkInterface != "" {
		out = append(out, fmt.Sprintf(`network_interface="%s"`, meta.NetworkInterface))
	}

	// Handle tags map specifically
	for k, v := range meta.Tags {
		out = append(out, fmt.Sprintf(`%s="%s"`, k, v))
	}

	sort.Strings(out)
	return strings.Join(out, ",")
}

func totalMetricCount(payloads []model.MetricPayload) int {
	count := 0
	for _, p := range payloads {
		count += len(p.Metrics)
	}
	return count
}

// / QueryInstant fetches the latest data points for a given metric with optional label filters.
func (v *VictoriaStore) QueryInstant(metric string, filters map[string]string) ([]model.MetricRow, error) {
	query := buildPromQL(metric, filters)

	fullURL := fmt.Sprintf("%s/api/v1/query?query=%s", v.url, url.QueryEscape(query))
	utils.Debug("📡 QueryInstant URL: %s", fullURL)

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
	query := buildPromQL(metric, filters)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", "15s")

	fullURL := fmt.Sprintf("%s/api/v1/query_range?%s", v.url, params.Encode())
	utils.Debug("📡 QueryRange URL: %s", fullURL)

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

func buildPromQL(metric string, filters map[string]string) string {
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
