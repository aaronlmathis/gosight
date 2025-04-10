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
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	debugEnabled = false

	infoLog   *log.Logger
	warnLog   *log.Logger
	errorLog  *log.Logger
	debugLog  *log.Logger
	accessLog *log.Logger
)

func InitLogger(appLogFile, errorLogFile, accessLogFile, logLevel string) error {
	var errorOutput io.Writer = os.Stdout
	var appOutput io.Writer = os.Stdout
	var accessOutput io.Writer = os.Stdout

	// Try to open file if provided
	if errorLogFile != "" {
		f, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		errorOutput = f
	}

	if appLogFile != "" {
		f, err := os.OpenFile(appLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		appOutput = f
	}

	if accessLogFile != "" {
		f, err := os.OpenFile(accessLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		accessOutput = f
	}

	infoLog = log.New(appOutput, "[INFO] ", log.LstdFlags)
	warnLog = log.New(errorOutput, "[WARN] ", log.LstdFlags)
	errorLog = log.New(errorOutput, "[ERROR] ", log.LstdFlags)
	accessLog = log.New(accessOutput, "[ACCESS] ", log.LstdFlags)

	multiDebugOutput := io.MultiWriter(os.Stdout, appOutput) // or appOutput if preferred
	debugLog = log.New(multiDebugOutput, "[DEBUG] ", log.LstdFlags)

	// Enable debug mode if requested
	if strings.ToLower(logLevel) == "debug" {
		debugEnabled = true
	}
	fmt.Printf("🧪 Log system initialized. Debug enabled: %v", debugEnabled)
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

}

func Debug(format string, args ...any) {
	if debugEnabled {
		debugLog.Printf(format, args...)
	}

}

func Access(format string, args ...any) {
	accessLog.Printf(format, args...)
}
