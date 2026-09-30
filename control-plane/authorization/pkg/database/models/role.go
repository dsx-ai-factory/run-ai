/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
)

type RoleScope string

const (
	ROLE_SCOPE_SYSTEM RoleScope = "system"
	ROLE_SCOPE_TENANT RoleScope = "tenant"
)

type Role struct {
	Common                  `gorm:"embedded"`
	ID                      uint             `gorm:"primarykey"`
	Name                    string           `gorm:"not null"`
	Permissions             []*Permission    `gorm:"constraint:OnDelete:CASCADE;"`
	Description             string           `gorm:"not null"`
	AccessRules             []*AccessRule    `gorm:"constraint:OnDelete:CASCADE;"`
	InternalOnly            bool             `gorm:"default:false"`
	ScopeType               RoleScope        `gorm:""`
	ScopeId                 *string          `gorm:""`
	ScopeName               *string          `gorm:""`
	Custom                  bool             `gorm:""`
	Enabled                 bool             `gorm:""`
	EffectiveEnabled        *bool            `gorm:"-"`
	ClusterRolePredefinedId *int64           `gorm:""`
	Deprecated              bool             `gorm:"default:false"`
	TenantId                *string          `gorm:""`
	NoPermissionSet         *bool            `gorm:"default:true"`
	PermissionSets          *[]PermissionSet `gorm:"many2many:role_to_permission_set;"`
}

func (r Role) SortAndFilterFieldsToColumn() queryapi.APIFieldsMapping {
	return queryapi.APIFieldsMapping{
		"name":      queryapi.NewFieldMapping("name"),
		"createdAt": queryapi.NewFieldMapping("created_at"),
		"createdBy": queryapi.NewFieldMapping("created_by"),
		"custom":    queryapi.NewFieldMapping("custom"),
		"scopeType": queryapi.NewFieldMapping("scope_type"),
		"enabled":   queryapi.NewFieldMapping("enabled"),
	}
}
