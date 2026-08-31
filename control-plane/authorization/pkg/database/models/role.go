/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"strings"

	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	authorization_config "github.com/run-ai/runai/control-plane/authorization/pkg/config"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-common-packages/config"

	"gorm.io/gorm"
)

const DGXCManagedPostfix = "-DGXC"

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
	RoleType                RoleType         `gorm:"default:ALL"`
	ShowForDGXC             bool             `gorm:"-"`
	Deprecated              bool             `gorm:"default:false"`
	TenantId                *string          `gorm:""`
	NoPermissionSet         *bool            `gorm:"default:true"`
	PermissionSets          *[]PermissionSet `gorm:"many2many:role_to_permission_set;"`
}

func (r *Role) BeforeUpdate(tx *gorm.DB) (err error) {
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		r.AddMissingResourceTypes()
	}
	return nil
}

func (r *Role) BeforeSave(tx *gorm.DB) (err error) {
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		r.AddMissingResourceTypes()
	}
	return nil
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if !config.Get[*authorization_config.General]().UsePermissionSets {
		r.AddMissingResourceTypes()
	}
	return nil
}

func (r *Role) AddMissingResourceTypes() (err error) {
	permissionsMap := map[api.ResourceType]bool{}
	for _, permission := range r.Permissions {
		permissionsMap[permission.ResourceType] = true
	}

	for _, resourceType := range api.AllowedResourceTypeEnumValues {
		if _, found := permissionsMap[resourceType]; !found {
			r.Permissions = append(r.Permissions, &Permission{
				ResourceType: resourceType,
				Create:       false,
				Read:         false,
				Update:       false,
				Delete:       false,
				Sync:         false,
			})
		}
	}

	return
}

func (r *Role) AfterFind(tx *gorm.DB) (err error) {
	r.Name, _ = strings.CutSuffix(r.Name, DGXCManagedPostfix)
	return
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

type RoleType string

// List of ResourceType
const (
	DGXC    RoleType = "DGXC"
	NonDGXC RoleType = "NON_DGXC"
	ALL     RoleType = "ALL"
)
