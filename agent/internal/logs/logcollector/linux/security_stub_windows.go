//go:build windows
// +build windows

/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis

This file is part of GoSight.

This is a Windows stub for the SecurityLogCollector to allow cross-platform compilation.
*/

package linuxcollector

import (
	"context"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
)

// SecurityLogCollector is a no-op stub for Windows.
type SecurityLogCollector struct{}

// Name returns the collector name.
func (c *SecurityLogCollector) Name() string {
	return "security"
}

// NewSecurityLogCollector returns a disabled stub.
func NewSecurityLogCollector(cfg *config.Config) *SecurityLogCollector {
	return &SecurityLogCollector{}
}

// Collect returns no logs on Windows.
func (c *SecurityLogCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	return nil, nil
}

// Close is a no-op.
func (c *SecurityLogCollector) Close() {}
