//go:build linux
// +build linux

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

package linuxcollector

import (
	"context"
	"fmt"
	"io" // Needed for tail.SeekInfo
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/nxadm/tail" // Import the tail library
)

type SecurityLogCollector struct {
	Config     *config.Config
	logPath    string
	maxMsgSize int
	batchSize  int

	tailer *tail.Tail          // The tail instance
	lines  chan model.LogEntry // Internal channel for collected lines
	stop   chan struct{}       // Channel to signal background goroutine stop
	wg     sync.WaitGroup      // WaitGroup to ensure clean shutdown
	mu     sync.Mutex          // Mutex to protect access during shutdown
}

func NewSecurityLogCollector(cfg *config.Config) *SecurityLogCollector {
	paths := []string{"/var/log/secure", "/var/log/auth.log"}
	var selected string
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			selected = p
			break
		}
	}

	// If no path is selected, return a non-functional collector
	if selected == "" {
		utils.Warn("No security log file found at expected paths (/var/log/secure, /var/log/auth.log). Collector disabled.")
		return &SecurityLogCollector{logPath: ""} // Mark as disabled (logPath is empty)
	}

	// Create the collector instance
	c := &SecurityLogCollector{
		Config:     cfg,
		logPath:    selected,
		maxMsgSize: cfg.Agent.LogCollection.MessageMax,
		batchSize:  cfg.Agent.LogCollection.BatchSize,
		// Buffer size: batchSize * some multiplier (e.g., 10) or configurable
		// Adjust buffer size based on expected log volume and collection interval
		lines: make(chan model.LogEntry, cfg.Agent.LogCollection.BatchSize*10),
		stop:  make(chan struct{}),
	}

	// Configure the tailer
	tailConfig := tail.Config{
		Location:  &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd}, // Start from the end of the file
		ReOpen:    true,                                          // Re-open file if recreated (log rotation)
		MustExist: false,                                         // Don't fail immediately if file missing, keep trying
		Follow:    true,                                          // Keep tracking the file
		Logger:    tail.DiscardingLogger,                         // Or use a logger compatible with your utils.Debug/Info
		// Poll: true, // Use polling if inotify doesn't work (e.g., NFS, some containers) - less efficient
	}

	var err error
	c.tailer, err = tail.TailFile(c.logPath, tailConfig)
	if err != nil {
		utils.Error("Failed to start tailing file %s: %v. Collector disabled.", c.logPath, err)
		// Mark as disabled by clearing the tailer
		c.tailer = nil
		c.logPath = "" // Mark as disabled
		return c
	}

	// Start the background goroutine to process lines from the tailer
	c.wg.Add(1)
	go c.runTailing()

	utils.Info("Started tailing security log file: %s", c.logPath)
	return c
}

// runTailing runs in a background goroutine, reading from the tailer
// and putting parsed entries onto the internal 'lines' channel.
func (c *SecurityLogCollector) runTailing() {
	defer c.wg.Done()
	defer func() {
		// Ensure the tailer is stopped if the goroutine exits unexpectedly
		// Use mutex to avoid race condition with Close()
		c.mu.Lock()
		if c.tailer != nil {
			// Attempt to stop, ignore error as we might be stopping due to it already being closed
			_ = c.tailer.Stop()
			c.tailer = nil // Mark as stopped
			utils.Info("Tailing stopped for: %s", c.logPath)
		}
		c.mu.Unlock()
		// Close the lines channel to signal Collect that no more lines will come
		close(c.lines)
	}()

	utils.Debug("Tailing goroutine started for: %s", c.logPath)

	for {
		select {
		case <-c.stop: // Check for stop signal first
			utils.Debug("Stop signal received for tailing: %s", c.logPath)
			return // Exit the loop and trigger deferred cleanup

		case line, ok := <-c.tailer.Lines:
			if !ok {
				// Channel closed, tailer likely stopped or encountered an error
				err := c.tailer.Err()
				if err != nil {
					utils.Error("Tailing error on %s: %v", c.logPath, err)
				} else {
					utils.Info("Tailer channel closed for %s.", c.logPath)
				}
				return // Exit loop
			}

			if line.Err != nil {
				utils.Warn("Error reading line from %s: %v", c.logPath, line.Err)
				continue // Skip this line
			}

			entry := c.parseLogLine(line.Text)
			if entry.Message == "" { // Skip empty/unparseable lines
				continue
			}

			// Try to send the parsed entry to the buffer channel.
			// Use a non-blocking send with select to avoid blocking
			// the tailer if the buffer is full (e.g., Collect not called often enough).
			select {
			case c.lines <- entry:
				// Successfully sent
			default:
				// Buffer is full, drop the log and warn
				utils.Warn("Log buffer full for %s. Dropping log entry: %s", c.logPath, entry.Message)
			}
		}
	}
}

// Close stops the tailing process gracefully.
func (c *SecurityLogCollector) Close() {
	c.mu.Lock()
	if c.tailer == nil {
		c.mu.Unlock()
		return // Already stopped or never started
	}
	// Signal the runTailing goroutine to stop *before* stopping the tailer
	// Closing the channel is the idiomatic way
	close(c.stop)
	c.mu.Unlock()

	// Wait for the runTailing goroutine to finish processing and cleanup
	c.wg.Wait()
	utils.Info("Security log collector closed for: %s", c.logPath)
}

func (c *SecurityLogCollector) Name() string {
	return "security"
}

// Collect drains the internal 'lines' channel and batches the entries.
// This is called periodically by the LogRunner.
func (c *SecurityLogCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	// If collector is disabled (no logPath or tailer failed)
	if c.logPath == "" {
		return nil, nil
	}

	var allBatches [][]model.LogEntry
	var currentBatch []model.LogEntry

	// Non-blockingly drain the lines channel
collectLoop:
	for {
		select {
		case entry, ok := <-c.lines:
			if !ok {
				// Channel closed, means tailing stopped. Might happen during shutdown.
				utils.Debug("Lines channel closed for %s during collect.", c.logPath)
				break collectLoop
			}

			currentBatch = append(currentBatch, entry)

			if len(currentBatch) >= c.batchSize {
				allBatches = append(allBatches, currentBatch)
				// Allocate new slice for the next batch
				// Important: Don't reuse the underlying array of currentBatch
				// as allBatches only stores the slice header.
				currentBatch = make([]model.LogEntry, 0, c.batchSize)
			}
		default:
			// No more lines available in the channel right now
			break collectLoop
		}
	}

	// Add any remaining logs in the current batch
	if len(currentBatch) > 0 {
		allBatches = append(allBatches, currentBatch)
	}

	if len(allBatches) > 0 {
		count := 0
		for _, b := range allBatches {
			count += len(b)
		}
		utils.Debug("Collected %d log entries in %d batches from %s", count, len(allBatches), c.logPath)
	}

	return allBatches, nil
}

// --- Keep your existing parseLogLine and detectSeverity ---
// (Copied from previous answer for completeness)
func (c *SecurityLogCollector) parseLogLine(line string) model.LogEntry {
	// Typical format: "Apr 26 07:15:01 myhost CRON[12345]: pam_unix(cron:session): session opened for user root by (uid=0)"
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return model.LogEntry{} // not a real log
	}

	// Parse timestamp (no year in log) - Robustness needed!
	tsString := strings.Join(parts[0:3], " ")
	ts, err := time.Parse("Jan 2 15:04:05", tsString)
	timestamp := time.Now() // Default to now if parse fails

	if err == nil {
		now := time.Now()
		logYear := now.Year()
		// If log month is Dec (12) and current month is Jan (1), assume log is from last year.
		if ts.Month() > now.Month() {
			logYear--
		}
		// Construct full timestamp
		timestamp = time.Date(logYear, ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), ts.Second(), 0, time.Local) // Use local time zone
	} else {
		utils.Debug("Failed to parse timestamp '%s': %v", tsString, err)
	}

	hostname := parts[3]
	sourceProc := parts[4]
	source := strings.SplitN(sourceProc, "[", 2)[0]
	source = strings.TrimRight(source, ":")

	msg := ""
	logPrefix := fmt.Sprintf("%s %s %s %s", parts[0], parts[1], parts[2], parts[3])
	procPartIndex := len(logPrefix) + 1 + len(sourceProc) + 1
	if procPartIndex < len(line) {
		msg = line[procPartIndex:]
	} else if len(parts) > 5 {
		msg = strings.Join(parts[5:], " ")
	}

	level := detectSeverity(msg)

	trimmedMsg := msg
	if len(trimmedMsg) > c.maxMsgSize {
		trimmedMsg = trimmedMsg[:c.maxMsgSize] + " [truncated]"
	}

	return model.LogEntry{
		Timestamp: timestamp,
		Level:     level,
		Message:   trimmedMsg,
		Source:    source,
		Category:  "auth",
		PID:       0, // PID extraction would need more parsing
		Tags: map[string]string{
			"log_path":     c.logPath,
			"hostname_log": hostname,
		},
		Meta: &model.LogMeta{
			Platform: "file",
			AppName:  source,
			Path:     c.logPath,
		},
	}
}

func detectSeverity(msg string) string {
	l := strings.ToLower(msg)
	switch {
	case strings.Contains(l, "failed") || strings.Contains(l, "invalid"):
		return "warn"
	case strings.Contains(l, "error") || strings.Contains(l, "denied"):
		return "error"
	case strings.Contains(l, "session opened") || strings.Contains(l, "accepted") || strings.Contains(l, "session closed"):
		return "info"
	default:
		return "info"
	}
}
