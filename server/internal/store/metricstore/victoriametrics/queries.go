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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

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
func (v *VictoriaStore) QueryRange(metric string, start, end time.Time, step string, filters map[string]string) ([]model.Point, error) {
	query := BuildPromQL(metric, filters)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", step)

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

func (v *VictoriaStore) GetAllKnownMetricNames() []string {
	return []string{}
}

func (v *VictoriaStore) QueryMultiInstant(metricNames []string, filters map[string]string) ([]model.MetricRow, error) {
	//utils.Debug("Executing VictoriaStore.QueryMultiInstant")
	if len(metricNames) == 0 {
		// Default to all known metric names if available
		metricNames = v.GetAllKnownMetricNames()
		//utils.Debug("No metric names provided, using all known metric names: %v", metricNames)
	}

	var query string
	if len(metricNames) == 1 {
		// Single metric → clean style: metric{label="value"}
		base := metricNames[0]
		if len(filters) > 0 {
			var parts []string
			for k, v := range filters {
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(parts)
			query = fmt.Sprintf(`%s{%s}`, base, strings.Join(parts, ","))
		} else {
			query = base
		}
	} else {
		// Multiple metrics → regex match on __name__
		nameSelector := fmt.Sprintf(`__name__=~"%s"`, strings.Join(metricNames, "|"))
		var parts []string
		parts = append(parts, nameSelector)
		for k, v := range filters {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
		}
		sort.Strings(parts)
		query = fmt.Sprintf("{%s}", strings.Join(parts, ","))
	}

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

func (v *VictoriaStore) QueryMultiRange(metrics []string, start, end time.Time, step string, filters map[string]string) ([]model.MetricRow, error) {
	if len(metrics) == 0 {
		return nil, nil
	}
	var query string
	if len(metrics) == 1 {
		// Single metric → clean form
		base := metrics[0]
		if len(filters) > 0 {
			var parts []string
			for k, v := range filters {
				if k == "step" {
					continue
				}
				parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
			}
			sort.Strings(parts)
			query = fmt.Sprintf(`%s{%s}`, base, strings.Join(parts, ","))
		} else {
			query = base
		}
	} else {
		// Multi-metric → __name__=~ form
		nameSelector := fmt.Sprintf(`__name__=~"%s"`, strings.Join(metrics, "|"))
		var parts []string
		parts = append(parts, nameSelector)
		for k, v := range filters {
			parts = append(parts, fmt.Sprintf(`%s="%s"`, k, v))
		}
		sort.Strings(parts)
		query = fmt.Sprintf("{%s}", strings.Join(parts, ","))
	}

	// Build URL
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))

	secs, err := parseDurationToSeconds(step)
	if err != nil {
		return nil, fmt.Errorf("invalid step: %v", err)
	}
	params.Set("step", strconv.Itoa(secs))

	fullURL := fmt.Sprintf("%s/api/v1/query_range?%s", v.url, params.Encode())
	//utils.Debug("QueryMultiRange URL: %s", fullURL)
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

// FetchDimensionsForMetric queries VictoriaMetrics for a given metric and extracts dimension keys.
func (v *VictoriaStore) FetchDimensionsForMetric(metric string) ([]string, error) {
	queryURL := fmt.Sprintf("%s/api/v1/query?query=%s", v.url, url.QueryEscape(metric))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("VictoriaMetrics returned %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Status string `json:"status"`
		Data   struct {
			ResultType string `json:"resultType"`
			Result     []struct {
				Metric map[string]string `json:"metric"`
				Value  []interface{}     `json:"value"`
			} `json:"result"`
		} `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	if parsed.Status != "success" {
		return nil, fmt.Errorf("VictoriaMetrics query status not success: %s", parsed.Status)
	}

	// Collect unique dimension keys
	dimSet := make(map[string]struct{})

	for _, series := range parsed.Data.Result {
		for key := range series.Metric {
			if key == "__name__" {
				continue // skip Prometheus internal field
			}
			dimSet[key] = struct{}{}
		}
	}

	if len(dimSet) == 0 {
		return nil, fmt.Errorf("no dimensions found for metric %s", metric)
	}

	var dims []string
	for k := range dimSet {
		dims = append(dims, k)
	}

	return dims, nil
}

func (v *VictoriaStore) ListLabelValues(label string, contains string) ([]string, error) {
	queryURL := fmt.Sprintf("%s/api/v1/label/%s/values", v.url, url.PathEscape(label))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(queryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to query VictoriaMetrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("victoria metrics returned %s", resp.Status)
	}

	var parsed struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	if contains != "" {
		contains = strings.ToLower(contains)
		filtered := make([]string, 0, len(parsed.Data))
		for _, val := range parsed.Data {
			if strings.Contains(strings.ToLower(val), contains) {
				filtered = append(filtered, val)
			}
		}
		return filtered, nil
	}

	return parsed.Data, nil
}
