/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

// AssetSpecRec is the persisted row of the generic assets_spec table.
// The table holds the raw Kubernetes spec (as JSON) for k8s-object-backed
// assets (starting with resource claim templates), keyed by asset id.
type AssetSpecRec struct {
	AssetId  string             `gorm:"primaryKey;column:asset_id"`
	TenantId int32              `gorm:"not null;column:tenant_id"`
	Spec     JsonBMapFromString `gorm:"not null;column:spec"`
}

func (AssetSpecRec) TableName() string {
	return dbtables.TableNameAssetsSpec
}

func (a AssetSpecRec) GetAssetId() string {
	return a.AssetId
}

// AssetSpecUpdate is the update-shaped view of the assets_spec row: only the
// spec column is user-updatable (asset_id and tenant_id are immutable identity
// columns). Kept as a separate type so GORM's Updates() targets exactly one
// column and cannot accidentally overwrite the identity.
type AssetSpecUpdate struct {
	Spec JsonBMapFromString `gorm:"not null;column:spec"`
}

func (AssetSpecUpdate) TableName() string {
	return dbtables.TableNameAssetsSpec
}

// AssetSpecCombinedRec holds the result of joining assets_metadata,
// assets_supported_workload_types and assets_spec. Used by list/get queries
// for k8s-object-backed assets (resource claim templates today) which read
// the spec as a JSON map and unmarshal it into a typed swagger struct at the
// DAO boundary. Mirrors ComputeFullRec / NFSCombinedRec structurally.
type AssetSpecCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
	Spec                       JsonBMapFromString `gorm:"column:spec"`
}

func (p AssetSpecCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p AssetSpecCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
