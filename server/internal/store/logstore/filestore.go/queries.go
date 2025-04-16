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

	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j] // newest first
	})

	var all []model.LogEntry
	for _, file := range files {
		if len(all) >= limit {
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
			all = append(all, payload.Logs...)
		}
		_ = gz.Close()
		_ = f.Close()
	}

	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
