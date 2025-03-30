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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package collector

import (
	"fmt"
	"syscall"
)

type DiskCollector struct {
	mountPoints []string
}

func NewDiskCollector() *DiskCollector {
	return &DiskCollector{
		mountPoints: []string{"/"}, // Customize as needed
	}
}

func (d *DiskCollector) Name() string {
	return "disk"
}

func (d *DiskCollector) Collect() MetricResult {
	result := make(map[string]float64)

	for _, mount := range d.mountPoints {
		var stat syscall.Statfs_t
		err := syscall.Statfs(mount, &stat)
		if err != nil {
			return MetricResult{
				Name: d.Name(),
				Err:  fmt.Errorf("failed to statfs %s: %w", mount, err),
			}
		}

		total := float64(stat.Blocks) * float64(stat.Bsize)
		free := float64(stat.Bfree) * float64(stat.Bsize)
		used := total - free

		label := sanitizeLabel(mount)
		result[fmt.Sprintf("disk_used_bytes_%s", label)] = used
		result[fmt.Sprintf("disk_total_bytes_%s", label)] = total
		result[fmt.Sprintf("disk_used_percent_%s", label)] = used / total * 100
	}

	return MetricResult{
		Name: d.Name(),
		Data: result,
	}
}

// Replace "/" with "root" etc. for metric names
func sanitizeLabel(s string) string {
	if s == "/" {
		return "root"
	}
	return s[1:] // remove leading slash
}
