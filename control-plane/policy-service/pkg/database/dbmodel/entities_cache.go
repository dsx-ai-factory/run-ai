/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dbtables"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

// EntityType is the kind of top-level entity whose computation is being cached.
// Mirrors the CHECK constraint on entities_cache.entity_type / entities_dependency.entity_type.
type EntityType string

const (
	EntityTypeBreakdown EntityType = "breakdown" // breakdown of a given policy
)

// EntitiesCache stores the cached, pre-computed artifact of an entity (policy
// or mapping). uniqueness on (tenant_id, entity_type, entity_id) ensures a
// single cache row per entity per tenant; the UUID id is what
// entities_dependency.cache_id references.
type EntitiesCache struct {
	ID uuid.UUID `gorm:"primaryKey"`

	TenantId   int32 `gorm:"not null"`
	EntityType EntityType
	EntityId   string

	CachedInfo dao_utils.JsonBMap `gorm:"type:jsonb;not null"`
}

func (EntitiesCache) TableName() string {
	return dbtables.EntitiesCacheTable
}
