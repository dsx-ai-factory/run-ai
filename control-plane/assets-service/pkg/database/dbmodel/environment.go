/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type EnvironmentAssetRec struct {
	AssetId                  string
	api.EnvironmentAssetSpec `gorm:"embedded"`
}

func (EnvironmentAssetRec) TableName() string {
	return dbtables.TableNameAssetsEnvironment
}

func (e EnvironmentAssetRec) GetAssetId() string {
	return e.AssetId
}

func ReplaceEnvironmentSpecIds(specs []EnvironmentAssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type EnvironmentAssetUpdate struct {
	api.EnvironmentAssetSpec `gorm:"embedded"`
}

func (EnvironmentAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsEnvironment
}

// EnvironmentCombinedRec holds result of join assets_metadata and assets_environment queries
type EnvironmentCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.EnvironmentAssetSpec   `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p EnvironmentCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p EnvironmentCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
