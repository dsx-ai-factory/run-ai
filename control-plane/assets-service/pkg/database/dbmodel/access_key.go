/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

type AccessKeyRecord struct {
	AssetId           string
	api.AccessKeySpec `gorm:"embedded"`
}

func (AccessKeyRecord) TableName() string {
	return dbtables.TableNameAssetsAccessKeys
}

type AccessKeyUpdate struct {
	api.AccessKeySpec `gorm:"embedded"`
}

func (AccessKeyUpdate) TableName() string {
	return dbtables.TableNameAssetsAccessKeys
}

// AccessKeyFullRecord is used to store joined data between assets_metadata and access_key tables
type AccessKeyFullRecord struct {
	api.AssetMeta              `gorm:"embedded"`
	PrevName                   *string
	api.AccessKeySpec          `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p AccessKeyFullRecord) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p AccessKeyFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
