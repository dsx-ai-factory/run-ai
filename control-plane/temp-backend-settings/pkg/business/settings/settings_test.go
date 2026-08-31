/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// stubAuthorizationUpdater is a minimal in-test implementation of AuthorizationUpdater
// that records which method was called with what value, optionally returning an error.
type stubAuthorizationUpdater struct {
	scopedAuthorizationCalls []interface{}
	cspManagedCalls          []interface{}
	err                      error
}

func (s *stubAuthorizationUpdater) UpdateEnableScopedAuthorization(_ context.Context, _ int, value interface{}) error {
	s.scopedAuthorizationCalls = append(s.scopedAuthorizationCalls, value)
	return s.err
}

func (s *stubAuthorizationUpdater) UpdateTenantIsCspManaged(_ context.Context, _ int, value interface{}) error {
	s.cspManagedCalls = append(s.cspManagedCalls, value)
	return s.err
}

// stubOrgUnitSettingsUpdater records the values forwarded to org-unit-service.
type stubOrgUnitSettingsUpdater struct {
	limitQuotaOverSubscriptionCalls []interface{}
	err                             error
}

func (s *stubOrgUnitSettingsUpdater) UpdateOrgUnitSettings(_ context.Context, _ int, _ string, _ string) error {
	return s.err
}

func (s *stubOrgUnitSettingsUpdater) UpdateLimitQuotaOverSubscription(_ context.Context, _ int, value interface{}) error {
	s.limitQuotaOverSubscriptionCalls = append(s.limitQuotaOverSubscriptionCalls, value)
	return s.err
}

// ==================== BuildSettingDefinitions ====================

func TestBuildSettingDefinitions_ReturnsAllSettings(t *testing.T) {
	settings := BuildSettingDefinitions(nil)
	assert.NotEmpty(t, settings)

	for key, value := range settings {
		assert.Equal(t, key, value.Key, "map key must match definition Key")
		assert.NotEmpty(t, value.Source, "setting %s must have a Source", key)
		assert.NotEmpty(t, value.Category, "setting %s must have a Category", key)
		assert.NotEmpty(t, value.Type, "setting %s must have a Type", key)
		assert.NotEmpty(t, value.Privilege, "setting %s must have a Privilege", key)
	}
}

func TestBuildSettingDefinitions_AllKeysPresent(t *testing.T) {
	settings := BuildSettingDefinitions(nil)

	expectedKeys := []string{
		SettingKeyLogoutURI,
		SettingKeyProjectManualOverQuota,
		SettingKeyLimitQuotaOverSubscription,
		SettingKeyEnableCPUDashboards,
		SettingKeyCPUResourcesQuota,
		SettingKeyUseOldInference,
		SettingKeyDeletedWorkloadsView,
		SettingKeyEnablePolicies,
		SettingKeyWorkloadSubmissionV2Enabled,
		SettingKeyEnableNewCliInstaller,
		SettingKeyAllowGPT2Model,
		SettingKeyDisableAudit,
		SettingKeyEnableInactivityLogout,
		SettingKeyInactivityMinutesTillLogout,
		SettingKeyExcludeViewerInactivityLogout,
		SettingKeyConsumptionReport,
		SettingKeyEnableAdvancedGPUMetrics,
		SettingKeyDashboardShowOld,
		SettingKeyAdminMessage,
		SettingKeyOptimizedAnalyticsQueries,
		SettingKeyEnableModelCatalog,
		SettingKeyEnableNIM,
		SettingKeyEnableHybridSubmission,
		SettingKeyEnableTemplatesV2,
		SettingKeyEnableDataVolumes,
		SettingKeyEnableNgcIntegration,
		SettingKeyGpuOverProvisioning,
		SettingKeyTenantIsCspManaged,
		SettingKeyEnableScopedAuthorization,
		SettingKeyNewNavigation,
		SettingKeyEnableCustomRoles,
	}

	for _, key := range expectedKeys {
		_, ok := settings[key]
		assert.True(t, ok, "expected setting key %q to be present", key)
	}
}

func TestBuildSettingDefinitions_NoDuplicateKeys(t *testing.T) {
	settings := BuildSettingDefinitions(nil)
	assert.Equal(t, len(settings), len(settings), "map guarantees no duplicates")
}

func TestBuildSettingDefinitions_SettingCounts(t *testing.T) {
	settings := BuildSettingDefinitions(nil)

	var migrated, nonMigrated int
	for _, value := range settings {
		if value.Migrated {
			migrated++
		} else {
			nonMigrated++
		}
	}

	assert.Equal(t, 31, len(settings), "total number of settings")
	assert.Equal(t, 21, migrated, "number of migrated settings")
	assert.Equal(t, 10, nonMigrated, "number of non-migrated settings")
}

func TestBuildSettingDefinitions_MigratedSettingsAreCorrect(t *testing.T) {
	settings := BuildSettingDefinitions(nil)

	expectedMigrated := []string{
		SettingKeyLogoutURI,
		SettingKeyDeletedWorkloadsView,
		SettingKeyAllowGPT2Model,
		SettingKeyInactivityMinutesTillLogout,
		SettingKeyConsumptionReport,
		SettingKeyAdminMessage,
		SettingKeyEnableModelCatalog,
		SettingKeyEnableNIM,
		SettingKeyEnableNgcIntegration,
		SettingKeyEnableDataVolumes,
		SettingKeyEnablePolicies,
		SettingKeyWorkloadSubmissionV2Enabled,
		SettingKeyEnableHybridSubmission,
		SettingKeyEnableTemplatesV2,
		SettingKeyGpuOverProvisioning,
		SettingKeyEnableAdvancedGPUMetrics,
		SettingKeyTenantIsCspManaged,
		SettingKeyEnableScopedAuthorization,
		SettingKeyLimitQuotaOverSubscription,
		SettingKeyProjectManualOverQuota,
		SettingKeyCPUResourcesQuota,
	}

	for _, key := range expectedMigrated {
		value, ok := settings[key]
		assert.True(t, ok, "setting %q should exist", key)
		assert.True(t, value.Migrated, "setting %q should be migrated", key)
	}

	for key, value := range settings {
		if value.Migrated {
			assert.Contains(t, expectedMigrated, key, "setting %q is migrated but not in expected list", key)
		}
	}
}

func TestBuildSettingDefinitions_MigratedNonLegacyHaveOnSet(t *testing.T) {
	settings := BuildSettingDefinitions(nil)

	for key, value := range settings {
		if value.Migrated && value.Stage != StageLegacy {
			assert.NotNil(t, value.OnSet, "migrated non-legacy setting %q must have OnSet defined", key)
		}
	}
}

func TestBuildSettingDefinitions_ValidatorsPresent(t *testing.T) {
	settings := BuildSettingDefinitions(nil)
	for key, value := range settings {
		if value.Type == TypeBoolean {
			assert.NotNil(t, value.Validate, "boolean setting %s should have a validator", key)
		}
	}
}

// ==================== filterNever ====================

func TestFilterNever(t *testing.T) {
	assert.False(t, filterNever(nil))
	assert.False(t, filterNever(true))
	assert.False(t, filterNever("anything"))
}

// ==================== filterShowOnlyWhenConfigured ====================

func TestFilterShowOnlyWhenConfigured_Nil(t *testing.T) {
	assert.False(t, filterShowOnlyWhenConfigured(nil))
}

func TestFilterShowOnlyWhenConfigured_NonNil(t *testing.T) {
	assert.True(t, filterShowOnlyWhenConfigured(true))
	assert.True(t, filterShowOnlyWhenConfigured(false))
	assert.True(t, filterShowOnlyWhenConfigured("value"))
}

// ==================== handleOptimizedAnalyticsSetting ====================

func TestHandleOptimizedAnalyticsSetting_ExplicitValue(t *testing.T) {
	allSettings := map[string]interface{}{SettingKeyOptimizedAnalyticsQueries: false}
	result := handleOptimizedAnalyticsSetting(allSettings, TenantStaticData{}, true)
	assert.Equal(t, false, result)
}

func TestHandleOptimizedAnalyticsSetting_OldTenantDefaultsFalse(t *testing.T) {
	old := TenantStaticData{CreatedAt: time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)}
	result := handleOptimizedAnalyticsSetting(map[string]interface{}{}, old, true)
	assert.Equal(t, false, result)
}

func TestHandleOptimizedAnalyticsSetting_NewTenantUsesDefault(t *testing.T) {
	newTenant := TenantStaticData{CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	result := handleOptimizedAnalyticsSetting(map[string]interface{}{}, newTenant, true)
	assert.Equal(t, true, result)
}

// ==================== handleNewCliSetting ====================

func TestHandleNewCliSetting_ExplicitValue(t *testing.T) {
	allSettings := map[string]interface{}{SettingKeyEnableNewCliInstaller: false}
	result := handleNewCliSetting(allSettings, TenantStaticData{}, true)
	assert.Equal(t, false, result)
}

func TestHandleNewCliSetting_NewTenantDefaultsTrue(t *testing.T) {
	newTenant := TenantStaticData{CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	result := handleNewCliSetting(map[string]interface{}{}, newTenant, false)
	assert.Equal(t, true, result)
}

func TestHandleNewCliSetting_OldTenantUsesDefault(t *testing.T) {
	old := TenantStaticData{CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
	result := handleNewCliSetting(map[string]interface{}{}, old, "fallback")
	assert.Equal(t, "fallback", result)
}

// ==================== handleNewNavigationSetting ====================

func TestHandleNewNavigationSetting_ExplicitValue(t *testing.T) {
	allSettings := map[string]interface{}{SettingKeyNewNavigation: true}
	result := handleNewNavigationSetting(allSettings, TenantStaticData{}, false)
	assert.Equal(t, true, result)
}

func TestHandleNewNavigationSetting_NewTenantDefaultsTrue(t *testing.T) {
	newTenant := TenantStaticData{CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	result := handleNewNavigationSetting(map[string]interface{}{}, newTenant, false)
	assert.Equal(t, true, result)
}

func TestHandleNewNavigationSetting_OldTenantUsesDefault(t *testing.T) {
	old := TenantStaticData{CreatedAt: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}
	result := handleNewNavigationSetting(map[string]interface{}{}, old, "default-val")
	assert.Equal(t, "default-val", result)
}

// ==================== EnableScopedAuthorization OnSet ====================

func TestEnableScopedAuthorization_OnSet_NoUpdater_Succeeds(t *testing.T) {
	defs := BuildSettingDefinitions(&OnSetDeps{})
	value := defs[SettingKeyEnableScopedAuthorization]

	result, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestEnableScopedAuthorization_OnSet_CallsUpdater(t *testing.T) {
	updater := &stubAuthorizationUpdater{}
	defs := BuildSettingDefinitions(&OnSetDeps{AuthorizationUpdater: updater})
	value := defs[SettingKeyEnableScopedAuthorization]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true}, updater.scopedAuthorizationCalls)
	assert.Empty(t, updater.cspManagedCalls)
}

func TestEnableScopedAuthorization_OnSet_PropagatesError(t *testing.T) {
	updater := &stubAuthorizationUpdater{err: errors.New("authorization service down")}
	defs := BuildSettingDefinitions(&OnSetDeps{AuthorizationUpdater: updater})
	value := defs[SettingKeyEnableScopedAuthorization]

	_, err := value.OnSet(context.Background(), 42, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "EnableScopedAuthorization")
	assert.Contains(t, err.Error(), "authorization service down")
}

// ==================== TenantIsCspManaged OnSet ====================

func TestTenantIsCspManaged_OnSet_NoUpdater_Succeeds(t *testing.T) {
	defs := BuildSettingDefinitions(&OnSetDeps{})
	value := defs[SettingKeyTenantIsCspManaged]

	result, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestTenantIsCspManaged_OnSet_CallsUpdater(t *testing.T) {
	updater := &stubAuthorizationUpdater{}
	defs := BuildSettingDefinitions(&OnSetDeps{AuthorizationUpdater: updater})
	value := defs[SettingKeyTenantIsCspManaged]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true}, updater.cspManagedCalls)
	assert.Empty(t, updater.scopedAuthorizationCalls)
}

func TestTenantIsCspManaged_OnSet_PropagatesError(t *testing.T) {
	updater := &stubAuthorizationUpdater{err: errors.New("authorization service down")}
	defs := BuildSettingDefinitions(&OnSetDeps{AuthorizationUpdater: updater})
	value := defs[SettingKeyTenantIsCspManaged]

	_, err := value.OnSet(context.Background(), 42, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TenantIsCspManaged")
	assert.Contains(t, err.Error(), "authorization service down")
}

// ==================== LimitQuotaOverSubscription OnSet ====================

func TestLimitQuotaOverSubscription_OnSet_NoUpdater_Succeeds(t *testing.T) {
	defs := BuildSettingDefinitions(&OnSetDeps{})
	value := defs[SettingKeyLimitQuotaOverSubscription]

	result, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestLimitQuotaOverSubscription_OnSet_CallsUpdater(t *testing.T) {
	updater := &stubOrgUnitSettingsUpdater{}
	defs := BuildSettingDefinitions(&OnSetDeps{OrgUnitSettingsUpdater: updater})
	value := defs[SettingKeyLimitQuotaOverSubscription]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true}, updater.limitQuotaOverSubscriptionCalls)
}

func TestLimitQuotaOverSubscription_OnSet_PropagatesError(t *testing.T) {
	updater := &stubOrgUnitSettingsUpdater{err: errors.New("org-unit service down")}
	defs := BuildSettingDefinitions(&OnSetDeps{OrgUnitSettingsUpdater: updater})
	value := defs[SettingKeyLimitQuotaOverSubscription]

	_, err := value.OnSet(context.Background(), 42, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), SettingKeyLimitQuotaOverSubscription)
	assert.Contains(t, err.Error(), "org-unit service down")
}

// ==================== Cluster-migrated node-pool settings OnSet ====================

// stubClusterUpdater is a minimal in-test implementation of ClusterUpdater
type stubClusterUpdater struct {
	err                         error
	advancedGPUMetricsCalls     []interface{}
	projectManualOverQuotaCalls []interface{}
	cpuResourcesQuotaCalls      []interface{}
}

func (s *stubClusterUpdater) UpdateEnableAdvancedGPUMetrics(_ context.Context, _ int, value interface{}) error {
	s.advancedGPUMetricsCalls = append(s.advancedGPUMetricsCalls, value)
	return s.err
}

func (s *stubClusterUpdater) UpdateProjectManualOverQuota(_ context.Context, _ int, value interface{}) error {
	s.projectManualOverQuotaCalls = append(s.projectManualOverQuotaCalls, value)
	return s.err
}

func (s *stubClusterUpdater) UpdateCPUResourcesQuota(_ context.Context, _ int, value interface{}) error {
	s.cpuResourcesQuotaCalls = append(s.cpuResourcesQuotaCalls, value)
	return s.err
}

func TestProjectManualOverQuota_OnSet_CallsClusterUpdater(t *testing.T) {
	updater := &stubClusterUpdater{}
	defs := BuildSettingDefinitions(&OnSetDeps{ClusterUpdater: updater})
	value := defs[SettingKeyProjectManualOverQuota]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true}, updater.projectManualOverQuotaCalls)
	assert.Empty(t, updater.cpuResourcesQuotaCalls)
}

func TestProjectManualOverQuota_OnSet_PropagatesError(t *testing.T) {
	updater := &stubClusterUpdater{err: errors.New("cluster service down")}
	defs := BuildSettingDefinitions(&OnSetDeps{ClusterUpdater: updater})
	value := defs[SettingKeyProjectManualOverQuota]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ProjectManualOverQuota")
	assert.Contains(t, err.Error(), "cluster service down")
}

func TestCPUResourcesQuota_OnSet_CallsClusterUpdater(t *testing.T) {
	updater := &stubClusterUpdater{}
	defs := BuildSettingDefinitions(&OnSetDeps{ClusterUpdater: updater})
	value := defs[SettingKeyCPUResourcesQuota]

	_, err := value.OnSet(context.Background(), 42, true)
	assert.NoError(t, err)
	assert.Equal(t, []interface{}{true}, updater.cpuResourcesQuotaCalls)
	assert.Empty(t, updater.projectManualOverQuotaCalls)
}

func TestCPUResourcesQuota_OnSet_PropagatesError(t *testing.T) {
	updater := &stubClusterUpdater{err: errors.New("cluster service down")}
	defs := BuildSettingDefinitions(&OnSetDeps{ClusterUpdater: updater})
	value := defs[SettingKeyCPUResourcesQuota]

	_, err := value.OnSet(context.Background(), 42, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CPUResourcesQuota")
	assert.Contains(t, err.Error(), "cluster service down")
}
