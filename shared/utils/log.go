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
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var (
	logger       zerolog.Logger
	debugEnabled bool
)

func InitLogger(appLogFile, errorLogFile, accessLogFile, logLevel string) error {

	// Set up the logger with the specified log files and log level
	var writers []io.Writer

	if appLogFile != "" {
		appOutput, err := os.OpenFile(appLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, appOutput)
	} else {
		writers = append(writers, os.Stdout)
	}

	if errorLogFile != "" && errorLogFile != appLogFile { // Avoid duplicate if error and app logs are the same file
		errorOutput, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, errorOutput)
	}

	if accessLogFile != "" {
		accessOutput, err := os.OpenFile(accessLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		writers = append(writers, accessOutput)
	}

	mw := io.MultiWriter(writers...)

	// For production, you would typically log to JSON instead of console output.
	// Remove the ConsoleWriter and use the MultiWriter directly with zerolog.New().
	consoleWriter := zerolog.ConsoleWriter{Out: mw, TimeFormat: "2006-01-02 15:04:05"}
	logger = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()

	// logger = zerolog.New(mw).With().Timestamp().Caller().Logger() // Production JSON output

	// Enable debug mode if requested
	if strings.ToLower(logLevel) == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		debugEnabled = true
	} else if strings.ToLower(logLevel) == "warn" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	} else if strings.ToLower(logLevel) == "error" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logger.Info().Msgf("Log system initialized. Debug enabled: %v, Log Level: %s", debugEnabled, zerolog.GlobalLevel().String())
	return nil
}

func Info(format string, args ...any) {
	logger.Info().Msgf(format, args...)
}

func Warn(format string, args ...any) {
	logger.Warn().Msgf(format, args...)
}

func Error(format string, args ...any) {
	logger.Error().Msgf(format, args...)
}

func Fatal(format string, args ...any) {
	logger.Fatal().Msgf(format, args...)
}

func Debug(format string, args ...any) {
	logger.Debug().Msgf(format, args...)
}

func Access(format string, args ...any) {
	logger.Info().Str("level", "access").Msgf(format, args...)
}

func Must(label string, err error) {
	if err != nil {
		logger.Fatal().Err(err).Msgf("%s init failed", label)
	}
}
