package common

import (
	"math"
	"time"
)

func RoundTo(v float64, decimals int) float64 {
	if decimals < 0 {
		decimals = 0
	}
	scale := math.Pow10(decimals)
	return math.Round(v*scale) / scale
}

func StringPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TimePtrIfNonZero lets optional timestamps serialize as `null` rather than
// `"0001-01-01T00:00:00Z"`.
func TimePtrIfNonZero(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
