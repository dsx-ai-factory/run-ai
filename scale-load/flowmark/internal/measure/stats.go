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
	SuccessRate       float64
	P50               time.Duration
	P90               time.Duration
	P99               time.Duration
	Mean              time.Duration
	Max               time.Duration
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
	if total := len(timings); total > 0 {
		summary.SuccessRate = float64(summary.SampleSize) / float64(total)
	}
	if summary.SampleSize == 0 {
		return summary
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	summary.P50 = percentile(durations, 50)
	summary.P90 = percentile(durations, 90)
	summary.P99 = percentile(durations, 99)
	summary.Max = durations[len(durations)-1]

	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	summary.Mean = sum / time.Duration(len(durations))

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

func (s Summary) Metrics() map[string]any {
	metrics := map[string]any{
		"sample_size":         s.SampleSize,
		"never_arrived_count": s.NeverArrivedCount,
		"failed_count":        s.FailedCount,
		"success_rate":        s.SuccessRate,
	}
	if s.SampleSize == 0 {
		return metrics
	}
	metrics["p50_ms"] = s.P50.Milliseconds()
	metrics["p90_ms"] = s.P90.Milliseconds()
	metrics["p99_ms"] = s.P99.Milliseconds()
	metrics["mean_ms"] = s.Mean.Milliseconds()
	metrics["max_ms"] = s.Max.Milliseconds()
	if !s.HasThroughput() {
		return metrics
	}
	metrics["makespan_ms"] = s.Makespan.Milliseconds()
	metrics["throughput_per_sec"] = s.Throughput
	return metrics
}
