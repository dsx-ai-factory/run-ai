package utils

import "sort"

// SortedKeys returns slice of sorted (string) keys of a map
// TODO: use slices.Sort (and make this function more generic with K any instated of string as key)
// once the run-ai operator moves to go1.21 and the performance is fix  https://github.com/golang/go/issues/61725
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Deref returns the value the pointer ref to or zero value
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// Ref returns a reference of a variable
func Ref[T any](t T) *T {
	return &t
}
