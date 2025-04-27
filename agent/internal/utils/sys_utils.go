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

// gosight/agent/internal/collector/system/sys_utils.go
// Package system provides utility functions for system collectors
// Package system provides collectors for system hardware (CPU/RAM/DISK/ETC)

package agentutils

import (
	"os"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func Metric(ns, sub, name string, value interface{}, typ, unit string, dims map[string]string, ts time.Time) model.Metric {
	return model.Metric{
		Namespace:    ns,
		SubNamespace: sub,
		Name:         name,
		Timestamp:    ts,
		Value:        ToFloat64(value),
		Type:         typ,
		Unit:         unit,
		Dimensions:   dims,
	}
}
func ToFloat64(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case uint64:
		return float64(n)
	case uint32:
		return float64(n)
	case int64:
		return float64(n)
	case int32:
		return float64(n)
	default:
		utils.Warn("ToFloat64: unknown type %T, defaulting to 0", v)
		return 0
	}
}

// GetHostname returns the system hostname, or "unknown" if it can't be determined.
func GetHostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "unknown"
	}
	return h
}
