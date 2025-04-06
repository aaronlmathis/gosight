package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Point struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

func queryAPI(metric string, labels map[string]string) (Point, error) {
	base := "http://localhost:8080/api/query"
	params := url.Values{}
	params.Set("metric", metric)
	params.Set("latest", "true")
	for k, v := range labels {
		params.Set(k, v)
	}

	reqURL := fmt.Sprintf("%s?%s", base, params.Encode())
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return Point{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Point{}, fmt.Errorf("non-200 status: %s", resp.Status)
	}

	var point Point
	err = json.NewDecoder(resp.Body).Decode(&point)
	if err != nil {
		return Point{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return point, nil
}

func main() {
	f, err := os.Create("test.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	tests := []struct {
		Metric string
		Labels map[string]string
	}{
		{
			Metric: "system.disk.used_percent",
			Labels: map[string]string{"job": "gosight-agent"},
		},
		{
			Metric: "system.diskio.io_time",
			Labels: map[string]string{"hostname": "DeepThought", "namespace": "system"},
		},
		{
			Metric: "system.memory.available",
			Labels: map[string]string{"ip_address": "192.168.0.40"},
		},
		{
			Metric: "system.memory.used_percent",
			Labels: map[string]string{"hostname": "RH-Podman-Cluster", "subnamespace": "memory"},
		},
		{
			Metric: "system.host.uptime",
			Labels: map[string]string{"endpoint_id": "host-ESLAPTOP-AARON-169-254-75-59"},
		},
	}

	for _, test := range tests {
		start := time.Now()
		point, err := queryAPI(test.Metric, test.Labels)
		line := fmt.Sprintf("[%s] Metric: %-32s Labels: %+v\n", start.Format(time.RFC3339), test.Metric, test.Labels)
		if err != nil {
			line += fmt.Sprintf("  ❌ Error: %v\n", err)
		} else {
			line += fmt.Sprintf("  ✅ Value: %v @ %s\n", point.Value, point.Timestamp)
		}
		f.WriteString(line)
	}
}
