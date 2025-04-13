package templates

import "fmt"

func FormatUptime(seconds float64) string {
	s := int64(seconds)
	days := s / 86400
	hours := (s % 86400) / 3600
	minutes := (s % 3600) / 60

	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}
