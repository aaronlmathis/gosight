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

// gosight/shared/utils/maps.go
// maps.go contains utility functions for working with maps

package utils

import "strings"

func MergeMaps(base, override map[string]string) map[string]string {
	result := make(map[string]string)

	// Copy base map
	for k, v := range base {
		result[k] = v
	}

	// Override with values from override
	for k, v := range override {
		result[k] = v
	}

	return result
}

func ParseTagString(input string) map[string]string {
	tags := make(map[string]string)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			k := strings.TrimSpace(kv[0])
			v := strings.TrimSpace(kv[1])
			if k != "" {
				tags[k] = v
			}
		}
	}
	return tags
}
