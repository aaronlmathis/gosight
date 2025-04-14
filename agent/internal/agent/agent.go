package gosightagent

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	metricrunner "github.com/aaronlmathis/gosight/agent/internal/metricrunner"
)

type Agent struct {
	Config       *config.Config
	MetricRunner *metricrunner.MetricRunner
}

func NewAgent(ctx context.Context, cfg *config.Config) (*Agent, error) {
	metricRunner, err := metricrunner.NewRunner(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %v", err)
	}
	return &Agent{
		Config:       cfg,
		MetricRunner: metricRunner,
	}, nil
}

func (a *Agent) Start(ctx context.Context) {
	// Start runner.
	a.MetricRunner.Run(ctx)
}
