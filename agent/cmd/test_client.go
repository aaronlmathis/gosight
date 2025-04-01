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

package main

/*
package main

import (
	"context"
	"log"

	"github.com/aaronlmathis/gosight/shared/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	conn, err := grpc.Dial("localhost:8080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := proto.NewMetricsServiceClient(conn)

	payload := &proto.MetricPayload{
		Host:      "test-agent",
		Timestamp: timestamppb.Now(),
		Metrics: []*proto.Metric{
			{
				Name:      "cpu_usage",
				Value:     42.5,
				Unit:      "percent",
				Timestamp: timestamppb.Now(),
			},
			{
				Name:      "mem_used",
				Value:     16384,
				Unit:      "MB",
				Timestamp: timestamppb.Now(),
			},
		},
	}

	resp, err := client.SubmitMetrics(context.Background(), payload)
	if err != nil {
		log.Fatalf("SubmitMetrics failed: %v", err)
	}

	log.Printf("Response from server: %s (code %d)", resp.Status, resp.StatusCode)
}
*/