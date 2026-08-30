package services_utils

import (
	"fmt"
)

// Bucket represents a time range bucket for pending time distribution.
type Bucket struct {
	MinSeconds int64
	MaxSeconds int64
	Name       string // Human-readable name like "(0)-(15m)" or "(6h)-(12h)"
}

// CalculatePendingTimeBuckets calculates 4 bucket ranges based on P95 percentile pending time.
func CalculatePendingTimeBuckets(p95PendingTimeSeconds int64) []Bucket {
	if p95PendingTimeSeconds == 0 {
		return []Bucket{
			{MinSeconds: 0, MaxSeconds: 900, Name: "(0)-(15m)"},
			{MinSeconds: 900, MaxSeconds: 1800, Name: "(15m)-(30m)"},
			{MinSeconds: 1800, MaxSeconds: 2700, Name: "(30m)-(45m)"},
			{MinSeconds: 2700, MaxSeconds: -1, Name: "(45m)+"},
		}
	}

	bucketSize := (p95PendingTimeSeconds + 3) / 4

	buckets := make([]Bucket, 4)
	for i := 0; i < 4; i++ {
		minSeconds := int64(i) * bucketSize
		var maxSeconds int64
		if i == 3 {
			// Last bucket is open-ended (no upper limit) to catch all remaining workloads, including outliers beyond P95
			maxSeconds = -1 // -1 means no upper limit
		} else {
			maxSeconds = int64(i+1) * bucketSize
		}

		buckets[i] = Bucket{
			MinSeconds: minSeconds,
			MaxSeconds: maxSeconds,
			Name:       FormatBucketName(minSeconds, maxSeconds),
		}
	}

	return buckets
}

// FormatBucketName returns a human-readable bucket name like "(0)-(15m)", "(1d-3h)-(2d-1h)", or "(45m)+"
// Format uses abbreviated units with parentheses: d=days, h=hours, m=minutes, s=seconds.
func FormatBucketName(minSeconds, maxSeconds int64) string {
	formatTime := func(seconds int64) string {
		if seconds == 0 {
			return "0"
		} else if seconds < 60 {
			// Less than 1 minute: show as seconds (e.g., "30s")
			return fmt.Sprintf("%ds", seconds)
		} else if seconds < 3600 {
			// Less than 1 hour: show as minutes (e.g., "15m")
			minutes := seconds / 60
			return fmt.Sprintf("%dm", minutes)
		} else if seconds < 86400 {
			// Less than 1 day: show as hours (e.g., "6h")
			hours := seconds / 3600
			return fmt.Sprintf("%dh", hours)
		} else {
			// 1 day or more: show as days with remaining hours (e.g., "1d-3h" for 1 day and 3 hours)
			days := seconds / 86400
			remainingHours := (seconds % 86400) / 3600
			if remainingHours > 0 {
				return fmt.Sprintf("%dd-%dh", days, remainingHours)
			}
			return fmt.Sprintf("%dd", days)
		}
	}

	if maxSeconds < 0 {
		// Last bucket has no upper bound, format as "(min)+" (e.g., "(45m)+")
		return fmt.Sprintf("(%s)+", formatTime(minSeconds))
	}

	// Regular bucket: format as "(min)-(max)" range (e.g., "(0)-(15m)", "(1d-3h)-(2d-1h)")
	return fmt.Sprintf("(%s)-(%s)", formatTime(minSeconds), formatTime(maxSeconds))
}
