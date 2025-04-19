package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// LogQueryParams represents the parameters for querying logs.
// It includes the limit of logs to return, the log levels to filter by,
// the unit of the logs, the source of the logs, a string to search for in the logs,
// and the start and end times for the logs.
type LogQueryParams struct {
	Limit    int
	Levels   map[string]bool
	Unit     string
	Source   string
	Contains string
	Start    *time.Time
	End      *time.Time
}

// HandleRecentLogs handles the HTTP request for recent logs.
// It retrieves the logs from the log store, applies any filters specified
// in the query parameters, and returns the logs as a JSON response.
// The limit for the number of logs returned can be specified in the query
// parameters, with a maximum of 1000 logs. If the limit is not specified,
// it defaults to 100 logs. The function also handles errors and returns
// appropriate HTTP status codes and messages.

func (s *HttpServer) HandleRecentLogs(w http.ResponseWriter, r *http.Request) {
	limit := 100

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			if parsed > 1000 {
				parsed = 1000
			}
			limit = parsed
		}
	}

	logs, err := s.LogStore.GetRecentLogs(limit)
	if err != nil {
		utils.Error("Failed to load logs: %v", err)
		http.Error(w, "failed to load logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(logs)

}

// HandleLogAPI handles the HTTP request for the log API.
// It retrieves the logs from the log store, applies any filters specified
// in the query parameters, and returns the logs as a JSON response.
// The function uses the LogQueryParams struct to parse the query parameters
// and filter the logs. It handles errors and returns appropriate HTTP status
// codes and messages. The logs are filtered based on the specified levels,
// unit, source, contains string, and start and end times. The function
// limits the number of logs returned to the specified limit in the query
// parameters, with a maximum of 1000 logs. If the limit is not specified,
// it defaults to 100 logs. The function also handles errors and returns
// appropriate HTTP status codes and messages.

func (s *HttpServer) HandleLogAPI(w http.ResponseWriter, r *http.Request) {
	params := parseLogQueryParams(r)

	all, err := s.LogStore.GetRecentLogs(1000) // load enough to filter
	if err != nil {
		http.Error(w, "failed to load logs", http.StatusInternalServerError)
		return
	}

	var filtered []model.LogEntry
	for _, log := range all {
		if strings.ToLower(log.Source) == "podman" && strings.ToLower(log.Level) == "debug" {
			continue // 🔕 skip noisy Podman debug
		}
		if len(filtered) >= params.Limit {
			break
		}
		if len(params.Levels) > 0 && !params.Levels[strings.ToLower(log.Level)] {
			continue
		}
		if params.Unit != "" && log.Category != params.Unit {
			continue
		}
		if params.Source != "" && log.Source != params.Source {
			continue
		}
		if params.Contains != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(params.Contains)) {
			continue
		}
		if params.Start != nil && log.Timestamp.Before(*params.Start) {
			continue
		}
		if params.End != nil && log.Timestamp.After(*params.End) {
			continue
		}
		filtered = append(filtered, log)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(filtered)
}

// parseLogQueryParams parses the query parameters from the HTTP request
// and returns a LogQueryParams struct. It handles the limit, levels, unit,
// source, contains, start, and end parameters. The limit is capped at 1000.
// The levels are stored in a map for quick lookup. The start and end times
// are parsed as RFC3339 formatted strings and returned as pointers to time.Time.
// If a parameter is not provided or invalid, it is ignored.
// The function also trims whitespace and converts levels to lowercase.

func parseLogQueryParams(r *http.Request) LogQueryParams {
	q := r.URL.Query()
	limit := 100
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		if l > 1000 {
			limit = 1000
		} else {
			limit = l
		}
	}

	levels := make(map[string]bool)
	for _, lvl := range strings.Split(q.Get("level"), ",") {
		if lvl != "" {
			levels[strings.ToLower(strings.TrimSpace(lvl))] = true
		}
	}

	var start, end *time.Time
	if s := q.Get("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			start = &t
		}
	}
	if s := q.Get("end"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			end = &t
		}
	}

	return LogQueryParams{
		Limit:    limit,
		Levels:   levels,
		Unit:     q.Get("unit"),
		Source:   q.Get("source"),
		Contains: q.Get("contains"),
		Start:    start,
		End:      end,
	}
}
