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
	"io" // Needed for Closer interface
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

// JournaldCollector streams log entries using an asynchronous background reader.
type JournaldCollector struct {
	Config *config.Config

	journal    *sdjournal.Journal
	lines      chan model.LogEntry // Internal channel for collected lines
	stop       chan struct{}       // Channel to signal background goroutine stop
	wg         sync.WaitGroup      // WaitGroup to ensure clean shutdown
	mu         sync.Mutex          // Mutex to protect access during shutdown
	once       sync.Once           // Add this field
	cleanupErr error
	batchSize  int
	maxSize    int
}

// Name returns the name of the collector.
func (j *JournaldCollector) Name() string {
	return "journald"
}

// NewJournaldCollector initializes a new JournaldCollector.
func NewJournaldCollector(cfg *config.Config) *JournaldCollector {
	utils.Info("Initializing journald collector...")
	j, err := sdjournal.NewJournal()
	if err != nil {
		utils.Error("Failed to open systemd journal: %v. Collector disabled.", err)
		return &JournaldCollector{} // Return disabled collector
	}

	// Filter for relevant priorities (e.g., INFO and higher)
	// Adjust priorities as needed (0=emerg, 1=alert, 2=crit, 3=err, 4=warn, 5=notice, 6=info, 7=debug)
	// Example: Include warning and higher
	for _, prio := range []string{"0", "1", "2", "3", "4"} {
		match := sdjournal.Match{Field: sdjournal.SD_JOURNAL_FIELD_PRIORITY, Value: prio}
		if err := j.AddMatch(match.String()); err != nil {
			utils.Warn("Failed to add journal priority match %s: %v", prio, err)
			// Continue anyway, might just get more logs
		}
		// Disjunction means OR - we want logs with PRIORITY=0 OR PRIORITY=1 OR ...
		if err := j.AddDisjunction(); err != nil {
			utils.Warn("Failed to add journal disjunction: %v", err)
		}
	}
	// Add more filters if needed (e.g., specific units)
	// j.AddMatch("_SYSTEMD_UNIT=nginx.service")

	// Seek to end to skip historical logs
	if err := j.SeekTail(); err != nil {
		utils.Error("Failed to seek journal to tail: %v. Collector might report old logs.", err)
		// Attempt to continue, but logs might be duplicated or old
	} else {
		// Seeking to the tail places the cursor *at* the last entry.
		// We need to move *past* it to only get new entries.
		// Calling Next() achieves this. Ignore result/error, just advance position.
		_, _ = j.Previous() // Move to the last entry
		// Note: Seeking tail and then immediately moving previous places cursor just before last entry
		// Waiting for the next event after this should fetch truly new logs.
		// Or alternatively, keep the j.Next() from the original code after SeekTail if that works better.
		// Let's stick with SeekTail and rely on Wait() picking up the next *new* event.
	}

	collector := &JournaldCollector{
		Config:  cfg,
		journal: j,
		// Buffer size: batchSize * some multiplier or configurable
		lines: make(chan model.LogEntry, cfg.Agent.LogCollection.BatchSize*10),
		stop:  make(chan struct{}),

		batchSize: cfg.Agent.LogCollection.BatchSize,
		maxSize:   cfg.Agent.LogCollection.MessageMax,
	}

	// Start the background reader goroutine
	collector.wg.Add(1)
	go collector.runReader()

	utils.Info("Journald collector initialized and reader started.")
	return collector
}

// runReader runs in the background, waiting for and processing journal entries.
func (j *JournaldCollector) runReader() {
	defer j.wg.Done()
	defer func() {
		// Ensure journal is closed if goroutine exits unexpectedly
		j.mu.Lock()
		if j.journal != nil {
			utils.Debug("Closing journal handle in runReader defer.")
			j.journal.Close()
			j.journal = nil
		}
		j.mu.Unlock()
		// Close the lines channel to signal Collect that no more lines will come
		close(j.lines)
		utils.Debug("Journald reader goroutine stopped.")
	}()

	utils.Debug("Journald reader goroutine started.")

	// Timeout for the Wait call. Needs to be short enough to allow
	// timely checking of the stop channel.
	waitTimeout := 2 * time.Second // Check stop channel every 2 seconds

	for {
		// Wait blocks until the journal changes, or the timeout occurs.
		// Returns 1 if journal changed, 0 if timeout, -1 on error.
		ret := j.journal.Wait(waitTimeout)

		// Check for stop signal *after* Wait returns, regardless of result.
		select {
		case <-j.stop:
			utils.Info("Stop signal received for journald reader.")
			return // Exit loop and trigger deferred cleanup
		default:
			// Continue processing if not stopped
		}

		if ret < 0 {
			utils.Error("Journal wait failed: %d. Stopping reader.", ret)
			// Consider if this error is recoverable or needs agent restart/alert
			return // Exit loop on error
		}

		// If Wait timed out (ret == 0) or journal changed (ret == 1),
		// try processing entries. This loop handles the case where multiple
		// entries arrived during the Wait or timeout.
		for {
			// Move cursor to the next entry. Returns > 0 if entry read, 0 if no more entries, < 0 on error.
			n, err := j.journal.Next()
			if err != nil {
				utils.Error("Failed reading next journal entry: %v. Stopping reader.", err)
				return // Exit loop on error
			}
			if n == 0 {
				// No more new entries currently available
				break // Exit inner processing loop, go back to Wait
			}

			// Successfully read an entry, get its data
			entry, err := j.journal.GetEntry()
			if err != nil {
				utils.Warn("Failed to get journal entry data: %v. Skipping entry.", err)
				continue // Skip this entry, try next
			}

			// Filter out kernel messages if desired (as in original code)
			// Could be made configurable
			if entry.Fields["SYSLOG_IDENTIFIER"] == "kernel" {
				continue
			}

			// Parse and build the log entry
			log := buildLogEntry(entry, j.maxSize)

			// Send parsed entry to buffer channel, non-blockingly
			select {
			case j.lines <- log:
				// Successfully sent
			case <-j.stop: // Check stop again in case it happened during processing
				utils.Info("Stop signal received while processing journal entry.")
				return
			default:
				// Buffer full, drop log and warn
				utils.Warn("Journald log buffer full. Dropping log entry: %s", log.Message)
			}
		} // End inner processing loop
	} // End outer wait loop
}

// Collect drains the internal 'lines' channel and batches the entries.
func (j *JournaldCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	// Check if collector is disabled (e.g., journal handle is nil)
	j.mu.Lock()
	isDisabled := j.journal == nil
	j.mu.Unlock()
	if isDisabled {
		// Return nil, nil to indicate no error but no data (collector is disabled)
		return nil, nil
	}

	var allBatches [][]model.LogEntry
	var currentBatch []model.LogEntry

	// Non-blockingly drain the lines channel
collectLoop:
	for {
		select {
		case entry, ok := <-j.lines:
			if !ok {
				// Channel closed, means reader stopped (likely during shutdown or error)
				utils.Warn("Journald lines channel closed during collect.")
				// Check if collector is now disabled due to reader error
				j.mu.Lock()
				isDisabled = j.journal == nil
				j.mu.Unlock()
				if isDisabled {
					// If reader stopped due to error and closed journal, report maybe?
					// For now, just break loop. Might need better error propagation.
					utils.Debug("Journald collector is disabled, stopping collection.")
				}
				break collectLoop
			}

			currentBatch = append(currentBatch, entry)

			if len(currentBatch) >= j.batchSize {
				allBatches = append(allBatches, currentBatch)
				// Allocate new slice for the next batch to avoid underlying array reuse issues
				currentBatch = make([]model.LogEntry, 0, j.batchSize)
			}
		case <-ctx.Done():
			// Context provided by the runner/registry was cancelled
			utils.Warn("Collect context cancelled for journald.")
			// Return what we have collected so far plus context error
			if len(currentBatch) > 0 {
				allBatches = append(allBatches, currentBatch)
			}
			return allBatches, ctx.Err()
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
		utils.Debug("Collected %d journald entries in %d batches", count, len(allBatches))
	}

	// Return nil error, as errors during collection itself are handled internally
	// Errors during reading are logged by the background goroutine.
	return allBatches, nil
}

// Close stops the background reader and closes the journal handle.
// Implements io.Closer.
func (j *JournaldCollector) Close() error {
	j.once.Do(func() {
		j.mu.Lock()
		if j.journal == nil {
			j.mu.Unlock()
			utils.Debug("Journald collector already closed or was never started.")
			return

		}
		utils.Info("Closing journald collector...")
		// Signal the runReader goroutine to stop
		close(j.stop)
		// The journal handle itself is closed in the runReader's defer func
		// just before wg.Done()
		j.mu.Unlock() // Unlock before waiting

		// Wait for the runReader goroutine to finish cleanly
		j.wg.Wait()

		utils.Info("Journald collector closed.")
	})
	return j.cleanupErr
}

// mapPriorityToLevel maps systemd journal priority levels to log levels.
func mapPriorityToLevel(priority string) string {
	switch priority {
	case "0", "1", "2": // emerg, alert, crit
		return "critical"
	case "3": // err
		return "error" // Often mapped to error as well
	case "4": // warning
		return "warn"
	case "5": // notice
		return "info" // Often mapped to info
	case "6": // informational
		return "info"
	case "7": // debug
		return "debug"
	default:
		return "unknown"
	}
}

// buildLogEntry constructs a LogEntry from a systemd journal entry.
// Fields: cleaned journald fields (no `_` prefix)
// Meta.Extra: original raw journald fields for advanced filtering/debugging
func buildLogEntry(entry *sdjournal.JournalEntry, maxSize int) model.LogEntry {

	timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond))

	msg := entry.Fields["MESSAGE"]
	// Ensure msg is valid UTF-8 *before* potentially truncating
	if !utf8.ValidString(msg) {
		msg = sanitizeUTF8(msg)
	}
	// Truncate after sanitizing
	if len(msg) > maxSize && maxSize > 0 { // Check maxSize > 0
		msg = msg[:maxSize] + " [truncated]"
	}

	// Filtered fields into Fields map
	wanted := []string{
		"_SYSTEMD_UNIT", "_SYSTEMD_SLICE", "_EXE",
		"_CMDLINE", "_PID", "_UID", "MESSAGE_ID",
		"SYSLOG_IDENTIFIER", "_COMM", "CONTAINER_ID",
		"CONTAINER_NAME"}

	fields := make(map[string]string)
	for _, k := range wanted {
		if v, ok := entry.Fields[k]; ok && v != "" { // Only add if value exists and is not empty
			fields[strings.TrimPrefix(k, "_")] = v // Trim leading _ for cleaner field names
		}
	}
	// Classify the category based on SYSLOG_IDENTIFIER or SYSTEMD_UNIT using the classifyCategory function
	category := classifyCategory(fields["SYSLOG_IDENTIFIER"], fields["SYSTEMD_UNIT"])

	// Try to extact user information from fields
	user := extractUserFromFields(entry.Fields)

	// Add priority and hostname if available
	if v := entry.Fields["_PRIORITY"]; v != "" {
		fields["PRIORITY"] = v
	}
	if v := entry.Fields["_HOSTNAME"]; v != "" {
		fields["HOSTNAME"] = v
	}

	// Simplified Tags - use Fields map for most details
	tags := map[string]string{
		// Add essential tags for quick filtering/grouping if needed
		// "unit": category, // Maybe redundant if in Fields
	}
	if cid := entry.Fields["CONTAINER_ID"]; cid != "" {
		tags["container_id"] = cid
	}
	if cname := entry.Fields["CONTAINER_NAME"]; cname != "" {
		tags["container_name"] = cname
	}

	// Keep the RAW  fields - may need?
	extra := map[string]string{}
	for _, k := range wanted {
		if v, ok := entry.Fields[k]; ok && v != "" {
			extra[k] = v
		}
	}

	return model.LogEntry{
		Timestamp: timestamp,
		Level:     mapPriorityToLevel(entry.Fields["PRIORITY"]),
		Message:   msg,
		Source:    "systemd", // e.g., systemd, auth, gosight
		Category:  category,  // e.g., systemd unit or slice
		PID:       parsePID(entry.Fields["_PID"]),
		Fields:    fields, // Richer metadata goes here
		Tags:      tags,   // Minimal, high-value tags
		Meta: &model.LogMeta{ // Keep essential routing/origin info here
			Platform:      "journald",
			AppName:       fields["SYSLOG_IDENTIFIER"],
			ContainerID:   fields["CONTAINER_ID"],
			ContainerName: fields["CONTAINER_NAME"],
			Unit:          fields["SYSTEMD_UNIT"], // Explicitly store unit if present
			User:          user,
			Executable:    fields["EXE"],
			Extra:         extra, // Keep all raw fields for potential future use
		},
	}
}

// parsePID converts a string representation of a PID to an integer.
func parsePID(pidStr string) int {
	pid, _ := strconv.Atoi(pidStr) // Ignore error, defaults to 0
	return pid
}

// sanitizeUTF8 ensures that the string is valid UTF-8.
func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// Replace invalid sequences with the replacement character ''
	return strings.ToValidUTF8(s, "\uFFFD")
}

// Ensure JournaldCollector implements io.Closer
var _ io.Closer = (*JournaldCollector)(nil)

// User mapping for journald logs based on best guess
// extractUserFromFields attempts to extract the user from journal fields.
// It checks for specific fields in a priority order and falls back to UID lookup if necessary.
func extractUserFromFields(fields map[string]string) string {
	// Priority order: named fields
	if user := fields["USERNAME"]; user != "" {
		return user
	}
	if user := fields["SUDO_USER"]; user != "" {
		return user
	}
	if user := fields["SSH_USER"]; user != "" {
		return user
	}

	// Try UID lookup
	var uidStr string
	if uidStr = fields["UID"]; uidStr == "" {
		uidStr = fields["_UID"]
	}
	if uidStr != "" {
		if uid, err := strconv.Atoi(uidStr); err == nil {
			if u, err := user.LookupId(strconv.Itoa(uid)); err == nil {
				return u.Username
			}
		}
	}

	// fallback
	return ""
}

// Category mapping for journald logs based on SYSLOG_IDENTIFIER or SYSTEMD_UNIT
// This mapping is used to categorize logs into predefined categories.

// journaldCategoryMap is a mapping of common journald identifiers to categories.
var journaldCategoryMap = []struct {
	MatchKey string // lowercased value of SYSLOG_IDENTIFIER or SYSTEMD_UNIT
	Category string
}{
	// Authentication
	{"sshd", "auth"},
	{"sudo", "auth"},
	{"login", "auth"},

	// Security
	{"firewalld", "security"},
	{"auditd", "security"},
	{"selinux", "security"},

	// Network
	{"networkmanager", "network"},
	{"dhclient", "network"},
	{"resolvconf", "network"},

	// Container
	{"podman", "container"},
	{"docker", "container"},
	{"containerd", "container"},

	// Application servers
	{"nginx", "app"},
	{"apache2", "app"},
	{"postgres", "app"},
	{"mysqld", "app"},

	// System / OS
	{"systemd", "system"},
	{"cron", "system"},
	{"crond", "system"},
	{"kernel", "system"},

	// Internal
	{"gosight", "gosight"},
}

// classifyCategory determines the category of a log entry based on its syslog identifier or systemd unit.
func classifyCategory(syslogID, unit string) string {
	inputs := []string{strings.ToLower(syslogID), strings.ToLower(unit)}
	for _, input := range inputs {
		for _, rule := range journaldCategoryMap {
			if strings.Contains(input, rule.MatchKey) {
				return rule.Category
			}
		}
	}
	return "system" // default fallback
}
