package httpserver

import (
	"net/http"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type HostsHandler struct {
	Store store.MetricStore
}

func (h *HostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	rows, err := h.Store.QueryInstant("system.host.uptime", nil)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to query hosts"})
		return
	}
	seen := make(map[string]bool)
	endpoints := []map[string]string{}
	for _, row := range rows {
		hostname := row.Tags["hostname"]
		ip := row.Tags["ip_address"]
		// Use hostname+ip as a unique key
		key := hostname + "|" + ip
		if seen[key] {
			continue
		}
		seen[key] = true
		endpoints = append(endpoints, map[string]string{
			"hostname": row.Tags["hostname"],
			"ip":       row.Tags["ip_address"],
			"os":       row.Tags["os"],
			"arch":     row.Tags["arch"],
			"status":   "online", // Future: derive from agent heartbeat
		})
	}

	utils.JSON(w, http.StatusOK, endpoints)
}
