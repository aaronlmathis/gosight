package filestore

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/aaronlmathis/gosight/shared/model"
)

func (v *FileStore) GetRecentLogs(limit int) ([]model.LogEntry, error) {
	files, err := filepath.Glob(filepath.Join(v.Dir, "*.json.gz"))
	if err != nil {
		return nil, err
	}

	// Sort files newest-first based on filename timestamp
	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	var all []model.LogEntry
	maxRead := limit * 5 // Safety cap to avoid loading too many logs

	for _, file := range files {
		if len(all) >= maxRead {
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
			for i := range payload.Logs {
				if payload.Logs[i].Tags == nil {
					payload.Logs[i].Tags = make(map[string]string)
				}
				payload.Logs[i].Tags["endpoint_id"] = payload.Meta.EndpointID
				payload.Logs[i].Tags["agent_id"] = payload.Meta.AgentID
				payload.Logs[i].Tags["host_id"] = payload.Meta.HostID
				payload.Logs[i].Tags["hostname"] = payload.Meta.Hostname
				payload.Logs[i].Tags["job"] = payload.Meta.Tags["job"]
				for k, v := range payload.Meta.Tags {
					payload.Logs[i].Tags[k] = v
				}
			}
			all = append(all, payload.Logs...)
		}
		_ = gz.Close()
		_ = f.Close()
	}

	// Sort logs by timestamp descending
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp.After(all[j].Timestamp)
	})

	if len(all) > limit {
		all = all[:limit]
	}

	return all, nil
}
