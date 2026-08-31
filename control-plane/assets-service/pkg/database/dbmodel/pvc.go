/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

const DefaultForAutoDelete = false
const ClaimNameColumn = "claim_name"

type PVCAssetRec struct {
	AssetId          string
	api.PvcAssetSpec `gorm:"embedded"`
}

func (PVCAssetRec) TableName() string {
	return dbtables.TableNameAssetsPVC
}

func (PVCAssetRec) AssetIdColumn() string {
	return dbtables.AssetIdColumnName
}

type PVCAssetUpdate struct {
	api.PvcAssetSpec `gorm:"embedded"`
}

func (PVCAssetUpdate) TableName() string {
	return dbtables.TableNameAssetsPVC
}

// PVCCombinedRec holds result of join assets_metadata and assets_pvc queries
type PVCCombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.PvcAssetSpec           `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p PVCCombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p PVCCombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}

type PVCDatavolumeRec struct {
	AssetId      string
	DatavolumeId string
}

func (PVCDatavolumeRec) TableName() string {
	return dbtables.TableNamePvcDatavolume
}
