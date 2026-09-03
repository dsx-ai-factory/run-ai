/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"gorm.io/gorm"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RevisionView struct {
	ID                            uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt                     time.Time
	DeletedAt                     gorm.DeletedAt
	Name                          string
	Type                          string
	WorkloadID                    uuid.UUID
	TenantID                      uint
	ClusterID                     uuid.UUID
	ProjectID                     string
	DepartmentID                  string
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
	Phase                         api.Phase
	PhaseMessage                  string
	PhaseUpdatedAt                time.Time
	RequestedPodsNumber           *int32
	RequestedPodsMin              *int32
	RequestedPodsMax              *int32
	RequestedPodsParallelism      *int32
	RequestedPodsCompletions      *int32
	RequestedNodepools            pq.StringArray `gorm:"type:text[]"`
	CurrentNodepools              pq.StringArray `gorm:"type:text[]"`
	NumberOfPods                  int32
	NumberOfRunningPods           int32
	Images                        pq.StringArray `gorm:"type:text[]"`
	Command                       string
	Arguments                     string
	EnvironmentVariables          EnvironmentVariables `gorm:"type:jsonb"`
	Conditions                    ConditionArray       `gorm:"type:jsonb"`
	AdditionalFields              AdditionalFields     `gorm:"type:jsonb"`
}

func (RevisionView) TableName() string {
	return "revisions_view"
}
