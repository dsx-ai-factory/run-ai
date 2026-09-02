/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
)

type WorkloadChild struct {
	WorkloadID uuid.UUID
	Workload   Workload `gorm:"foreignKey:WorkloadID"`
	ChildID    uuid.UUID
	ChildType  string
}

func (WorkloadChild) TableName() string {
	return "workloads_children"
}
