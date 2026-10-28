/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Pod struct {
	ID                         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	CompletedAt                *time.Time
	DeletedAt                  gorm.DeletedAt
	WorkloadID                 uuid.UUID
	RevisionID                 *string
	Containers                 []*Container
	Name                       string
	PriorityClassName          string
	TenantID                   uint
	ClusterID                  uuid.UUID
	ProjectID                  string
	NodeName                   *string
	CurrentNodepool            *string
	RequestedNodepools         pq.StringArray `gorm:"type:text[]"`
	RequestedGPULimit          *float64
	RequestedGPURequest        *float64
	RequestedGPUMemoryLimit    *string
	RequestedGPUMemoryRequest  *string
	RequestedCPULimit          *float64
	RequestedCPURequest        *float64
	RequestedCPUMemoryLimit    *string
	RequestedCPUMemoryRequest  *string
	RequestedMigDevice         *string
	AllocatedGPU               *float64
	AllocatedGPUMemory         *string
	AllocatedCPU               *float64
	AllocatedCPUMemory         *string
	AllocatedMigDevice         *string
	Phase                      string
	PhaseUpdatedAt             time.Time
	Tolerations                TolerationArray `gorm:"type:jsonb"`
	Ip                         *string
	Task                       *string
	NodeAffinity               NodeAffinity          `gorm:"type:jsonb"`
	NetworkTopologyLabelValues *pq.StringArray       `gorm:"type:text[]"`
	ResourceClaims             PodResourceClaimArray `gorm:"type:jsonb"`
	Elements                   []*PodElement
}

type PodFull struct {
	Pod
	ProjectName  *string `gorm:"column:project_name"`
	WorkloadName *string `gorm:"column:workload_name"`
}

func (Pod) TableName() string {
	return "pods"
}

func (p Pod) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":            queryapi.NewFieldMapping("pods.name"),
		"clusterId":       queryapi.NewFieldMapping("pods.cluster_id"),
		"workloadId":      queryapi.NewFieldMapping("pods.workload_id"),
		"revisionId":      queryapi.NewFieldMapping("pods.revision_id"),
		"nodeName":        queryapi.NewFieldMapping("node_name"),
		"currentNodepool": queryapi.NewFieldMapping("current_nodepool"),
		"createdAt":       queryapi.NewFieldMapping("pods.created_at"),
		"completedAt":     queryapi.NewFieldMapping("completed_at"),
		"deletedAt":       queryapi.NewFieldMapping("pods.deleted_at"),
		"task":            queryapi.NewFieldMapping("task"),
		"elementId":       queryapi.NewFieldMapping("pod_elements.element_id"),
	}
}
