/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatBucketName(t *testing.T) {
	tests := []struct {
		name       string
		minSeconds int64
		maxSeconds int64
		expected   string
	}{
		// Seconds
		{
			name:       "seconds range",
			minSeconds: 0,
			maxSeconds: 30,
			expected:   "(0)-(30s)",
		},
		{
			name:       "seconds to minutes",
			minSeconds: 30,
			maxSeconds: 120,
			expected:   "(30s)-(2m)",
		},
		// Minutes
		{
			name:       "minutes range",
			minSeconds: 60,
			maxSeconds: 300,
			expected:   "(1m)-(5m)",
		},
		{
			name:       "zero to minutes",
			minSeconds: 0,
			maxSeconds: 900,
			expected:   "(0)-(15m)",
		},
		{
			name:       "minutes to hours",
			minSeconds: 1800,
			maxSeconds: 7200,
			expected:   "(30m)-(2h)",
		},
		// Hours
		{
			name:       "hours range",
			minSeconds: 3600,
			maxSeconds: 7200,
			expected:   "(1h)-(2h)",
		},
		{
			name:       "hours range larger",
			minSeconds: 21600,
			maxSeconds: 43200,
			expected:   "(6h)-(12h)",
		},
		{
			name:       "hours to days",
			minSeconds: 43200,
			maxSeconds: 172800,
			expected:   "(12h)-(2d)",
		},
		// Days
		{
			name:       "days range",
			minSeconds: 86400,
			maxSeconds: 172800,
			expected:   "(1d)-(2d)",
		},
		{
			name:       "days with hours",
			minSeconds: 86400,
			maxSeconds: 194400, // 2d6h
			expected:   "(1d)-(2d-6h)",
		},
		{
			name:       "days range larger",
			minSeconds: 259200, // 3d
			maxSeconds: 432000, // 5d
			expected:   "(3d)-(5d)",
		},
		// Open-ended (maxSeconds < 0)
		{
			name:       "open-ended with -1",
			minSeconds: 2700,
			maxSeconds: -1,
			expected:   "(45m)+",
		},
		{
			name:       "open-ended hours with -1",
			minSeconds: 64800, // 18h
			maxSeconds: -1,
			expected:   "(18h)+",
		},
		{
			name:       "open-ended days with -1",
			minSeconds: 172800, // 2d
			maxSeconds: -1,
			expected:   "(2d)+",
		},
		{
			name:       "open-ended days with hours with -1",
			minSeconds: 194400, // 2d6h
			maxSeconds: -1,
			expected:   "(2d-6h)+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBucketName(tt.minSeconds, tt.maxSeconds)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePendingTimeBuckets(t *testing.T) {
	t.Run("zero max pending time returns default buckets", func(t *testing.T) {
		buckets := CalculatePendingTimeBuckets(0)

		assert.Len(t, buckets, 4, "should return 4 buckets")

		// Check first bucket
		assert.Equal(t, int64(0), buckets[0].MinSeconds)
		assert.Equal(t, int64(900), buckets[0].MaxSeconds)
		assert.Equal(t, "(0)-(15m)", buckets[0].Name)

		// Check second bucket
		assert.Equal(t, int64(900), buckets[1].MinSeconds)
		assert.Equal(t, int64(1800), buckets[1].MaxSeconds)
		assert.Equal(t, "(15m)-(30m)", buckets[1].Name)

		// Check third bucket
		assert.Equal(t, int64(1800), buckets[2].MinSeconds)
		assert.Equal(t, int64(2700), buckets[2].MaxSeconds)
		assert.Equal(t, "(30m)-(45m)", buckets[2].Name)

		// Check fourth bucket (open-ended)
		assert.Equal(t, int64(2700), buckets[3].MinSeconds)
		assert.Equal(t, int64(-1), buckets[3].MaxSeconds)
		assert.Equal(t, "(45m)+", buckets[3].Name)
	})

	t.Run("non-zero max pending time creates 4 equal buckets", func(t *testing.T) {
		maxTime := int64(3600) // 1 hour
		buckets := CalculatePendingTimeBuckets(maxTime)

		assert.Len(t, buckets, 4, "should return 4 buckets")

		expectedBucketSize := (maxTime + 3) / 4 // 900 seconds (15 minutes)

		// Check first three buckets have equal size
		for i := 0; i < 3; i++ {
			assert.Equal(t, int64(i)*expectedBucketSize, buckets[i].MinSeconds)
			assert.Equal(t, int64(i+1)*expectedBucketSize, buckets[i].MaxSeconds)
		}

		// Check last bucket is open-ended
		assert.Equal(t, int64(3)*expectedBucketSize, buckets[3].MinSeconds)
		assert.Equal(t, int64(-1), buckets[3].MaxSeconds)
		assert.Contains(t, buckets[3].Name, ")+", "last bucket should be open-ended")
	})

	t.Run("large max pending time creates appropriate buckets", func(t *testing.T) {
		maxTime := int64(86400 * 7) // 7 days
		buckets := CalculatePendingTimeBuckets(maxTime)

		assert.Len(t, buckets, 4, "should return 4 buckets")

		expectedBucketSize := (maxTime + 3) / 4 // ~1.75 days per bucket

		// Verify bucket ranges are correct
		assert.Equal(t, int64(0), buckets[0].MinSeconds)
		assert.Equal(t, expectedBucketSize, buckets[0].MaxSeconds)

		assert.Equal(t, expectedBucketSize, buckets[1].MinSeconds)
		assert.Equal(t, expectedBucketSize*2, buckets[1].MaxSeconds)

		assert.Equal(t, expectedBucketSize*2, buckets[2].MinSeconds)
		assert.Equal(t, expectedBucketSize*3, buckets[2].MaxSeconds)

		assert.Equal(t, expectedBucketSize*3, buckets[3].MinSeconds)
		assert.Equal(t, int64(-1), buckets[3].MaxSeconds)

		// Verify all buckets have names
		for i, bucket := range buckets {
			assert.NotEmpty(t, bucket.Name, "bucket %d should have a name", i)
		}
	})

	t.Run("bucket names are human-readable", func(t *testing.T) {
		maxTime := int64(43200) // 12 hours
		buckets := CalculatePendingTimeBuckets(maxTime)

		// Each bucket should be ~3 hours
		// Names should be in format like "(0)-(3h)", "(3h)-(6h)", "(6h)-(9h)", "(9h)+"
		for i, bucket := range buckets {
			assert.NotEmpty(t, bucket.Name, "bucket %d should have a name", i)
			if i == 3 {
				assert.Contains(t, bucket.Name, ")+", "last bucket should contain ')+'")
			} else {
				assert.Contains(t, bucket.Name, ")-(", "bucket %d should contain ')-(' (closing then opening parenthesis)", i)
			}
		}
	})

	t.Run("always returns exactly 4 buckets", func(t *testing.T) {
		testCases := []int64{0, 100, 1000, 10000, 100000, 1000000}

		for _, maxTime := range testCases {
			buckets := CalculatePendingTimeBuckets(maxTime)
			assert.Len(t, buckets, 4, "should always return 4 buckets for maxTime=%d", maxTime)
		}
	})

	t.Run("last bucket is always open-ended with -1", func(t *testing.T) {
		testCases := []int64{0, 3600, 86400, 604800} // 0, 1h, 1d, 1w

		for _, maxTime := range testCases {
			buckets := CalculatePendingTimeBuckets(maxTime)
			assert.Equal(t, int64(-1), buckets[3].MaxSeconds,
				"last bucket should have MaxSeconds=-1 for maxTime=%d", maxTime)
		}
	})
}
