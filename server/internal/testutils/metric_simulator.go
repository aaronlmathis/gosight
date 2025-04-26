package testutils

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/server/internal/sys"
	"github.com/aaronlmathis/gosight/shared/model"
)

// TriggerTestAlert simulates a high memory usage metric to trigger a test rule.
func TriggerTestAlert(sys *sys.SystemContext) {
	metrics := []model.Metric{
		{
			Namespace:    "System",
			SubNamespace: "Memory",
			Name:         "used_percent",
			Value:        92.4,
			Type:         "gauge",
			Unit:         "percent",
			Timestamp:    time.Now(),
		},
	}

	meta := &model.Meta{
		EndpointID: "host-test-123",
		Tags: map[string]string{
			"env":  "test",
			"team": "ops",
		},
	}

	sys.Tele.Evaluator.EvaluateMetric(context.Background(), metrics, meta)
}
