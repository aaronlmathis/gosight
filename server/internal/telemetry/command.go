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

// internal/grpc/command.go

package telemetry

import (
	"context"

	"github.com/aaronlmathis/gosight/shared/proto"
)

func (h *StreamHandler) EnqueueCommandToAgent(ctx context.Context, endpointID string, commandType, command string, args []string) error {
	cmdReq := &proto.CommandRequest{
		EndpointId:  endpointID,
		CommandType: commandType,
		Command:     command,
		Args:        args,
	}
	h.Sys.Tracker.EnqueueCommand(endpointID, cmdReq)
	return nil
}
