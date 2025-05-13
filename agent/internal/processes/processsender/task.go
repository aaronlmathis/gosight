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
// Package model contains the data structures used in GoSight.
// agent/processes/processsender/task.go

package processsender

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StartWorkerPool starts a pool of workers to process incoming process payloads.
// Each worker will attempt to send the payload to the gRPC server.
// The number of workers is determined by the workerCount parameter.
// The workers will run until the context is done or an error occurs.
// The function uses a goroutine for each worker, allowing them to run concurrently.
func (s *ProcessSender) StartWorkerPool(_ context.Context, queue <-chan *model.ProcessPayload, workerCount int) {
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go func(id int) {
			defer s.wg.Done()
			for {
				select {
				case <-s.streamCtx.Done():
					utils.Info("Process worker %d shutting down", id)
					return
				case payload := <-queue:
					if err := s.trySendWithBackoff(payload); err != nil {
						utils.Error("Process worker %d failed to send payload: %v", id, err)
					}
				}
			}
		}(i + 1)
	}
}

// trySendWithBackoff attempts to send the process payload to the gRPC server.
// It uses exponential backoff for retries in case of transient errors.
// The function will retry sending the payload up to 5 times with increasing backoff times.
// If the send fails after 5 attempts, it returns an error.
func (s *ProcessSender) trySendWithBackoff(payload *model.ProcessPayload) error {
	var err error
	backoff := 500 * time.Millisecond
	maxBackoff := 10 * time.Second

	for attempt := 1; attempt <= 5; attempt++ {
		err = s.SendSnapshot(payload)
		if err == nil {
			return nil
		}

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
				utils.Warn("Transient process send error (%s) — retrying in %v [attempt %d/5]", st.Code(), backoff, attempt)
			default:
				utils.Error("Permanent process send error (%s): %v", st.Code(), err)
				return err
			}
		} else {
			utils.Warn("Unknown process send error — retrying in %v [attempt %d/5]: %v", backoff, attempt, err)
		}

		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
	return fmt.Errorf("process send failed after 5 attempts: %w", err)
}
