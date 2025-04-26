package templates

import (
	"context"
	"time"

	gosightauth "github.com/aaronlmathis/gosight/server/internal/auth"
	"github.com/aaronlmathis/gosight/server/internal/store/metastore"
	"github.com/aaronlmathis/gosight/server/internal/store/metricstore"
	"github.com/aaronlmathis/gosight/server/internal/usermodel"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

var HostMetrics = []model.MetricSelector{
	{Name: "system.cpu.count_physical", Namespace: "System", SubNamespace: "System.CPU", Instant: true},
	{Name: "system.cpu.usage_percent", Namespace: "System", SubNamespace: "System.CPU", Instant: true},
	{Name: "system.cpu.time_user", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_system", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_idle", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_nice", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_guest", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_guest_nice", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_iowait", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_irq", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_softirq", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.time_steal", Namespace: "System", SubNamespace: "System.CPU", Instant: false},
	{Name: "system.cpu.clock_mhz", Namespace: "System", SubNamespace: "System.CPU", Instant: true},
	{Name: "system.cpu.count_logical", Namespace: "System", SubNamespace: "System.CPU", Instant: true},

	{Name: "system.memory.available", Namespace: "System", SubNamespace: "System.Memory", Instant: true},
	{Name: "system.memory.used", Namespace: "System", SubNamespace: "System.Memory", Instant: true},
	{Name: "system.memory.used_percent", Namespace: "System", SubNamespace: "System.Memory", Instant: false},
	{Name: "system.memory.total", Namespace: "System", SubNamespace: "System.Memory", Instant: true},

	{Name: "system.host.procs", Namespace: "System", SubNamespace: "System.Host", Instant: true},
	{Name: "system.host.users_loggedin", Namespace: "System", SubNamespace: "System.Host", Instant: true},
	{Name: "system.host.info", Namespace: "System", SubNamespace: "System.Host", Instant: true},
	{Name: "system.host.uptime", Namespace: "System", SubNamespace: "System.Host", Instant: true},

	{Name: "system.diskio.read_bytes", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.read_time", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.write_time", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.merged_read_count", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.merged_write_count", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.weighted_io", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.read_count", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.write_count", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.write_bytes", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},
	{Name: "system.diskio.io_time", Namespace: "System", SubNamespace: "System.DiskIO", Instant: false},

	{Name: "system.disk.used_percent", Namespace: "System", SubNamespace: "System.Disk", Instant: false},
	{Name: "system.disk.inodes_total", Namespace: "System", SubNamespace: "System.Disk", Instant: true},
	{Name: "system.disk.total", Namespace: "System", SubNamespace: "System.Disk", Instant: true},
	{Name: "system.disk.used", Namespace: "System", SubNamespace: "System.Disk", Instant: true},
	{Name: "system.disk.inodes_used", Namespace: "System", SubNamespace: "System.Disk", Instant: true},
	{Name: "system.disk.inodes_free", Namespace: "System", SubNamespace: "System.Disk", Instant: true},
	{Name: "system.disk.inodes_used_percent", Namespace: "System", SubNamespace: "System.Disk", Instant: false},
	{Name: "system.disk.free", Namespace: "System", SubNamespace: "System.Disk", Instant: true},

	{Name: "system.network.packets_recv", Namespace: "System", SubNamespace: "System.Network", Instant: false},
	{Name: "system.network.err_in", Namespace: "System", SubNamespace: "System.Network", Instant: false},
	{Name: "system.network.err_out", Namespace: "System", SubNamespace: "System.Network", Instant: false},
	{Name: "system.network.bytes_sent", Namespace: "System", SubNamespace: "System.Network", Instant: false},
	{Name: "system.network.bytes_recv", Namespace: "System", SubNamespace: "System.Network", Instant: false},
	{Name: "system.network.packets_sent", Namespace: "System", SubNamespace: "System.Network", Instant: false},
}

var ContainerMetrics = []model.MetricSelector{
	{Name: "container.podman.uptime_seconds", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.running", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.cpu_percent", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.mem_usage_bytes", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.mem_limit_bytes", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.net_rx_bytes", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
	{Name: "container.podman.net_tx_bytes", Namespace: "Container", SubNamespace: "Container.Podman", Instant: true},
}

// FlattenInstant flattens the MetricRows returned by store and makes it usable for template
func FlattenInstant(rows []model.MetricRow) map[string]float64 {
	result := make(map[string]float64)
	for _, row := range rows {
		if name, ok := row.Tags["__name__"]; ok {
			result[name] = row.Value
		}
	}
	return result
}

func FlattenRange(rows []model.MetricRow) map[string][]model.MetricPoint {
	result := make(map[string][]model.MetricPoint)

	for _, row := range rows {
		name, ok := row.Tags["__name__"]
		if !ok {
			continue
		}

		// Optional: add other label filters here if needed

		result[name] = append(result[name], model.MetricPoint{
			Timestamp: row.Timestamp,
			Value:     row.Value,
		})
	}

	return result
}

func GetMetricNames(metrics []model.MetricSelector, instantOnly bool) []string {
	var out []string
	for _, m := range metrics {
		if m.Instant == instantOnly {
			out = append(out, m.Name)
		}
	}
	return out
}

func BuildHostDashboardData(ctx context.Context, ms metricstore.MetricStore, metaTracker *metastore.MetaTracker, user *usermodel.User, endpointID string) (*TemplateData, error) {

	utils.Debug("Building host dashboard data for endpoint: %s", endpointID)

	// Basic labels
	labels := map[string]string{
		"endpoint_id": endpointID,
	}

	meta, _ := metaTracker.Get(endpointID)

	utils.Debug("MetricStore concrete type: %T\n", ms)
	// Pull all instant metrics for stat boxes
	instantNames := GetMetricNames(HostMetrics, true)

	instantRows, err := ms.QueryMultiInstant(instantNames, labels)
	utils.Debug("instantRows: %v", instantRows)
	if err != nil {
		return nil, err
	}

	metrics := FlattenInstant(instantRows)

	// Pull range metrics for charts (10 min window)
	start := time.Now().Add(-10 * time.Minute)
	end := time.Now()
	rangeNames := GetMetricNames(HostMetrics, false)
	step := "15s"
	rangeRows, err := ms.QueryMultiRange(rangeNames, start, end, step, labels)
	if err != nil {
		return nil, err
	}
	timeseries := FlattenRange(rangeRows)

	// Extract tags from the first metric row
	tags := map[string]string{}
	if len(instantRows) > 0 {
		for k, v := range instantRows[0].Tags {
			tags[k] = v
		}
	}

	// 🔍 Determine latest timestamp to infer status
	var latestTs int64
	for _, row := range instantRows {
		if row.Timestamp > latestTs {
			latestTs = row.Timestamp
		}
	}

	if latestTs > 0 {
		age := time.Since(time.UnixMilli(latestTs))
		if age > 60*time.Second {
			labels["status"] = "offline"
		} else {
			labels["status"] = "online"
		}
		labels["last_report"] = time.UnixMilli(latestTs).Format(time.RFC3339)
	} else {
		labels["status"] = "unknown"
		labels["last_report"] = ""
	}

	labels["platform"] = tags["platform"]
	labels["platform_version"] = tags["platform_version"]
	labels["os"] = tags["os"]
	utils.Debug("Labels %v", labels)

	breadcrumbs := []Breadcrumb{
		{Label: "Endpoints", URL: "/endpoints"},
		{Label: "Host Overview", URL: ""}, // No URL for current page
	}

	return &TemplateData{
		Title:       "Host Dashboard",
		User:        user,
		Permissions: gosightauth.FlattenPermissions(user.Roles),
		Labels:      labels,
		Tags:        tags,
		Metrics:     metrics,
		Meta:        meta,
		Timeseries:  timeseries,
		Breadcrumbs: breadcrumbs,
	}, nil
}
