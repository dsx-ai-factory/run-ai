/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
)

type Department struct {
	IntBaseModel
	TenantId        string
	ClusterId       uuid.UUID
	Name            string
	Description     string
	ParentID        *int64
	OrgUnitDataID   uuid.UUID
	OrgUnitData     OrgUnitData `gorm:"foreignKey:OrgUnitDataID"`
	ProjectChildren []Project   `gorm:"foreignKey:ParentID"`
	DepartmentStatus
}

type DepartmentStatus struct {
	StatusLastUpdatedTime *time.Time
	QuotaStatus           QuotaStatus `gorm:"-"`
}

type DepartmentView struct {
	IntBaseModel
	TenantId      string
	ClusterId     uuid.UUID
	Name          string
	Description   string
	ParentID      *int64
	ParentName    *string
	OrgUnitDataID uuid.UUID
	OrgUnitDataView
}

type DepartmentExtendedView struct {
	IntBaseModel
	TenantId      string
	ClusterId     uuid.UUID
	Name          string
	Description   string
	ParentID      *int64
	ParentName    *string
	OrgUnitDataID uuid.UUID
	OrgUnitDataView
	QuotaStatusID uuid.UUID
	QuotaStatus
}

type DepartmentSyncView struct {
	IntBaseModel
	TenantId              string
	ClusterId             uuid.UUID
	Name                  string
	CalculatedUpdatedAt   time.Time
	ParentID              *int64
	ParentName            *string
	OrgUnitDataID         uuid.UUID
	StatusLastUpdatedTime *time.Time
	ResourcesJson         interface{} `gorm:"serializer:json;->"`
}

func (Department) TableName() string {
	return DepartmentTable
}

func (Department) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":                       queryapi.NewFieldMapping("name"),
		"clusterId":                  queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
		"totalGpuQuota":              queryapi.NewFieldMapping("total_gpu_quota"),
		"totalCpuQuota":              queryapi.NewFieldMapping("total_cpu_quota"),
		"totalMemoryQuota":           queryapi.NewFieldMapping("total_memory_quota"),
		"gpuAllocated":               queryapi.NewFieldMapping("gpu_allocated"),
		"createdAt":                  queryapi.NewFieldMapping("created_at", queryapi.DateValidator{}),
		"avgGpuAllocation24h":        queryapi.NewFieldMapping("avg_gpu_allocation24h"),
		"avgGpuUtilization24h":       queryapi.NewFieldMapping("avg_gpu_utilization24h"),
		"avgGpuMemoryUtilization24h": queryapi.NewFieldMapping("avg_gpu_memory_utilization24h"),
		"avgGpuAllocation7d":         queryapi.NewFieldMapping("avg_gpu_allocation7d"),
		"avgGpuUtilization7d":        queryapi.NewFieldMapping("avg_gpu_utilization7d"),
		"avgGpuMemoryUtilization7d":  queryapi.NewFieldMapping("avg_gpu_memory_utilization7d"),
		"avgGpuAllocation30d":        queryapi.NewFieldMapping("avg_gpu_allocation30d"),
		"avgGpuUtilization30d":       queryapi.NewFieldMapping("avg_gpu_utilization30d"),
		"avgGpuMemoryUtilization30d": queryapi.NewFieldMapping("avg_gpu_memory_utilization30d"),
	}
}
