package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// HandleEventsAPI returns the most recent events from the event store.
func (s *HttpServer) HandleEventsAPI(w http.ResponseWriter, r *http.Request) {
	// Default limit
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if parsed, err := strconv.Atoi(q); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	events := s.Sys.Stores.Events.GetRecent(limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, "failed to encode events", http.StatusInternalServerError)
		return
	}
}
