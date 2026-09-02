/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"

	"k8s.io/utils/ptr"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"
)

const (
	// ConfigMapNameColumn is column's name of the config map's name
	ConfigMapNameColumn = "name"
	// configmapsTable is the name of the table that contains configmap
	configmapsTable = "configmaps"
	// configMapsTenantIdColumn is the column's name of the config map's tenant id
	configMapsTenantIdColumn = "tenant_id"
	// configMapsClusterIdColumn is the column's name of the config map's cluster id
	configMapsClusterIdColumn = "cluster_id"
	// configMapsClusterIdColumn is the column's name of the config map's cluster id
	configMapsScopeTypeColumn = "scope_type"
	// configMapsScopeTypeColumn is the column's name of the config map's scope type
	configMapsScopeIdColumn = "scope_id"
)

// ConfigMapFilters is a filter on the configmap table also used as part of the ConfigMap ORM
type ConfigMapFilters struct {
	TenantId   int32
	ClusterId  *string
	Name       *string
	ScopeQuery *ScopeQueryParams `gorm:"embedded"`
}

func (cf *ConfigMapFilters) Clone() *ConfigMapFilters {
	return &ConfigMapFilters{
		TenantId:   cf.TenantId,
		ClusterId:  cf.ClusterId,
		Name:       cf.Name,
		ScopeQuery: cf.ScopeQuery,
	}
}

func (cf *ConfigMapFilters) Apply(ctx context.Context, db orm.DB, includeAncestors bool, includeDescendants bool) (orm.DB, *api.Error) {
	// we need all the filters except the scope query that we handle separately
	clonedFilters := cf.Clone()
	clonedFilters.ScopeQuery = nil

	db = db.Where(clonedFilters)

	if cf.ScopeQuery != nil {
		scopeQuery, args, err := cf.ScopeQuery.BuildScopeQuery(ctx, includeAncestors, includeDescendants)
		if err != nil {
			return nil, err
		}
		db = db.Where(scopeQuery, args...)
	}

	return db, nil
}

// TableName returns the name of the table
func (f *ConfigMapFilters) TableName() string {
	return configmapsTable
}

// ConfigMap represents the ConfigMap ORM
type ConfigMap struct {
	ID               uuid.UUID `gorm:"primaryKey"`
	ConfigMapFilters `gorm:"embedded"`
}

// NewConfigMap return a new ConfigMap
func NewConfigMap(filters *ConfigMapFilters) *ConfigMap {
	var result ConfigMap
	if filters != nil {
		result.ConfigMapFilters = *filters
	}

	return &result
}

// BeforeCreate implements the BeforeCreateInterface by adding an id to the object
func (cm *ConfigMap) BeforeCreate(_ *gorm.DB) error {
	cm.ID = uuid.New()
	return nil
}

// TableName implements the tabler interface
func (cm *ConfigMap) TableName() string {
	return configmapsTable
}

func (cm *ConfigMap) TenantIdColumn() string {
	return configMapsTenantIdColumn
}

func (cm *ConfigMap) ClusterIdColumn() string {
	return configMapsClusterIdColumn
}

func (cm *ConfigMap) ScopeTypeColumn() string {
	return configMapsScopeTypeColumn
}

func (cm *ConfigMap) ScopeIdColumn() string {
	return configMapsScopeIdColumn
}

// ErrorMessage decorates the title message with ConfigMap's details
func (cm *ConfigMap) ErrorMessage(title string) string {
	return fmt.Sprintf("failed to %s configmap %s", title, cm.details())
}

// details returns details on the ConfigMap
func (cm *ConfigMap) details() string {
	var scopeType string
	var scopeId string

	if cm.ScopeQuery != nil {
		scopeType = cm.ScopeQuery.ScopeType
		scopeId = cm.ScopeQuery.ScopeId
	}

	return fmt.Sprintf("tenant[%d] cluster[%s] %s[%s] configmap [%s]",
		cm.TenantId,
		ptr.Deref(cm.ClusterId, ""),
		scopeType,
		scopeId,
		ptr.Deref(cm.Name, ""))
}
