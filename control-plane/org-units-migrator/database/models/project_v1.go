/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/google/uuid"
)

type ProjectV1 struct {
	Id                                  int       `gorm:"column:project_id;primaryKey"`
	Name                                string    `gorm:"column:project_name"`
	TenantId                            uint      `gorm:"column:tenant_id"`
	ClusterUuid                         uuid.UUID `gorm:"column:cluster_uuid"`
	CreatedAt                           time.Time `gorm:"column:created_at"`
	DepartmentId                        int64     `gorm:"column:department_id"`
	DepartmentName                      string    `gorm:"column:department_name"`
	Status                              *string   `gorm:"type:jsonb;column:status"`
	Phase                               *string   `gorm:"column:phase"`
	Namespace                           string    `gorm:"column:namespace"`
	CreatedBy                           string    `gorm:"column:created_by"`
	UpdatedBy                           string    `gorm:"column:updated_by"`
	InteractiveJobTimeLimitSec          int64     `gorm:"column:interactive_job_time_limit_secs"`
	InteractiveJobMaxIdleSec            int64     `gorm:"column:interactive_job_max_idle_duration_secs"`
	InteractivePreemptibleJobMaxIdleSec int64     `gorm:"column:interactive_preemptible_job_max_idle_duration_secs"`
	TrainingJobMaxIdleSec               int64     `gorm:"column:training_job_max_idle_duration_secs"`
	TrainingJobTimeLimitSec             int64     `gorm:"column:training_job_time_limit_secs"`
	NodeAffinityTrainingIds             *string   `gorm:"type:jsonb;column:node_affinity.train.selected_type_ids"`
	NodeAffinityWorkspaceIds            *string   `gorm:"type:jsonb;column:node_affinity.interactive.selected_type_ids"`
	CPUDeserved                         *float32  `gorm:"column:node_pool_resources.cpu_deserved"`
	CPUMaxAllowed                       *float32  `gorm:"column:node_pool_resources.cpu_max_allowed"`
	CPUOverQuotaWeight                  *float32  `gorm:"column:node_pool_resources.cpu_over_quota_weight"`
	MemoryDeserved                      *float32  `gorm:"column:node_pool_resources.memory_deserved"`
	MemoryMaxAllowed                    *float32  `gorm:"column:node_pool_resources.memory_max_allowed"`
	MemoryOverQuotaWeight               *float32  `gorm:"column:node_pool_resources.memory_over_quota_weight"`
	GPUDeserved                         *float32  `gorm:"column:node_pool_resources.gpu_deserved"`
	GPUMaxAllowed                       *float32  `gorm:"column:node_pool_resources.gpu_max_allowed"`
	GPUOverQuotaWeight                  *float32  `gorm:"column:node_pool_resources.gpu_over_quota_weight"`
	CPUAllocated                        float32   `gorm:"column:quota_resources_status.cpu_allocated"`
	CPUAllocatedNonPreemptible          float32   `gorm:"column:quota_resources_status.cpu_allocated_non_preemptible"`
	CPURequested                        float32   `gorm:"column:quota_resources_status.cpu_requested"`
	GPUAllocated                        float32   `gorm:"column:quota_resources_status.gpu_allocated"`
	GPUAllocatedNonPreemptible          float32   `gorm:"column:quota_resources_status.gpu_allocated_non_preemptible"`
	GPURequested                        float32   `gorm:"column:quota_resources_status.gpu_requested"`
	MemoryAllocated                     float32   `gorm:"column:quota_resources_status.memory_allocated"`
	MemoryAllocatedNonPreemptible       float32   `gorm:"column:quota_resources_status.memory_allocated_non_preemptible"`
	MemoryRequested                     float32   `gorm:"column:quota_resources_status.memory_requested"`
	NodePoolID                          uint      `gorm:"column:node_pool_resources.node_pool_id"`
	NodePoolName                        string    `gorm:"column:node_pool_resources.node_pool_name"`
	DefaultNodePoolIndex                int32     `gorm:"column:project_to_node_pool.default_node_pool_list_index"`
}

func (ProjectV1) TableName() string {
	return "projects_new_view"
}

// ProjectV1Grouped In projects view there is a row per project-nodepool, so we want to group it to a single model per project
type ProjectV1Grouped struct {
	Id                                  int                       `gorm:"column:project_id;primaryKey"`
	Name                                string                    `gorm:"column:project_name"`
	TenantId                            uint                      `gorm:"column:tenant_id"`
	ClusterUuid                         uuid.UUID                 `gorm:"column:cluster_uuid"`
	CreatedAt                           time.Time                 `gorm:"column:created_at"`
	DepartmentId                        int64                     `gorm:"column:department_id"`
	DepartmentName                      string                    `gorm:"column:department_name"`
	Status                              *string                   `gorm:"column:status"`
	Phase                               *string                   `gorm:"column:phase"`
	Namespace                           string                    `gorm:"column:namespace"`
	CreatedBy                           string                    `gorm:"column:created_by"`
	UpdatedBy                           string                    `gorm:"column:updated_by"`
	InteractiveJobTimeLimitSec          int64                     `gorm:"column:interactive_job_time_limit_secs"`
	InteractiveJobMaxIdleSec            int64                     `gorm:"column:interactive_job_max_idle_duration_secs"`
	InteractivePreemptibleJobMaxIdleSec int64                     `gorm:"column:interactive_preemptible_job_max_idle_duration_secs"`
	TrainingJobMaxIdleSec               int64                     `gorm:"column:training_job_max_idle_duration_secs"`
	TrainingJobTimeLimitSec             int64                     `gorm:"column:training_job_time_limit_secs"`
	NodeAffinityTrainingIds             []string                  `gorm:"column:node_affinity.train.selected_type_ids"`
	NodeAffinityWorkspaceIds            []string                  `gorm:"column:node_affinity.interactive.selected_type_ids"`
	NodepoolResources                   []AssignedResource        `gorm:"-"`
	QuotaStatuses                       *[]api.QuotaStatusV1Inner `gorm:"-"`
	DefaultNodePools                    *[]string                 `gorm:"-"`
}

// ProjectV1ForTest gorm auto migrate does not work well with columns that has a dot in their name.
// We duplicated the struct for test purposes only, it allow us to create and insert data to the table, but fetching the data is using the regular struct.
type ProjectV1ForTest struct {
	Id                                  int       `gorm:"column:project_id;uniqueIndex:p_np"`
	Name                                string    `gorm:"column:project_name"`
	TenantId                            uint      `gorm:"column:tenant_id"`
	ClusterUuid                         uuid.UUID `gorm:"column:cluster_uuid"`
	CreatedAt                           time.Time `gorm:"column:created_at"`
	DepartmentId                        int64     `gorm:"column:department_id"`
	Status                              *string   `gorm:"column:status"`
	Phase                               *string   `gorm:"column:phase"`
	Namespace                           string    `gorm:"column:namespace"`
	CreatedBy                           string    `gorm:"column:created_by"`
	UpdatedBy                           string    `gorm:"column:updated_by"`
	InteractiveJobTimeLimitSec          int64     `gorm:"column:interactive_job_time_limit_secs"`
	InteractiveJobMaxIdleSec            int64     `gorm:"column:interactive_job_max_idle_duration_secs"`
	InteractivePreemptibleJobMaxIdleSec int64     `gorm:"column:interactive_preemptible_job_max_idle_duration_secs"`
	TrainingJobMaxIdleSec               int64     `gorm:"column:training_job_max_idle_duration_secs"`
	TrainingJobTimeLimitSec             int64     `gorm:"column:training_job_time_limit_secs"`
	NodeAffinityTrainingIds             *string   `gorm:"column:\"node_affinity.train.selected_type_ids\""`
	NodeAffinityWorkspaceIds            *string   `gorm:"column:\"node_affinity.interactive.selected_type_ids\""`
	CPUDeserved                         *float32  `gorm:"column:\"node_pool_resources.cpu_deserved\""`
	CPUMaxAllowed                       *float32  `gorm:"column:\"node_pool_resources.cpu_max_allowed\""`
	CPUOverQuotaWeight                  *float32  `gorm:"column:\"node_pool_resources.cpu_over_quota_weight\""`
	MemoryDeserved                      *float32  `gorm:"column:\"node_pool_resources.memory_deserved\""`
	MemoryMaxAllowed                    *float32  `gorm:"column:\"node_pool_resources.memory_max_allowed\""`
	MemoryOverQuotaWeight               *float32  `gorm:"column:\"node_pool_resources.memory_over_quota_weight\""`
	GPUDeserved                         *float32  `gorm:"column:\"node_pool_resources.gpu_deserved\""`
	GPUMaxAllowed                       *float32  `gorm:"column:\"node_pool_resources.gpu_max_allowed\""`
	GPUOverQuotaWeight                  *float32  `gorm:"column:\"node_pool_resources.gpu_over_quota_weight\""`
	CPUAllocated                        float32   `gorm:"column:\"quota_resources_status.cpu_allocated\""`
	CPUAllocatedNonPreemptible          float32   `gorm:"column:\"quota_resources_status.cpu_allocated_non_preemptible\""`
	CPURequested                        float32   `gorm:"column:\"quota_resources_status.cpu_requested\""`
	GPUAllocated                        float32   `gorm:"column:\"quota_resources_status.gpu_allocated\""`
	GPUAllocatedNonPreemptible          float32   `gorm:"column:\"quota_resources_status.gpu_allocated_non_preemptible\""`
	GPURequested                        float32   `gorm:"column:\"quota_resources_status.gpu_requested\""`
	MemoryAllocated                     float32   `gorm:"column:\"quota_resources_status.memory_allocated\""`
	MemoryAllocatedNonPreemptible       float32   `gorm:"column:\"quota_resources_status.memory_allocated_non_preemptible\""`
	MemoryRequested                     float32   `gorm:"column:\"quota_resources_status.memory_requested\""`
	NodePoolID                          uint      `gorm:"column:\"node_pool_resources.node_pool_id\";uniqueIndex:p_np"`
	NodePoolName                        string    `gorm:"column:\"node_pool_resources.node_pool_name\""`
	DefaultNodePoolIndex                int32     `gorm:"column:\"project_to_node_pool.default_node_pool_list_index\""`
}

func (ProjectV1ForTest) TableName() string {
	return "projects_new_view"
}
