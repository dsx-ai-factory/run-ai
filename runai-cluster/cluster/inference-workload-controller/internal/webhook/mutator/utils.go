/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mutator

import (
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/status/aggregator"
	inferenceworkload "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadStatusUpdater struct {
}

func (as *WorkloadStatusUpdater) SetPhaseIfNeeded(jobWorkload client.Object) error {
	switch workloadCast := jobWorkload.(type) {
	case *inferenceworkload.InferenceWorkload:
		status := &workloadCast.Status.InferenceStatus
		if status != nil {
			status.Phase = aggregator.Creating
		}
	default:
		return fmt.Errorf("unknown workload type %T", jobWorkload)
	}

	return nil
}
