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
// agent/internal/utils/cursor.go
package agentutils

import (
	"os"
	"strings"
)

// LoadCursor reads the last saved journald cursor from a file.
// It returns an empty string and nil error if the file does not exist.
func LoadCursor(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist, return empty cursor and no error
		}
		return "", err // Other errors should be returned
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveCursor writes the given journald cursor to a file.
func SaveCursor(path, cursor string) error {
	return os.WriteFile(path, []byte(cursor), 0644)
}
