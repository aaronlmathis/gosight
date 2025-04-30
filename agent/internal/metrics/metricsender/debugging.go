package metricsender

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/aaronlmathis/gosight/shared/model"
)

func AppendMetricsToFile(payload *model.MetricPayload, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n')) // newline-delimited JSON
	return err
}
