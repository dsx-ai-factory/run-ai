/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

import (
	"github.com/google/uuid"
)

const (
	TemplatesAssetsTableName = "crud_templates_assets"

	TemplatesAssetsColumnTenantId    = "tenant_id"
	TemplatesAssetsColumnTemplateId  = "template_id"
	TemplatesAssetsColumnAssetId     = "asset_id"
	TemplatesAssetsColumnReplicaType = "replica_type"
	TemplatesAssetsColumnCategory    = "category"
)

type TemplateAssets struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	TenantId uint

	AssetId     string
	TemplateId  string
	ReplicaType string
	Category    string
}

func (TemplateAssets) TableName() string {
	return TemplatesAssetsTableName
}
