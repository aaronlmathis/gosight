package httpserver

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

// HandleEventsAPI handles api calls to /api/events.
// It returns a list of events based on the provided filters.
// The filters are:
// - level: the level of the event (e.g., "info", "error")
// - type: the type of the event (e.g., "alert", "notification")
// - category: the category of the event (e.g., "system", "application")
// - source: the source of the event (e.g., "agent", "server")
// - contains: a string that must be contained in the event message
// - scope: the scope of the event (e.g., "global", "local")
// - target: the target of the event (e.g., "user", "system")
// - start: the start timestamp for the event (RFC3339 format)
// - end: the end timestamp for the event (RFC3339 format)
// - limit: the maximum number of events to return (default: 100)
// The response is a JSON array of event entries.
// If no filters are provided, all events are returned.

func (s *HttpServer) HandleEventsAPI(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	utils.Debug("Received request for events API with query: %s", q.Encode())

	filter := model.EventFilter{}

	if limit := q.Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			filter.Limit = n
		}
	} else {
		filter.Limit = 100
	}
	if v := q.Get("level"); v != "" {
		filter.Level = v
	}
	if v := q.Get("type"); v != "" {
		filter.Type = v
	}
	if v := q.Get("category"); v != "" {
		filter.Category = v
	}
	if v := q.Get("scope"); v != "" {
		filter.Scope = v
	}
	if v := q.Get("target"); v != "" {
		filter.Target = v
	}
	if v := q.Get("source"); v != "" {
		filter.Source = v
	}
	if v := q.Get("contains"); v != "" {
		filter.Contains = v
	}
	if v := q.Get("start"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filter.Start = &t
		}
	}
	if v := q.Get("hostID"); v != "" {
		filter.HostID = v
	}

	if v := q.Get("endpointID"); v != "" {
		filter.EndpointID = v
	}
	if v := q.Get("end"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			filter.End = &t
		}
	}
	if v := q.Get("sort"); v != "" {
		filter.SortOrder = v
	}
	utils.Debug("Calling GetRecentEvents with filter: %+v", filter)
	results, err := s.Sys.Stores.Events.GetRecentEvents(r.Context(), filter)
	if err != nil {
		http.Error(w, "Failed to query events", http.StatusInternalServerError)
		utils.Error("Error querying events: %v", err)
		return
	}
	if filter.SortOrder == "asc" {
		// Sort by ascending Timestamp
		sort.Slice(results, func(i, j int) bool {
			return results[i].Timestamp.Before(results[j].Timestamp)
		})
	} else if filter.SortOrder == "desc" {
		// Sort by descending Timestamp
		sort.Slice(results, func(i, j int) bool {
			return results[i].Timestamp.After(results[j].Timestamp)
		})
	}

	if results == nil {
		results = []model.EventEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}
