/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
)

type AssetResourceIdsRecord struct {
	AssetId      string    `json:"assetId" gorm:"not null;"`
	ResourceId   string    `json:"resourceId" gorm:"not null;"`
	TenantId     int32     `json:"tenantId" gorm:"not null;"`
	LastSyncedAt time.Time `json:"lastSyncedAt" gorm:"not null;"`
}

func (AssetResourceIdsRecord) TableName() string { return dbtables.TableNameAssetsResourcesIds }
