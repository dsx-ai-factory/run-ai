/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"fmt"
	"testing"
	"time"
)

func arrivals(base time.Time, durations ...time.Duration) []Timing {
	timings := make([]Timing, 0, len(durations))
	for i, d := range durations {
		created := base.Add(time.Duration(i) * time.Second)
		timings = append(timings, Timing{Name: fmt.Sprintf("w%d", i), Outcome: Arrived, Created: created, Arrived: created.Add(d), Duration: d})
	}
	return timings
}

func pendings(n int) []Timing {
	timings := make([]Timing, 0, n)
	for i := range n {
		timings = append(timings, Timing{Name: fmt.Sprintf("p%d", i), Outcome: Pending})
	}
	return timings
}

func seconds(values ...int) []time.Duration {
	out := make([]time.Duration, 0, len(values))
	for _, v := range values {
		out = append(out, time.Duration(v)*time.Second)
	}
	return out
}

func TestStatsPercentiles(t *testing.T) {
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		durations     []time.Duration
		p50, p90, p99 time.Duration
	}{
		{"odd count", seconds(1, 2, 3, 4, 5), 3 * time.Second, 5 * time.Second, 5 * time.Second},
		{"even count", seconds(1, 2, 3, 4), 2 * time.Second, 4 * time.Second, 4 * time.Second},
		{"single", seconds(7), 7 * time.Second, 7 * time.Second, 7 * time.Second},
		{"unsorted input", seconds(9, 1, 5, 3, 7, 2, 8, 4, 6, 10), 5 * time.Second, 9 * time.Second, 10 * time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Stats(arrivals(base, tc.durations...))
			if got.P50 != tc.p50 || got.P90 != tc.p90 || got.P99 != tc.p99 {
				t.Fatalf("p50/p90/p99 = %s/%s/%s, want %s/%s/%s", got.P50, got.P90, got.P99, tc.p50, tc.p90, tc.p99)
			}
			if got.SampleSize != len(tc.durations) || got.SuccessRate != 1 {
				t.Fatalf("sample_size %d success_rate %v", got.SampleSize, got.SuccessRate)
			}
		})
	}
}

func TestStatsPendingIsExcludedNotZero(t *testing.T) {
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	timings := append(arrivals(base, seconds(60, 60, 60, 60, 60)...), pendings(5)...)

	got := Stats(timings)

	if got.P50 != 60*time.Second || got.P99 != 60*time.Second || got.Mean != 60*time.Second {
		t.Fatalf("pending items must not enter the distribution: %+v", got)
	}
	if got.PendingCount != 5 || got.SampleSize != 5 || got.SuccessRate != 0.5 {
		t.Fatalf("expected 5 pending, 5 samples, 50%% success, got %+v", got)
	}
	if v, ok := got.Metrics()["p50_ms"]; !ok || v != int64(60000) {
		t.Fatalf("p50_ms must be 60000, got %v", v)
	}
}

func TestStatsFailedIsCountedNotTimed(t *testing.T) {
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	timings := append(arrivals(base, seconds(2, 4)...), Timing{Name: "x", Outcome: Failed, Err: fmt.Errorf("boom")})

	got := Stats(timings)

	if got.FailedCount != 1 || got.SampleSize != 2 || got.P50 != 2*time.Second {
		t.Fatalf("failed item must be counted only: %+v", got)
	}
}

func TestStatsEmptyAndAllPending(t *testing.T) {
	for name, timings := range map[string][]Timing{"empty": nil, "all pending": pendings(3)} {
		t.Run(name, func(t *testing.T) {
			got := Stats(timings)
			if got.SampleSize != 0 || got.P50 != 0 || got.Throughput != 0 {
				t.Fatalf("no samples must mean no latencies: %+v", got)
			}
			metrics := got.Metrics()
			for _, key := range []string{"p50_ms", "p90_ms", "p99_ms", "mean_ms", "max_ms", "throughput_per_sec"} {
				if _, present := metrics[key]; present {
					t.Fatalf("%s must be omitted when nothing arrived", key)
				}
			}
			if metrics["pending_count"] != len(timings) {
				t.Fatalf("pending_count must still be reported: %v", metrics)
			}
		})
	}
}

func TestStatsThroughputAndMakespan(t *testing.T) {
	base := time.Date(2026, 9, 17, 12, 0, 0, 0, time.UTC)
	timings := arrivals(base, seconds(2, 2, 2, 2)...)

	got := Stats(timings)

	if got.Makespan != 5*time.Second {
		t.Fatalf("makespan from first creation to last arrival must be 5s, got %s", got.Makespan)
	}
	if got.Throughput != 4.0/5.0 {
		t.Fatalf("throughput must be 4 items / 5s, got %v", got.Throughput)
	}
	if got.Max != 2*time.Second || got.Mean != 2*time.Second {
		t.Fatalf("max/mean wrong: %+v", got)
	}
}
