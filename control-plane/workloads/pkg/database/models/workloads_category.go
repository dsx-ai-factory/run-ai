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

type WorkloadsCategory struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt   time.Time `gorm:"not null;default:now()"`
	CreatedBy   string    `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"default:now()"`
	UpdatedBy   *string
	DeletedAt   gorm.DeletedAt
	DeletedBy   *string
	Name        string `gorm:"not null"`
	Description *string
	TenantID    int64  `gorm:"not null"`
	ScopeType   string `gorm:"not null;default:'tenant'"`
}

func (WorkloadsCategory) TableName() string {
	return "workloads_categories"
}

// ExternalWorkloadsTypeConfig represents the external_workloads_type_config table
// See: 02220005_categories.up.sql

type ExternalWorkloadsTypeConfig struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	TenantID   int32     `gorm:"not null;default:-1"`
	Type       string    `gorm:"not null"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null"`
	ScopeType  string    `gorm:"not null;default:'tenant'"`
	UpdatedAt  time.Time `gorm:"default:now()"`
	UpdatedBy  *string
}

func (ExternalWorkloadsTypeConfig) TableName() string {
	return "external_workloads_type_config"
}
