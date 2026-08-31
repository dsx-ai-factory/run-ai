/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodepoolV1 struct {
	Id                   int32 `gorm:"primaryKey"`
	Name                 string
	LabelKey             *string
	LabelValue           *string
	Status               string
	StatusMessage        *string
	ClusterId            uuid.UUID `gorm:"not null"`
	CreatedAt            time.Time
	UpdatedAt            time.Time
	CreatedBy            *string
	UpdatedBy            *string
	IsDefault            bool
	TenantId             int32 `gorm:"not null"`
	Nodes                *string
	GpuPlacementStrategy string
	CpuPlacementStrategy string
	DeletedAt            gorm.DeletedAt
}

func (NodepoolV1) TableName() string {
	return "node_pool"
}
