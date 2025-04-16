package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func (s *HttpServer) HandleRecentLogs(w http.ResponseWriter, r *http.Request) {

	logs, err := s.LogStore.GetRecentLogs(100)
	if err != nil {
		http.Error(w, "Failed to load logs", http.StatusInternalServerError)

	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ") // <== this is the magic
	err = enc.Encode(logs)
	if err != nil {
		utils.Error("❌ Failed to write JSON response: %v", err)
	}

}
