/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cache "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	settingDal "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/database"
)

type cachedSetting []byte
type cachedSettingsList []byte

const defaultCacheExpirationMinutes = 60

func initCache(cacheExpirationMinutes *int) {
	var cacheExpirationMinutesValue int
	if cacheExpirationMinutes == nil || *cacheExpirationMinutes <= 0 {
		cacheExpirationMinutesValue = defaultCacheExpirationMinutes
	} else {
		cacheExpirationMinutesValue = *cacheExpirationMinutes
	}
	cacheExpirationTime := time.Duration(cacheExpirationMinutesValue) * time.Minute
	_ = cache.RegisterStore[cachedSetting](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(10))
	_ = cache.RegisterStore[cachedSettingsList](options.Ttl(cacheExpirationTime), options.MaxRecordSizeKB(100))
}

func getCachedSetting(ctx context.Context, tenantId, settingKey string, registeredSetting Settinger) Settinger {
	cacheKey := buildCacheKey(tenantId, settingKey)
	cachedData, err := cache.Get[cachedSetting](ctx, cacheKey)
	if err != nil || cachedData == nil {
		return nil
	}
	newInstance := createNewInstance(registeredSetting)
	if err := json.Unmarshal(*cachedData, newInstance); err != nil {
		_ = cache.Del[cachedSetting](ctx, cacheKey)
		return nil
	}
	return newInstance
}

func setCachedSetting(ctx context.Context, tenantId, settingKey string, setting Settinger) {
	dataBytes, err := json.Marshal(setting)
	if err != nil {
		return
	}
	cacheKey := buildCacheKey(tenantId, settingKey)
	_ = cache.Set(ctx, cacheKey, cachedSetting(dataBytes))
}

func getCachedSettingsList(ctx context.Context, tenantId, category string) *[]settingDal.Setting {
	cacheKey := buildCacheKeyForAll(tenantId, category)
	cachedData, err := cache.Get[cachedSettingsList](ctx, cacheKey)
	if err != nil || cachedData == nil {
		return nil
	}

	var result []settingDal.Setting
	if err := json.Unmarshal(*cachedData, &result); err != nil {
		_ = cache.Del[cachedSettingsList](ctx, cacheKey)
		return nil
	}

	return &result
}

func setCachedSettingsList(ctx context.Context, tenantId, category string, dbSettings *[]settingDal.Setting) {
	dataBytes, err := json.Marshal(dbSettings)
	if err != nil {
		return
	}
	cacheKey := buildCacheKeyForAll(tenantId, category)
	_ = cache.Set(ctx, cacheKey, cachedSettingsList(dataBytes))
}

func invalidateSettingCache(ctx context.Context, tenantId, settingKey string) {
	cacheKey := buildCacheKey(tenantId, settingKey)
	_ = cache.Del[cachedSetting](ctx, cacheKey)
}

func invalidateAllSettingsCache(ctx context.Context, tenantId string, categories []string) {
	_ = cache.Del[cachedSettingsList](ctx, buildCacheKeyForAll(tenantId, ""))
	for _, category := range categories {
		_ = cache.Del[cachedSettingsList](ctx, buildCacheKeyForAll(tenantId, category))
	}
}

func buildCacheKey(tenantId, settingKey string) string {
	return fmt.Sprintf("setting:%s:%s", tenantId, settingKey)
}

func buildCacheKeyForAll(tenantId, category string) string {
	if category == "" {
		return fmt.Sprintf("settings:all:%s", tenantId)
	}
	return fmt.Sprintf("settings:all:%s:%s", tenantId, category)
}

func createNewInstance(original Settinger) Settinger {
	return createNewInstanceFromType(original)
}
