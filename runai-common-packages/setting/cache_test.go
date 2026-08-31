/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setting

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	dal "github.com/run-ai/runai/runai-common-packages/setting/database"
	managerMock "github.com/run-ai/runai/runai-common-packages/setting/manager/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initTestWithCache(t *testing.T) {
	_ = cache.Close()
	once = sync.Once{}
	settingsManager = nil

	enabled := true
	cacheConfig := &cache_config.Cache{
		Type:            cache_config.CTInMemory,
		Enabled:         &enabled,
		TtlSeconds:      3600,
		MaxStoreSizeMB:  100,
		MaxRecordSizeKB: 1000,
		ServiceName:     fmt.Sprintf("test-cache-%d", time.Now().UnixNano()),
	}
	if err := cache.Init(cacheConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize cache: %v", err))
	}

	settingsDal = managerMock.NewSettingsDal(t)
	config := &SettingConfig{
		CacheEnabled: true,
	}
	Init(config, settingsDal)
}

func TestGet_WithCache_CacheHit(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-cache-hit")

	// First call - cache miss, fetches from DB
	settingValue := TestSetting1{Name: "from-db", Count: 88}
	valueBytes, _ := json.Marshal(settingValue)
	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             valueBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil).Once()

	result1, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "from-db", (*result1).Name)

	result2, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "from-db", (*result2).Name)

	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 1)
}

func TestGet_WithCache_CacheMiss(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-cache-miss")
	settingValue := TestSetting1{Name: "from-db", Count: 88}
	valueBytes, _ := json.Marshal(settingValue)

	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             valueBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}

	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil).Once()

	result, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "from-db", (*result).Name)
	assert.Equal(t, 88, (*result).Count)

	settingsDal.AssertExpectations(t)
}

func TestGetAll_WithCache_CacheHit(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-getall-hit")

	setting1Value := TestSetting1{Name: "db-setting1", Count: 10}
	setting1Bytes, _ := json.Marshal(setting1Value)

	setting2Value := TestSetting2{Enabled: true, Message: "db-setting2"}
	setting2Bytes, _ := json.Marshal(setting2Value)

	dbSettings := &[]dal.Setting{
		{
			Key:               "test-setting1",
			Value:             setting1Bytes,
			TenantID:          uniqueTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: true,
		},
		{
			Key:               "test-setting2",
			Value:             setting2Bytes,
			TenantID:          uniqueTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: false,
		},
	}

	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(dbSettings, nil).Once()
	results1, err := GetAll(uniqueTenantID, "")
	assert.NoError(t, err)
	assert.Len(t, results1, 2)

	results2, err := GetAll(uniqueTenantID, "")
	assert.NoError(t, err)
	assert.Len(t, results2, 2)

	settingsDal.AssertNumberOfCalls(t, "GetSettingsByTenantId", 1)
}

func TestGetAll_WithCache_CacheMiss(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)
	err = Register[*TestSetting2]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-getall-miss")

	setting1Value := TestSetting1{Name: "db-setting1", Count: 10}
	setting1Bytes, _ := json.Marshal(setting1Value)

	setting2Value := TestSetting2{Enabled: true, Message: "db-setting2"}
	setting2Bytes, _ := json.Marshal(setting2Value)

	dbSettings := &[]dal.Setting{
		{
			Key:               "test-setting1",
			Value:             setting1Bytes,
			TenantID:          uniqueTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: true,
		},
		{
			Key:               "test-setting2",
			Value:             setting2Bytes,
			TenantID:          uniqueTenantID,
			UpdatedAt:         time.Now(),
			UpdatedByCustomer: false,
		},
	}

	settingsDal.On("GetSettingsByTenantId", uniqueTenantID).Return(dbSettings, nil).Once()

	results, err := GetAll(uniqueTenantID, "")
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	settingsDal.AssertExpectations(t)
}

func TestUpdate_WithCache_InvalidatesCache(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-tenant-update")

	settingValue := TestSetting1{Name: "old-value", Count: 1}
	valueBytes, _ := json.Marshal(settingValue)
	dbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             valueBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil).Once()

	result1, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "old-value", (*result1).Name)

	updateValue := TestSetting1{Name: "new-value", Count: 2}
	updateValueBytes, _ := json.Marshal(updateValue)
	settingsDal.On("SaveSetting", "test-setting1", uniqueTenantID, true, updateValueBytes).Return(nil).Once()

	err = Update[*TestSetting1](uniqueTenantID, updateValue).Execute()
	assert.NoError(t, err)

	newDbSetting := &dal.Setting{
		Key:               "test-setting1",
		Value:             updateValueBytes,
		TenantID:          uniqueTenantID,
		UpdatedAt:         time.Now(),
		UpdatedByCustomer: true,
	}
	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(newDbSetting, nil).Once()

	result2, err := Get[*TestSetting1](uniqueTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "new-value", (*result2).Name)

	settingsDal.AssertExpectations(t)
	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 2)
}

func TestCache_DifferentTenantsHaveSeparateCaches(t *testing.T) {
	initTestWithCache(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	tenant1 := generateUniqueTenantID("tenant1")
	tenant2 := generateUniqueTenantID("tenant2")

	setting1Value := TestSetting1{Name: "tenant1-value", Count: 1}
	setting1Bytes, _ := json.Marshal(setting1Value)
	dbSetting1 := &dal.Setting{
		Key:      "test-setting1",
		Value:    setting1Bytes,
		TenantID: tenant1,
	}

	setting2Value := TestSetting1{Name: "tenant2-value", Count: 2}
	setting2Bytes, _ := json.Marshal(setting2Value)
	dbSetting2 := &dal.Setting{
		Key:      "test-setting1",
		Value:    setting2Bytes,
		TenantID: tenant2,
	}

	settingsDal.On("GetSettingByKey", "test-setting1", tenant1).Return(dbSetting1, nil).Once()
	settingsDal.On("GetSettingByKey", "test-setting1", tenant2).Return(dbSetting2, nil).Once()

	result1, err := Get[*TestSetting1](tenant1)
	assert.NoError(t, err)
	assert.Equal(t, "tenant1-value", (*result1).Name)

	result2, err := Get[*TestSetting1](tenant2)
	assert.NoError(t, err)
	assert.Equal(t, "tenant2-value", (*result2).Name)

	result1Again, err := Get[*TestSetting1](tenant1)
	assert.NoError(t, err)
	assert.Equal(t, "tenant1-value", (*result1Again).Name)

	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 2)
}

func TestCacheDisabled_NoCache(t *testing.T) {
	initTest(t)

	err := Register[*TestSetting1]()
	require.NoError(t, err)

	uniqueTenantID := generateUniqueTenantID("test-no-cache")
	settingValue := TestSetting1{Name: "from-db", Count: 1}
	valueBytes, _ := json.Marshal(settingValue)
	dbSetting := &dal.Setting{
		Key:      "test-setting1",
		Value:    valueBytes,
		TenantID: uniqueTenantID,
	}

	settingsDal.On("GetSettingByKey", "test-setting1", uniqueTenantID).Return(dbSetting, nil)

	_, _ = Get[*TestSetting1](uniqueTenantID)
	_, _ = Get[*TestSetting1](uniqueTenantID)
	_, _ = Get[*TestSetting1](uniqueTenantID)

	settingsDal.AssertNumberOfCalls(t, "GetSettingByKey", 3)
}

func initTestWithCacheForGlobalAPI(t *testing.T) {
	_ = cache.Close()
	once = sync.Once{}
	settingsManager = nil

	enabled := true
	cacheConfig := &cache_config.Cache{
		Type:            cache_config.CTInMemory,
		Enabled:         &enabled,
		TtlSeconds:      3600,
		MaxStoreSizeMB:  100,
		MaxRecordSizeKB: 1000,
		ServiceName:     fmt.Sprintf("test-cache-%d", time.Now().UnixNano()),
	}
	if err := cache.Init(cacheConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize cache: %v", err))
	}

	settingsDal = managerMock.NewSettingsDal(t)

	config := &SettingConfig{
		CacheEnabled: true,
	}
	Init(config, settingsDal)
}
