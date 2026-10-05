/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import "time"

type Submission struct {
	Name            string
	RequestedAt     time.Time
	ApiAcceptedAt   time.Time
	ClusterSyncedAt time.Time
	PodCreatedAt    time.Time
	Err             error
}

func (s Submission) Accepted() bool {
	return s.Err == nil && !s.ApiAcceptedAt.IsZero()
}

type Spans struct {
	ApiAccepted   []Timing
	ClusterSynced []Timing
	PodCreated    []Timing
	TTJ           []Timing
}

func CollectSpans(submissions []Submission) Spans {
	var collected Spans
	for _, s := range submissions {
		if !s.Accepted() {
			collected.ApiAccepted = append(collected.ApiAccepted, Timing{Name: s.Name, Outcome: Failed, Created: s.RequestedAt, Err: s.Err})
			continue
		}
		collected.ApiAccepted = append(collected.ApiAccepted, arrived(s.Name, s.RequestedAt, s.ApiAcceptedAt))
		collected.ClusterSynced = append(collected.ClusterSynced, span(s.Name, s.ApiAcceptedAt, s.ClusterSyncedAt))
		collected.PodCreated = append(collected.PodCreated, span(s.Name, s.ClusterSyncedAt, s.PodCreatedAt))
		collected.TTJ = append(collected.TTJ, span(s.Name, s.ApiAcceptedAt, s.PodCreatedAt))
	}
	return collected
}

func span(name string, from, to time.Time) Timing {
	if from.IsZero() || to.IsZero() {
		return Timing{Name: name, Outcome: NeverArrived, Created: from}
	}
	return arrived(name, from, to)
}

func arrived(name string, from, to time.Time) Timing {
	return Timing{Name: name, Outcome: Arrived, Created: from, Arrived: to, Duration: max(to.Sub(from), 0)}
}

func OutOfOrder(submissions []Submission, tolerance time.Duration) []string {
	var names []string
	for _, s := range submissions {
		if !s.Accepted() {
			continue
		}
		if seenTooEarly(s.ClusterSyncedAt, s.ApiAcceptedAt, tolerance) || seenTooEarly(s.PodCreatedAt, s.ClusterSyncedAt, tolerance) {
			names = append(names, s.Name)
		}
	}
	return names
}

func seenTooEarly(event, previous time.Time, tolerance time.Duration) bool {
	return !event.IsZero() && !previous.IsZero() && previous.Sub(event) > tolerance
}

func SuccessRate(submissions []Submission) float64 {
	if len(submissions) == 0 {
		return 0
	}
	accepted := 0
	for _, s := range submissions {
		if s.Accepted() {
			accepted++
		}
	}
	return float64(accepted) / float64(len(submissions))
}
