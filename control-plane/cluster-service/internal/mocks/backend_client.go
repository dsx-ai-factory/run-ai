/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package mocks holds in-process stand-ins for cluster-service dependencies that
// are not deployed in integration tests.
package mocks

import (
	"context"

	settings_manager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

// BackendClient reports no settings in temp-backend-settings. The startup
// settings migration calls the backend on a fresh DB and exits if it fails, so
// an empty answer lets it complete with nothing to migrate.
type BackendClient struct{}

func (BackendClient) GetSettingByKeyForAllTenants(_ context.Context, key string) (*settings_manager.SettingValuesForAllTenantsBackendResponse, error) {
	return &settings_manager.SettingValuesForAllTenantsBackendResponse{Key: key}, nil
}
