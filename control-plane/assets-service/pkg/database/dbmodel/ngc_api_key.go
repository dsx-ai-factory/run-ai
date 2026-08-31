/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type NgcApiKeyRecord struct {
	AssetId           string
	api.NgcApiKeySpec `gorm:"embedded"`
}

func (NgcApiKeyRecord) TableName() string { return dbtables.TableNameAssetsNgcApiKey }

type NgcApiKeyUpdate struct {
	api.NgcApiKeySpec `gorm:"embedded"`
}

func (NgcApiKeyUpdate) TableName() string { return dbtables.TableNameAssetsNgcApiKey }

type NgcApiKeyFullRecord struct {
	AssetMeta                  `gorm:"embedded"`
	PrevName                   *string
	api.NgcApiKeySpec          `gorm:"embedded"`
	api.WorkloadSupportedTypes `gorm:"embedded"`
}

func (a *NgcApiKeyFullRecord) GetApiAssetMeta() AssetMeta {
	return a.AssetMeta
}

func (a *NgcApiKeyFullRecord) GetApiWorkloadSupportedTypes() *api.WorkloadSupportedTypes {
	return &a.WorkloadSupportedTypes
}
