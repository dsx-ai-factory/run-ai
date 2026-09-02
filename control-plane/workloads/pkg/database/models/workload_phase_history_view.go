/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type WorkloadPhasesView struct {
	WorkloadID uuid.UUID
	ClusterID  uuid.UUID
	TenantID   uint
	CreatedAt  time.Time

	Phase        string
	PhaseMessage string
	OldPhase     string

	Name        string
	Type        string
	ProjectName string
	SubmittedBy string
}

func (WorkloadPhasesView) TableName() string {
	return "workloads_phases_view"
}
