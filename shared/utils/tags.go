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

package utils

import "github.com/aaronlmathis/gosight/shared/model"

// MatchAllTags returns true if all required tags exist in actualTags and match.
func MatchAllTags(required map[string]string, actual map[string]string) bool {
	for k, v := range required {
		if actualVal, ok := actual[k]; !ok || actualVal != v {
			return false
		}
	}
	return true
}

// SafeCopyTags ensures that meta.Tags is never nil.
// It returns a non-nil map[string]string for tags, even if it was nil originally.
func SafeCopyTags(meta *model.Meta) map[string]string {
	if meta == nil {
		return make(map[string]string)
	}
	if meta.Tags == nil {
		meta.Tags = make(map[string]string)
	}
	return meta.Tags
}
