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
package metricsender

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/aaronlmathis/gosight/shared/model"
)

// AppendMetricsToFile appends the given MetricPayload to a file in JSON format.
// It creates the directory if it doesn't exist.
// The file is opened in append mode, and the payload is written as a newline-delimited JSON object.
// This is useful for debugging purposes, allowing you to inspect the metrics being sent.
func AppendMetricsToFile(payload *model.MetricPayload, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n')) // newline-delimited JSON
	return err
}
