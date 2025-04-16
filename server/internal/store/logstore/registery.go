package logstore

import (
	"context"
	"fmt"

	"github.com/aaronlmathis/gosight/server/internal/config"
	"github.com/aaronlmathis/gosight/server/internal/store/logstore/filestore.go"

	"github.com/aaronlmathis/gosight/shared/utils"
)

func InitLogStore(ctx context.Context, cfg *config.Config) (LogStore, error) {
	engine := cfg.LogStore.Engine
	if engine == "" {
		engine = "file"
	}
	utils.Debug("InitLogStore selected engine: %s", engine)
	switch engine {
	case "file":
		utils.Debug("Bootstrapping JSON File Store with %d workers", cfg.LogStore.Workers) // TODO give separate configs for workers
		s := filestore.NewFileStore(ctx, cfg)
		utils.Debug("Returning JSON Filestore at: %p", s)
		return s, nil
	default:
		return nil, fmt.Errorf("unsupported storage engine: %s", engine)
	}
}
