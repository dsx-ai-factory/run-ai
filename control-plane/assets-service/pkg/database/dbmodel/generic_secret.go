/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

type GenericSecretRecord struct {
	AssetId               string
	api.GenericSecretSpec `gorm:"embedded"`
}

func (GenericSecretRecord) TableName() string {
	return dbtables.TableNameAssetsGenericSecret
}

type GenericSecretUpdate struct {
	api.GenericSecretSpec `gorm:"embedded"`
}

func (GenericSecretUpdate) TableName() string {
	return dbtables.TableNameAssetsGenericSecret
}

// GenericSecretFullRecord is used to store joined data between assets_metadata and generic_secret tables
type GenericSecretFullRecord struct {
	AssetMeta                  `gorm:"embedded"`
	PrevName                   *string
	api.GenericSecretSpec      `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p GenericSecretFullRecord) GetApiAssetMeta() AssetMeta {
	return p.AssetMeta
}

func (p GenericSecretFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
