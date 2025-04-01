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

// gosight/shared/utils
// log.go - Simple logging utility for the agent

/*
Example Usage

utils.Info("gRPC server is up and running on %s", cfg.ListenAddr)

utils.Error("Could not create default config: %v", err)
os.Exit(1)
*/

package utils

import (
	"io"
	"log"
	"os"
	"strings"
)

var (
	debugEnabled = false

	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
)

func InitLogger(logFile string, level string) error {
	var output io.Writer = os.Stdout

	// Try to open file if provided
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		output = f
	}

	infoLog = log.New(output, "[INFO] ", log.LstdFlags)
	warnLog = log.New(output, "[WARN] ", log.LstdFlags)
	errorLog = log.New(output, "[ERROR] ", log.LstdFlags)
	debugLog = log.New(output, "[DEBUG] ", log.LstdFlags)

	// Enable debug mode if requested
	if strings.ToLower(level) == "debug" {
		debugEnabled = true
	}

	return nil
}

func Info(format string, args ...any) {
	infoLog.Printf(format, args...)
}

func Warn(format string, args ...any) {
	warnLog.Printf(format, args...)
}

func Error(format string, args ...any) {
	errorLog.Printf(format, args...)
}

func Fatal(format string, args ...any) {
	errorLog.Fatalf(format, args...)
	os.Exit(1)
}

func Debug(format string, args ...any) {
	if debugEnabled {
		debugLog.Printf(format, args...)
	}
}