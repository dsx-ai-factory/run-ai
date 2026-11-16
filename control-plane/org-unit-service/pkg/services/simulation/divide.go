/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation

import (
	"cmp"
	"math"
	"slices"
	"time"

	"github.com/kai-scheduler/KAI-scheduler/pkg/fairshare"
)

const Unlimited = fairshare.Unlimited

// Unit is one org unit on one level of the queue tree, already mapped the way the cluster sync sends it to the
// scheduler.
type Unit struct {
	ID        string
	Priority  int
	CreatedAt time.Time
	Deserved  float64
	Limit     float64
	Weight    float64
	Demand    float64
}

type Result struct {
	Demand            float64
	FairShare         float64
	Guaranteed        float64
	OverQuota         float64
	Unmet             float64
	WeightShareInRank float64
}

// RequestableDemand is the part of the demand the scheduler can give: the demand capped by the limit.
func (u Unit) RequestableDemand() float64 {
	if u.Limit == Unlimited {
		return u.Demand
	}
	return math.Min(u.Demand, u.Limit)
}

// Divide splits capacity between the units of one level of the queue tree with the scheduler's own division.
func Divide(capacity float64, units []Unit) map[string]Result {
	// A round can run out of capacity midway, so the units visited first get more; a fixed order keeps the same
	// scenario from giving different results.
	units = slices.SortedFunc(slices.Values(units), func(a, b Unit) int {
		if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})
	queues := make([]*fairshare.Queue, len(units))
	weightsPerPriority := map[int]float64{}
	for i, u := range units {
		queues[i] = &fairshare.Queue{
			ID: u.ID, Priority: u.Priority, CreationTimestamp: u.CreatedAt,
			Deserved: u.Deserved, MaxAllowed: u.Limit, OverQuotaWeight: u.Weight, Request: u.Demand,
		}
		weightsPerPriority[u.Priority] += u.Weight
	}
	fairshare.Divide(capacity, 0, queues)

	results := make(map[string]Result, len(units))
	for i, u := range units {
		demand := u.RequestableDemand()
		// KAI's remainder round hands out whole GPUs without checking the remaining request, so a fractional demand or
		// limit can get more than it asks for. The scheduler never allocates past that, so neither does the simulation.
		fairShare := math.Min(queues[i].FairShare, demand)
		deserved := u.Deserved
		if deserved == Unlimited {
			deserved = capacity
		}
		guaranteed := math.Min(math.Min(deserved, demand), fairShare)
		weightShare := 0.0
		if total := weightsPerPriority[u.Priority]; total > 0 {
			weightShare = u.Weight / total
		}
		results[u.ID] = Result{
			Demand:            demand,
			FairShare:         fairShare,
			Guaranteed:        guaranteed,
			OverQuota:         fairShare - guaranteed,
			Unmet:             math.Max(0, demand-fairShare),
			WeightShareInRank: weightShare,
		}
	}
	return results
}
