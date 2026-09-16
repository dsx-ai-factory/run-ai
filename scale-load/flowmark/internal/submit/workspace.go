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

func Workspace(ctx context.Context, clients *client.Clients, spec WorkloadSpec) (WorkloadRef, error) {
	workspace, _, err := clients.Workloads.WorkspacesAPI.CreateWorkspace(ctx).
		WorkspaceCreationRequest(workspaceRequest(spec)).Execute()
	if err != nil {
		return WorkloadRef{}, fmt.Errorf("creating workspace %q: %w", spec.Name, err)
	}
	return WorkloadRef{ID: workspace.WorkloadId, Name: spec.Name}, nil
}

func workspaceRequest(spec WorkloadSpec) workloadclient.WorkspaceCreationRequest {
	return workloadclient.WorkspaceCreationRequest{
		Name:      spec.Name,
		ProjectId: spec.Project.ID,
		ClusterId: spec.Project.ClusterID,
		Spec: &workloadclient.WorkspaceSpecSpec{
			Command:         nullableString(spec.Command),
			Image:           nullableString(spec.Image),
			ImagePullPolicy: *workloadclient.NewNullableImagePullPolicy(workloadclient.IMAGEPULLPOLICY_IF_NOT_PRESENT.Ptr()),
			Compute:         compute(spec.GPUs),
			Labels:          labels(spec.Labels),
		},
	}
}
