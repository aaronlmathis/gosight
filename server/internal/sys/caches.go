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

// gosight/agent/internal/sys/caches.go
// Package sys provides system-level caches and context for the agent.
// This includes caches for metrics, logs, and processes.
package sys

import "github.com/aaronlmathis/gosight/server/internal/cache"

type CacheModule struct {
	Process cache.ProcessCache
	Metric  cache.MetricCache
	// Log     cache.LogCache
}
