/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type InternalCreatedWorkload struct {
	Datasources         DatasourceArray         `gorm:"type:jsonb"`
	Environments        EnvironmentArray        `gorm:"type:jsonb"`
	ExternalConnections ExternalConnectionArray `gorm:"type:jsonb"`
	SourceApi           *api.SourceApi
	Preemptibility      *string
}

func (InternalCreatedWorkload) TableName() string {
	return "workloads"
}

type Workload struct {
	ID                          uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt                   time.Time `gorm:"<-:create"`
	UpdatedAt                   time.Time
	DeletedAt                   gorm.DeletedAt
	SubmittedBy                 string
	Name                        string
	ExternalWorkloadName        string
	Source                      string
	PriorityClassName           string
	Preemptibility              *string
	Type                        string
	TenantID                    uint
	ClusterID                   uuid.UUID
	ProjectName                 string
	ProjectID                   string
	DepartmentName              string
	DepartmentID                string
	CategoryID                  *uuid.UUID `gorm:"type:uuid"`
	Namespace                   string
	RequestedGPULimit           *float64
	RequestedGPURequest         *float64
	RequestedGPUMemoryLimit     *string
	RequestedGPUMemoryRequest   *string
	RequestedCPULimit           *float64
	RequestedCPURequest         *float64
	RequestedCPUMemoryLimit     *string
	RequestedCPUMemoryRequest   *string
	RequestedMigDevice          pq.StringArray `gorm:"type:text[]"`
	SupportDelete               bool
	SupportSuspend              bool
	Phase                       api.Phase
	PhaseMessage                string
	PhaseUpdatedAt              time.Time
	K8sPhase                    string
	K8sPhaseUpdatedAt           time.Time
	RequestedPodsNumber         *int32
	RequestedPodsMin            *int32
	RequestedPodsMax            *int32
	RequestedPodsParallelism    *int32
	RequestedPodsCompletions    *int32
	RequestedNodepools          pq.StringArray `gorm:"type:text[]"`
	Images                      pq.StringArray `gorm:"type:text[]"`
	DistributedFramework        string
	Urls                        pq.StringArray             `gorm:"type:text[]"`
	NetworkStatus               *api.WorkloadNetworkStatus `gorm:"serializer:json;type:jsonb"`
	Conditions                  ConditionArray             `gorm:"type:jsonb"`
	Priority                    *int32
	AdditionalFields            AdditionalFields `gorm:"type:jsonb"`
	FullySaasWorkload           bool
	Command                     string
	Arguments                   string
	EnvironmentVariables        EnvironmentVariables     `gorm:"type:jsonb"`
	SchedulingConditions        SchedulingConditionArray `gorm:"type:jsonb"`
	PhaseReason                 *string
	GuaranteedRuntimeEndsAt     *time.Time
	ApplicationId               *string
	ApplicationName             *string
	Topology                    *Topology `gorm:"serializer:json; type:jsonb"`
	IdleSince                   *time.Time
	ExternalConnections         ExternalConnectionArray   `gorm:"type:jsonb"`
	StructureElementsTotalCount *int                      `gorm:"column:structure_elements_total_count"`
	Endpoints                   EndpointArray             `gorm:"type:jsonb"`
	NodeAffinities              WorkloadNodeAffinityArray `gorm:"serializer:json;type:jsonb"`
	K8sObjectGroup              *string                   `gorm:"column:k8s_object_group"`
	K8sObjectVersion            *string                   `gorm:"column:k8s_object_version"`
	ResourceRefs                WorkloadResourceRefArray  `gorm:"type:jsonb"`
}

func (w Workload) TableName() string {
	return "workloads"
}

// filterableColumns -
// out of all filterable column options (see WorkloadView.FilterFieldsToColumn),
// returns the columns that are filterable from the table itself (because some columns are only in the view).
func (Workload) filterableColumns() map[string]bool {
	return map[string]bool{
		"type":                    true,
		"name":                    true,
		"projectName":             true,
		"createdAt":               true,
		"deletedAt":               true,
		"submittedBy":             true,
		"phase":                   true,
		"projectId":               true,
		"departmentId":            true,
		"departmentName":          true,
		"clusterId":               true,
		"distributedFramework":    true,
		"phaseUpdatedAt":          true,
		"priority":                true,
		"priorityClassName":       true,
		"guaranteedRuntimeEndsAt": true,
	}
}

// AreFilterableFromTable - checks if the requested filters are filterable from the table,
// if not - means we need to query the view instead of the table.
func (Workload) AreFilterableFromTable(filterBy []string) bool {
	filterableColumns := Workload{}.filterableColumns()

	for _, filter := range filterBy {
		if _, ok := filterableColumns[filter]; !ok {
			return false
		}
	}
	return true
}
