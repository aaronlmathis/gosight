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
along with LeetScraper. If not, see https://www.gnu.org/licenses/.
*/

package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gosight/internal/shared"
)

func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	var payload shared.MetricPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	fmt.Printf("[%s] Received %d metrics from %s\n", time.Now().Format(time.RFC3339), len(payload.Metrics), payload.Host)
	// TODO: Store metrics in DB or memory

	w.WriteHeader(http.StatusAccepted)
}
