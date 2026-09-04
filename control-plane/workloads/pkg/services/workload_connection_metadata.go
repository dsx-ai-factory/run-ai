/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/google/uuid"
)

// WorkloadConnectionMetadata contains the connection fields persisted from native request translation.
type WorkloadConnectionMetadata struct {
	Environments        []api.Environment
	ExternalConnections []api.Connection
}

// UpdateWorkloadConnectionMetadata stores create-time connection correlation metadata.
// Empty slices are ignored so minimal creation notifications do not clear existing metadata.
func (s *WorkloadsService) UpdateWorkloadConnectionMetadata(ctx context.Context, workloadId string, metadata WorkloadConnectionMetadata) *api.Error {
	if len(metadata.Environments) == 0 && len(metadata.ExternalConnections) == 0 {
		return nil
	}

	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return apiErr
	}

	parsedWorkloadId, err := uuid.Parse(workloadId)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("invalid workload id %q", workloadId), err)
	}

	return s.WorkloadsDao.DbUpdateWorkloadConnectionMetadata(
		ctx,
		tenantId,
		parsedWorkloadId,
		convutils.ConvertSliceTypes(metadata.Environments, environmentApiToModel),
		convutils.ConvertSliceTypes(metadata.ExternalConnections, connectionApiToModel),
	)
}

// EnsureWorkloadConnectionMetadata stores connection metadata even when the
// workloads-manager creation notification has not inserted the row yet.
func (s *WorkloadsService) EnsureWorkloadConnectionMetadata(
	ctx context.Context,
	workloadCreationFields api.WorkloadCreationFields,
	metadata WorkloadConnectionMetadata,
) *api.Error {
	if len(metadata.Environments) == 0 && len(metadata.ExternalConnections) == 0 {
		return nil
	}

	updateErr := s.UpdateWorkloadConnectionMetadata(ctx, workloadCreationFields.Id, metadata)
	if updateErr == nil {
		return nil
	}
	if int(updateErr.Code) != http.StatusNotFound {
		return updateErr
	}

	workloadCreationFields.Environments = metadata.Environments
	workloadCreationFields.ExternalConnections = metadata.ExternalConnections

	if _, createErr := s.InternalCreateWorkloadHelper(ctx, workloadCreationFields); createErr != nil {
		return createErr
	}

	return s.UpdateWorkloadConnectionMetadata(ctx, workloadCreationFields.Id, metadata)
}
