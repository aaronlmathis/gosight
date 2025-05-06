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

// gosight/shared/utils/helpers.go
//

package utils

import (
	"strconv"
	"unicode/utf8"
)

func ParseIntOrDefault(input string, def int) int {
	if val, err := strconv.Atoi(input); err == nil {
		return val
	}
	return def
}

// Truncate safely trims a string to the given length in runes.
// If truncation occurs, it appends "…" (ellipsis).
func Truncate(s string, max int) string {
	if max <= 0 || s == "" {
		return ""
	}

	if utf8.RuneCountInString(s) <= max {
		return s
	}

	runes := []rune(s)
	if max <= 1 {
		return "…"
	}

	return string(runes[:max-1]) + "…"
}
