/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/dao"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

type PostSaveError struct {
	Err error
}

func (e *PostSaveError) Error() string {
	return e.Err.Error()
}

type SettingsManager struct {
	settingsDao      *dao.TenantSettingsDAO
	settings         map[string]SettingDefinition
	settingProviders []SettingProvider
	showAlphaFlags   bool
}

func NewSettingsManager(settingsDao *dao.TenantSettingsDAO, settingProviders []SettingProvider, showAlphaFlags bool, deps *OnSetDeps) *SettingsManager {
	return &SettingsManager{
		settingsDao:      settingsDao,
		settings:         BuildSettingDefinitions(deps),
		settingProviders: settingProviders,
		showAlphaFlags:   showAlphaFlags,
	}
}

func (m *SettingsManager) GetAllSettings(ctx context.Context, tenantID int, tenantData TenantStaticData, clusterVersions []string, excludeMigrated bool) ([]api.Setting, error) {
	allSetting, err := m.getTenantSettings(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant settings: %w", err)
	}

	settingOverrides := m.fetchProviderOverrides(ctx, tenantID, allSetting)

	result := m.buildSettingsResponse(allSetting, settingOverrides, clusterVersions, tenantData, excludeMigrated)

	return result, nil
}

func (m *SettingsManager) SetSetting(ctx context.Context, tenantID int, key string, value interface{}, clusterVersions []string) error {
	setting, ok := m.settings[key]
	if !ok {
		return fmt.Errorf("unknown setting key: %s", key)
	}

	if !isPermitted(setting) {
		return fmt.Errorf("insufficient permissions for setting %s", key)
	}

	if err := validateSettingValue(setting, value); err != nil {
		return err
	}

	if clusterVersions != nil {
		if !filterUnsupportedSingleClustersFlags(setting, clusterVersions, isFalsy(value)) {
			return fmt.Errorf("can't enable unsupported cluster version flag. Minimum cluster version is %s", setting.MinimumSingleClusterVersion)
		}
	}
	if err := m.processAndSaveSetting(ctx, tenantID, value, setting); err != nil {
		return err
	}

	if setting.PostSave != nil {
		revertValue, err := setting.PostSave(ctx, tenantID, value)
		if err != nil {
			if revertValue != nil {
				revertErr := m.processAndSaveSetting(ctx, tenantID, revertValue, setting)
				if revertErr != nil {
					logging.Errorf("failed to revert setting %s after PostSave failure: %v", key, revertErr)
				}
			}
			return &PostSaveError{Err: err}
		}
	}

	return nil
}

func (m *SettingsManager) processAndSaveSetting(ctx context.Context, tenantID int, value interface{}, setting SettingDefinition) error {
	valueToSave := value
	skipDaoSave := false

	if setting.OnSet != nil {
		result, err := setting.OnSet(ctx, tenantID, value)
		if err != nil {
			return err
		}
		if result != nil {
			if result.SkipDaoSave {
				skipDaoSave = true
			}
			if result.HasNewValue {
				valueToSave = result.Value
			}
		}
	}

	if !skipDaoSave {
		finalValue := valueToSave
		if setting.IsJson {
			b, err := json.Marshal(valueToSave)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON setting value: %w", err)
			}
			finalValue = string(b)
		}
		if err := m.settingsDao.Set(ctx, tenantID, map[string]interface{}{setting.Key: finalValue}); err != nil {
			return fmt.Errorf("failed to save setting: %w", err)
		}
	}

	return nil
}

func validateSettingValue(setting SettingDefinition, value interface{}) error {
	if setting.Validate != nil {
		if msg := setting.Validate(value); msg != "" {
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}

func isFalsy(value interface{}) bool {
	if value == nil {
		return true
	}
	if b, ok := value.(bool); ok {
		return !b
	}
	return false
}

func (m *SettingsManager) getTenantSettings(ctx context.Context, tenantID int) (map[string]interface{}, error) {
	return m.settingsDao.GetAllByTenantID(ctx, tenantID)
}

func (m *SettingsManager) fetchProviderOverrides(ctx context.Context, tenantID int, allSetting map[string]interface{}) map[string]SettingOverride {
	overrides := make(map[string]SettingOverride)

	type providerResult struct {
		settings map[string]SettingOverride
		err      error
	}

	results := make([]providerResult, len(m.settingProviders))
	var wg sync.WaitGroup

	for i, provider := range m.settingProviders {
		wg.Add(1)
		go func(idx int, p SettingProvider) {
			defer wg.Done()
			settings, err := p.FetchSettings(ctx, tenantID)
			results[idx] = providerResult{settings: settings, err: err}
		}(i, provider)
	}

	wg.Wait()

	for _, r := range results {
		if r.err != nil {
			continue
		}
		for key, override := range r.settings {
			if override.Value != nil {
				allSetting[key] = override.Value
			}
			existing, ok := overrides[key]
			if ok {
				merged := mergeOverrides(existing, override)
				overrides[key] = merged
			} else {
				overrides[key] = override
			}
		}
	}

	return overrides
}

func mergeOverrides(existing, incoming SettingOverride) SettingOverride {
	if incoming.Value != nil {
		existing.Value = incoming.Value
	}
	if incoming.Description != "" {
		existing.Description = incoming.Description
	}
	if incoming.DisplayName != "" {
		existing.DisplayName = incoming.DisplayName
	}
	if incoming.Stage != "" {
		existing.Stage = incoming.Stage
	}
	return existing
}

func (m *SettingsManager) buildSettingsResponse(allSetting map[string]interface{}, settingOverrides map[string]SettingOverride, clusterVersions []string, tenantData TenantStaticData, excludeMigrated bool) []api.Setting {
	var result []api.Setting

	for _, setting := range m.settings {
		if !isPermitted(setting) {
			continue
		}

		settingValue := resolveValue(setting.Key, setting.DefaultValue, allSetting)

		if !m.filterDisabledAlphaFlag(setting, settingValue) {
			continue
		}
		if !filterUnsupportedSingleClustersFlags(setting, clusterVersions, false) {
			continue
		}
		if !filterMigrated(setting, excludeMigrated) {
			continue
		}

		value := m.getSettingValue(setting.Key, setting.DefaultValue, allSetting, tenantData)

		active := true
		if setting.Filter != nil {
			active = setting.Filter(allSetting[setting.Key])
		}

		label := setting.Label
		description := setting.Description
		stage := setting.Stage

		if override, ok := settingOverrides[setting.Key]; ok {
			if override.DisplayName != "" {
				label = override.DisplayName
			}
			if override.Description != "" {
				description = override.Description
			}
			if override.Stage != "" {
				stage = override.Stage
			}
		}

		if stage == StageGA {
			stage = ""
		}

		setting := api.Setting{
			Key:         setting.Key,
			Source:      setting.Source,
			Category:    setting.Category,
			Label:       label,
			Description: description,
			Stage:       stage,
			Type:        setting.Type,
			Active:      active,
		}
		if value != nil {
			setting.Value = &value
		}
		result = append(result, setting)
	}

	return result
}

func resolveValue(key string, defaultValue interface{}, allSetting map[string]interface{}) interface{} {
	if val, ok := allSetting[key]; ok {
		return val
	}
	return defaultValue
}

func (m *SettingsManager) getSettingValue(key string, defaultValue interface{}, allSetting map[string]interface{}, tenantData TenantStaticData) interface{} {
	setting := m.settings[key]

	if setting.IsJson {
		if val, ok := allSetting[key]; ok {
			switch v := val.(type) {
			case string:
				var parsed interface{}
				if err := json.Unmarshal([]byte(v), &parsed); err != nil {
					return val
				}
				return parsed
			default:
				return val
			}
		}
		return defaultValue
	}

	switch key {
	case SettingKeyOptimizedAnalyticsQueries:
		return handleOptimizedAnalyticsSetting(allSetting, tenantData, defaultValue)
	case SettingKeyEnableNewCliInstaller:
		return handleNewCliSetting(allSetting, tenantData, defaultValue)
	case SettingKeyNewNavigation:
		return handleNewNavigationSetting(allSetting, tenantData, defaultValue)
	default:
		return resolveValue(key, defaultValue, allSetting)
	}
}

func isPermitted(setting SettingDefinition) bool {
	return setting.Privilege == PrivilegeAdmin
}

func (m *SettingsManager) filterDisabledAlphaFlag(setting SettingDefinition, value interface{}) bool {
	if setting.Stage != StageAlpha {
		return true
	}
	if m.showAlphaFlags {
		return true
	}
	boolVal, ok := value.(bool)
	return ok && boolVal
}

func filterUnsupportedSingleClustersFlags(setting SettingDefinition, clusterVersions []string, isValueSetToFalse bool) bool {
	if isValueSetToFalse {
		return true
	}
	minVersion := setting.MinimumSingleClusterVersion
	if minVersion == "" || len(clusterVersions) == 0 {
		return true
	}
	unsupported := getUnsupportedVersions(clusterVersions, minVersion)
	return len(unsupported) < len(clusterVersions)
}

func filterMigrated(setting SettingDefinition, excludeMigrated bool) bool {
	return !excludeMigrated || !setting.Migrated
}

func getUnsupportedVersions(clusterVersions []string, minimumVersion string) []string {
	var unsupported []string
	for _, v := range clusterVersions {
		if v == "" || !IsVersionSupported(v, minimumVersion) {
			unsupported = append(unsupported, v)
		}
	}
	return unsupported
}

func handleOptimizedAnalyticsSetting(allSetting map[string]interface{}, tenantData TenantStaticData, defaultValue interface{}) interface{} {
	if val, ok := allSetting[SettingKeyOptimizedAnalyticsQueries]; ok {
		return val
	}
	if tenantData.CreatedAt.Before(OptimizedAnalyticsDefaultFalseUntil) {
		return false
	}
	return defaultValue
}

func handleNewCliSetting(allSetting map[string]interface{}, tenantData TenantStaticData, defaultValue interface{}) interface{} {
	if val, ok := allSetting[SettingKeyEnableNewCliInstaller]; ok {
		return val
	}
	if tenantData.CreatedAt.After(NewCliDefaultFalseUntil) {
		return true
	}
	return defaultValue
}

func handleNewNavigationSetting(allSetting map[string]interface{}, tenantData TenantStaticData, defaultValue interface{}) interface{} {
	if val, ok := allSetting[SettingKeyNewNavigation]; ok {
		return val
	}
	if tenantData.CreatedAt.After(NewNavigationDefaultFalseUntil) {
		return true
	}
	return defaultValue
}
