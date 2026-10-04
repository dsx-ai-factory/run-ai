/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
)

type RealmSettingsDal interface {
	SaveRealmSetting(ctx context.Context, realmSetting *models.RealmSettings) error
	GetRealmSettingByKey(ctx context.Context, tenantId string, settingKey string) (*models.RealmSettings, error)
	GetRealmSettingsByTenantId(ctx context.Context, tenantId string) (*[]models.RealmSettings, error)
}
