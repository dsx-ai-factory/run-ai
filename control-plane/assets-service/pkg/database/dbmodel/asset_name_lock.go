/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"gorm.io/gorm"
	"time"
)

type AssetNameLockRecord struct {
	// Unique identifier of the asset.
	Id string `json:"id" gorm:"not null;"`

	// The name of the asset.
	Name string `json:"name" gorm:"not null;"`

	// The kind of the asset.
	Kind api.AssetKind `json:"kind" gorm:"not null;"`

	// The id of the tenant.
	TenantId int32 `json:"tenantId" gorm:"not null;"`

	// The user who created the asset.
	CreatedBy string `json:"createdBy" gorm:"not null;"`

	// The time at which the asset were created
	CreatedAt time.Time `json:"createdAt" gorm:"not null;"`
}

func (AssetNameLockRecord) TableName() string { return dbtables.TableNameAssetsNameLocks }

func (AssetNameLockRecord) BeforeCreate(tx *gorm.DB) (err error) {
	// Remove all rows created more than 1 minute ago
	cutoffTime := time.Now().Add(-1 * time.Minute)
	tx.Where("created_at < ?", cutoffTime).Delete(&AssetNameLockRecord{})
	return nil
}
