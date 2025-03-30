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

import "time"

func StartCollectors(collectors []MetricCollector, interval time.Duration, out chan<- MetricResult) {
	for _, c := range collectors {
		go func(c MetricCollector) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				result := c.Collect()
				out <- result
				<-ticker.C
			}
		}(c)
	}
}
