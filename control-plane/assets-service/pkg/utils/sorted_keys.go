package utils

import "sort"

// SortedMapKeys returns a sorted list of keys for any map with string kind of keys
func SortedMapKeys[T any](unsorted map[string]T) []string {
	result := make([]string, 0, len(unsorted))
	for k := range unsorted {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}
