/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

// RestartRequestUIDLabel is the label that must be set on DistributedWorkload when managedBy is restart-controller.
const RestartRequestUIDLabel = "run.ai/request-ref"

// RestartControllerManagedBy is the ManagedBy value for the restart handoff controller.
const RestartControllerManagedBy = "restart-controller"

// ValidateDistributedWorkloadRestartManagedBy validates updates to DistributedWorkload
// managedBy and the restart-request-uid label according to:
// 1. ManagedBy can only be "" or "restart-controller".
// 2. If managedBy is "restart-controller", label run.ai/request-ref must be present with a value.
// 3. If managedBy is "", label run.ai/request-ref must be absent.
// 4. If the label exists it cannot be modified (only add when setting managedBy to restart-controller, remove when clearing).
func ValidateDistributedWorkloadRestartManagedBy(old, new *jobs.JobWorkload) error {
	newManagedBy := new.Spec.ManagedBy
	newLabelVal, newHasLabel := new.Labels[RestartRequestUIDLabel]
	oldLabelVal, oldHasLabel := old.Labels[RestartRequestUIDLabel]

	if newManagedBy != "" && newManagedBy != RestartControllerManagedBy {
		return fmt.Errorf("spec.managedBy can only be %q or empty, got %q", RestartControllerManagedBy, newManagedBy)
	}

	if newManagedBy == RestartControllerManagedBy && newLabelVal == "" {
		return fmt.Errorf("when spec.managedBy is %q, label %q must be set with a non-empty value", RestartControllerManagedBy, RestartRequestUIDLabel)
	}

	if newManagedBy == "" && newHasLabel {
		return fmt.Errorf("when spec.managedBy is empty, label %q must be removed", RestartRequestUIDLabel)
	}

	if oldHasLabel && newHasLabel && oldLabelVal != newLabelVal {
		return fmt.Errorf("label %q cannot be modified once set", RestartRequestUIDLabel)
	}

	return nil
}
