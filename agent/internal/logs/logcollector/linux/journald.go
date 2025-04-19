package linuxcollector

import (
	"context"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

// JournaldTailCollector streams only new log entries since start.
type JournaldCollector struct {
	journal *sdjournal.Journal
	Config  *config.Config
}

func (j *JournaldCollector) Name() string {
	return "journald"
}

func NewJournaldCollector(cfg *config.Config) *JournaldCollector {
	j, err := sdjournal.NewJournal()
	if err != nil {
		// Optional: log or panic if initialization fails
		utils.Debug("Failed to open systemd journal: %v", err)
	}

	j.AddMatch("PRIORITY<=4")
	// Seek to end to skip historical logs
	if err := j.SeekTail(); err != nil {
		utils.Debug("failed to seek to tail: %v\n", err)

	}
	// Skip current tail entry (so we get only new logs)
	_, err = j.Next()
	if err != nil {
		utils.Debug("failed to get next entry: %v\n", err)

	}
	return &JournaldCollector{
		Config:  cfg,
		journal: j,
	}
}
func (j *JournaldCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	var results []model.LogEntry

	// Wait for new logs up to 2 seconds
	j.journal.Wait(time.Second * 2)

	// Read new logs
	for {
		n, err := j.journal.Next()
		if err != nil || n == 0 {
			break
		}
		entry, err := j.journal.GetEntry()
		if err != nil {
			continue
		}

		log := buildLogEntry(entry, 1000)
		results = append(results, log)
	}

	if len(results) == 0 {
		return nil, nil
	}
	return [][]model.LogEntry{results}, nil
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
func parsePID(pidStr string) int {
	pid, _ := strconv.Atoi(pidStr)
	return pid
}
