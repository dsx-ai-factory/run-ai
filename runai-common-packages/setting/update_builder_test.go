/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setting

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"

	manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
	managerMock "github.com/run-ai/runai/runai-common-packages/setting/manager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenantID = "test-tenant-123"

func setupUpdateBuilderTest(t *testing.T) {
	once = sync.Once{}
	settingsManager = nil

	settingsDal = &managerMock.SettingsDal{}
	var dalInterface manager.SettingsDal = settingsDal
	config := &SettingConfig{}
	Init(config, dalInterface)
}

func TestUpdateBuilder_Constructor(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	assert.NotNil(t, builder, "Update should return a non-nil updateBuilder")
	assert.Equal(t, testTenantID, builder.tenantId, "updateBuilder should store tenantId")
	assert.Equal(t, value, builder.value, "updateBuilder should store value")
	assert.True(t, builder.updateByUser, "updateBuilder should default to updateByUser=true")
}

func TestUpdateBuilder_UpdatedByUser_Method(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	result := builder.UpdatedByUser()
	assert.Same(t, builder, result, "UpdatedByUser should return same instance for chaining")
	assert.True(t, builder.updateByUser, "UpdatedByUser should set updateByUser to true")
}

func TestUpdateBuilder_UpdatedBySystem_Method(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	result := builder.UpdatedBySystem()
	assert.Same(t, builder, result, "UpdatedBySystem should return same instance for chaining")
	assert.False(t, builder.updateByUser, "UpdatedBySystem should set updateByUser to false")
}

func TestUpdateBuilder_MethodChaining_UserThenSystem(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	result := builder.UpdatedByUser().UpdatedBySystem()
	assert.Same(t, builder, result, "Method chaining should return same instance")
	assert.False(t, builder.updateByUser, "Final call should determine updateByUser value")
}

func TestUpdateBuilder_MethodChaining_SystemThenUser(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	result := builder.UpdatedBySystem().UpdatedByUser()
	assert.Same(t, builder, result, "Method chaining should return same instance")
	assert.True(t, builder.updateByUser, "Final call should determine updateByUser value")
}

func TestUpdateBuilder_Execute_Success_DefaultUserUpdate(t *testing.T) {
	setupUpdateBuilderTest(t)

	err := Register[*TestSetting]()
	require.NoError(t, err)

	value := TestSetting{Name: "test", Count: 42}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting", testTenantID, true, valueBytes).Return(nil).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err = builder.Execute()
	assert.NoError(t, err, "Execute should succeed")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Success_ExplicitUserUpdate(t *testing.T) {
	setupUpdateBuilderTest(t)

	err := Register[*TestSetting]()
	require.NoError(t, err)

	value := TestSetting{Name: "user-update", Count: 100}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting", testTenantID, true, valueBytes).Return(nil).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err = builder.UpdatedByUser().Execute()

	assert.NoError(t, err, "Execute should succeed")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Success_SystemUpdate_Direct(t *testing.T) {
	setupUpdateBuilderTest(t)
	require.NoError(t, Register[*TestSetting]())

	value := TestSetting{Name: "system-update", Count: 200}
	valueBytes, _ := json.Marshal(value)
	settingsDal.On("SaveSetting", "test-setting", testTenantID, false, valueBytes).Return(nil).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err := builder.UpdatedBySystem().Execute()

	assert.NoError(t, err, "Execute should succeed")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Error_UnregisteredSetting(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "test", Count: 42}
	builder := Update[*TestSetting](testTenantID, value)

	err := builder.Execute()

	assert.Error(t, err, "Execute should fail for unregistered setting")
	assert.Contains(t, err.Error(), "setting is not registered", "Error should mention unregistered setting")
}

func TestUpdateBuilder_Execute_Error_InvalidValue(t *testing.T) {
	setupUpdateBuilderTest(t)

	err := Register[*TestSetting]()
	require.NoError(t, err)

	invalidValue := TestSetting{Name: "", Count: -1}
	builder := Update[*TestSetting](testTenantID, invalidValue)

	err = builder.Execute()

	assert.Error(t, err, "Execute should fail for invalid value")
	assert.Contains(t, err.Error(), "not valid", "Error should mention validation failure")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_Error_DatabaseError(t *testing.T) {
	setupUpdateBuilderTest(t)

	err := Register[*TestSetting]()
	require.NoError(t, err)

	value := TestSetting{Name: "test", Count: 42}
	valueBytes, _ := json.Marshal(value)

	dbError := errors.New("database connection failed")
	settingsDal.On("SaveSetting", "test-setting", testTenantID, true, valueBytes).Return(dbError).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err = builder.Execute()

	assert.Error(t, err, "Execute should fail when database operation fails")
	assert.Contains(t, err.Error(), "database connection failed", "Error should contain original database error")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_ComplexChaining(t *testing.T) {
	setupUpdateBuilderTest(t)
	require.NoError(t, Register[*TestSetting]())

	value := TestSetting{Name: "complex", Count: 500}
	valueBytes, _ := json.Marshal(value)
	settingsDal.On("SaveSetting", "test-setting", testTenantID, false, valueBytes).Return(nil).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err := builder.UpdatedByUser().UpdatedBySystem().UpdatedByUser().UpdatedBySystem().Execute()

	assert.NoError(t, err, "Execute should succeed after complex chaining")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_Execute_NilReturn(t *testing.T) {
	setupUpdateBuilderTest(t)

	err := Register[*TestSetting]()
	require.NoError(t, err)

	value := TestSetting{Name: "nil-test", Count: 1}
	valueBytes, _ := json.Marshal(value)

	settingsDal.On("SaveSetting", "test-setting", testTenantID, true, valueBytes).Return(nil).Once()

	builder := Update[*TestSetting](testTenantID, value)
	err = builder.Execute()

	assert.NoError(t, err, "Execute should return nil on success")
	settingsDal.AssertExpectations(t)
}

func TestUpdateBuilder_StructIntegrity(t *testing.T) {
	setupUpdateBuilderTest(t)

	value := TestSetting{Name: "integrity-test", Count: 999}
	builder := Update[*TestSetting](testTenantID, value)

	assert.Equal(t, testTenantID, builder.tenantId)
	assert.Equal(t, value, builder.value)
	assert.True(t, builder.updateByUser)

	builder.UpdatedBySystem()
	assert.False(t, builder.updateByUser)

	builder.UpdatedByUser()
	assert.True(t, builder.updateByUser)

	assert.Equal(t, testTenantID, builder.tenantId)
	assert.Equal(t, value, builder.value)
}

// ============================================================================
// TEST MOCK DATA
// ============================================================================

type TestSetting struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t TestSetting) Key() string {
	return "test-setting"
}

func (t *TestSetting) SetDefault() {
	t.Name = "default"
	t.Count = 0
}

func (t TestSetting) DisplayName() string {
	return "Test Setting"
}

func (t TestSetting) Description() string {
	return "A test setting"
}

func (t TestSetting) Category() string {
	return "test"
}

func (t TestSetting) Stage() manager.Stage {
	return manager.Alpha
}

func (t TestSetting) Validate(value interface{}) bool {
	if ts, ok := value.(TestSetting); ok {
		return len(ts.Name) > 0 && ts.Count >= 0
	}
	return false
}
