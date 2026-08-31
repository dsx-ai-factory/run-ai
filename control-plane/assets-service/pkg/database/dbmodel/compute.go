/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

// ComputeFullRec holds result of join assets_metadata and assets_compute queries
type ComputeFullRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.ComputeAssetSpec       `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p ComputeFullRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p ComputeFullRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}

type ComputeAssetRec struct {
	AssetId              string
	api.ComputeAssetSpec `gorm:"embedded"`
}

func (ComputeAssetRec) TableName() string {
	return dbtables.TableNameAssetsCompute
}

func (c ComputeAssetRec) GetAssetId() string {
	return c.AssetId
}

func ReplaceComputeSpecIds(specs []ComputeAssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type ComputeAssetUpdate struct {
	api.ComputeAssetSpec `gorm:"embedded"`
}

func (ComputeAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsCompute
}
