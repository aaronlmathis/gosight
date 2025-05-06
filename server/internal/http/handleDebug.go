package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/aaronlmathis/gosight/server/debugtools"
)

func (s *HttpServer) HandleCacheAudit(w http.ResponseWriter, r *http.Request) {

	report := debugtools.AuditCaches(s.Sys.Cache.Tags, s.Sys.Cache.Metrics)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}
