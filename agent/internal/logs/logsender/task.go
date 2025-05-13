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

// gosight/agent/internal/logsender/task.go
//

package logsender

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// StartWorkerPool launches N workers and processes metric payloads with retries
// in case of transient errors. Each worker will attempt to send the payload
// to the gRPC server. The number of workers is determined by the workerCount
// parameter. The workers will run until the context is done or an error occurs.
func (s *LogSender) StartWorkerPool(ctx context.Context, queue <-chan *model.LogPayload, workerCount int) {
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1) // track worker
		go func(id int) {
			defer s.wg.Done() // signal on exit
			for {
				select {
				case <-ctx.Done():
					utils.Info("Worker %d shutting down", id)
					return
				case payload := <-queue:
					if err := s.trySendWithBackoff(payload); err != nil {
						utils.Error("Worker %d failed to send payload: %v", id, err)
					}
				}
			}
		}(i + 1)
	}
}

// trySendWithBackoff attempts to send the log payload to the server with exponential backoff.
// It retries sending the payload up to 5 times with increasing wait times between attempts.
// If all attempts fail, it returns the last error encountered.
// The backoff starts at 500ms and doubles each time, up to a maximum of 10 seconds.
func (s *LogSender) trySendWithBackoff(payload *model.LogPayload) error {
	var err error
	backoff := 500 * time.Millisecond
	maxBackoff := 10 * time.Second

	for retries := 0; retries < 5; retries++ {
		err = s.SendLogs(payload)
		if err == nil {
			return nil
		}
		utils.Warn("Retrying in %v: %v", backoff, err)
		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return err
}
