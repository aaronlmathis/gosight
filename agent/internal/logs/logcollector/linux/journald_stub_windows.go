//go:build windows
// +build windows

package collector

import (
	"context"

	"github.com/aaronlmathis/gosight/shared/model"
)

type JournaldCollector struct{}

func (jc *JournaldCollector) Name() string { return "journald" }
func (jc *JournaldCollector) Collect(ctx context.Context) [][]model.LogEntry {
	return nil
}
