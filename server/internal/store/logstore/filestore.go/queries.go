package filestore

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

func (v *FileStore) GetLogs(filter model.LogFilter) ([]model.LogEntry, error) {
	files, err := filepath.Glob(filepath.Join(v.dir, "*.json.gz"))
	if err != nil {
		return nil, err
	}

	if filter.Order == "asc" {
		sort.Strings(files)
	} else {
		sort.Sort(sort.Reverse(sort.StringSlice(files)))
	}

	var result []model.LogEntry

	maxScan := 20000

	count := 0
	for _, file := range files {
		if count >= maxScan {
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
			// Optional: sort entries within file
			sort.Slice(payload.Logs, func(i, j int) bool {
				if filter.Order == "asc" {
					return payload.Logs[i].Timestamp.Before(payload.Logs[j].Timestamp)
				}
				return payload.Logs[i].Timestamp.After(payload.Logs[j].Timestamp)
			})

			for _, entry := range payload.Logs {
				count++
				ts := entry.Timestamp

				// Enrich tags
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

				// Cursor filtering
				if !filter.Cursor.IsZero() {
					cursor := filter.Cursor.Add(-1 * time.Nanosecond)
					if filter.Order == "asc" && !ts.After(cursor) {
						continue
					}
					if filter.Order != "asc" && !ts.Before(cursor) {
						continue
					}
				}

				// Time range filtering
				if !filter.Start.IsZero() && ts.Before(filter.Start) {
					continue
				}
				if !filter.End.IsZero() && ts.After(filter.End) {
					continue
				}

				if filter.Level != "" && !strings.EqualFold(entry.Level, filter.Level) {
					continue
				}
				if filter.Source != "" && !strings.EqualFold(entry.Source, filter.Source) {
					continue
				}

				if filter.Category != "" && !strings.EqualFold(entry.Category, filter.Category) {
					continue
				}

				// Flat field filter
				match := func(key, want string) bool {
					if want == "" {
						return true
					}
					return strings.EqualFold(entry.Tags[key], want)
				}
				if !match("endpoint_id", filter.EndpointID) ||

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

				// Meta match
				meta := entry.Meta

				// Tags match
				for k, v := range filter.Tags {
					if actual, ok := entry.Tags[k]; !ok || !strings.EqualFold(actual, v) {
						goto skip
					}
				}

				// Fields match
				for k, v := range filter.Fields {
					if actual, ok := entry.Fields[k]; !ok || !strings.EqualFold(actual, v) {
						goto skip
					}
				}

				for k, v := range filter.Meta {
					switch strings.ToLower(k) {
					case "platform":
						if !strings.EqualFold(meta.Platform, v) {
							goto skip
						}
					case "app_name":
						if !strings.EqualFold(meta.AppName, v) {
							goto skip
						}
					case "app_version":
						if !strings.EqualFold(meta.AppVersion, v) {
							goto skip
						}
					case "container_id":
						if !strings.EqualFold(meta.ContainerID, v) {
							goto skip
						}
					case "container_name":
						if !strings.EqualFold(meta.ContainerName, v) {
							goto skip
						}
					case "unit":
						if !strings.EqualFold(meta.Unit, v) {
							goto skip
						}
					case "service":
						if !strings.EqualFold(meta.Service, v) {
							goto skip
						}
					case "event_id":
						if !strings.EqualFold(meta.EventID, v) {
							goto skip
						}
					case "user":
						if !strings.EqualFold(meta.User, v) {
							goto skip
						}
					case "exe":
						if !strings.EqualFold(meta.Executable, v) {
							goto skip
						}
					case "path":
						if !strings.EqualFold(meta.Path, v) {
							goto skip
						}
					default:
						if actual, ok := meta.Extra[k]; !ok || !strings.EqualFold(actual, v) {
							goto skip
						}
					}
				}

				// Add to result
				result = append(result, entry)
				if filter.Limit > 0 && len(result) >= filter.Limit {
					break
				}
			skip:
			}
		}
		_ = gz.Close()
		_ = f.Close()
	}

	// Final safety sort
	sort.Slice(result, func(i, j int) bool {
		if filter.Order == "asc" {
			return result[i].Timestamp.Before(result[j].Timestamp)
		}
		return result[i].Timestamp.After(result[j].Timestamp)
	})
	// Trim after sort
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}

	return result, nil
}
