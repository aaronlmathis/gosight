/*
SPDX-License-Identifier: GPL-3.0-or-later

Copyright (C) 2025 Aaron Mathis aaron.mathis@gmail.com

This file is part of GoSight.

GoSight is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

GoSight is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with GoSight. If not, see https://www.gnu.org/licenses/.
*/

// gosight/server/internal/store/logstore/filestore.go
// Defines filestore

package filestore

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type FileStore struct {
	dir string
}

func New(dir string) *FileStore {
	if err := os.MkdirAll(dir, 0755); err != nil {
		utils.Error("Failed to create log store directory: %v", err)
	}
	return &FileStore{dir: dir}
}

func (f *FileStore) Name() string {
	return "JSON FileStore"
}

func (f *FileStore) Write(payloads []model.LogPayload) error {
	for _, payload := range payloads {
		if len(payload.Logs) == 0 {
			continue
		}

		timestamp := time.Now().UTC().Format("20060102T150405Z")
		filename := fmt.Sprintf("logs_%s_%s.json.gz", payload.EndpointID, timestamp)
		path := filepath.Join(f.dir, filename)

		file, err := os.Create(path)
		if err != nil {
			utils.Error("Failed to create log file %s: %v", path, err)
			continue
		}

		gz := gzip.NewWriter(file)
		enc := json.NewEncoder(gz)
		enc.SetIndent("", "  ")

		if err := enc.Encode(payload); err != nil {
			utils.Error("Failed to encode log payload to file %s: %v", path, err)
		}

		_ = gz.Close()
		_ = file.Close()
	}
	return nil
}

func (f *FileStore) Close() error {
	// No-op for file store currently
	return nil
}

// HELPERS

func totalLogCount(payloads []model.LogPayload) int {
	count := 0
	for _, p := range payloads {
		count += len(p.Logs)
	}
	return count
}
