/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation

import (
	"math"
	"sort"
)

const Unlimited = -1.0

type Queue struct {
	ID              string
	Deserved        float64
	MaxAllowed      float64
	OverQuotaWeight float64
	Priority        int
	Request         float64
	CreatedAt       int64
}

type leftover struct {
	queue     *Queue
	remaining float64
}

// DivideResource splits totalAmount of one resource between queues the way KAI-scheduler v0.18.2 does
// (plugins/proportion/resource_division, kValue 0). It must stay in step with the scheduler version the
// clusters run; kai_division_cases.json holds results produced by KAI's own code.
func DivideResource(totalAmount float64, queues []Queue) map[string]float64 {
	fair := make(map[string]float64, len(queues))
	remaining := totalAmount
	for i := range queues {
		q := &queues[i]
		deserved := q.Deserved
		if deserved == Unlimited {
			deserved = totalAmount
		}
		amount := math.Min(deserved, requestable(q))
		fair[q.ID] += amount
		remaining -= amount
	}
	if remaining <= 0 {
		return fair
	}

	byPriority := map[int][]*Queue{}
	for i := range queues {
		byPriority[queues[i].Priority] = append(byPriority[queues[i].Priority], &queues[i])
	}
	priorities := make([]int, 0, len(byPriority))
	for p := range byPriority {
		priorities = append(priorities, p)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(priorities)))

	leftovers := map[int]map[string]*leftover{}
	for _, p := range priorities {
		remaining, leftovers[p] = divideUpToFairShare(remaining, byPriority[p], fair)
	}
	for _, p := range priorities {
		if remaining <= 0 {
			break
		}
		if len(leftovers[p]) > 0 {
			remaining = divideRemaining(remaining, leftovers[p], fair)
		}
	}
	return fair
}

func requestable(q *Queue) float64 {
	if q.MaxAllowed == Unlimited {
		return q.Request
	}
	return math.Min(q.MaxAllowed, q.Request)
}

func remainingRequested(q *Queue, fair float64) float64 {
	r := requestable(q)
	if r < fair {
		return 0
	}
	return r - fair
}

func isSatisfied(q *Queue, fair float64) bool {
	if q.Request <= fair {
		return true
	}
	return q.MaxAllowed != Unlimited && q.MaxAllowed <= fair
}

func divideUpToFairShare(total float64, queues []*Queue, fair map[string]float64) (float64, map[string]*leftover) {
	left := map[string]*leftover{}
	for {
		anotherRound := false
		amountThisRound := total

		totalWeights := 0.0
		for _, q := range queues {
			if remainingRequested(q, fair[q.ID]) > 0 {
				totalWeights += q.OverQuotaWeight
			}
		}
		if totalWeights == 0 {
			break
		}
		shareWeights := map[string]float64{}
		sum := 0.0
		for _, q := range queues {
			if isSatisfied(q, fair[q.ID]) {
				continue
			}
			w := math.Max(0, q.OverQuotaWeight/totalWeights)
			shareWeights[q.ID] = w
			sum += w
		}
		if sum == 0 {
			break
		}

		for _, q := range queues {
			if total == 0 {
				break
			}
			if isSatisfied(q, fair[q.ID]) || q.OverQuotaWeight == 0 {
				continue
			}
			requested := remainingRequested(q, fair[q.ID])
			fairShare := amountThisRound * (shareWeights[q.ID] / sum)
			give := 0.0
			if requested <= fairShare {
				give = requested
				delete(left, q.ID)
			} else {
				if rounded := math.Floor(fairShare); rounded > 0 {
					give = rounded
				}
				if fairShare-give > 0 {
					left[q.ID] = &leftover{queue: q, remaining: fairShare - give}
				}
			}
			if give == 0 {
				continue
			}
			fair[q.ID] += give
			total -= give
			anotherRound = anotherRound || requested < fairShare
		}
		if !anotherRound || total == 0 {
			break
		}
	}
	return total, left
}

func divideRemaining(total float64, left map[string]*leftover, fair map[string]float64) float64 {
	sorted := make([]*leftover, 0, len(left))
	for _, l := range left {
		sorted = append(sorted, l)
	}
	sort.Slice(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if a.remaining != b.remaining {
			return a.remaining > b.remaining
		}
		if a.queue.CreatedAt != b.queue.CreatedAt {
			return a.queue.CreatedAt < b.queue.CreatedAt
		}
		return a.queue.ID < b.queue.ID
	})
	for _, l := range sorted {
		if total == 0 {
			break
		}
		give := math.Min(1, total)
		fair[l.queue.ID] += give
		total -= give
	}
	return total
}
