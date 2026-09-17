/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"fmt"

	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

func Training(ctx context.Context, clients *client.Clients, spec WorkloadSpec) (WorkloadRef, error) {
	training, _, err := clients.Workloads.TrainingsAPI.CreateTraining(ctx).
		TrainingCreationRequest(trainingRequest(spec)).Execute()
	if err != nil {
		return WorkloadRef{}, fmt.Errorf("creating training %q: %w", spec.Name, err)
	}
	return WorkloadRef{ID: training.WorkloadId, Name: spec.Name}, nil
}

func trainingRequest(spec WorkloadSpec) workloadclient.TrainingCreationRequest {
	return workloadclient.TrainingCreationRequest{
		Name:      spec.Name,
		ProjectId: spec.Project.ID,
		ClusterId: spec.Project.ClusterID,
		Spec: &workloadclient.TrainingSpecSpec{
			Command:         nullableString(spec.Command),
			Image:           nullableString(spec.Image),
			ImagePullPolicy: *workloadclient.NewNullableImagePullPolicy(workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT.Ptr()),
			Compute:         compute(spec.GPUs),
			Labels:          labels(spec.Labels),
		},
	}
}
