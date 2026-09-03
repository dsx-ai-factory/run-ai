/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusUpdater interface {
	UpdateStatus(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error
}

type IsStatusChanged func(ctx context.Context, modifiedWorkload *jobs.JobWorkload, originalWorkload *jobs.JobWorkload) bool

type PreUpdateStatusFunc func(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error

type StatusUpdaterFactory interface {
	CreateStatusUpdater(client client.Client) StatusUpdater
}

func GetStatusUpdaterFactory(workloadType constants.WorkloadType) (StatusUpdaterFactory, error) {
	switch workloadType {
	case constants.InteractiveWorkload:
		return &InteractiveWorkloadStatusUpdaterFactory{}, nil
	case constants.TrainingWorkload:
		return &TrainingWorkloadStatusUpdaterFactory{}, nil
	case constants.DistributedWorkload:
		return &DistributedWorkloadStatusUpdaterFactory{}, nil
	default:
		return nil, fmt.Errorf("unknown workload type: %s", workloadType)
	}
}
