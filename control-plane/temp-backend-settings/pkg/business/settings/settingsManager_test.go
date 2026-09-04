/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== helpers ====================

type fakeSettingProvider struct {
	overrides map[string]SettingOverride
	err       error
}

func (f *fakeSettingProvider) FetchSettings(_ context.Context, _ int) (map[string]SettingOverride, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.overrides, nil
}

// ==================== isFalsy ====================

func TestIsFalsy(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil is falsy", nil, true},
		{"false is falsy", false, true},
		{"true is not falsy", true, false},
		{"string is not falsy", "hello", false},
		{"zero int is not falsy", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isFalsy(tt.value))
		})
	}
}

// ==================== validateSettingValue ====================

func TestValidateSettingValue_NoValidator(t *testing.T) {
	setting := SettingDefinition{Key: "test"}
	assert.NoError(t, validateSettingValue(setting, "anything"))
}

func TestValidateSettingValue_ValidValue(t *testing.T) {
	setting := SettingDefinition{Key: "test", Validate: ValidateBoolean}
	assert.NoError(t, validateSettingValue(setting, true))
}

func TestValidateSettingValue_InvalidValue(t *testing.T) {
	setting := SettingDefinition{Key: "test", Validate: ValidateBoolean}
	err := validateSettingValue(setting, "not-a-bool")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "true or false")
}

// ==================== isPermitted ====================

func TestIsPermitted_Admin(t *testing.T) {
	assert.True(t, isPermitted(SettingDefinition{Privilege: PrivilegeAdmin}))
}

func TestIsPermitted_NonAdmin(t *testing.T) {
	assert.False(t, isPermitted(SettingDefinition{Privilege: PrivilegeReader}))
	assert.False(t, isPermitted(SettingDefinition{Privilege: PrivilegeWriter}))
	assert.False(t, isPermitted(SettingDefinition{Privilege: PrivilegeTest}))
}

// ==================== filterDisabledAlphaFlag ====================

func TestFilterDisabledAlphaFlag_NonAlpha(t *testing.T) {
	m := &SettingsManager{showAlphaFlags: false}
	assert.True(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageBeta}, false))
	assert.True(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageGA}, false))
}

func TestFilterDisabledAlphaFlag_AlphaShowEnabled(t *testing.T) {
	m := &SettingsManager{showAlphaFlags: true}
	assert.True(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageAlpha}, false))
}

func TestFilterDisabledAlphaFlag_AlphaShowDisabled_ValueTrue(t *testing.T) {
	m := &SettingsManager{showAlphaFlags: false}
	assert.True(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageAlpha}, true))
}

func TestFilterDisabledAlphaFlag_AlphaShowDisabled_ValueFalse(t *testing.T) {
	m := &SettingsManager{showAlphaFlags: false}
	assert.False(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageAlpha}, false))
}

func TestFilterDisabledAlphaFlag_AlphaShowDisabled_NonBoolValue(t *testing.T) {
	m := &SettingsManager{showAlphaFlags: false}
	assert.False(t, m.filterDisabledAlphaFlag(SettingDefinition{Stage: StageAlpha}, "string"))
}

// ==================== filterUnsupportedSingleClustersFlags ====================

func TestFilterUnsupported_NoMinVersion(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: ""}
	assert.True(t, filterUnsupportedSingleClustersFlags(s, []string{"2.10"}, false))
}

func TestFilterUnsupported_NoClusters(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: "2.10"}
	assert.True(t, filterUnsupportedSingleClustersFlags(s, []string{}, false))
}

func TestFilterUnsupported_ValueFalse(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: "2.10"}
	assert.True(t, filterUnsupportedSingleClustersFlags(s, []string{"1.0"}, true))
}

func TestFilterUnsupported_AllSupported(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: "2.10"}
	assert.True(t, filterUnsupportedSingleClustersFlags(s, []string{"2.10", "2.11"}, false))
}

func TestFilterUnsupported_SomeSupported(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: "2.10"}
	assert.True(t, filterUnsupportedSingleClustersFlags(s, []string{"2.10", "2.9"}, false))
}

func TestFilterUnsupported_AllUnsupported(t *testing.T) {
	s := SettingDefinition{MinimumSingleClusterVersion: "2.10"}
	assert.False(t, filterUnsupportedSingleClustersFlags(s, []string{"2.9", "2.8"}, false))
}

// ==================== filterMigrated ====================

func TestFilterMigrated_ExcludeFalse(t *testing.T) {
	assert.True(t, filterMigrated(SettingDefinition{Migrated: true}, false))
	assert.True(t, filterMigrated(SettingDefinition{Migrated: false}, false))
}

func TestFilterMigrated_ExcludeTrue_Migrated(t *testing.T) {
	assert.False(t, filterMigrated(SettingDefinition{Migrated: true}, true))
}

func TestFilterMigrated_ExcludeTrue_NotMigrated(t *testing.T) {
	assert.True(t, filterMigrated(SettingDefinition{Migrated: false}, true))
}

// ==================== getUnsupportedVersions ====================

func TestGetUnsupportedVersions(t *testing.T) {
	result := getUnsupportedVersions([]string{"2.10", "2.9", "", "2.11"}, "2.10")
	assert.ElementsMatch(t, []string{"2.9", ""}, result)
}

func TestGetUnsupportedVersions_AllSupported(t *testing.T) {
	result := getUnsupportedVersions([]string{"2.10", "2.11"}, "2.10")
	assert.Empty(t, result)
}

func TestGetUnsupportedVersions_AllUnsupported(t *testing.T) {
	result := getUnsupportedVersions([]string{"1.0", "2.5"}, "2.10")
	assert.Len(t, result, 2)
}

// ==================== resolveValue ====================

func TestResolveValue_FromSettings(t *testing.T) {
	allSettings := map[string]interface{}{"key1": "override"}
	result := resolveValue("key1", "default", allSettings)
	assert.Equal(t, "override", result)
}

func TestResolveValue_Default(t *testing.T) {
	allSettings := map[string]interface{}{}
	result := resolveValue("key1", "default", allSettings)
	assert.Equal(t, "default", result)
}

// ==================== mergeOverrides ====================

func TestMergeOverrides_IncomingOverrides(t *testing.T) {
	existing := SettingOverride{Value: "old", Description: "old-desc", DisplayName: "old-name", Stage: "Beta"}
	incoming := SettingOverride{Value: "new", Description: "new-desc", DisplayName: "new-name", Stage: "GA"}
	merged := mergeOverrides(existing, incoming)
	assert.Equal(t, "new", merged.Value)
	assert.Equal(t, "new-desc", merged.Description)
	assert.Equal(t, "new-name", merged.DisplayName)
	assert.Equal(t, "GA", merged.Stage)
}

func TestMergeOverrides_EmptyIncoming(t *testing.T) {
	existing := SettingOverride{Value: "old", Description: "old-desc", DisplayName: "old-name", Stage: "Beta"}
	incoming := SettingOverride{}
	merged := mergeOverrides(existing, incoming)
	assert.Equal(t, "old", merged.Value)
	assert.Equal(t, "old-desc", merged.Description)
	assert.Equal(t, "old-name", merged.DisplayName)
	assert.Equal(t, "Beta", merged.Stage)
}

func TestMergeOverrides_PartialIncoming(t *testing.T) {
	existing := SettingOverride{Value: "old", Description: "old-desc"}
	incoming := SettingOverride{Description: "new-desc"}
	merged := mergeOverrides(existing, incoming)
	assert.Equal(t, "old", merged.Value)
	assert.Equal(t, "new-desc", merged.Description)
}

// ==================== PostSaveError ====================

func TestPostSaveError(t *testing.T) {
	err := &PostSaveError{Err: errors.New("some failure")}
	assert.Equal(t, "some failure", err.Error())

	var postSaveErr *PostSaveError
	assert.True(t, errors.As(err, &postSaveErr))
}

// ==================== fetchProviderOverrides ====================

func TestFetchProviderOverrides_NoProviders(t *testing.T) {
	m := &SettingsManager{settingProviders: nil}
	allSettings := map[string]interface{}{}
	overrides := m.fetchProviderOverrides(context.Background(), 1, allSettings)
	assert.Empty(t, overrides)
}

func TestFetchProviderOverrides_SingleProvider(t *testing.T) {
	provider := &fakeSettingProvider{
		overrides: map[string]SettingOverride{
			"key1": {Value: "provider-value", Description: "provider-desc"},
		},
	}
	m := &SettingsManager{settingProviders: []SettingProvider{provider}}
	allSettings := map[string]interface{}{}

	overrides := m.fetchProviderOverrides(context.Background(), 1, allSettings)

	assert.Equal(t, "provider-value", allSettings["key1"])
	assert.Equal(t, "provider-desc", overrides["key1"].Description)
}

func TestFetchProviderOverrides_ProviderError(t *testing.T) {
	provider := &fakeSettingProvider{err: errors.New("fetch failed")}
	m := &SettingsManager{settingProviders: []SettingProvider{provider}}
	allSettings := map[string]interface{}{}

	overrides := m.fetchProviderOverrides(context.Background(), 1, allSettings)

	assert.Empty(t, overrides)
	assert.Empty(t, allSettings)
}

func TestFetchProviderOverrides_MultipleProvidersMerge(t *testing.T) {
	p1 := &fakeSettingProvider{
		overrides: map[string]SettingOverride{
			"key1": {Value: "v1", Description: "d1"},
		},
	}
	p2 := &fakeSettingProvider{
		overrides: map[string]SettingOverride{
			"key1": {Description: "d2", DisplayName: "name2"},
		},
	}
	m := &SettingsManager{settingProviders: []SettingProvider{p1, p2}}
	allSettings := map[string]interface{}{}

	overrides := m.fetchProviderOverrides(context.Background(), 1, allSettings)

	assert.Equal(t, "d2", overrides["key1"].Description)
	assert.Equal(t, "name2", overrides["key1"].DisplayName)
}

func TestFetchProviderOverrides_NilValueNotWrittenToAllSettings(t *testing.T) {
	provider := &fakeSettingProvider{
		overrides: map[string]SettingOverride{
			"key1": {Value: nil, Description: "desc-only"},
		},
	}
	m := &SettingsManager{settingProviders: []SettingProvider{provider}}
	allSettings := map[string]interface{}{}

	m.fetchProviderOverrides(context.Background(), 1, allSettings)

	_, exists := allSettings["key1"]
	assert.False(t, exists)
}
