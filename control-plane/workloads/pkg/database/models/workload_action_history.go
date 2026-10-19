/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

// Authorization scope survives hard deletion of the canonical workload.
type WorkloadActionScope struct {
	TenantID     uint
	ClusterID    uuid.UUID
	ProjectID    string
	DepartmentID string
}

type WorkloadActionHistory struct {
	ID                  uuid.UUID `gorm:"primaryKey"`
	WorkloadActionScope `gorm:"embedded"`
	WorkloadID          uuid.UUID
	WorkloadName        string
	Subject             string
	SubjectType         string
	Action              string
	AcceptedAt          time.Time
}

func (WorkloadActionHistory) TableName() string { return "workloads_action_history" }
