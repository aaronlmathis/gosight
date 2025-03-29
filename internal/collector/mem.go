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
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MemoryCollector struct{}

func (m *MemoryCollector) Name() string {
	return "memory"
}

func (m *MemoryCollector) Collect() (map[string]float64, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer file.Close()

	stats := make(map[string]uint64)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		stats[key] = value // kb
	}

	total := stats["MemTotal"]
	available := stats["MemAvailable"]
	if available == 0 {
		available = stats["MemFree"] + stats["Buffers"] + stats["Cached"]
	}

	used := total - available

	return map[string]float64{
		"mem_total_kb":     float64(total),
		"mem_used_kb":      float64(used),
		"mem_free_kb":      float64(available),
		"mem_used_percent": float64(used) / float64(total) * 100,
	}, nil
}
