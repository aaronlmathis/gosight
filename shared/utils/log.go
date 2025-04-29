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

// Writers for different log types
var (
	infoWriter   io.Writer
	errorWriter  io.Writer
	accessWriter io.Writer
	debugWriter  io.Writer
)

func InitLogger(appLogFile, errorLogFile, accessLogFile, debugLogFile, logLevel string) error {
	// Default: discard everything (safe fallback)
	infoWriter = io.Discard
	errorWriter = io.Discard
	accessWriter = io.Discard
	debugWriter = io.Discard

	// Open files
	var appFile, errFile, accessFile, debugFile *os.File
	var err error

	if appLogFile != "" {
		appFile, err = os.OpenFile(appLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		infoWriter = appFile
	}

	if errorLogFile != "" {
		errFile, err = os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		errorWriter = errFile
	}

	if accessLogFile != "" {
		accessFile, err = os.OpenFile(accessLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		accessWriter = accessFile
	}

	if debugLogFile != "" {
		debugFile, err = os.OpenFile(debugLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		debugWriter = debugFile
	}

	// Enable debug mode?
	if strings.ToLower(logLevel) == "debug" {
		debugEnabled = true
		zerolog.SetGlobalLevel(zerolog.DebugLevel)

		// Pretty console output
		console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006-01-02 15:04:05"}

		infoWriter = io.MultiWriter(infoWriter, console)
		errorWriter = io.MultiWriter(errorWriter, console)
		accessWriter = io.MultiWriter(accessWriter, console)
		debugWriter = io.MultiWriter(debugWriter, console)

	} else {
		// Production levels
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Base logger writes nowhere directly; Info/Warn/Error/Debug will use split writers
	logger = zerolog.New(io.Discard).With().Timestamp().Caller().Logger()

	return nil
}

// INFO (normal server app messages)
func Info(format string, args ...any) {
	event := zerolog.New(infoWriter).With().Timestamp().Logger()
	event.Info().Msgf(format, args...)
}

// WARNING
func Warn(format string, args ...any) {
	event := zerolog.New(errorWriter).With().Timestamp().Logger()
	event.Warn().Msgf(format, args...)
}

// ERROR
func Error(format string, args ...any) {
	event := zerolog.New(errorWriter).With().Timestamp().Logger()
	event.Error().Msgf(format, args...)
}

// FATAL
func Fatal(format string, args ...any) {
	event := zerolog.New(errorWriter).With().Timestamp().Logger()
	event.Error().Msgf(format, args...)
	os.Exit(1)
}

// DEBUG
func Debug(format string, args ...any) {
	if debugEnabled {
		event := zerolog.New(debugWriter).With().Timestamp().Logger()
		event.Debug().Msgf(format, args...)
	}
}

// ACCESS LOGGING (special traffic / auth / API hits)
func Access(format string, args ...any) {
	event := zerolog.New(accessWriter).With().Timestamp().Str("level", "access").Logger()
	event.Info().Msgf(format, args...)
}

// MUST (fatal init failures)
func Must(label string, err error) {
	if err != nil {
		event := zerolog.New(errorWriter).With().Timestamp().Logger()
		event.Fatal().Err(err).Msgf("%s init failed", label)
	}
}
