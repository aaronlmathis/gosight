package linuxcollector

import (
	"bufio"
	"context"
	"os"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type SecurityLogCollector struct {
	Config     *config.Config
	logPath    string
	maxMsgSize int
	batchSize  int
}

func NewSecurityLogCollector(cfg *config.Config) *SecurityLogCollector {
	// Try both common paths
	paths := []string{"/var/log/secure", "/var/log/auth.log"}
	var selected string
	for _, p := range paths {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			selected = p
			break
		}
	}

	return &SecurityLogCollector{
		Config:     cfg,
		logPath:    selected,
		maxMsgSize: cfg.Agent.LogCollection.MessageMax,
		batchSize:  cfg.Agent.LogCollection.BatchSize,
	}
}

func (c *SecurityLogCollector) Name() string {
	return "security"
}

func (c *SecurityLogCollector) Collect(ctx context.Context) ([][]model.LogEntry, error) {
	utils.Debug("🟢 SecurityLogCollector starting tail of %s", c.logPath)

	file, err := os.Open(c.logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.Seek(0, os.SEEK_END); err != nil {
		return nil, err
	}

	//	reader := bufio.NewReader(file)
	var allBatches [][]model.LogEntry
	var current []model.LogEntry
	ticker := time.NewTicker(c.Config.Agent.Interval)
	defer ticker.Stop()

	fileInfo, _ := file.Stat()
	lastSize := fileInfo.Size()
	scanner := bufio.NewScanner(file)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case <-ticker.C:
			if len(current) > 0 {
				utils.Debug("Interval triggered flush with %d logs", len(current))
				allBatches = append(allBatches, current)
				current = nil
			}
		default:
			if scanner.Scan() {
				line := scanner.Text()
				utils.Debug("🔍 Got new log line: %s", line)

				entry := c.parseLogLine(line)
				if entry.Message == "" {
					continue
				}
				current = append(current, entry)

				if len(current) >= c.batchSize {
					utils.Debug("📦 Batch size reached (%d), flushing", c.batchSize)
					allBatches = append(allBatches, current)
					current = nil
				}
			} else {
				// Check for new content
				newInfo, _ := file.Stat()
				if newInfo.Size() > lastSize {
					lastSize = newInfo.Size()
				} else {
					time.Sleep(200 * time.Millisecond)
				}
			}
		}
	}

	if len(current) > 0 {
		allBatches = append(allBatches, current)
	}

	return allBatches, nil
}

func (c *SecurityLogCollector) parseLogLine(line string) model.LogEntry {
	// Typical format: "Apr 17 19:45:36 hostname sshd[123]: Failed password for invalid user root"
	parts := strings.Fields(line)
	if len(parts) < 5 {
		return model.LogEntry{} // not a real log
	}

	// Parse timestamp (no year in log)
	ts, _ := time.Parse("Jan 2 15:04:05", strings.Join(parts[0:3], " "))
	timestamp := ts
	if ts.IsZero() {
		timestamp = time.Now()
	} else {
		timestamp = timestamp.AddDate(time.Now().Year(), 0, 0)
	}

	source := parts[4]
	msg := strings.Join(parts[5:], " ")
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
		PID:       0,
		Tags: map[string]string{
			"log_path": c.logPath,
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
	case strings.Contains(l, "session opened") || strings.Contains(l, "accepted"):
		return "info"
	default:
		return "debug"
	}
}
