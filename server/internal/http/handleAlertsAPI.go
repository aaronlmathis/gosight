package httpserver

import (
	"encoding/json"
	"net/http"
)

func (s *HttpServer) HandleAlertsAPI(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	alerts := s.AlertsMgr.ListActive()
	if err := json.NewEncoder(w).Encode(alerts); err != nil {
		http.Error(w, "failed to encode alerts", http.StatusInternalServerError)
		return
	}
}
