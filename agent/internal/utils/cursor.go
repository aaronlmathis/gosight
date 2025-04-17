package agentutils

import (
	"os"
	"strings"
)

// LoadCursor reads the last saved journald cursor from a file.
func LoadCursor(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveCursor writes the given journald cursor to a file.
func SaveCursor(path, cursor string) error {
	return os.WriteFile(path, []byte(cursor), 0644)
}
