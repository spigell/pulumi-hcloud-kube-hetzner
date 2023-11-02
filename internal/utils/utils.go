package utils

import (
	"sort"
)

func SortedMapKeys[V any](unsorted map[string]V) []string {
	keys := make([]string, 0, len(unsorted))

	for k := range unsorted {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
