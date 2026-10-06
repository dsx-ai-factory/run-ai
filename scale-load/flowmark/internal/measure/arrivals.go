/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package measure

import (
	"context"
	"time"

	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Arrivals struct {
	runners *Watcher
	pods    *Watcher
}

func OpenArrivals(ctx context.Context, kube k8sclient.WithWatch, namespace string) (*Arrivals, error) {
	runners, err := WatchWorkloads(ctx, kube, namespace)
	if err != nil {
		return nil, err
	}
	pods, err := WatchWorkloadPods(ctx, kube, namespace)
	if err != nil {
		runners.Stop()
		return nil, err
	}
	return &Arrivals{runners: runners, pods: pods}, nil
}

func (a *Arrivals) MarkAccepted(name string, at time.Time) {
	a.runners.MarkCreated(name, at)
	a.pods.MarkCreated(name, at)
}

func (a *Arrivals) Fill(submissions []Submission, timeout time.Duration) []Submission {
	synced := arrivalStamps(a.runners.WaitAll(timeout))
	created := arrivalStamps(a.pods.WaitAll(timeout))
	filled := make([]Submission, len(submissions))
	for i, s := range submissions {
		s.ClusterSyncedAt = synced[s.Name]
		s.PodCreatedAt = created[s.Name]
		filled[i] = s
	}
	return filled
}

func (a *Arrivals) Reconnects() int {
	return a.runners.Reconnects() + a.pods.Reconnects()
}

func (a *Arrivals) Stop() {
	a.pods.Stop()
	a.runners.Stop()
}

func arrivalStamps(timings []Timing) map[string]time.Time {
	stamps := make(map[string]time.Time, len(timings))
	for _, timing := range timings {
		if timing.Outcome == Arrived {
			stamps[timing.Name] = timing.Arrived
		}
	}
	return stamps
}
