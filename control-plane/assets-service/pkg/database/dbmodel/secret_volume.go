/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

// SecretAssetRec represents a Secret asset record in the table
type SecretAssetRec struct {
	AssetId             string
	api.SecretAssetSpec `gorm:"embedded"`
}

// TableName implements gorm tabler interface
func (SecretAssetRec) TableName() string {
	return dbtables.TableNameAssetsSecretVolume
}

// SecretAssetUpdate used as a db model for updating Secret record
type SecretAssetUpdate struct {
	api.SecretAssetSpec `gorm:"embedded"`
}

// TableName implements gorm tabler interface
func (SecretAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsSecretVolume
}

// SecretCombinedRec holds result of join assets_metadata and assets_secret_volume queries
type SecretAssetCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.SecretAssetSpec        `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

// GetApiAssetMeta implements the AssetRecord interface
func (s *SecretAssetCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return s.AssetMeta
}

// GetApiWorkloadSupportedTypes implements the AssetRecord interface
func (s *SecretAssetCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &s.WorkloadSupportedTypes
}
