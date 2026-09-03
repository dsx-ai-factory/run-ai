/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/utils"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

const (
	ErrUnknownSettingKey = "unknown setting key"
)

type SettingsManager struct{}

func NewSettingsManager() *SettingsManager {
	return &SettingsManager{}
}

func (s *SettingsManager) GetSettings(tenantID string, category utils.CategoryKey) ([]manager.Settinger, error) {
	return setting_pkg.GetAll(tenantID, string(category))
}

func (s *SettingsManager) GetSettingByKey(key string, tenantID string) (manager.Settinger, error) {
	handler, exists := settingRegistry[key]
	if !exists {
		return nil, fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
	return handler.get(tenantID)
}

func (s *SettingsManager) UpdateSettingByKey(ctx context.Context, key string, tenantID string, updateRequest any) error {
	handler, exists := settingRegistry[key]
	if !exists {
		return fmt.Errorf("%s: %s", ErrUnknownSettingKey, key)
	}
	return handler.update(tenantID, updateRequest)
}

func (s *SettingsManager) TransformSettingsToResponse(settingsList []manager.Settinger, category utils.CategoryKey) (any, error) {
	transformer, exists := transformerRegistry[category]
	if !exists {
		return nil, fmt.Errorf("no transformer found for category: %s", category)
	}
	return transformer(settingsList), nil
}
