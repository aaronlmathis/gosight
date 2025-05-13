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

// gosight/agent/internal/sender/task.go
//

package metricsender

import (
	"context"
	"fmt"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StartWorkerPool launches N workers and processes metric payloads with retries
// in case of transient errors. Each worker will attempt to send the payload
// to the gRPC server. The number of workers is determined by the workerCount
// parameter. The workers will run until the context is done or an error occurs.
// The function uses a goroutine for each worker, allowing them to run concurrently.
func (s *MetricSender) StartWorkerPool(ctx context.Context, queue <-chan *model.MetricPayload, workerCount int) {
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

// trySendWithBackoff attempts to send the metric payload to the gRPC server.
// It uses exponential backoff for retries in case of transient errors.
// The function will retry sending the payload up to 5 times with increasing
// backoff times. If the payload is successfully sent, it returns nil.
// If the send fails after 5 attempts, it returns an error.
func (s *MetricSender) trySendWithBackoff(payload *model.MetricPayload) error {
	var err error
	backoff := 500 * time.Millisecond
	maxBackoff := 10 * time.Second

	for attempt := 1; attempt <= 5; attempt++ {
		err = s.SendMetrics(payload)
		if err == nil {
			return nil
		}

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
				utils.Warn("Transient error (%s) — retrying in %v [attempt %d/5]", st.Code(), backoff, attempt)
			default:
				utils.Error("Permanent send error (%s): %v", st.Code(), err)
				return err // Do not retry permanent errors
			}
		} else {
			utils.Warn("Unknown error — retrying in %v [attempt %d/5]: %v", backoff, attempt, err)
		}

		time.Sleep(backoff)
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}

	return fmt.Errorf("send failed after 5 attempts: %w", err)
}
