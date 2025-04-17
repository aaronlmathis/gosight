package linuxcollector

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

type JournaldCollector struct {
	journal *sdjournal.Journal
	Config  *config.Config
}

func NewJournaldCollector(cfg *config.Config) *JournaldCollector {
	j, err := sdjournal.NewJournal()
	if err != nil {
		// Optional: log or panic if initialization fails
		utils.Debug("Failed to open systemd journal: %v", err)
	}

	return &JournaldCollector{
		Config:  cfg,
		journal: j,
	}
}

func (c *JournaldCollector) Name() string {
	return "journald"
}

func (c *JournaldCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {

	var allBatches [][]model.LogEntry
	var current []model.LogEntry

	batchSize := c.Config.Agent.LogCollection.BatchSize
	maxMessageSize := c.Config.Agent.LogCollection.MessageMax

	_ = c.journal.SeekTail() // read most recent first

	// Clear existing matches and apply default filter: only important log levels
	c.journal.FlushMatches()
	_ = c.journal.AddMatch("PRIORITY<=4")

	wait := c.journal.Wait(500 * time.Millisecond)
	if wait != sdjournal.SD_JOURNAL_APPEND {
		return allBatches, nil
	}

	for {
		select {
		case <-ctx.Done():
			if len(current) > 0 {
				allBatches = append(allBatches, current)
			}
			return allBatches, ctx.Err()

		default:
			n, err := c.journal.Next()
			if err != nil || n == 0 {
				if len(current) > 0 {
					allBatches = append(allBatches, current)
				}
				return allBatches, err
			}

			entry, err := c.journal.GetEntry()
			if err != nil {
				continue
			}

			log := buildLogEntry(entry, maxMessageSize)
			current = append(current, log)

			if len(current) >= batchSize {
				allBatches = append(allBatches, current)
				current = nil
			}
		}

	}
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
		Host:      entry.Fields["_HOSTNAME"],
		Fields:    fields,
		Tags: map[string]string{
			"unit":           entry.Fields["_SYSTEMD_UNIT"],
			"container_id":   entry.Fields["CONTAINER_ID"],
			"container_name": entry.Fields["CONTAINER_NAME"],
		},
		Meta: &model.LogMeta{
			OS:            "linux",
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
