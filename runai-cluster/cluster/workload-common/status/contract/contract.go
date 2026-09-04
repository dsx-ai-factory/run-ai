/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package contract contains status contract constants shared by workload status calculators and adapters.
package contract

import (
	"fmt"
	"strings"
)

type Phase string

const (
	AdditionalFieldReadyPods        = "readyPods"
	AdditionalFieldRunningPods      = "runningPods"
	AdditionalFieldPendingPods      = "pendingPods"
	AdditionalFieldInitializingPods = "initializingPods"
	AdditionalFieldSucceededPods    = "succeededPods"
	AdditionalFieldFailedPods       = "failedPods"
)

const (
	ConditionTypeReady                 = "Ready"
	ConditionTypeGPUNetworkAccelerated = "GPUNetworkAccelerated"
)

const (
	Creating     Phase = "Creating"
	Initializing Phase = "Initializing"
	Running      Phase = "Running"
	Ready        Phase = "Ready"
	Degraded     Phase = "Degraded"
	Pending      Phase = "Pending"
	Suspended    Phase = "Suspended"
	Suspending   Phase = "Suspending"
	Resuming     Phase = "Resuming"
	Completed    Phase = "Completed"
	Failed       Phase = "Failed"
	Deleting     Phase = "Deleting"
	Terminating  Phase = "Terminating"
)

const (
	MsgInitializingSinglePod = "Workload is initializing, the pod is still not ready"
	MsgInitializingMultiPods = "Workload is initializing, minimum ready pods not reached, waiting for %d pods to be ready"
	MsgGenericJobFailed      = "job reached failed state"
	MsgGenericDegraded       = "Workload is in degraded state"

	MsgFailedDueToIssueSingular   = "Workload failed due to a %s issue"
	MsgFailedDueToIssuesPlural    = "Workload failed due to %s issues"
	MsgDegradedDueToIssueSingular = "Workload is in degraded state due to a %s issue"
	MsgDegradedDueToIssuesPlural  = "Workload is in degraded state due to %s issues"

	MsgMnnvlNotReady = "MNNVL support is not ready"
)

const (
	ReasonMnnvlNotReady = "MNNVLSupportNotReady"
)

func FormatIssueCause(groups []string, singular, plural string) string {
	switch len(groups) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf(singular, groups[0])
	default:
		return fmt.Sprintf(plural, strings.Join(groups, ", "))
	}
}
