package logstore

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/cache"
	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore/filestore.go"
	victorialogstore "github.com/aaronlmathis/gosight/server/internal/store/logstore/victoriametrics"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitLogStore(ctx context.Context, cfg *config.Config, logCache cache.LogCache) (LogStore, error) {
	engine := cfg.LogStore.Engine
	if engine == "" {
		engine = "file"
	}
	utils.Debug("InitLogStore selected engine: %s", engine)
	switch engine {
	case "file":
		utils.Debug("Bootstrapping JSON File LogStore.")
		s := filestore.New(cfg.LogStore.Dir)
		utils.Debug("Returning JSON Filestore at: %p", s)
		return s, nil
	case "victoriametrics":
		utils.Debug("Bootstrapping VictoriaMetrics LogStore.")
		s, err := victorialogstore.NewVictoriaLogStore(cfg.LogStore.Url, cfg.LogStore.Dir, logCache)
		if err != nil {
			return nil, fmt.Errorf("failed to create VictoriaMetrics LogStore: %w", err)
		}
		utils.Debug("Returning VictoriaMetric LogStore at: %p", s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", engine)
	}
}
