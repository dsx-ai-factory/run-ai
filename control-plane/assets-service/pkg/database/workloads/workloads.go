/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
)

const ColumnTenantId = "tenant_id"

func JoinWorkloadsAndWorkloadsAssets() string {
	return fmt.Sprintf("join %s on %s.id = %s.workload_id", dbtables.TableNameWorkloads, dbtables.TableNameWorkloads, dbtables.TableNameWorkloadsAssets)
}

// SearchWorkloads find workload assets by criteria
func SearchWorkloads(ctx context.Context, workloadType string,
	id *string, kind *api.AssetKind, allowedProjectIds *[]string,
) ([]string, []dbmodel.CombinedWorkloadsRec, *api.Error) {
	workloads := make([]dbmodel.CombinedWorkloadsRec, 0)

	query := dbmodel.CombinedWorkloadsRec{}
	if id != nil {
		query.WorkloadsAssetsRec.AssetId = *id
	}
	if kind != nil {
		query.WorkloadsAssetsRec.AssetKind = *kind
	}

	combinedRec := dbmodel.CombinedWorkloadsRec{}
	result := runaidb.DB().WithContext(ctx).Model(combinedRec).
		Select("*").
		Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsColumnWorkloadType), workloadType).
		Joins(JoinWorkloadsAndWorkloadsAssets()).
		Where(query)

	if allowedProjectIds != nil {
		result.Where(fmt.Sprintf("%s IN (?)", dbtables.WorkloadsColumnProjectId), *allowedProjectIds)
	}

	result.Where(fmt.Sprintf("%s.%s = ?", dbtables.TableNameWorkloads, ColumnTenantId), auth.TenantIdOf(ctx))

	_, scanErr := result.
		Order(&combinedRec.WorkloadsAssetsRec.AssetKind).
		Order(&combinedRec.WorkloadsAssetsRec.AssetId).
		Order(&combinedRec.WorkloadRec.Id).
		Scan(&workloads)
	if scanErr != nil {
		return nil, nil, api.NewError(http.StatusServiceUnavailable, "failed to list workloads", scanErr)
	}

	names := make([]string, 0, len(workloads))
	for _, workspace := range workloads {
		names = append(names, workspace.WorkloadRec.Name)
	}
	return names, workloads, nil
}

// UpdateWorkloadCurrentState sets workloads.current_state for the given workload id (same as the legacy backend).
func UpdateWorkloadCurrentState(ctx context.Context, clusterId, workloadId, currentState string) *api.Error {
	db := runaidb.DB().WithContext(ctx)
	updRows, updErr := db.Model(&dbmodel.WorkloadRec{}).
		Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
			dbtables.WorkloadsColumnTenantId,
			dbtables.WorkloadsColumnClusterId,
			dbtables.WorkloadsColumnId),
			auth.TenantIdOf(ctx), clusterId, workloadId).
		Updates(map[string]interface{}{
			dbtables.WorkloadsColumnCurrentState: currentState,
			dbtables.WorkloadsColumnUpdatedBy:    auth.SubjectIdOf(ctx),
		})
	if updErr != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to update current state of %s to %s per update from cluster %s", workloadId, currentState, clusterId),
			updErr)
	}
	if updRows == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("%s does not exist, handling status update from %s failed", workloadId, clusterId), nil)
	}
	return nil
}

// DeleteWorkloadAndAssetsById deletes the workload row by id, then all workloads_assets rows for that workload id,
// matching the legacy backend functions.
func DeleteWorkloadAndAssetsById(ctx context.Context, clusterId string, workloadId string) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rowsAffected, err := tx.Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
		dbtables.WorkloadsColumnTenantId,
		dbtables.WorkloadsColumnClusterId,
		dbtables.WorkloadsColumnId),
		auth.TenantIdOf(ctx), clusterId, workloadId).
		Delete(&dbmodel.WorkloadRec{})
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to delete %s per update from %s", workloadId, clusterId), err)
	}
	if rowsAffected == 0 {
		return api.NewError(http.StatusNotFound,
			fmt.Sprintf("%s does not exist, ignore update from %s", workloadId, clusterId), nil)
	}

	_, err = tx.Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsAssetsColumnWorkloadId), workloadId).
		Delete(&dbmodel.WorkloadsAssetsRec{})
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to delete assets of %s per update from %s", workloadId, clusterId), err)
	}
	if err = tx.Commit(); err != nil {
		return api.NewError(http.StatusServiceUnavailable,
			fmt.Sprintf("failed to commit delete of %s following update from %s", workloadId, clusterId), err)
	}

	committed = true
	return nil
}
