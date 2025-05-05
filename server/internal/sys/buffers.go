package sys

import "github.com/aaronlmathis/gosight/server/internal/bufferengine"

type BufferModule struct {
	Metrics bufferengine.BufferedStore
	Logs    bufferengine.BufferedStore
	Data    bufferengine.BufferedStore
	Events  bufferengine.BufferedStore
	Alerts  bufferengine.BufferedStore
}
