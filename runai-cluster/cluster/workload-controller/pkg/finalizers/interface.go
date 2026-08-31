/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package finalizers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type RemoveFinalizer interface {
	// RemoveFinalizer removes the finalizer from the jobWorkload
	// only for backward compatibility with workload pre 1.13.0
	// should be removed in the future
	RemoveFinalizer(ctx context.Context, jobWorkload *jobs.JobWorkload) error
}

type RemoveFinalizerFactory interface {
	CreateRemoveFinalizer(client k8sClient.Client) RemoveFinalizer
}

func GetRemoveFinalizerFactory(workloadType constants.WorkloadType) (RemoveFinalizerFactory, error) {
	switch workloadType {
	case constants.InteractiveWorkload:
		return &InteractiveWorkloadRemoveFinalizerFactory{}, nil
	case constants.TrainingWorkload:
		return &TrainingWorkloadRemoveFinalizerFactory{}, nil
	case constants.DistributedWorkload:
		return &DistributedWorkloadRemoveFinalizerFactory{}, nil

	default:
		return nil, fmt.Errorf("unknown workload type: %s", workloadType)
	}
}
