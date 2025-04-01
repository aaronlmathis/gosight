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

package model

import "time"

type StatisticValues struct {
	Minimum     float64 `json:"min"`
	Maximum     float64 `json:"max"`
	SampleCount int     `json:"count"`
	Sum         float64 `json:"sum"`
}

type Metric struct {
	Namespace         string            `json:"namespace,omitempty"`
	Name              string            `json:"name"`
	Timestamp         time.Time         `json:"timestamp,omitempty"`
	Value             float64           `json:"value,omitempty"`
	StatisticValues   *StatisticValues  `json:"stats,omitempty"`
	Unit              string            `json:"unit,omitempty"`
	Dimensions        map[string]string `json:"dimensions,omitempty"`
	StorageResolution int               `json:"resolution,omitempty"`
	Type              string            `json:"type,omitempty"`
}

type MetricPayload struct {
	Host      string            `json:"host"`
	Timestamp time.Time         `json:"timestamp"`
	Metrics   []Metric          `json:"metrics"`
	Meta      map[string]string `json:"meta,omitempty"`
}