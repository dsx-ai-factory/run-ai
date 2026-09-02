/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
	"gorm.io/gorm"
	"k8s.io/utils/ptr"
)

type Project struct {
	IntBaseModel
	TenantId              string
	ClusterId             uuid.UUID
	Name                  string
	Description           string
	RequestedNamespace    string
	EnforceRunaiScheduler bool
	ParentID              *int64
	ParentName            *string
	DeletionRequest       *time.Time
	DeletionType          *api.DeletionTypes
	ForceDelete           *bool
	OrgUnitDataID         uuid.UUID
	OrgUnitData           OrgUnitData `gorm:"foreignKey:OrgUnitDataID"`
	ProjectStatus
}

type ProjectView struct {
	IntBaseModel
	TenantId              string
	ClusterId             uuid.UUID
	Name                  string
	Description           string
	RequestedNamespace    string
	EnforceRunaiScheduler bool
	ParentID              *int64
	ParentName            *string
	DeletionRequest       *time.Time
	DeletionType          *api.DeletionTypes
	ForceDelete           *bool
	OrgUnitDataID         uuid.UUID
	OrgUnitDataView
	ProjectStatusView
}

type ProjectSyncView struct {
	IntBaseModel
	TenantId              string
	ClusterId             uuid.UUID
	Name                  string
	CalculatedUpdatedAt   time.Time
	RequestedNamespace    string
	EnforceRunaiScheduler bool
	ParentID              *int64
	ParentName            *string
	DeletionRequest       *time.Time
	DeletionType          *api.DeletionTypes
	ForceDelete           *bool
	OrgUnitDataID         uuid.UUID
	StatusLastUpdatedTime *time.Time
	OrgUnitDataSyncView
}

type ProjectStatus struct {
	StatusLastUpdatedTime *time.Time
	Phase                 *string
	PhaseMessage          *string
	StatusNamespace       *string
	Conditions            ConditionArray `gorm:"type:jsonb"`
	QuotaStatusID         uuid.UUID
	QuotaStatus           QuotaStatus            `gorm:"foreignKey:QuotaStatusID"`
	QuotaStatusRaw        map[string]interface{} `gorm:"serializer:json;->"`
}

type ProjectStatusView struct {
	Phase           *string
	PhaseMessage    *string
	StatusNamespace *string
	Conditions      ConditionArray `gorm:"type:jsonb"`
	QuotaStatusID   uuid.UUID
	QuotaStatus
}

func (Project) TableName() string {
	return ProjectTable
}

func (Project) FieldsConfiguration() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"name":                       validators.NewAlwaysValidFieldValidation("name"),
		"clusterId":                  validators.NewUUIDFieldValidation("cluster_id"),
		"parentId":                   validators.NewAlwaysValidFieldValidation("parent_id"),
		"parentName":                 validators.NewAlwaysValidFieldValidation("parent_name"),
		"phase":                      validators.NewAlwaysValidFieldValidation("phase"),
		"totalGpuQuota":              validators.NewAlwaysValidFieldValidation("total_gpu_quota"),
		"totalCpuQuota":              validators.NewAlwaysValidFieldValidation("total_cpu_quota"),
		"totalMemoryQuota":           validators.NewAlwaysValidFieldValidation("total_memory_quota"),
		"gpuAllocated":               validators.NewAlwaysValidFieldValidation("gpu_allocated"),
		"createdAt":                  validators.NewDateFieldValidation("created_at"),
		"avgGpuAllocation24h":        validators.NewAlwaysValidFieldValidation("avg_gpu_allocation24h"),
		"avgGpuUtilization24h":       validators.NewAlwaysValidFieldValidation("avg_gpu_utilization24h"),
		"avgGpuMemoryUtilization24h": validators.NewAlwaysValidFieldValidation("avg_gpu_memory_utilization24h"),
		"avgGpuAllocation7d":         validators.NewAlwaysValidFieldValidation("avg_gpu_allocation7d"),
		"avgGpuUtilization7d":        validators.NewAlwaysValidFieldValidation("avg_gpu_utilization7d"),
		"avgGpuMemoryUtilization7d":  validators.NewAlwaysValidFieldValidation("avg_gpu_memory_utilization7d"),
		"avgGpuAllocation30d":        validators.NewAlwaysValidFieldValidation("avg_gpu_allocation30d"),
		"avgGpuUtilization30d":       validators.NewAlwaysValidFieldValidation("avg_gpu_utilization30d"),
		"avgGpuMemoryUtilization30d": validators.NewAlwaysValidFieldValidation("avg_gpu_memory_utilization30d"),
	}
}

func (p Project) Clone() Project {
	return Project{
		IntBaseModel: IntBaseModel{
			ID:        p.ID,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
			DeletedAt: gorm.DeletedAt{
				Time:  p.DeletedAt.Time,
				Valid: p.DeletedAt.Valid,
			},
			CreatedBy: p.CreatedBy,
			UpdatedBy: p.UpdatedBy,
			DeletedBy: p.DeletedBy,
		},
		TenantId:              p.TenantId,
		ClusterId:             p.ClusterId,
		Name:                  p.Name,
		EnforceRunaiScheduler: p.EnforceRunaiScheduler,
		ParentID:              CopyPtr(p.ParentID),
		ParentName:            CopyPtr(p.ParentName),
		OrgUnitDataID:         p.OrgUnitDataID,
		OrgUnitData:           p.OrgUnitData.Clone(),
	}
}

func CopyPtr[T any](source *T) *T {
	if source == nil {
		return nil
	}

	val := *source
	return ptr.To(val)
}

func ProjectsEqual(p1, p2 Project) bool {
	if p1.TenantId != p2.TenantId ||
		p1.ClusterId != p2.ClusterId ||
		p1.Name != p2.Name ||
		p1.RequestedNamespace != p2.RequestedNamespace ||
		p1.OrgUnitDataID != p2.OrgUnitDataID {
		return false
	}

	// Compare OrgUnitData
	if !OrgUnitDataEqual(p1.OrgUnitData, p2.OrgUnitData) {
		return false
	}

	return true
}

func ProjectsV1Equal(p1, p2 Project) bool {
	if p1.TenantId != p2.TenantId ||
		p1.ClusterId != p2.ClusterId ||
		p1.Name != p2.Name ||
		p1.RequestedNamespace != p2.RequestedNamespace ||
		p1.OrgUnitDataID != p2.OrgUnitDataID {
		return false
	}

	// Compare OrgUnitData
	if !OrgUnitDataV1Equal(p1.OrgUnitData, p2.OrgUnitData) {
		return false
	}

	return true
}
