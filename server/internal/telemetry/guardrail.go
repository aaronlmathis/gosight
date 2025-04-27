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

// Provide safety and stability while accepting untrusted, variable, or user generated data.
// server/internal/telemetry/guardrail.go

package telemetry

import (
	"runtime/debug"

	"github.com/aaronlmathis/gosight/shared/utils"
)

// SafeHandlePayload wraps a handler function to recover from any panics that occur during its execution.
func SafeHandlePayload(handler func()) {
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			utils.Error("Panic recovered in payload handler: %v\n%s", r, string(stack))
		}
	}()
	handler()
}
