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

// gosight/agent/api/convert.go
// convert.go - converts internal metric payloads to protobuf format for gRPC.

package api

import (
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToProtoPayload(payload model.MetricPayload) *proto.MetricPayload {
	metrics := make([]*proto.Metric, 0, len(payload.Metrics))
	for _, m := range payload.Metrics {
		pm := &proto.Metric{
			Namespace:         m.Namespace,
			Name:              m.Name,
			Timestamp:         timestamppb.New(m.Timestamp),
			Value:             m.Value,
			Unit:              m.Unit,
			Dimensions:        m.Dimensions,
			StorageResolution: int32(m.StorageResolution),
			Type:              m.Type,
		}
		if m.StatisticValues != nil {
			pm.StatisticValues = &proto.StatisticValues{
				Minimum:     m.StatisticValues.Minimum,
				Maximum:     m.StatisticValues.Maximum,
				SampleCount: int32(m.StatisticValues.SampleCount),
				Sum:         m.StatisticValues.Sum,
			}
		}
		metrics = append(metrics, pm)
	}
	return &proto.MetricPayload{
		Host:      payload.Host,
		Timestamp: timestamppb.New(payload.Timestamp),
		Metrics:   metrics,
		Meta:      payload.Meta,
	}
}