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

// server/internal/http/templates/helpers.go

package gosighttemplate

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

// HumanizeBytes converts a byte size in bytes to a human-readable string.
func HumanizeBytes(b float64) string {
	const KB = 1024
	const MB = KB * 1024
	const GB = MB * 1024
	switch {
	case b > GB:
		return fmt.Sprintf("%.1f GB", b/GB)
	case b > MB:
		return fmt.Sprintf("%.1f MB", b/MB)
	case b > KB:
		return fmt.Sprintf("%.1f KB", b/KB)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}

// Template functions

// Marshal converts a Go value to a JSON string and returns it as a template.JS type.
func Marshal(v interface{}) template.JS {
	data, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(data)
}

// FormatUptime formats a duration in seconds into a human-readable string

func FormatUptime(seconds float64) string {
	s := int64(seconds)
	days := s / 86400
	hours := (s % 86400) / 3600
	minutes := (s % 3600) / 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

// Seq generates a slice of integers from `from` to `to`, inclusive.
func Seq(from, to int) []int {
	s := make([]int, to-from+1)
	for i := range s {
		s[i] = from + i
	}
	return s
}

// Div performs safe division, returning 0 if the divisor is zero.
func Div(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}

// Since calculates the time duration since a given timestamp in seconds.
func Since(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "unknown"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

// SafeHTML returns a template.HTML type to safely render HTML content.
func SafeHTML(s string) template.HTML { return template.HTML(s) }

// HasPermission checks if a user has a specific permission.
func HasPermission(userPermissions []string, permission string) bool {
	for _, p := range userPermissions {
		if p == permission {
			return true
		}
	}
	return false
}

func HasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
