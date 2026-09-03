/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import "sort"

type WorkloadIssues struct {
	hasIssues          bool
	hasCritical        bool
	hasCriticalCompute bool
	groupsWithCritical []string
}

func NewWorkloadIssues() WorkloadIssues {
	return WorkloadIssues{
		groupsWithCritical: make([]string, 0),
	}
}

func (wi *WorkloadIssues) Set(current WorkloadIssues) {
	wi.hasIssues = current.hasIssues
	wi.hasCritical = current.hasCritical
	wi.hasCriticalCompute = current.hasCriticalCompute

	wi.groupsWithCritical = current.groupsWithCritical
	sort.Strings(wi.groupsWithCritical)
}
