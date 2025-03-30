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
	"log"
	"os"
	"strconv"
	"strings"
)

type NetCollector struct{}

func (n *NetCollector) Name() string {
	return "network"
}

func (n *NetCollector) Collect() MetricResult {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return MetricResult{
			Name: n.Name(),
			Err:  fmt.Errorf("failed to open /proc/net/dev: %w", err),
		}
	}
	defer f.Close()

	log.Printf("[network] collected interfaces")
	scanner := bufio.NewScanner(f)
	results := make(map[string]float64)

	for i := 0; scanner.Scan(); i++ {
		if i < 2 {
			continue // skip headers
		}
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 10 {
			continue
		}

		iface := strings.TrimSuffix(parts[0], ":")
		rxBytes, _ := strconv.ParseFloat(parts[1], 64)
		txBytes, _ := strconv.ParseFloat(parts[9], 64)

		results[fmt.Sprintf("net_rx_bytes_%s", iface)] = rxBytes
		results[fmt.Sprintf("net_tx_bytes_%s", iface)] = txBytes
	}

	return MetricResult{
		Name: n.Name(),
		Data: results,
	}
}
