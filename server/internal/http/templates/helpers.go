package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
)

func FormatUptime(seconds float64) string {
	s := int64(seconds)
	days := s / 86400
	hours := (s % 86400) / 3600
	minutes := (s % 3600) / 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

func HumanizeBytes(b float64) string {
	const KB = 1024
	const MB = KB * 1024
	const GB = MB * 1024
	switch {
	case b > GB:
		return fmt.Sprintf("%.1f GB", b/GB)
	case b > MB:
		return fmt.Sprintf("%.1f MB", b/MB)
	case b > KB:
		return fmt.Sprintf("%.1f KB", b/KB)
	default:
		return fmt.Sprintf("%.0f B", b)
	}
}
func Marshal(v interface{}) template.JS {
	data, err := json.Marshal(v)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(data)
}
