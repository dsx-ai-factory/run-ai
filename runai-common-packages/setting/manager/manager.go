/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	settingDal "github.com/run-ai/runai/runai-common-packages/setting/database"
	errHandler "github.com/run-ai/runai/runai-common-packages/setting/error"
	"gorm.io/gorm"
)

type Stage string

const (
	GA     Stage = "GA"
	Beta   Stage = "Beta"
	Alpha  Stage = "Alpha"
	Legacy Stage = "Legacy"
)

type Settinger interface {
	Key() string
	SetDefault()
	DisplayName() string
	Description() string
	Category() string
	Stage() Stage
	Validate(interface{}) bool
}

type SettingerWithTransform interface {
	Settinger
	TransformForMigration(interface{}) interface{}
}

type SettingsDal interface {
	GetSettingByKey(key, tenantId string) (*settingDal.Setting, error)
	GetSettingsByTenantId(tenantId string) (*[]settingDal.Setting, error)
	SaveSetting(key, tenantId string, updateByUser bool, value []byte) error
	SaveSettingByKeyForTenantsList(key string, tenantSettings []settingDal.TenantSettingInput) error
	ShouldMigrateSetting(key string) (bool, error)
}

type TenantSettingValueForBackendResponse struct {
	TenantId int
	Value    interface{}
}
type SettingValuesForAllTenantsBackendResponse struct {
	Key    string
	Values []TenantSettingValueForBackendResponse
}

type CacheConfig struct {
	Enabled           bool
	ExpirationMinutes *int
}

type Manager struct {
	settingsMap  map[string]Settinger
	settingsDal  SettingsDal
	cacheEnabled bool
}

func NewManager(settingsDal SettingsDal, cacheConfig *CacheConfig) *Manager {
	var settingsDalForUsage SettingsDal
	if settingsDal != nil {
		settingsDalForUsage = settingsDal
	} else {
		settingsDalForUsage = settingDal.NewSettingDAL()
	}

	cacheEnabled := false
	if cacheConfig != nil && cacheConfig.Enabled {
		initCache(cacheConfig.ExpirationMinutes)
		cacheEnabled = true
	}

	return &Manager{
		settingsMap:  make(map[string]Settinger),
		settingsDal:  settingsDalForUsage,
		cacheEnabled: cacheEnabled,
	}
}

func (m *Manager) Register(setting Settinger) error {
	key := setting.Key()
	if _, exists := m.settingsMap[key]; !exists {
		m.settingsMap[key] = setting
	}
	return nil
}

func (m *Manager) Get(ctx context.Context, tenantId string, key string) (Settinger, error) {
	registeredSetting, ok := m.settingsMap[key]
	if !ok {
		return nil, errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to get setting by key: '%s'. setting is not registered", key),
		}
	}

	if m.cacheEnabled {
		cached := getCachedSetting(ctx, tenantId, key, registeredSetting)
		if cached != nil {
			return cached, nil
		}
	}

	dbSetting, err := m.settingsDal.GetSettingByKey(key, tenantId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newInstance := m.createNewInstance(registeredSetting)
			newInstance.SetDefault()
			if m.cacheEnabled {
				setCachedSetting(ctx, tenantId, key, newInstance)
			}
			return newInstance, nil
		} else {
			return nil, errHandler.ErrorInternal{
				Message: fmt.Sprintf("failed to get setting by key: '%s'. db error: %v", key, err),
			}
		}
	}

	if dbSetting == nil || dbSetting.Value == nil {
		newInstance := m.createNewInstance(registeredSetting)
		newInstance.SetDefault()
		if m.cacheEnabled {
			setCachedSetting(ctx, tenantId, key, newInstance)
		}
		return newInstance, nil
	}
	newInstance := m.createNewInstance(registeredSetting)
	err = json.Unmarshal(dbSetting.Value, newInstance)
	if err != nil {
		return nil, errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to get setting by key: '%s'. failed to unmarshal value. err: %v", key, err),
		}
	}
	if m.cacheEnabled {
		setCachedSetting(ctx, tenantId, key, newInstance)
	}
	return newInstance, nil
}

func (m *Manager) GetAll(ctx context.Context, tenantId, category string) ([]Settinger, error) {
	if m.cacheEnabled {
		cachedDBSettings := getCachedSettingsList(ctx, tenantId, category)
		if cachedDBSettings != nil {
			dbSettingsMap := m.convertSettingsSliceToMap(cachedDBSettings)
			return m.createSettingListResponse(category, dbSettingsMap), nil
		}
	}

	dbSettings, err := m.settingsDal.GetSettingsByTenantId(tenantId)
	if err != nil {
		return nil, errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to get settings by tenantId: '%s' from db. error: %v", tenantId, err),
		}
	}

	categoryDBSettings := m.filterDBSettingsByCategory(dbSettings, category)

	if m.cacheEnabled {
		setCachedSettingsList(ctx, tenantId, category, categoryDBSettings)
	}

	dbSettingsMap := m.convertSettingsSliceToMap(categoryDBSettings)
	return m.createSettingListResponse(category, dbSettingsMap), nil
}

func (m *Manager) Update(ctx context.Context, tenantId string, key string, updateByCustomer bool, value interface{}) error {
	registeredSetting, ok := m.settingsMap[key]
	if !ok {
		return errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. setting is not registered", key),
		}
	}
	isValid := registeredSetting.Validate(value)
	if !isValid {
		return errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. input value: '%v' is not valid", key, value),
		}
	}
	dataBytes, err := json.Marshal(value)
	if err != nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. failed to marshal value: '%v'. err: %v", key, value, err),
		}
	}
	err = m.settingsDal.SaveSetting(key, tenantId, updateByCustomer, dataBytes)
	if err != nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. db error: %v", key, err),
		}
	}

	if m.cacheEnabled {
		invalidateSettingCache(ctx, tenantId, key)
		invalidateAllSettingsCache(ctx, tenantId, m.RegisteredCategories())
	}

	return nil
}

func (m *Manager) ShouldMigrateSetting(key string) (bool, error) {
	return m.settingsDal.ShouldMigrateSetting(key)
}

func (m *Manager) UpdateAllTenantsByKey(key string, tenantValues []TenantSettingValueForBackendResponse) error {
	registeredSetting, ok := m.settingsMap[key]
	if !ok {
		return errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. setting is not registered", key),
		}
	}
	transformer, ok := registeredSetting.(SettingerWithTransform)
	if !ok {
		return errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to update setting by key: '%s'. setting does not support migration transformation", key),
		}
	}
	tenantSettings, err := m.buildTenantSettings(key, tenantValues, registeredSetting, transformer)
	if err != nil {
		return err
	}
	if err := m.settingsDal.SaveSettingByKeyForTenantsList(key, tenantSettings); err != nil {
		return errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to batch update setting by key: '%s' for %d tenants. db error: %v", key, len(tenantValues), err),
		}
	}
	return nil
}

func (m *Manager) buildTenantSettings(key string, tenantValues []TenantSettingValueForBackendResponse, registeredSetting Settinger, transformer SettingerWithTransform) ([]settingDal.TenantSettingInput, error) {
	tenantSettings := make([]settingDal.TenantSettingInput, 0, len(tenantValues))

	for _, tv := range tenantValues {
		input, err := m.transformAndMarshalValue(key, tv.TenantId, tv.Value, registeredSetting, transformer)
		if err != nil {
			return nil, err
		}
		tenantSettings = append(tenantSettings, input)
	}
	return tenantSettings, nil
}

func (m *Manager) transformAndMarshalValue(key string, tenantId int, value interface{}, registeredSetting Settinger, transformer SettingerWithTransform) (settingDal.TenantSettingInput, error) {
	transformedValue := transformer.TransformForMigration(value)

	if !registeredSetting.Validate(transformedValue) {
		return settingDal.TenantSettingInput{}, errHandler.ErrorBadRequest{
			Message: fmt.Sprintf("failed to update setting by key: '%s' for tenant '%d'. input value: '%v' is not valid", key, tenantId, value),
		}
	}
	dataBytes, err := json.Marshal(transformedValue)
	if err != nil {
		return settingDal.TenantSettingInput{}, errHandler.ErrorInternal{
			Message: fmt.Sprintf("failed to update setting by key: '%s' for tenant '%d'. failed to marshal value: '%v'. err: %v", key, tenantId, value, err),
		}
	}
	return settingDal.TenantSettingInput{
		TenantID: strconv.Itoa(tenantId),
		Value:    dataBytes,
	}, nil
}

func (m *Manager) createSettingListResponse(category string, dbSettingsMap map[string]settingDal.Setting) []Settinger {
	settingsList := make([]Settinger, 0, len(m.settingsMap))

	for key, registeredSetting := range m.settingsMap {
		if category != "" && registeredSetting.Category() != category {
			continue
		}
		newInstance := m.createNewInstance(registeredSetting)
		if dbSetting, ok := dbSettingsMap[key]; ok && dbSetting.Value != nil {
			if err := json.Unmarshal(dbSetting.Value, newInstance); err != nil {
				newInstance.SetDefault()
			}
		} else {
			newInstance.SetDefault()
		}
		settingsList = append(settingsList, newInstance)
	}
	return settingsList
}

func (m *Manager) createNewInstance(original Settinger) Settinger {
	return createNewInstanceFromType(original)
}

func createNewInstanceFromType(original Settinger) Settinger {
	originalType := reflect.TypeOf(original)
	return reflect.New(originalType.Elem()).Interface().(Settinger)
}

func (m *Manager) filterDBSettingsByCategory(dbSettings *[]settingDal.Setting, category string) *[]settingDal.Setting {
	if category == "" || dbSettings == nil {
		return dbSettings
	}
	filtered := make([]settingDal.Setting, 0, len(*dbSettings))
	for i := range *dbSettings {
		dbSetting := (*dbSettings)[i]
		if registeredSetting, ok := m.settingsMap[dbSetting.Key]; ok && registeredSetting.Category() == category {
			filtered = append(filtered, dbSetting)
		}
	}
	return &filtered
}

func (m *Manager) RegisteredCategories() []string {
	seen := make(map[string]struct{})
	categories := make([]string, 0)
	for _, registeredSetting := range m.settingsMap {
		category := registeredSetting.Category()
		if category == "" {
			continue
		}
		if _, ok := seen[category]; ok {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	return categories
}

func (m *Manager) convertSettingsSliceToMap(settingsSlice *[]settingDal.Setting) map[string]settingDal.Setting {
	settingsMap := make(map[string]settingDal.Setting)
	if settingsSlice == nil {
		return settingsMap
	}
	for i := range *settingsSlice {
		setting := (*settingsSlice)[i]
		settingsMap[setting.Key] = setting
	}
	return settingsMap
}
