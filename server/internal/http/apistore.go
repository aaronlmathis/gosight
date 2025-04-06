//gosight/server/internal/http/apistore.go

package httpserver

import (
	"time"

	"github.com/aaronlmathis/gosight/server/internal/store"
	"github.com/aaronlmathis/gosight/shared/model"
	"github.com/aaronlmathis/gosight/shared/utils"
)

type APIMetricStore struct {
	Store store.MetricStore
}

func (a *APIMetricStore) Write(metrics []model.MetricPayload) error {
	return a.Store.Write(metrics)
}

func (a *APIMetricStore) Close() error {
	return a.Store.Close()
}

func (a *APIMetricStore) QueryInstant(metric string, filters map[string]string) ([]model.MetricRow, error) {
	utils.Debug("🔍 QueryInstant: %s", metric)
	return a.Store.QueryInstant(metric, filters)
}

func (a *APIMetricStore) QueryRange(metric string, start, end time.Time, filters map[string]string) ([]model.Point, error) {
	utils.Debug("📈 QueryRange: %s (%s - %s)", metric, start.Format(time.RFC3339), end.Format(time.RFC3339))
	return a.Store.QueryRange(metric, start, end, filters)
}
