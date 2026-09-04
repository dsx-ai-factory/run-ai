/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models_crud

const (
	MigrationsTableName = "crud_migrations"

	MigrationsColumnTenantId              = "tenant_id"
	MigrationsColumnAssetsToFlatTemplates = "assets_to_flat_templates"
)

type Migrations struct {
	TenantId              int64 `gorm:"primaryKey"`
	AssetsToFlatTemplates int32
}

func (Migrations) TableName() string {
	return MigrationsTableName
}
