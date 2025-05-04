package filestore

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

func (v *FileStore) GetRecentLogs(ctx context.Context, filter model.LogFilter) ([]model.LogEntry, error) {
	files, err := filepath.Glob(filepath.Join(v.Dir, "*.json.gz"))
	if err != nil {
		return nil, err
	}

	// Sort files newest-first
	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	var result []model.LogEntry
	maxRead := filter.Limit * 5
	if maxRead == 0 {
		maxRead = 5000
	}

	for _, file := range files {
		if len(result) >= maxRead {
			break
		}

		f, err := os.Open(file)
		if err != nil {
			continue
		}
		gz, err := gzip.NewReader(f)
		if err != nil {
			_ = f.Close()
			continue
		}

		var payload model.LogPayload
		if err := json.NewDecoder(gz).Decode(&payload); err == nil {
			for _, entry := range payload.Logs {
				if entry.Tags == nil {
					entry.Tags = make(map[string]string)
				}

				// Add metadata
				entry.Tags["endpoint_id"] = payload.Meta.EndpointID
				entry.Tags["agent_id"] = payload.Meta.AgentID
				entry.Tags["host_id"] = payload.Meta.HostID
				entry.Tags["hostname"] = payload.Meta.Hostname
				entry.Tags["job"] = payload.Meta.Tags["job"]
				for k, v := range payload.Meta.Tags {
					entry.Tags[k] = v
				}

				// Timestamp filtering
				ts := entry.Timestamp

				if !filter.Start.IsZero() && ts.Before(filter.Start) {
					continue
				}
				if !filter.End.IsZero() && ts.After(filter.End) {
					continue
				}

				// Apply additional filters
				if filter.EndpointID != "" && entry.Tags["endpoint_id"] != filter.EndpointID {
					continue
				}
				if filter.Target != "" && entry.Tags["target"] != filter.Target {
					continue
				}
				if filter.Level != "" && entry.Level != filter.Level {
					continue
				}
				if filter.Category != "" && entry.Category != filter.Category {
					continue
				}
				if filter.Contains != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(filter.Contains)) {
					continue
				}

				// Append valid entry to result
				result = append(result, entry)
				if filter.Limit > 0 && len(result) >= filter.Limit {
					break
				}
			}
		}

		_ = gz.Close()
		_ = f.Close()
	}

	// Sort final result
	if filter.Order == "asc" {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Timestamp.Before(result[j].Timestamp)
		})
	} else {
		sort.Slice(result, func(i, j int) bool {
			return result[i].Timestamp.After(result[j].Timestamp)
		})
	}

	return result, nil
}

func contains(s, substr string) bool {
	return substr == "" || strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
