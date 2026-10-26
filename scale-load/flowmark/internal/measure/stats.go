/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"math"
	"sort"
	"time"
)

type Summary struct {
	SampleSize        int
	NeverArrivedCount int
	FailedCount       int
	P50               time.Duration
	P90               time.Duration
	P99               time.Duration
	Makespan          time.Duration
	Throughput        float64
}

func Stats(timings []Timing) Summary {
	var summary Summary
	var durations []time.Duration
	var first, last time.Time

	for _, timing := range timings {
		switch timing.Outcome {
		case Arrived:
			durations = append(durations, timing.Duration)
			if first.IsZero() || timing.Created.Before(first) {
				first = timing.Created
			}
			if timing.Arrived.After(last) {
				last = timing.Arrived
			}
		case NeverArrived:
			summary.NeverArrivedCount++
		case Failed:
			summary.FailedCount++
		}
	}

	summary.SampleSize = len(durations)
	if summary.SampleSize == 0 {
		return summary
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	summary.P50 = percentile(durations, 50)
	summary.P90 = percentile(durations, 90)
	summary.P99 = percentile(durations, 99)

	summary.Makespan = max(last.Sub(first), 0)
	if summary.HasThroughput() {
		summary.Throughput = float64(summary.SampleSize) / summary.Makespan.Seconds()
	}
	return summary
}

func (s Summary) HasThroughput() bool {
	return s.SampleSize >= 2 && s.Makespan > 0
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	rank := int(math.Ceil(p / 100 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	return sorted[rank-1]
}
