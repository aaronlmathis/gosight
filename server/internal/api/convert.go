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

package api

import (
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
)

func ConvertToModelPayload(pbPayload *proto.MetricPayload) model.MetricPayload {
	metrics := make([]model.Metric, 0, len(pbPayload.Metrics))
	for _, m := range pbPayload.Metrics {
		metric := model.Metric{
			Namespace:         m.Namespace,
			Name:              m.Name,
			Timestamp:         m.Timestamp.AsTime(),
			Value:             m.Value,
			Unit:              m.Unit,
			Dimensions:        m.Dimensions,
			StorageResolution: int(m.StorageResolution),
			Type:              m.Type,
		}
		if m.StatisticValues != nil {
			metric.StatisticValues = &model.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		metrics = append(metrics, metric)
	}
	return model.MetricPayload{
		Host:      pbPayload.Host,
		Timestamp: pbPayload.Timestamp.AsTime(),
		Metrics:   metrics,
		Meta:      pbPayload.Meta,
	}
}