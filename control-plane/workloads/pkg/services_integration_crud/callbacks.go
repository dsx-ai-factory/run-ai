/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_integration_crud

import (
	"context"
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"
	"net/http"
)

// UpdateOnDeletedWorkload - this function gets updates from workloads to workloads_crud and delete a workload that has been deleted not through the fullySaas.
func UpdateCrudOnDeletedWorkload(ctx context.Context, workloadId uuid.UUID) *api.Error {
	err := dao_crud.DeleteWorkloadById(ctx, workloadId)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, err.Error(), err)
	}

	return nil
}

// UpdateCrudOnSyncWorkloads - This function is used to update the crud_workloads when cluster sync is syncing the workloads.
// existingWorkloadIds is an array of the ids of the workloads that we shouldn't delete.
// deletionFilters is filters that we get only from clusters older than 2.16, when deployments and the rest of the workloads didn't sync together so we don't want to mistakenly delete the other type of the workloads
// to & from is the uuids that are synced in this cycle, we don't want to delete ids that are not in this range.
func UpdateCrudOnSyncWorkloads(ctx context.Context,
	tenantId uint,
	projectId string,
	clusterId string,
	to *uuid.UUID,
	from *uuid.UUID,
	existingWorkloadIds []string,
	deletionFilters []validators.ParameterizedFilter) *api.Error {
	err := dao_crud.DbSyncWorkloads(ctx, tenantId, projectId, clusterId, to, from, existingWorkloadIds, deletionFilters)
	if err != nil {
		return err
	}
	return nil
}
