/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// QuotaStatus - represents the status of the quota for a project, updated from the cluster using the cluster-sync.
// GPU - in units
// CPU - in millicores
// Memory - in MB
type QuotaStatus struct {
	ID        uuid.UUID
	DeletedAt gorm.DeletedAt

	GpuAllocated                  *float64 `json:"gpu_allocated,omitempty"`
	GpuAllocatedNonPreemptible    *float64 `json:"gpu_allocated_non_preemptible,omitempty"`
	GpuRequested                  *float64 `json:"gpu_requested,omitempty"`
	CpuAllocated                  *float64 `json:"cpu_allocated,omitempty"`
	CpuAllocatedNonPreemptible    *float64 `json:"cpu_allocated_non_preemptible,omitempty"`
	CpuRequested                  *float64 `json:"cpu_requested,omitempty"`
	MemoryAllocated               *float64 `json:"memory_allocated,omitempty"`
	MemoryAllocatedNonPreemptible *float64 `json:"memory_allocated_non_preemptible,omitempty"`
	MemoryRequested               *float64 `json:"memory_requested,omitempty"`
}

func (QuotaStatus) TableName() string {
	return QuotaStatusTable
}

func (m *QuotaStatus) BeforeCreate(tx *gorm.DB) (err error) {
	emptyId := uuid.UUID{}
	if m.ID == emptyId {
		m.ID = uuid.New()
	}
	return
}
