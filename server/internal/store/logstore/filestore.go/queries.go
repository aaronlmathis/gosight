package filestore

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aaronlmathis/gosight/shared/model"
)

func (v *FileStore) GetLogs(filter model.LogFilter) ([]model.LogEntry, error) {
	files, err := filepath.Glob(filepath.Join(v.dir, "*.json.gz"))
	if err != nil {
		return nil, err
	}

	// Sort files by order (affects load direction)
	if filter.Order == "asc" {
		sort.Strings(files)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}

	var result []model.LogEntry
	maxScan := filter.Limit * 5
	if maxScan == 0 {
		maxScan = 5000
	}

	count := 0
	for _, file := range files {
		if count >= maxScan || (filter.Limit > 0 && len(result) >= filter.Limit) {
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
				count++
				ts := entry.Timestamp

				// Add enriched tags
				if entry.Tags == nil {
					entry.Tags = make(map[string]string)
				}
				entry.Tags["endpoint_id"] = payload.Meta.EndpointID
				entry.Tags["agent_id"] = payload.Meta.AgentID
				entry.Tags["host_id"] = payload.Meta.HostID
				entry.Tags["hostname"] = payload.Meta.Hostname
				entry.Tags["job"] = payload.Meta.Tags["job"]
				for k, v := range payload.Meta.Tags {
					entry.Tags[k] = v
				}

				// Cursor filter
				if !filter.Cursor.IsZero() {
					if filter.Order == "asc" && !ts.After(filter.Cursor) {
						continue
					}
					if filter.Order != "asc" && !ts.Before(filter.Cursor) {
						continue
					}
				}

				// Time range
				if !filter.Start.IsZero() && ts.Before(filter.Start) {
					continue
				}
				if !filter.End.IsZero() && ts.After(filter.End) {
					continue
				}

				// Property filters
				match := func(key, want string) bool {
					if want == "" {
						return true
					}
					return strings.EqualFold(entry.Tags[key], want)
				}

				if !match("endpoint_id", filter.EndpointID) ||
					!match("target", filter.Target) ||
					!match("level", filter.Level) ||
					!match("category", filter.Category) ||
					!match("source", filter.Source) ||
					!match("unit", filter.Unit) ||
					!match("app_name", filter.AppName) ||
					!match("service", filter.Service) ||
					!match("event_id", filter.EventID) ||
					!match("user", filter.User) ||
					!match("container_id", filter.ContainerID) ||
					!match("container_name", filter.ContainerName) ||
					!match("platform", filter.Platform) {
					continue
				}

				if filter.Contains != "" && !strings.Contains(strings.ToLower(entry.Message), strings.ToLower(filter.Contains)) {
					continue
				}

				result = append(result, entry)

				if filter.Limit > 0 && len(result) >= filter.Limit {
					break
				}
			}
		}
		_ = gz.Close()
		_ = f.Close()
	}

	// Final sort
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

