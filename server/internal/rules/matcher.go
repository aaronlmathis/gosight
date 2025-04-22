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

package rules

import (
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// ruleMatches determines if a rule applies to the given metric + meta.
func ruleMatches(rule model.AlertRule, m model.Metric, meta *model.Meta) bool {
	sel := rule.Match

	// Basic namespace and metric match
	if sel.Namespace != "" && sel.Namespace != m.Namespace {
		return false
	}
	if sel.SubNamespace != "" && sel.SubNamespace != m.SubNamespace {
		return false
	}
	if sel.Metric != "" && sel.Metric != m.Name {
		return false
	}

	// Match on metric labels
	for k, v := range sel.Labels {
		if m.Dimensions[k] != v {
			return false
		}
	}

	// Match on specific endpoint ID(s)
	if len(sel.EndpointIDs) > 0 {
		match := false
		for _, id := range sel.EndpointIDs {
			if id == meta.EndpointID {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Match on meta tags (like hostname, environment, etc.)
	for k, v := range sel.TagSelectors {
		if meta.Tags[k] != v {
			return false
		}
	}
	utils.Debug("🔍 Rule %s matched %s on %s", rule.ID, m.Name, meta.EndpointID)
	return true
}
