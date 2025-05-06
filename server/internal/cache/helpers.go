package cache

import "strings"

type StringSet map[string]struct{}

func addToSet(m map[string]StringSet, key, val string) {
	if m[key] == nil {
		m[key] = make(StringSet)
	}
	m[key][val] = struct{}{}
}

func containsMatch(value, substr string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(substr))
}
