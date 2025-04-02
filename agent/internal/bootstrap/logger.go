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

// server/internal/bootstrap/logger.go
// Initializes logger.
package bootstrap

import (
	"fmt"
	"os"

	"github.com/aaronlmathis/gosight/agent/internal/config"
	"github.com/aaronlmathis/gosight/shared/utils"
)

func SetupLogging(cfg *config.AgentConfig) {
	if err := utils.InitLogger(cfg.LogFile, cfg.LogLevel); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

}
