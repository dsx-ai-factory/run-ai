/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	api "github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
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

func (WorkloadView) FilterFieldsToColumn() map[string]string {
	return map[string]string{
		"type":                    "type",
		"name":                    "name",
		"projectName":             "project_name",
		"createdAt":               "created_at",
		"deletedAt":               "deleted_at",
		"submittedBy":             "submitted_by",
		"phase":                   "phase",
		"projectId":               "project_id",
		"departmentId":            "department_id",
		"departmentName":          "department_name",
		"clusterId":               "cluster_id",
		"completedAt":             "completed_at",
		"allocatedGPU":            "allocated_gpu",
		"nodepool":                "current_nodepools[1]",
		"distributedFramework":    "distributed_framework",
		"idleAllocatedGpus":       "idle_allocated_gpus",
		"phaseUpdatedAt":          "phase_updated_at",
		"totalPendingTimeSeconds": "total_pending_time_seconds",
		"totalRunningTimeSeconds": "total_running_time_seconds",
		"category":                "category",
		"priority":                "priority",
		"priorityClassName":       "priority_class_name",
		"guaranteedRuntimeEndsAt": "guaranteed_runtime_ends_at",
		"aiApplicationId":         "application_id",
		"aiApplicationName":       "application_name",
	}
}

func (wv WorkloadView) SortFieldsToColumn() map[string]string {
	filterMap := wv.FilterFieldsToColumn()
	filterMap["phase"] = "phase %s, phase_updated_at %s"
	return filterMap
}
