package utils

import (
	"cmp"
	"slices"
)

// SortOrder represents the sorting direction
type SortOrder int

const (
	Ascending SortOrder = iota
	Descending
)

// SortColumn defines a column to sort by and its order
type SortColumn struct {
	Index int
	Order SortOrder
}

// MultiColumnSort sorts a 2D slice by multiple columns using modern slices.SortFunc
func MultiColumnSort(data [][]string, columns []SortColumn) {
	slices.SortFunc(data, func(a, b []string) int {
		for _, col := range columns {
			// Bounds check
			if col.Index >= len(a) || col.Index >= len(b) {
				continue
			}

			valA := a[col.Index]
			valB := b[col.Index]

			result := cmp.Compare(valA, valB)
			if result != 0 {
				if col.Order == Descending {
					return -result
				}
				return result
			}
			// If values are equal, continue to next column
		}
		return 0 // All compared columns are equal
	})
}
