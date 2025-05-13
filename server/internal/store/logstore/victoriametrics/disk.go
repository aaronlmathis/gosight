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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// server/internal/store/logstore/victoriametrics/disk.go

package victorialogstore

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
)

// writeCompressedWrappedLogs writes StoredLogs to disk
func (v *VictoriaLogStore) writeCompressedWrappedLogs(logs []*model.StoredLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Use first timestamp to determine file path
	t := logs[0].Log.Timestamp
	if t.IsZero() {
		t = time.Now()
	}

	dir := filepath.Join(v.logsPath, "logs", t.Format("2006"), t.Format("01"), t.Format("02"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	hour := t.Format("15")
	endpointID := logs[0].Meta.EndpointID
	if endpointID == "" {
		endpointID = "unknown"
	}
	filename := fmt.Sprintf("%s_%s.json.gz", endpointID, hour)
	fullPath := filepath.Join(dir, filename)

	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	enc := json.NewEncoder(gz)
	for _, entry := range logs {
		if err := enc.Encode(entry); err != nil {
			return fmt.Errorf("write log: %w", err)
		}
	}

	return nil
}
