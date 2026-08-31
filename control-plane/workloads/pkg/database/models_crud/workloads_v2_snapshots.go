/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

import (
	"time"

	"github.com/google/uuid"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const (
	WorkloadV2SnapshotsTableName      = "workloads_v2_snapshots"
	WorkloadV2SnapshotsColumnTenantId = "tenant_id"
	WorkloadV2SnapshotsColumnId       = "id"
	WorkloadV2SnapshotsColumnType     = "workload_type"
)

type WorkloadV2Snapshot struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
	CreatedBy string
	UpdatedAt time.Time `gorm:"autoUpdateTime:false"`
	UpdatedBy string

	TenantId     int32
	ClusterId    string
	DepartmentId string
	ProjectId    string

	WorkloadType string

	RequestedName string

	Name      string
	Namespace string

	Spec       *workloadsApi.SupersetSpec `gorm:"serializer:json; type:jsonb"`
	MasterSpec *workloadsApi.MasterSpec   `gorm:"serializer:json; type:jsonb"`

	MasterSameAsWorker *bool

	AssetsMapping []workloadsApi.AssetsMapping `gorm:"serializer:json; type:jsonb"`
}

func (WorkloadV2Snapshot) TableName() string {
	return WorkloadV2SnapshotsTableName
}

func (s WorkloadV2Snapshot) ToWorkload() *Workload {
	return &Workload{
		ID:                 s.ID,
		CreatedAt:          s.CreatedAt,
		CreatedBy:          s.CreatedBy,
		UpdatedAt:          s.UpdatedAt,
		UpdatedBy:          s.UpdatedBy,
		TenantId:           s.TenantId,
		ClusterId:          s.ClusterId,
		DepartmentId:       s.DepartmentId,
		ProjectId:          s.ProjectId,
		WorkloadType:       s.WorkloadType,
		RequestedName:      s.RequestedName,
		Name:               s.Name,
		Namespace:          s.Namespace,
		Spec:               s.Spec,
		MasterSpec:         s.MasterSpec,
		MasterSameAsWorker: s.MasterSameAsWorker,
		AssetsMapping:      s.AssetsMapping,
	}
}
