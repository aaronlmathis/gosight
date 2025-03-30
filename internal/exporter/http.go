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

package exporter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/internal/collector"
	"github.com/aaronlmathis/gosight/internal/config"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

// WSMessage defines the JSON structure sent via WebSocket
type WSMessage struct {
	Metrics    map[string]float64          `json:"metrics"`
	Thresholds map[string]config.Threshold `json:"thresholds"`
	Meta       map[string]string           `json:"meta"`
}

func StartHTTPServer(addr string, store *collector.MetricStore, thresholds map[string]config.Threshold, enablePrometheus bool, enableDashboard bool) {
	mux := http.NewServeMux()

	// JSON endpoint
	mux.HandleFunc("/metrics/json", func(w http.ResponseWriter, r *http.Request) {
		metrics, meta := store.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"metrics": metrics,
			"meta":    meta,
		})
	})

	if enablePrometheus {
		// Prometheus text endpoint
		mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
			metrics, _ := store.Snapshot()
			w.Header().Set("Content-Type", "text/plain")
			keys := make([]string, 0, len(metrics))
			for k := range metrics {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				name := sanitizePrometheusKey(k)
				fmt.Fprintf(w, "%s %f\n", name, metrics[k])
			}
		})
	}

	if enableDashboard {
		// WebSocket endpoint
		mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("WebSocket upgrade error:", err)
				return
			}
			defer conn.Close()

			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for {
				<-ticker.C
				metrics, meta := store.Snapshot()
				msg := WSMessage{
					Metrics:    metrics,
					Thresholds: thresholds,
					Meta:       meta,
				}
				log.Printf("WS sending metrics: %v", metrics)
				log.Printf("WS sending meta: %v", meta)
				if err := conn.WriteJSON(msg); err != nil {
					log.Println("WebSocket write error:", err)
					break
				}
			}
		})

		// HTML dashboard endpoint
		mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			metrics, _ := store.Snapshot()
			tmplPath := filepath.Join("web", "templates", "dashboard.html")
			tmpl := template.Must(template.ParseFiles(tmplPath))
			err := tmpl.Execute(w, metrics)
			if err != nil {
				http.Error(w, "Template error", http.StatusInternalServerError)
			}
		})
	}

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// Prometheus metrics must match regex [a-zA-Z_:][a-zA-Z0-9_:]*
func sanitizePrometheusKey(k string) string {
	k = strings.ReplaceAll(k, "/", "_")
	k = strings.ReplaceAll(k, "-", "_")
	k = strings.ReplaceAll(k, ".", "_")
	return k
}
