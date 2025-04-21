package container

import (
	"context"
	"strings"
	"time"

	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
)

type DockerCollector struct {
	socketPath string
}

func NewDockerCollector() *DockerCollector {
	return &DockerCollector{socketPath: "/var/run/docker.sock"}
}
func NewDockerCollectorWithSocket(path string) *DockerCollector {
	return &DockerCollector{socketPath: path}
}
func (c *DockerCollector) Name() string {
	return "docker"
}

type DockerContainer struct {
	ID      string            `json:"Id"`
	Names   []string          `json:"Names"`
	Image   string            `json:"Image"`
	State   string            `json:"State"`
	Labels  map[string]string `json:"Labels"`
	Ports   []PortMapping     `json:"Ports"`
	Created int64             `json:"Created"` // unix timestamp
}

func (c *DockerCollector) Collect(ctx context.Context) ([]model.Metric, error) {
	containers, err := fetchContainersFromSocket[DockerContainer](c.socketPath, "/v1.41/containers/json?all=true")
	if err != nil {
		return nil, err
	}

	var metrics []model.Metric
	now := time.Now()

	for _, ctr := range containers {
		stats, err := fetchContainerStatsFromSocket[PodmanStats](c.socketPath, "/v1.41/containers/"+ctr.ID+"/stats?stream=false")
		if err != nil {
			continue
		}

		uptime := 0.0
		if strings.ToLower(ctr.State) == "running" && ctr.Created > 0 {
			startTime := time.Unix(ctr.Created, 0)
			uptime = now.Sub(startTime).Seconds()
			if uptime > 1e6 || uptime < 0 {
				uptime = 0
			}
		}

		running := 0.0
		if strings.ToLower(ctr.State) == "running" {
			running = 1.0
		}

		dims := map[string]string{
			"container_id": ctr.ID[:12],
			"name":         strings.TrimPrefix(ctr.Names[0], "/"),
			"image":        ctr.Image,
			"status":       ctr.State,
			"runtime":      "docker",
		}
		for k, v := range ctr.Labels {
			dims["label."+k] = v
		}
		if ports := formatPorts(ctr.Ports); ports != "" {
			dims["ports"] = ports
		}

		cpu := calculateCPUPercent(ctr.ID, &stats)
		rx, tx := calculateNetRate(ctr.ID, now, sumNetRxRaw(&stats), sumNetTxRaw(&stats))

		metrics = append(metrics,
			agentutils.Metric("Container", "Docker", "uptime_seconds", uptime, "gauge", "seconds", dims, now),
			agentutils.Metric("Container", "Docker", "running", running, "gauge", "bool", dims, now),
			agentutils.Metric("Container", "Docker", "cpu_percent", cpu, "gauge", "percent", dims, now),
			agentutils.Metric("Container", "Docker", "mem_usage_bytes", float64(stats.MemoryStats.Usage), "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Docker", "mem_limit_bytes", float64(stats.MemoryStats.Limit), "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Docker", "net_rx_bytes", rx, "gauge", "bytes", dims, now),
			agentutils.Metric("Container", "Docker", "net_tx_bytes", tx, "gauge", "bytes", dims, now),
		)
	}

	return metrics, nil
}
