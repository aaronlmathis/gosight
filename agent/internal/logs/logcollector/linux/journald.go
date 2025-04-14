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

func (c *JournaldCollector) Collect(ctx context.Context) ([]model.LogEntry, error) {
	var logs []model.LogEntry

	_ = c.journal.SeekTail() // read most recent first

	filters := BuildJournaldFilterList(c.Config)
	for _, match := range filters {
		_ = c.journal.AddMatch(match)
	}

	wait := c.journal.Wait(500 * time.Millisecond)
	if wait != sdjournal.SD_JOURNAL_APPEND {
		return logs, nil
	}

	for {
		select {
		case <-ctx.Done():
			return logs, ctx.Err()

		default:
			n, err := c.journal.Next()
			if err != nil || n == 0 {
				return logs, err
			}

			entry, err := c.journal.GetEntry()
			if err != nil {
				continue
			}

			timestamp := time.Unix(0, int64(entry.RealtimeTimestamp)*int64(time.Microsecond))

			log := model.LogEntry{
				Timestamp: timestamp,
				Level:     mapPriorityToLevel(entry.Fields["PRIORITY"]),
				Message:   entry.Fields["MESSAGE"],
				Source:    entry.Fields["SYSLOG_IDENTIFIER"],
				Category:  "", // Optional, can derive from unit
				Host:      "", // Fill in at runner level
				PID:       parsePID(entry.Fields["_PID"]),
				Fields:    entry.Fields,
				Tags: map[string]string{
					"unit": entry.Fields["_SYSTEMD_UNIT"],
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
					Extra:         map[string]string{},
				},
			}

			logs = append(logs, log)

			if len(logs) >= 500 {
				break
			}
		}

		return logs, nil
	}
}

func BuildJournaldFilterList(cfg *config.Config) []string {
	var filters []string

	// Optional: focus on this host
	if cfg.Agent.HostOverride != "" {
		filters = append(filters, fmt.Sprintf("_HOSTNAME=%s", cfg.Agent.HostOverride))
	}

	// Based on enabled services
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
			filters = append(filters, "_SYSTEMD_UNIT=docker.service")
		case "kernel":
			filters = append(filters, "_TRANSPORT=kernel")
		default:
			// Assume it's a valid SYSLOG_IDENTIFIER or service
			filters = append(filters, fmt.Sprintf("SYSLOG_IDENTIFIER=%s", name))
		}
	}

	// Optional fallback: collect nothing if no filters specified
	if len(filters) == 0 {
		filters = append(filters, "PRIORITY<=4") // info, warn, error
	}

	return filters
}

func parsePID(pidStr string) int {
	if pidStr == "" {
		return 0
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}
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
