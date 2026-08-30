package reports

import (
	"sort"
)

func iterateSortedResultMap(data map[string][]string) ResultMapIterator {
	sortedKeys := sortKeys(data)
	return func(yield func(string, []string) bool) {
		for _, key := range sortedKeys {
			if !yield(key, data[key]) {
				return
			}
		}
	}
}

func sortKeys(data map[string][]string) []string {
	keys := make([]string, len(data))
	i := 0
	for key := range data {
		keys[i] = key
		i++
	}
	sort.Strings(keys)
	return keys
}
