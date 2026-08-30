package models

import (
	"time"

	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

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

func (Department) FieldsConfiguration() validators.FieldsValidation {
	return map[string]validators.FieldValidator{
		"name":                       validators.NewAlwaysValidFieldValidation("name"),
		"clusterId":                  validators.NewUUIDFieldValidation("cluster_id"),
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
