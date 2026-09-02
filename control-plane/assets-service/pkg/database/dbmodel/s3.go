/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
)

type S3AssetRec struct {
	AssetId         string
	api.S3AssetSpec `gorm:"embedded"`
}

func (S3AssetRec) TableName() string {
	return dbtables.TableNameAssetsS3
}

func (s S3AssetRec) GetAssetId() string {
	return s.AssetId
}

func ReplaceS3SpecIds(specs []S3AssetRec, translationMap map[string]string) {
	for idx := range specs {
		specs[idx].AssetId = translationMap[specs[idx].AssetId]
	}
}

type S3AssetUpdate struct {
	api.S3AssetSpec `gorm:"embedded"`
}

func (S3AssetUpdate) TableName() string {
	return dbtables.TableNameAssetsS3
}

// S3CombinedRec holds result of join assets_metadata and assets_s3 queries
type S3CombinedRec struct {
	api.AssetMeta              `gorm:"embedded"`
	api.S3AssetSpec            `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (p S3CombinedRec) GetApiAssetMeta() api.AssetMeta {
	return p.AssetMeta
}

func (p S3CombinedRec) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &p.WorkloadSupportedTypes
}
