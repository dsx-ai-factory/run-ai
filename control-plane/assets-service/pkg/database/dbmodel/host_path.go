/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

type HostPathAssetRec struct {
	AssetId          string
	api.HostPathSpec `gorm:"embedded"`
}

func (HostPathAssetRec) TableName() string {
	return dbtables.TableNameAssetsHostPath
}

func (h HostPathAssetRec) GetAssetId() string {
	return h.AssetId
}

func ReplaceHostPathSpecIds(specs []HostPathAssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type HostPathAssetUpdate struct {
	api.HostPathSpec `gorm:"embedded"`
}

func (HostPathAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsHostPath
}

// HostPathCombinedRec holds result of join assets_metadata and assets_hostPath queries
type HostPathCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.HostPathSpec           `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p HostPathCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p HostPathCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
