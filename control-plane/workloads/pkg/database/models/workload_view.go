/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type WorkloadView struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
	//nolint:staticcheck // pq.NullTime matches the existing DB column mapping
	CompletedAt                   pq.NullTime
	SubmittedBy                   string
	Name                          string
	ExternalWorkloadName          string
	Source                        string
	PriorityClassName             string
	Priority                      *int32
	Preemptibility                *string
	Type                          string
	TenantID                      uint
	ClusterID                     uuid.UUID
	ProjectName                   string
	ProjectID                     string
	DepartmentName                string
	DepartmentID                  string
	Namespace                     string
	NumberOfRunningPods           int32
	RequestedGPULimit             *float64
	RequestedGPURequest           *float64
	RequestedGPUMemoryLimit       *string
	RequestedGPUMemoryRequest     *string
	RequestedCPULimit             *float64
	RequestedCPURequest           *float64
	RequestedCPUMemoryLimit       *string
	RequestedCPUMemoryRequest     *string
	RequestedMigDevice            *pq.StringArray `gorm:"type:text[]"`
	PodsRequestedGPULimit         *float64
	PodsRequestedGPURequest       *float64
	PodsRequestedGPUMemoryLimit   *pq.StringArray `gorm:"type:text[]"`
	PodsRequestedGPUMemoryRequest *pq.StringArray `gorm:"type:text[]"`
	PodsRequestedCPULimit         *float64
	PodsRequestedCPURequest       *float64
	PodsRequestedCPUMemoryLimit   *pq.StringArray `gorm:"type:text[]"`
	PodsRequestedCPUMemoryRequest *pq.StringArray `gorm:"type:text[]"`
	PodsRequestedMigDevice        *pq.StringArray `gorm:"type:text[]"`
	AllocatedGPU                  *float64
	AllocatedGPUMemory            *pq.StringArray `gorm:"type:text[]"`
	AllocatedCPU                  *float64
	AllocatedCPUMemory            *pq.StringArray `gorm:"type:text[]"`
	AllocatedMigDevice            *pq.StringArray `gorm:"type:text[]"`
	SupportDelete                 bool
	SupportSuspend                bool
	Phase                         api.Phase
	PhaseMessage                  string
	PhaseUpdatedAt                time.Time
	K8sPhase                      string
	K8sPhaseUpdatedAt             time.Time
	RequestedPodsNumber           *int32
	RequestedPodsMin              *int32
	RequestedPodsMax              *int32
	RequestedPodsParallelism      *int32
	RequestedPodsCompletions      *int32
	RequestedNodepools            pq.StringArray             `gorm:"type:text[]"`
	CurrentNodepools              pq.StringArray             `gorm:"type:text[]"`
	Images                        pq.StringArray             `gorm:"type:text[]"`
	Urls                          pq.StringArray             `gorm:"type:text[]"`
	NetworkStatus                 *api.WorkloadNetworkStatus `gorm:"serializer:json;type:jsonb"`
	Conditions                    ConditionArray             `gorm:"type:jsonb"`
	Datasources                   DatasourceArray            `gorm:"type:jsonb"`
	Environments                  EnvironmentArray           `gorm:"type:jsonb"`
	ExternalConnections           ExternalConnectionArray    `gorm:"type:jsonb"`
	Endpoints                     EndpointArray              `gorm:"type:jsonb"`
	DistributedFramework          string
	AdditionalFields              AdditionalFields `gorm:"type:jsonb"`
	Preemptible                   *bool
	FullySaasWorkload             bool
	Command                       string
	Arguments                     string
	EnvironmentVariables          EnvironmentVariables     `gorm:"type:jsonb"`
	SchedulingConditions          SchedulingConditionArray `gorm:"type:jsonb"`
	PhaseReason                   *string
	IdleAllocatedGpus             *float32
	IdleSince                     *time.Time
	TotalPendingTimeSeconds       *int32
	TotalRunningTimeSeconds       *int32
	Category                      string
	GuaranteedRuntimeEndsAt       *time.Time
	ApplicationId                 *string
	ApplicationName               *string
	SourceApi                     *api.SourceApi
	Topology                      *Topology                 `gorm:"type:jsonb"`
	StructureElementsTotalCount   *int                      `gorm:"column:structure_elements_total_count"`
	NodeAffinities                WorkloadNodeAffinityArray `gorm:"serializer:json;type:jsonb"`
	K8sObjectGroup                *string                   `gorm:"column:k8s_object_group"`
	K8sObjectVersion              *string                   `gorm:"column:k8s_object_version"`
}

func (WorkloadView) TableName() string {
	return "workloads_view"
}

func (WorkloadView) FilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"type":                    queryapi.NewFieldMapping("type"),
		"name":                    queryapi.NewFieldMapping("name"),
		"projectName":             queryapi.NewFieldMapping("project_name"),
		"createdAt":               queryapi.NewFieldMapping("created_at"),
		"deletedAt":               queryapi.NewFieldMapping("deleted_at"),
		"submittedBy":             queryapi.NewFieldMapping("submitted_by"),
		"phase":                   queryapi.NewFieldMapping("phase"),
		"projectId":               queryapi.NewFieldMapping("project_id"),
		"departmentId":            queryapi.NewFieldMapping("department_id"),
		"departmentName":          queryapi.NewFieldMapping("department_name"),
		"clusterId":               queryapi.NewFieldMapping("cluster_id"),
		"completedAt":             queryapi.NewFieldMapping("completed_at"),
		"allocatedGPU":            queryapi.NewFieldMapping("allocated_gpu"),
		"nodepool":                queryapi.NewFieldMapping("current_nodepools[1]"),
		"distributedFramework":    queryapi.NewFieldMapping("distributed_framework"),
		"idleAllocatedGpus":       queryapi.NewFieldMapping("idle_allocated_gpus"),
		"phaseUpdatedAt":          queryapi.NewFieldMapping("phase_updated_at"),
		"totalPendingTimeSeconds": queryapi.NewFieldMapping("total_pending_time_seconds"),
		"totalRunningTimeSeconds": queryapi.NewFieldMapping("total_running_time_seconds"),
		"category":                queryapi.NewFieldMapping("category"),
		"priority":                queryapi.NewFieldMapping("priority"),
		"priorityClassName":       queryapi.NewFieldMapping("priority_class_name"),
		"preemptibility":          queryapi.NewFieldMapping("preemptibility"),
		"guaranteedRuntimeEndsAt": queryapi.NewFieldMapping("guaranteed_runtime_ends_at"),
		"aiApplicationId":         queryapi.NewFieldMapping("application_id"),
		"aiApplicationName":       queryapi.NewFieldMapping("application_name"),
	}
}

func (wv WorkloadView) SortFieldsToColumn() queryapi.APIFieldsMapping {
	mapping := wv.FilterFieldsToColumn()
	mapping["phase"] = queryapi.NewFieldMapping("phase %s, phase_updated_at %s")
	return mapping
}
