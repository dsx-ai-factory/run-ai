/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"testing"
	"time"
)

func TestDivide(t *testing.T) {
	unit := func(id string, priority int, deserved, limit, weight, demand float64) Unit {
		return Unit{ID: id, Priority: priority, CreatedAt: time.Unix(int64(len(id)), 0),
			Deserved: deserved, Limit: limit, Weight: weight, Demand: demand}
	}
	for _, tc := range []struct {
		name     string
		capacity float64
		units    []Unit
		expected map[string]Result
	}{
		{
			name:     "quota first, then the rest by weight within the rank, the leftover GPU going to the older unit",
			capacity: 20,
			units:    []Unit{unit("a", 2, 4, Unlimited, 1, 20), unit("bb", 2, 2, Unlimited, 3, 20)},
			expected: map[string]Result{
				"a":  {Demand: 20, FairShare: 8, Guaranteed: 4, OverQuota: 4, Unmet: 12, WeightShareInRank: 0.25},
				"bb": {Demand: 20, FairShare: 12, Guaranteed: 2, OverQuota: 10, Unmet: 8, WeightShareInRank: 0.75},
			},
		},
		{
			name:     "a lower rank gets only its quota while a higher rank still asks for more",
			capacity: 10,
			units:    []Unit{unit("high", 3, 2, Unlimited, 1, 10), unit("low", 1, 3, Unlimited, 1, 10)},
			expected: map[string]Result{
				"high": {Demand: 10, FairShare: 7, Guaranteed: 2, OverQuota: 5, Unmet: 3, WeightShareInRank: 1},
				"low":  {Demand: 10, FairShare: 3, Guaranteed: 3, OverQuota: 0, Unmet: 7, WeightShareInRank: 1},
			},
		},
		{
			name:     "the demand is capped by the limit",
			capacity: 20,
			units:    []Unit{unit("capped", 1, 2, 5, 1, 20)},
			expected: map[string]Result{
				"capped": {Demand: 5, FairShare: 5, Guaranteed: 2, OverQuota: 3, Unmet: 0, WeightShareInRank: 1},
			},
		},
		{
			name:     "quotas are granted even when they exceed the capacity",
			capacity: 0,
			units:    []Unit{unit("a", 1, 4, Unlimited, 1, 10), unit("bb", 1, 2, Unlimited, 1, 1)},
			expected: map[string]Result{
				"a":  {Demand: 10, FairShare: 4, Guaranteed: 4, OverQuota: 0, Unmet: 6, WeightShareInRank: 0.5},
				"bb": {Demand: 1, FairShare: 1, Guaranteed: 1, OverQuota: 0, Unmet: 0, WeightShareInRank: 0.5},
			},
		},
		{
			name:     "an unlimited quota covers the whole demand, and no weight gives no weight share",
			capacity: 8,
			units:    []Unit{unit("unlimited", 1, Unlimited, Unlimited, 0, 6)},
			expected: map[string]Result{
				"unlimited": {Demand: 6, FairShare: 6, Guaranteed: 6, OverQuota: 0, Unmet: 0, WeightShareInRank: 0},
			},
		},
		{
			name:     "an org unit that asks for nothing gets nothing, even with quota",
			capacity: 10,
			units:    []Unit{unit("idle", 1, 5, Unlimited, 1, 0)},
			expected: map[string]Result{
				"idle": {Demand: 0, FairShare: 0, Guaranteed: 0, OverQuota: 0, Unmet: 0, WeightShareInRank: 1},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			results := Divide(tc.capacity, tc.units)
			for id, want := range tc.expected {
				if got := results[id]; got != want {
					t.Errorf("%s: got %+v, want %+v", id, got, want)
				}
			}
		})
	}
}

func TestDivideIgnoresInputOrder(t *testing.T) {
	random := rand.New(rand.NewPCG(1, 2))
	for scenario := range 200 {
		units := make([]Unit, 40)
		for i := range units {
			units[i] = Unit{
				ID: fmt.Sprintf("unit-%02d", i), Priority: random.IntN(3), CreatedAt: time.Unix(int64(random.IntN(5)), 0),
				Deserved: float64(random.IntN(3)), Limit: Unlimited, Weight: float64(1 + random.IntN(3)),
				Demand: float64(random.IntN(20)),
			}
		}
		capacity := float64(10 + random.IntN(90))
		expected := Divide(capacity, units)
		for range 5 {
			random.Shuffle(len(units), func(i, j int) { units[i], units[j] = units[j], units[i] })
			if got := Divide(capacity, units); !maps.Equal(got, expected) {
				t.Fatalf("scenario %d: the result changed with the order of the units", scenario)
			}
		}
	}
}
