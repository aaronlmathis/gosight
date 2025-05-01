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

package gosighttemplate

import (
	"github.com/aaronlmathis/gosight/shared/model"
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
