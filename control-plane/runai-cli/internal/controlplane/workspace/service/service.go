/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
)

type BackendWorkspaceService struct {
	workspaceClient workload.WithSuspendableAction
}

func NewBackendWorkspaceService(workspaceClient workload.WithSuspendableAction) *BackendWorkspaceService {
	return &BackendWorkspaceService{workspaceClient: workspaceClient}
}

func (bws *BackendWorkspaceService) DeleteByUuid(ctx context.Context, uuid string) error {
	return bws.workspaceClient.Delete(ctx, uuid)
}

func (bws *BackendWorkspaceService) SuspendByUuid(ctx context.Context, uuid string) error {
	return bws.workspaceClient.Suspend(ctx, uuid)

}

func (bws *BackendWorkspaceService) ResumeByUuid(ctx context.Context, uuid string) error {
	return bws.workspaceClient.Resume(ctx, uuid)

}

func (bws *BackendWorkspaceService) Submit(ctx context.Context, name, projectId, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bws.workspaceClient.Submit(ctx, name, projectId, clusterId, spec)
}
