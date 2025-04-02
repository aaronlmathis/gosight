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
along with GoBright. If not, see https://www.gnu.org/licenses/.
*/

// gosight/agent/internal/sender/task.go
//

package sender

import (
	"context"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// StartWorkerPool launches N workers and processes metric payloads with retries
func StartWorkerPool(ctx context.Context, s *Sender, queue <-chan model.MetricPayload, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for {
				select {
				case <-ctx.Done():
					utils.Info("Worker %d shutting down", id)
					return
				case payload := <-queue:
					if err := trySendWithBackoff(s, payload); err != nil {
						utils.Error("Worker %d failed to send payload: %v", id, err)
					}
				}
			}
		}(i + 1)
	}
}

func trySendWithBackoff(s *Sender, payload model.MetricPayload) error {
	var err error
	backoff := 500 * time.Millisecond
	maxBackoff := 10 * time.Second

	for retries := 0; retries < 5; retries++ {
		err = s.SendMetrics(payload)
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
