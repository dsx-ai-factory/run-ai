/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_MaxValidationFailed_MaxStoreSizeMB(t *testing.T) {
	defer config.FlushAll()
	t.Setenv("CACHE_TYPE", "inmemory")
	t.Setenv("CACHE_URL", "localhost:1234")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")

	t.Setenv("CACHE_MAX_STORE_SIZE_MB", "4001")

	err := InitConfig()
	require.Equal(t, "failed to register cache config: failed to validate config struct Cache, error: Key: 'Cache.MaxStoreSizeMB' Error:Field validation for 'MaxStoreSizeMB' failed on the 'max' tag", err.Error())
}

func TestInitConfig_MinValidationFailed_MaxStoreSizeMB(t *testing.T) {
	defer config.FlushAll()
	t.Setenv("CACHE_TYPE", "inmemory")
	t.Setenv("CACHE_URL", "localhost:1234")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")

	t.Setenv("CACHE_MAX_STORE_SIZE_MB", "1")

	err := InitConfig()
	require.Equal(t, "failed to register cache config: failed to validate config struct Cache, error: Key: 'Cache.MaxStoreSizeMB' Error:Field validation for 'MaxStoreSizeMB' failed on the 'min' tag", err.Error())
}

func TestInitConfig_ValidationFailed_MaxRecordSizeKB(t *testing.T) {
	defer config.FlushAll()
	t.Setenv("CACHE_TYPE", "inmemory")
	t.Setenv("CACHE_URL", "localhost:1234")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")

	t.Setenv("CACHE_MAX_RECORD_SIZE_KB", "30001")

	err := InitConfig()
	require.Equal(t, "failed to register cache config: failed to validate config struct Cache, error: Key: 'Cache.MaxRecordSizeKB' Error:Field validation for 'MaxRecordSizeKB' failed on the 'max' tag", err.Error())
}

func TestInitConfig_UrlRequired(t *testing.T) {
	for _, cacheType := range []CacheType{CTRedis, CTNats, CTKeyDB} {
		defer config.FlushAll()
		t.Setenv("CACHE_TYPE", string(cacheType))
		t.Setenv("CACHE_URL", "")
		t.Setenv("SERVICE_NAME", "test-service")
		t.Setenv("SECRETS_MOUNT_PATH", ".")
		err := InitConfig()
		require.Equal(t, "failed to register cache config: failed to validate config struct Cache, error: Key: 'Cache.Url' Error:Field validation for 'Url' failed on the 'required_unless' tag", err.Error())
	}
}

func TestInitConfig_UrlNotRequired_Inmemory(t *testing.T) {
	defer config.FlushAll()
	t.Setenv("CACHE_TYPE", "inmemory")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")
	err := InitConfig()
	require.NoError(t, err)
}

func TestInitConfig_SupportedType(t *testing.T) {
	for _, cacheType := range []CacheType{CTRedis, CTNats, CTInMemory, CTKeyDB} {
		t.Run(string(cacheType), func(t *testing.T) {
			defer config.FlushAll()
			t.Setenv("CACHE_TYPE", string(cacheType))
			t.Setenv("CACHE_URL", "localhost:1234")
			t.Setenv("SERVICE_NAME", "test-service")
			t.Setenv("SECRETS_MOUNT_PATH", ".")
			err := InitConfig()
			require.NoError(t, err)
		})
	}
}

func TestInitConfig_UnsupportedType(t *testing.T) {
	defer config.FlushAll()
	t.Setenv("CACHE_TYPE", "unsupported")
	t.Setenv("CACHE_URL", "localhost:1234")
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")
	err := InitConfig()
	require.Equal(t, "failed to register cache config: failed to validate config struct Cache, error: Key: 'Cache.Type' Error:Field validation for 'Type' failed on the 'oneof' tag", err.Error())
}

func TestInitConfig_Disabled(t *testing.T) {
	defer config.FlushAll()
	e := false
	err := config.Set(&Cache{Enabled: &e})
	require.NoError(t, err)
	conf := config.Get[*Cache]()
	require.False(t, *conf.Enabled)
}

func TestInitConfig_EnabledByDefault(t *testing.T) {
	defer config.FlushAll()
	c := Cache{Type: CTInMemory}
	err := config.SetWithDefaults(&c)
	require.NoError(t, err)
	conf := config.Get[*Cache]()
	require.True(t, *conf.Enabled)
}

func TestInitConfig_Enabled(t *testing.T) {
	defer config.FlushAll()
	e := true
	err := config.Set(&Cache{Enabled: &e})
	require.NoError(t, err)
	conf := config.Get[*Cache]()
	require.True(t, *conf.Enabled)
}
