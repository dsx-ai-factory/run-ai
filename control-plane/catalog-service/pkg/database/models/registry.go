/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type Registry struct {
	ID uuid.UUID `gorm:"primaryKey"`

	ScopeType string `gorm:"not null"`
	ScopeId   string `json:"scopeId,omitempty" gorm:""`

	TenantId     *string    `json:"tenantId,omitempty" gorm:""`
	ClusterId    *uuid.UUID `json:"clusterId,omitempty" gorm:""`
	DepartmentId *string    `json:"departmentId,omitempty" gorm:""`
	ProjectId    *string    `json:"projectId,omitempty" gorm:""`

	CreatedAt time.Time `json:"createdAt" gorm:"not null;"`
	CreatedBy string    `json:"createdBy" gorm:"not null;"`

	UpdatedAt time.Time `json:"updatedAt" gorm:"not null;"`
	UpdatedBy string    `json:"updatedBy" gorm:"not null;"`

	Url         string  `gorm:"not null"`
	Name        string  `gorm:"not null"`
	Description *string `gorm:""`

	Version *string `gorm:""`
}

const (
	RegistriesTableName = "registries"

	RegistryScopeTypeColumn    = "scope_type"
	RegistryScopeIdColumn      = "scope_id"
	RegistryTenantIdColumn     = "tenant_id"
	RegistryClusterIdColumn    = "cluster_id"
	RegistryDepartmentIdColumn = "department_id"
	RegistryProjectIdColumn    = "project_id"
	RegistryUrlColumn          = "url"
	RegistryNameColumn         = "name"
	RegistryDescriptionColumn  = "description"
	RegistryVersionColumn      = "version"

	RegistryCreatedByColumn = "created_by"
	RegistryCreatedAtColumn = "created_at"
	RegistryUpdatedByColumn = "updated_by"
	RegistryUpdatedAtColumn = "updated_at"
)

func (r *Registry) TableName() string {
	return RegistriesTableName
}
