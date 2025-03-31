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

// gosight/agent/internal/runner/runner.go

package runner

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/agent/internal/collector"
	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/agent/internal/utils"
)

// RunOnce collects and prints metrics based on the current config
func RunOnce(cfg *config.AgentConfig) {
	ctx := context.Background()
	reg := collector.NewRegistry(cfg)

	metrics, err := reg.Collect(ctx)
	if err != nil {
		utils.Error("Failed to collect metrics: %v", err)
		return
	}

	for _, m := range metrics {
		fmt.Printf("[%s] %s = %.2f %s\n", m.Namespace, m.Name, m.Value, m.Unit)
	}
}
