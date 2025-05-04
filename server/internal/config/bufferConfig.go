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

// server/internal/config/bufferConfig.go

package config

import "time"

// Buff
type BufferEngineConfig struct {
	Enabled              bool               `yaml:"enabled"`
	FlushInterval        time.Duration      `yaml:"flush_interval"`
	ShutdownFlushTimeout time.Duration      `yaml:"shutdown_flush_timeout"`
	MaxWorkers           int                `yaml:"max_workers"`
	Metrics              MetricBufferConfig `yaml:"metrics"`
	Logs                 LogBufferConfig    `yaml:"logs"`
	Events               EventBufferConfig  `yaml:"events"`
	Alerts               AlertBufferConfig  `yaml:"alerts"`
}

type MetricBufferConfig struct {
	Enabled           bool             `yaml:"enabled"`
	BufferSize        int              `yaml:"buffer_size"`
	FlushInterval     time.Duration    `yaml:"flush_interval"`
	DropOnOverflow    bool             `yaml:"drop_on_overflow"`
	RetryFailedFlush  bool             `yaml:"retry_failed_flush"`
	FlushOnDisconnect bool             `yaml:"flush_on_disconnect"`
	FallbackDisk      DiskBufferConfig `yaml:"fallback_disk"`
}

type LogBufferConfig struct {
	Enabled          bool             `yaml:"enabled"`
	BufferSize       int              `yaml:"buffer_size"`
	FlushInterval    time.Duration    `yaml:"flush_interval"`
	DropOnOverflow   bool             `yaml:"drop_on_overflow"`
	RetryFailedFlush bool             `yaml:"retry_failed_flush"`
	FallbackDisk     DiskBufferConfig `yaml:"fallback_disk"`
}

type EventBufferConfig struct {
	Enabled          bool          `yaml:"enabled"`
	BufferSize       int           `yaml:"buffer_size"`
	FlushInterval    time.Duration `yaml:"flush_interval"`
	RetryFailedFlush bool          `yaml:"retry_failed_flush"`
}

type AlertBufferConfig struct {
	Enabled          bool          `yaml:"enabled"`
	BufferSize       int           `yaml:"buffer_size"`
	FlushInterval    time.Duration `yaml:"flush_interval"`
	DropOnOverflow   bool          `yaml:"drop_on_overflow"`
	RetryFailedFlush bool          `yaml:"retry_failed_flush"`
}

type DiskBufferConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Path          string `yaml:"path"`
	MaxDiskSizeMB int    `yaml:"max_disk_size_mb"`
}
