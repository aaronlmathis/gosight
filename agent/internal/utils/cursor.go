package agentutils

import (
	"os"
	"strings"
)

// LoadCursor reads the last saved journald cursor from a file.
// It returns an empty string and nil error if the file does not exist.
func LoadCursor(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist, return empty cursor and no error
		}
		return "", err // Other errors should be returned
	}
	return strings.TrimSpace(string(data)), nil
}

// SaveCursor writes the given journald cursor to a file.
func SaveCursor(path, cursor string) error {
	return os.WriteFile(path, []byte(cursor), 0644)
}
