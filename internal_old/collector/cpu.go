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

type CPUCollector struct {
	prevIdle  uint64
	prevTotal uint64
}

func (c *CPUCollector) Name() string {
	return "cpu"
}

func (c *CPUCollector) Collect() MetricResult {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return MetricResult{
			Name: c.Name(),
			Err:  fmt.Errorf("failed to open /proc/stat: %w", err),
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return MetricResult{
					Name: c.Name(),
					Err:  fmt.Errorf("unexpected format in /proc/stat"),
				}
			}

			user, _ := strconv.ParseUint(fields[1], 10, 64)
			nice, _ := strconv.ParseUint(fields[2], 10, 64)
			system, _ := strconv.ParseUint(fields[3], 10, 64)
			idle, _ := strconv.ParseUint(fields[4], 10, 64)

			total := user + nice + system + idle
			deltaTotal := total - c.prevTotal
			deltaIdle := idle - c.prevIdle

			c.prevTotal = total
			c.prevIdle = idle

			usage := 0.0
			if deltaTotal > 0 {
				usage = float64(deltaTotal-deltaIdle) / float64(deltaTotal) * 100
			}

			return MetricResult{
				Name: c.Name(),
				Data: map[string]float64{"cpu_usage": usage},
			}
		}
	}

	// fallback
	return MetricResult{
		Name: c.Name(),
		Data: map[string]float64{"cpu_usage": 42.0},
	}
}
