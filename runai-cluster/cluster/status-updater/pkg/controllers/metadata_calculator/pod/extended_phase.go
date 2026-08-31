/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	v1 "k8s.io/api/core/v1"
)

const (
	pending   = "Pending"
	allocated = "Allocated"
	pipelined = "Pipelined"
	binding   = "Binding"
	bound     = "Bound"
	running   = "Running"
	releasing = "Releasing"
	deleted   = "Deleted"
	unknown   = "Unknown"
	failed    = "Failed"
)

var (
	activeUsedExtendedPhases = []string{allocated, pipelined, binding, bound, running, releasing}
	boundExtendedPhases      = []string{allocated, bound, running, releasing}
)

func GetExtendedPhase(pod *v1.Pod) string {
	switch pod.Status.Phase {
	case v1.PodRunning:
		if pod.DeletionTimestamp != nil {
			return releasing
		}
		return running
	case v1.PodPending:
		if pod.DeletionTimestamp != nil {
			return releasing
		}
		if len(pod.Spec.NodeName) == 0 {
			return pending
		}
		return bound
	case v1.PodUnknown:
		return unknown
	case v1.PodSucceeded:
		return succeeded
	case v1.PodFailed:
		return failed
	default:
		return unknown
	}
}

func IsActiveUsedStatus(podExtendedPhase string) bool {
	for _, status := range activeUsedExtendedPhases {
		if podExtendedPhase == status {
			return true
		}
	}
	return false
}

func IsPodPending(podPhase v1.PodPhase, podExtendedPhase string) bool {
	return podPhase == v1.PodPending && podExtendedPhase != deleted
}

func IsPodBound(podExtendedPhase string) bool {
	for _, status := range boundExtendedPhases {
		if podExtendedPhase == status {
			return true
		}
	}
	return false
}
