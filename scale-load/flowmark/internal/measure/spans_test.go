/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"errors"
	"testing"
	"time"
)

var start = time.Date(2026, 10, 5, 9, 0, 0, 0, time.UTC)

func at(seconds float64) time.Time {
	return start.Add(time.Duration(seconds * float64(time.Second)))
}

func TestCollectSpansSplitsOneWorkspaceIntoFourSpans(t *testing.T) {
	s := CollectSpans([]Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)}})

	want := map[string]time.Duration{"apiAccepted": 300 * time.Millisecond, "clusterSynced": 2 * time.Second, "podCreated": 3 * time.Second, "ttj": 5 * time.Second}
	for name, timings := range map[string][]Timing{"apiAccepted": s.ApiAccepted, "clusterSynced": s.ClusterSynced, "podCreated": s.PodCreated, "ttj": s.TTJ} {
		if len(timings) != 1 || timings[0].Outcome != Arrived || timings[0].Duration != want[name] {
			t.Errorf("%s span: want one arrival of %s, got %+v", name, want[name], timings)
		}
	}
}

func TestFailedSubmissionCountsOnlyInTheApiAcceptedSpanAndNeverAsZero(t *testing.T) {
	s := CollectSpans([]Submission{{Name: "w1", RequestedAt: at(0), Err: errors.New("503 from the gateway")}})

	if len(s.ApiAccepted) != 1 || s.ApiAccepted[0].Outcome != Failed || s.ApiAccepted[0].Duration != 0 {
		t.Fatalf("a failed POST must be a Failed api timing, got %+v", s.ApiAccepted)
	}
	if len(s.ClusterSynced)+len(s.PodCreated)+len(s.TTJ) != 0 {
		t.Fatalf("a failed Submission has no cluster Spans, got runner=%d job=%d ttj=%d", len(s.ClusterSynced), len(s.PodCreated), len(s.TTJ))
	}
	if got := Stats(s.ApiAccepted); got.FailedCount != 1 || got.SampleSize != 0 {
		t.Fatalf("stats must count it as failed and keep it out of the latencies: %+v", got)
	}
}

func TestMissingArrivalsArePendingNotZero(t *testing.T) {
	s := CollectSpans([]Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3)}})

	if s.ClusterSynced[0].Outcome != Arrived {
		t.Fatalf("the runner did arrive: %+v", s.ClusterSynced[0])
	}
	for name, timings := range map[string][]Timing{"podCreated": s.PodCreated, "ttj": s.TTJ} {
		if timings[0].Outcome != Pending || timings[0].Duration != 0 {
			t.Errorf("%s span without a pod must be pending with no duration, got %+v", name, timings[0])
		}
	}
	if got := Stats(s.TTJ); got.PendingCount != 1 || got.SampleSize != 0 {
		t.Fatalf("pending must stay out of the percentiles: %+v", got)
	}
}

func TestSpansAreConsistentByConstructionAndACorrelationBugIsCaught(t *testing.T) {
	good := CollectSpans([]Submission{
		{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3), ClusterSyncedAt: at(2.3), PodCreatedAt: at(5.3)},
		{Name: "w2", RequestedAt: at(0.5), ApiAcceptedAt: at(0.7), ClusterSyncedAt: at(1.0), PodCreatedAt: at(1.4)},
	})
	if names := good.Inconsistent(0); len(names) != 0 {
		t.Fatalf("Spans from the same stamps must add up exactly, got mismatches %v", names)
	}

	bad := good
	bad.PodCreated = []Timing{arrived("w1", at(2.3), at(9.3)), good.PodCreated[1]}
	if names := bad.Inconsistent(time.Second); len(names) != 1 || names[0] != "w1" {
		t.Fatalf("a job span that does not belong to its workspace must be reported, got %v", names)
	}
}

func TestClusterSyncedBeforeTheAcceptIsClampedAndToleratedWithinTheGap(t *testing.T) {
	s := CollectSpans([]Submission{{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(1.0), ClusterSyncedAt: at(0.8), PodCreatedAt: at(3.0)}})

	if s.ClusterSynced[0].Duration != 0 {
		t.Fatalf("a WorkloadRunner that appears before the 200 is a zero span, never negative: %+v", s.ClusterSynced[0])
	}
	if names := s.Inconsistent(200 * time.Millisecond); len(names) != 0 {
		t.Fatalf("a tolerance covering the clamped 200ms must accept it, got %v", names)
	}
	if names := s.Inconsistent(100 * time.Millisecond); len(names) != 1 {
		t.Fatalf("a tighter tolerance must report it, got %v", names)
	}
}

func TestSuccessRate(t *testing.T) {
	subs := []Submission{
		{Name: "w1", RequestedAt: at(0), ApiAcceptedAt: at(0.3)},
		{Name: "w2", RequestedAt: at(0), Err: errors.New("boom")},
		{Name: "w3", RequestedAt: at(0), ApiAcceptedAt: at(0.3)},
		{Name: "w4", RequestedAt: at(0), ApiAcceptedAt: at(0.3)},
	}
	if got := SuccessRate(subs); got != 0.75 {
		t.Fatalf("3 of 4 accepted is 0.75, got %v", got)
	}
	if got := SuccessRate(nil); got != 0 {
		t.Fatalf("no submissions is not a rate, got %v", got)
	}
}
