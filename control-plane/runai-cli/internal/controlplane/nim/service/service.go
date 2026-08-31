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

type BackendNimService struct {
	nimClient workload.WithUpdateAction
}

func NewBackendNimService(nimClient workload.WithUpdateAction) *BackendNimService {
	return &BackendNimService{nimClient: nimClient}
}

func (bns *BackendNimService) Update(ctx context.Context, uuid, clusterId string, spec *spec.GeneralSpec) (string, error) {
	return bns.nimClient.Update(ctx, uuid, clusterId, spec)
}
