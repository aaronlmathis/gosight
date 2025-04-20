package linuxcollector

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	agentutils "github.com/aaronlmathis/gosight/agent/internal/utils"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

type JournaldCollector struct {
	journal    *sdjournal.Journal
	Config     *config.Config
	lastCursor string
}

func NewJournaldCollector(cfg *config.Config) *JournaldCollector {
	j, err := sdjournal.NewJournal()
	if err != nil {
		// Optional: log or panic if initialization fails
		utils.Debug("Failed to open systemd journal: %v", err)
	}
	lastCursor, err := agentutils.LoadCursor(cfg.Agent.LogCollection.CursorFile)
	if err != nil {
		utils.Warn("⚠️ Error loading last cursor: %v", err)
		lastCursor = "" // Start from now if loading fails
	} else {
		utils.Debug("Loaded last cursor: %s", lastCursor)
	}
	return &JournaldCollector{
		Config:     cfg,
		journal:    j,
		lastCursor: lastCursor,
	}
}

func (c *JournaldCollector) Name() string {
	return "journald"
}

func (c *JournaldCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	utils.Debug("🟢 Entered Collect() for JournaldCollector")

	var allBatches [][]model.LogEntry
	var current []model.LogEntry

	if c.journal == nil {
		utils.Warn("Journal not initialized")
		return allBatches, nil
	}

	batchSize := c.Config.Agent.LogCollection.BatchSize
	maxMsgSize := c.Config.Agent.LogCollection.MessageMax
	utils.Debug("maxMsgSize: %d, batchSize: %d", maxMsgSize, batchSize)

	// Always apply a default filter: only priority 0–4 (emerg to warning)
	c.journal.FlushMatches()
	_ = c.journal.AddMatch("PRIORITY<=4")

	if c.lastCursor == "" {
		utils.Debug("📭 No prior cursor loaded — seeking to tail (most recent)")
		err := c.journal.SeekTail()
		if err != nil {
			utils.Warn("⚠️ Failed to seek to tail: %v", err)
		} else {
			utils.Debug("Successfully called SeekTail()")
			n, err := c.journal.Next()
			if err == nil && n > 0 {
				entry, err := c.journal.GetEntry()
				if err == nil && entry != nil {
					entryTimestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond)).UTC()
					utils.Debug("First entry after SeekTail (no cursor): %s | Timestamp (UTC): %s", entry.Fields["MESSAGE"], entryTimestamp.String())
				} else {
					utils.Warn("Error getting first entry after SeekTail: %v, entry: %v", err, entry)
				}
			} else {
				utils.Debug("No entry immediately after SeekTail: n=%d, err=%v", n, err)
			}
		}
	} else {
		utils.Debug("Seeking to last known cursor: %s", c.lastCursor)
		err := c.journal.SeekCursor(c.lastCursor)
		if err != nil {
			utils.Warn("Failed to seek to saved cursor (%s): %v. Falling back to seeking tail.", c.lastCursor, err)
			_ = c.journal.SeekTail()
			c.lastCursor = "" // Reset lastCursor
		}
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			r := c.journal.Wait(250 * time.Millisecond)
			utils.Debug("journal.Wait() returned: %d", r)
			if r != sdjournal.SD_JOURNAL_APPEND {
				break loop
			}

			n, err := c.journal.Next()
			if err != nil {
				utils.Error("journal.Next() failed: %v", err)
				break loop
			}
			utils.Debug("journal.Next() advanced by: %d", n)
			if n == 0 {
				utils.Debug("journal.Next() returned 0, breaking loop.")
				break loop
			}

			entry, err := c.journal.GetEntry()
			if err != nil || entry == nil {
				utils.Warn("Failed to get journal entry: %v, entry: %v", err, entry)
				continue
			}

			entryTimestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond)).UTC()
			utils.Debug("Processing entry: %s | %s | Cursor: %s | Timestamp (UTC): %s",
				entry.Fields["SYSLOG_IDENTIFIER"], entry.Fields["MESSAGE"], entry.Cursor, entryTimestamp.String())
			utils.Debug("maxMsgSize: %d, batchSize: %d", maxMsgSize, batchSize)
			log := buildLogEntry(entry, maxMsgSize)
			current = append(current, log)

			if len(current) >= batchSize {
				allBatches = append(allBatches, current)
				current = nil
			}

			// Save cursor for next run
			if cursor := entry.Cursor; cursor != "" {
				utils.Debug("About to save cursor: %s", cursor)
				c.lastCursor = cursor
				err := agentutils.SaveCursor(c.Config.Agent.LogCollection.CursorFile, cursor)
				if err != nil {
					utils.Error("Error saving cursor: %v", err)
				} else {
					utils.Debug("Saved cursor: %s", cursor)
				}
			}
		}
	}

	if len(current) > 0 {
		allBatches = append(allBatches, current)
	}

	utils.Debug("Collect() returning %d batches", len(allBatches))
	return allBatches, nil
}

func buildLogEntry(entry *sdjournal.JournalEntry, maxSize int) model.LogEntry {
	timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond))
	msg := entry.Fields["MESSAGE"]
	if len(msg) > maxSize {
		msg = msg[:maxSize] + " [truncated]"
	}

	// Filtered fields
	wanted := []string{"_SYSTEMD_UNIT", "_EXE", "_CMDLINE", "_PID", "_UID", "MESSAGE_ID", "SYSLOG_IDENTIFIER", "CONTAINER_ID", "CONTAINER_NAME"}
	fields := make(map[string]string)
	for _, k := range wanted {
		if v, ok := entry.Fields[k]; ok {
			fields[k] = v
		}
	}

	return model.LogEntry{
		Timestamp: timestamp,
		Level:     mapPriorityToLevel(entry.Fields["PRIORITY"]),
		Message:   msg,
		Source:    entry.Fields["SYSLOG_IDENTIFIER"],
		Category:  entry.Fields["_SYSTEMD_UNIT"],
		PID:       parsePID(entry.Fields["_PID"]),
		Fields:    fields,
		Tags: map[string]string{
			"unit":           entry.Fields["_SYSTEMD_UNIT"],
			"container_id":   entry.Fields["CONTAINER_ID"],
			"container_name": entry.Fields["CONTAINER_NAME"],
		},
		Meta: &model.LogMeta{
			Platform:      "journald",
			AppName:       entry.Fields["SYSLOG_IDENTIFIER"],
			ContainerID:   entry.Fields["CONTAINER_ID"],
			ContainerName: entry.Fields["CONTAINER_NAME"],
			Unit:          entry.Fields["_SYSTEMD_UNIT"],
			Service:       entry.Fields["SYSLOG_IDENTIFIER"],
			EventID:       entry.Fields["MESSAGE_ID"],
			User:          entry.Fields["_UID"],
			Executable:    entry.Fields["_EXE"],
			Path:          entry.Fields["_CMDLINE"],
		},
	}
}

func BuildJournaldFilterList(cfg *config.Config) []string {
	var filters []string

	if cfg.Agent.HostOverride != "" {
		filters = append(filters, fmt.Sprintf("_HOSTNAME=%s", cfg.Agent.HostOverride))
	}

	for _, name := range cfg.Agent.LogCollection.Services {
		switch name {
		case "nginx":
			filters = append(filters, "SYSLOG_IDENTIFIER=nginx")
		case "httpd":
			filters = append(filters, "SYSLOG_IDENTIFIER=httpd")
		case "sshd":
			filters = append(filters, "_SYSTEMD_UNIT=sshd.service")
		case "docker":
			filters = append(filters, "_SYSTEMD_UNIT=docker.service")
		case "podman":
			filters = append(filters, "_SYSTEMD_UNIT=podman.service")
		case "kernel":
			filters = append(filters, "_TRANSPORT=kernel")
		default:
			filters = append(filters, fmt.Sprintf("SYSLOG_IDENTIFIER=%s", name))
		}
	}

	if len(filters) == 0 {
		filters = append(filters, "PRIORITY<=4")
	}

	return filters
}

func parsePID(pidStr string) int {
	pid, _ := strconv.Atoi(pidStr)
	return pid
}

func mapPriorityToLevel(priority string) string {
	switch priority {
	case "0", "1", "2":
		return "error"
	case "3":
		return "warn"
	case "4":
		return "info"
	case "5":
		return "notice"
	case "6":
		return "debug"
	default:
		return "unknown"
	}
}
