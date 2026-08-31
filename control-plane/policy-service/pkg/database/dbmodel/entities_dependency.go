/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
)

// ComponentType is the kind of building block an entity may depend on.
// Mirrors the CHECK constraint on entities_dependency.component_type.
type ComponentType string

const (
	ComponentTypeMapping    ComponentType = "mapping"
	ComponentTypeFieldGroup ComponentType = "field_group"
	ComponentTypeFormatter  ComponentType = "formatter"
)

// EntitiesDependency records a single (cache -> component) dependency edge.
// the entity that owns the cache is reachable via the entities_cache row
// pointed at by CacheId, so it is not duplicated here.
//
// the composite primary key (component_type, component_id, cache_id) gives:
//   - uniqueness of edges
//   - a "by component" prefix index used by dao.OnComponentDeleted
//
// CacheId has a foreign key to entities_cache(id) with no ON DELETE CASCADE:
// deletion is handled explicitly in code.
//
// TenantId mirrors the tenant_id of the owning entities_cache row and is
// included here so all queries can be tenant-scoped without a join.
type EntitiesDependency struct {
	CacheId       uuid.UUID     `gorm:"primaryKey"`
	TenantId      int32         `gorm:"not null"`
	ComponentType ComponentType `gorm:"primaryKey"`
	ComponentId   string        `gorm:"primaryKey"`
}

func (EntitiesDependency) TableName() string {
	return dbtables.EntitiesDependencyTable
}
