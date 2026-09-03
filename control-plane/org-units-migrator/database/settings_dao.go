/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	ProjectManualOverQuota = "project.quota.manual_over_quota"
	CPUResourcesQuota      = "project.quota.cpu_resources"
)

var MigrationSettings = []string{
	ProjectManualOverQuota,
	CPUResourcesQuota,
}

func GetTenantSettingsMap(tenantId int32) (map[string]bool, error) {
	settings := []models.TenantSetting{}

	if err := sourceClient.
		Where("tenant_id = ?", tenantId).
		Where("key IN ?", MigrationSettings).
		Find(&settings); err != nil {
		logging.Errorf("Failed to query tenant settings for tenant %d. err: %v", tenantId, err)
		return nil, err
	}

	settingsMap := map[string]bool{}

	for _, key := range MigrationSettings {
		settingsMap[key] = false
	}

	for _, setting := range settings {
		settingsMap[setting.Key] = strings.ToLower(setting.Value) == "true"
	}
	return settingsMap, nil
}
