/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const ResponseFieldSpec = "spec"

func NewWorkloadFromWorkloadParams(params *createWorkloadParams, assetsMapping []workloadsApi.AssetsMapping) *models_crud.Workload {
	result := models_crud.Workload{
		ID: params.WorkloadId,

		CreatedAt: params.Response.CreatedAt.Truncate(time.Microsecond),
		CreatedBy: params.Response.CreatedBy,
		UpdatedAt: params.Response.CreatedAt.Truncate(time.Microsecond),
		UpdatedBy: params.Response.CreatedBy,

		TenantId:     params.TenantId,
		ClusterId:    params.Response.ClusterId,
		ProjectId:    params.Response.ProjectId,
		DepartmentId: convutils.Deref(params.Response.DepartmentId, ""),

		WorkloadType: params.WorkloadType,

		RequestedName: params.Response.RequestedName,

		Name: params.Response.Name,

		Spec:       params.Response.Spec,
		MasterSpec: params.Response.MasterSpec.Get(),

		MasterSameAsWorker: params.Response.MasterSpecSameAsWorker.Get(),

		DesiredPhase: &params.Response.DesiredPhase,

		AssetsMapping: assetsMapping,
	}

	return &result
}

func NewWorkloadV2SnapshotFromWorkloadParams(params *createWorkloadParams, assetsMapping []workloadsApi.AssetsMapping, namespace string) *models_crud.WorkloadV2Snapshot {
	result := models_crud.WorkloadV2Snapshot{
		ID: params.WorkloadId,

		CreatedAt: params.Response.CreatedAt.Truncate(time.Microsecond),
		CreatedBy: params.Response.CreatedBy,
		UpdatedAt: params.Response.CreatedAt.Truncate(time.Microsecond),
		UpdatedBy: params.Response.CreatedBy,

		TenantId:     params.TenantId,
		ClusterId:    params.Response.ClusterId,
		ProjectId:    params.Response.ProjectId,
		DepartmentId: convutils.Deref(params.Response.DepartmentId, ""),

		WorkloadType: params.WorkloadType,

		RequestedName: params.Response.RequestedName,

		Name:      params.Response.Name,
		Namespace: namespace,

		Spec:       params.Response.Spec,
		MasterSpec: params.Response.MasterSpec.Get(),

		MasterSameAsWorker: params.Response.MasterSpecSameAsWorker.Get(),

		AssetsMapping: assetsMapping,
	}

	return &result
}

func WorkloadFromDbRecord(record *models_crud.Workload, phase workloadsApi.Phase) workloadsApi.SupersetWorkload {
	response := workloadsApi.SupersetWorkload{
		Name:          record.Name,
		RequestedName: record.RequestedName,
		WorkloadId:    record.ID.String(),
		ProjectId:     record.ProjectId,
		DepartmentId:  &record.DepartmentId,
		ClusterId:     record.ClusterId,
		CreatedBy:     record.CreatedBy,
		CreatedAt:     record.CreatedAt.Truncate(time.Microsecond),
		DesiredPhase:  convutils.Deref(record.DesiredPhase, ""),
		ActualPhase:   &phase,

		AssetsMapping: record.AssetsMapping,
	}

	if record.DeletedAt.Valid && !record.DeletedAt.Time.IsZero() {
		response.DeletedAt = *workloadsApi.NewNullableTime(&record.DeletedAt.Time)
	}

	if record.Spec != nil {
		response.Spec = gpucompat.LegacyCompatibleWorkloadSpec(record.Spec)
	}

	if record.MasterSpec != nil {
		response.MasterSpec.Set(gpucompat.LegacyCompatibleMasterSpec(record.MasterSpec))
	}

	if convutils.BoolIsTrue(record.MasterSameAsWorker) {
		response.MasterSpecSameAsWorker.Set(convutils.Ref(true))
	}

	return response
}

// workloadFromV2Snapshot builds the public response from the typed snapshot.
// Desired phase is supplied separately because it is owned by workloads-manager
// and is not persisted in the snapshot.
func workloadFromV2Snapshot(snapshot *models_crud.WorkloadV2Snapshot, desiredPhase workloadsApi.WorkloadDesiredPhase) workloadsApi.SupersetWorkload {
	response := WorkloadFromDbRecord(snapshot.ToWorkload(), workloadsApi.PHASE_UNKNOWN)
	response.ActualPhase = nil
	response.DesiredPhase = desiredPhase
	return response
}

func WorkloadSyncUpdateFromDbRecord(record *models_crud.WorkloadSyncUpdate) api.SupersetWorkload {
	result := api.SupersetWorkload{
		Name:          record.Name,
		RequestedName: record.RequestedName,
		WorkloadId:    record.ID.String(),
		ProjectId:     record.ProjectId,
		DepartmentId:  record.DepartmentId,
		ClusterId:     record.ClusterId,
		CreatedBy:     record.CreatedBy,
		CreatedAt:     record.CreatedAt.Truncate(time.Microsecond),
		DesiredPhase:  convutils.Deref(record.DesiredPhase, ""),

		Spec: record.Spec,

		MasterSpecSameAsWorker: record.MasterSameAsWorker,

		AssetsMapping: convutils.Ref(record.AssetsMapping),
	}

	if record.MasterSpec != nil {
		result.MasterSpec = &api.MasterSpec{}
		convutils.CopySpecs(result.MasterSpec, record.MasterSpec)
	}

	return result
}
