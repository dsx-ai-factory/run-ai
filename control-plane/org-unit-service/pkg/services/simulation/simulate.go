/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package simulation

import "math"

// OrgUnit holds one department or project as the scheduler sees it after the cluster sync's mapping.
type OrgUnit struct {
	ID        string
	ParentID  string
	Deserved  float64
	Limit     float64
	Weight    float64
	Priority  int
	Demand    float64
	CreatedAt int64
}

type Share struct {
	FairShare  float64
	Guaranteed float64
	OverQuota  float64
	Unmet      float64
}

// Simulate splits capacity between departments and then each department's fair share between its projects, as the
// scheduler's queue hierarchy does. A department asks for the sum of its projects' demand, and a project whose
// demand is math.Inf(1) asks for as much as it can get. Departments' Demand is ignored.
func Simulate(capacity float64, departments, projects []OrgUnit) (departmentShares, projectShares map[string]Share) {
	projectsOf := map[string][]OrgUnit{}
	for _, p := range projects {
		projectsOf[p.ParentID] = append(projectsOf[p.ParentID], p)
	}

	departmentQueues := make([]Queue, 0, len(departments))
	departmentRequests := make(map[string]float64, len(departments))
	for _, d := range departments {
		request := 0.0
		for _, p := range projectsOf[d.ID] {
			request += p.Demand
		}
		departmentRequests[d.ID] = request
		departmentQueues = append(departmentQueues, toQueue(d, request))
	}
	departmentFairShares := DivideResource(capacity, departmentQueues)

	departmentShares = make(map[string]Share, len(departments))
	projectShares = make(map[string]Share, len(projects))
	for _, d := range departments {
		fairShare := departmentFairShares[d.ID]
		departmentShares[d.ID] = share(d, departmentRequests[d.ID], fairShare, capacity, capacity)

		children := projectsOf[d.ID]
		queues := make([]Queue, 0, len(children))
		for _, p := range children {
			queues = append(queues, toQueue(p, p.Demand))
		}
		projectFairShares := DivideResource(fairShare, queues)
		for _, p := range children {
			projectShares[p.ID] = share(p, p.Demand, projectFairShares[p.ID], fairShare, capacity)
		}
	}
	return departmentShares, projectShares
}

func toQueue(u OrgUnit, request float64) Queue {
	return Queue{
		ID: u.ID, Deserved: u.Deserved, MaxAllowed: u.Limit, OverQuotaWeight: u.Weight, Priority: u.Priority,
		Request: request, CreatedAt: u.CreatedAt,
	}
}

// share reports an unbounded request's unmet part against the most the org unit could use: its limit, else the
// node pool's capacity.
func share(u OrgUnit, request, fairShare, levelCapacity, capacity float64) Share {
	requestable := request
	if u.Limit != Unlimited {
		requestable = math.Min(u.Limit, request)
	}
	deserved := u.Deserved
	if deserved == Unlimited {
		deserved = levelCapacity
	}
	guaranteed := math.Min(math.Min(deserved, requestable), fairShare)

	demand := request
	if math.IsInf(demand, 1) {
		demand = capacity
	}
	if u.Limit != Unlimited {
		demand = math.Min(u.Limit, demand)
	}
	return Share{
		FairShare:  fairShare,
		Guaranteed: guaranteed,
		OverQuota:  fairShare - guaranteed,
		Unmet:      math.Max(0, demand-fairShare),
	}
}
