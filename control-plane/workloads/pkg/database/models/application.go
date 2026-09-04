/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const (
	ApplicationTableName = "applications"
	ApplicationViewName  = "applications_view"

	ApplicationColumnId           = "id"
	ApplicationColumnTenantId     = "tenant_id"
	ApplicationColumnClusterId    = "cluster_id"
	ApplicationColumnUpdatedAt    = "updated_at"
	ApplicationColumnDesiredPhase = "desired_phase"
	ApplicationColumnActualPhase  = "actual_phase"
	ApplicationColumnMessage      = "message"
	ApplicationColumnDeletedAt    = "deleted_at"
	ApplicationColumnDeletedBy    = "deleted_by"
	ApplicationColumnUpdatedBy    = "updated_by"

	ApplicationPhaseCreating   = "creating"
	ApplicationPhaseValidating = "validating"
	ApplicationPhaseDeployed   = "deployed"
	ApplicationPhaseDeleting   = "deleting"
	ApplicationPhaseDeleted    = "deleted"
	ApplicationPhaseFailed     = "failed"
	ApplicationSourceRunai     = "runai"

	ApplicationCrudSyncTableName         = "application_crud_sync"
	ApplicationCrudSyncSequence          = "application_crud_sync_sequence"
	ApplicationCrudSyncSequenceLastVal   = "last_value"
	ApplicationCrudSyncColumnAppId       = "application_id"
	ApplicationCrudSyncColumnUpdateCount = "update_count"
)

func (Application) TableName() string {
	return ApplicationTableName
}

type Application struct {
	dbmodels.BaseModel
	Name              string
	Source            string
	Type              string
	ActualPhase       string
	TenantId          string
	ClusterId         string
	DepartmentId      string
	DepartmentName    string
	ProjectId         string
	ProjectName       string
	ApplicationStatus `gorm:"embedded"`

	Message *string
}

type ApplicationResource struct {
	Name        string
	Id          string
	ClusterName string
	ClusterId   string
	Kind        string
}

type ApplicationResourceList []ApplicationResource

type ApplicationStatus struct {
	K8sPhase  string
	Resources ApplicationResourceList `json:"resources" gorm:"type:jsonb"`
}

func (r ApplicationResourceList) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ApplicationResourceList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to convert database value to []byte")
	}
	return json.Unmarshal(bytes, r)
}

func (ApplicationView) TableName() string {
	return ApplicationViewName
}

type ApplicationView struct {
	dbmodels.BaseModel
	Name                string
	Source              string
	Type                string
	TenantId            string
	ClusterId           string
	DepartmentId        string
	DepartmentName      string
	ProjectId           string
	ProjectName         string
	Endpoints           WorkloadEndpointsArray           `gorm:"type:jsonb"`
	ResourceConnections WorkloadResourceConnectionsArray `gorm:"type:jsonb"`
	ApplicationStatus
	ApplicationStatistics

	Spec         *api.AiApplicationSpec `gorm:"serializer:json;type:jsonb"`
	ActualPhase  string
	DesiredPhase *string
	UpdateCount  *int64
	Message      *string
}

type ApplicationSyncUpdate struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	CreatedBy *string

	UpdateCount int64

	Name         string
	Source       string
	Type         string
	TenantId     string
	ClusterId    string
	DepartmentId string
	ProjectId    string
	ProjectName  string

	Spec         *api.AiApplicationSpec `gorm:"serializer:json;type:jsonb"`
	ActualPhase  string
	DesiredPhase *string
}

type ApplicationCrudSync struct {
	ApplicationId uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantId      string
	ClusterId     string
	Spec          *api.AiApplicationSpec `gorm:"serializer:json;type:jsonb"`
	DesiredPhase  *string
	UpdateCount   int64
}

func (ApplicationCrudSync) TableName() string {
	return ApplicationCrudSyncTableName
}

type ApplicationStatistics struct {
	RequestedGpuLimit         *float64
	RequestedGpuRequest       *float64
	RequestedGpuMemoryLimit   *pq.StringArray `gorm:"type:text[]"`
	RequestedGpuMemoryRequest *pq.StringArray `gorm:"type:text[]"`
	RequestedCpuLimit         *float64
	RequestedCpuRequest       *float64
	RequestedCpuMemoryLimit   *pq.StringArray `gorm:"type:text[]"`
	RequestedCpuMemoryRequest *pq.StringArray `gorm:"type:text[]"`
	AllocatedGpu              *float64
	AllocatedGpuMemory        *pq.StringArray `gorm:"type:text[]"`
	AllocatedCpu              *float64
	AllocatedCpuMemory        *pq.StringArray `gorm:"type:text[]"`
}

func (Application) FieldsConfiguration() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"clusterId":      queryapi.NewFieldMapping("cluster_id", queryapi.UUIDValidator{}),
		"name":           queryapi.NewFieldMapping("name"),
		"projectId":      queryapi.NewFieldMapping("project_id"),
		"projectName":    queryapi.NewFieldMapping("project_name"),
		"departmentId":   queryapi.NewFieldMapping("department_id"),
		"departmentName": queryapi.NewFieldMapping("department_name"),
		"source":         queryapi.NewFieldMapping("source"),
		"type":           queryapi.NewFieldMapping("type"),
		"createdAt":      queryapi.NewFieldMapping("created_at", queryapi.DateValidator{}),
	}
}
