/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"time"

	"github.com/google/uuid"
)

type HelmRepo struct {
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
	IsPrivate   bool    `gorm:"not null;default:false"`
	Description *string `gorm:""`

	Version *string `gorm:""`
}

const (
	HelmReposTableName = "helm_repos"

	HelmRepoScopeTypeColumn    = "scope_type"
	HelmRepoScopeIdColumn      = "scope_id"
	HelmRepoTenantIdColumn     = "tenant_id"
	HelmRepoClusterIdColumn    = "cluster_id"
	HelmRepoDepartmentIdColumn = "department_id"
	HelmRepoProjectIdColumn    = "project_id"
	HelmRepoUrlColumn          = "url"
	HelmRepoNameColumn         = "name"
	HelmRepoIsPrivateColumn    = "is_private"
	HelmRepoDescriptionColumn  = "description"

	HelmRepoVersionColumn = "version"

	HelmRepoCreatedByColumn = "created_by"
	HelmRepoCreatedAtColumn = "created_at"
	HelmRepoUpdatedByColumn = "updated_by"
	HelmRepoUpdatedAtColumn = "updated_at"
)

func (r *HelmRepo) TableName() string {
	return HelmReposTableName
}
