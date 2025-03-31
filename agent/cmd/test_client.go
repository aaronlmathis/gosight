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
