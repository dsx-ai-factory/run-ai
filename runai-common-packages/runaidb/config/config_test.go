/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

func setupRequiredConfigVars(t *testing.T) {
	secretsDirPath := t.TempDir()
	os.WriteFile(filepath.Join(secretsDirPath, "DB_PASSWORD"), []byte("password"), 0600)
	os.WriteFile(filepath.Join(secretsDirPath, "DB_USER"), []byte("user"), 0600)
	t.Setenv("SECRETS_MOUNT_PATH", secretsDirPath)
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_SCHEMA", "test-schema")
}

func TestSetDefaults(t *testing.T) {
	conf := &Config{}
	conf.SetDefaults()

	require.True(t, *conf.UseServiceSession)
	require.Equal(t, LogLevelInfo, conf.LogLevel)

	require.Equal(t, 5432, conf.Connection.Port)
	require.Equal(t, "backend", conf.Connection.Name)
	require.Equal(t, "2min", conf.Connection.StatementTimeout)
	require.Equal(t, SslModeDisable, conf.Connection.SslMode)

	require.Equal(t, 5432, conf.ReadConnection.Port)
	require.Equal(t, SslModeDisable, conf.ReadConnection.SslMode)

	require.Equal(t, 20, conf.Pool.MaxConnections)
	require.Equal(t, 2, conf.Pool.MaxIdleConnections)
	require.Equal(t, time.Hour, conf.Pool.MaxIdleConnectionTimeDuration)

	require.Equal(t, "file:///migrations", conf.Migration.Files)
	require.Equal(t, MigrationDirectionUp, conf.Migration.Direction)
	require.Equal(t, 30, conf.Migration.LockRetries)
	require.NotNil(t, conf.Migration.Auto)
	require.True(t, *conf.Migration.Auto)
}

func TestInitConfig(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("DB_ENCRYPTION_KEY", "1234567890123456") // 16 bytes
	t.Setenv("LOG_LEVEL", "Debug")
	t.Setenv("DB_USE_SERVICE_SESSION", "false")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.NotNil(t, conf)
	require.Equal(t, "test-service", conf.ServiceName)
	require.Equal(t, "localhost", conf.Connection.Host)
	require.Equal(t, "test-schema", conf.Connection.Schema)
	require.Equal(t, "password", conf.Connection.Password)
	require.Equal(t, "user", conf.Connection.User)
	require.Equal(t, "1234567890123456", conf.EncryptionKey)
	require.Equal(t, LogLevelDebug, conf.LogLevel)
	require.False(t, *conf.UseServiceSession)
}

func TestInitConfig_CustomConnection(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_SCHEMA", "custom-schema")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_NAME", "custom-name")
	t.Setenv("DB_SSL_MODE", "require")
	t.Setenv("DB_STATEMENT_TIMEOUT", "3sec")
	t.Setenv("DB_MAX_CONNECTIONS", "15")
	t.Setenv("DB_MAX_IDLE_CONNECTIONS", "2")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.NotNil(t, conf)
	require.Equal(t, "localhost", conf.Connection.Host)
	require.Equal(t, "custom-schema", conf.Connection.Schema)
	require.Equal(t, "password", conf.Connection.Password)
	require.Equal(t, "user", conf.Connection.User)
	require.Equal(t, 5433, conf.Connection.Port)
	require.Equal(t, "custom-name", conf.Connection.Name)
	require.Equal(t, SslModeRequire, conf.Connection.SslMode)
	require.Equal(t, "3sec", conf.Connection.StatementTimeout)
}

func TestInitConfig_CustomPool(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("DB_MAX_CONNECTIONS", "15")
	t.Setenv("DB_MAX_IDLE_CONNECTIONS", "2")
	t.Setenv("DB_MAX_IDLE_CONNECTIONS_TIME_DURATION", "10m")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.NotNil(t, conf)
	require.Equal(t, 15, conf.Pool.MaxConnections)
	require.Equal(t, 2, conf.Pool.MaxIdleConnections)
	require.Equal(t, 10*time.Minute, conf.Pool.MaxIdleConnectionTimeDuration)
}

func TestInitConfig_CustomMigration(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("DB_MIGRATION_FILES", "file:///custom-migrations")
	t.Setenv("DB_MIGRATION_DIRECTION", "down")
	t.Setenv("DB_MIGRATION_LOCK_RETRIES", "10")
	t.Setenv("DB_MIGRATION_AUTO", "false")
	t.Setenv("DB_MIGRATION_TARGET_VERSION", "1.0.0")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.NotNil(t, conf)
	require.False(t, *conf.Migration.Auto)
	require.Equal(t, "file:///custom-migrations", conf.Migration.Files)
	require.Equal(t, MigrationDirectionDown, conf.Migration.Direction)
	require.Equal(t, 10, conf.Migration.LockRetries)
	require.Equal(t, "1.0.0", *conf.Migration.TargetVersion)
}

func TestInitConfig_Fails_WithoutSecretsFolder(t *testing.T) {
	defer config.FlushAll()

	t.Setenv("SECRETS_MOUNT_PATH", "non-existent-path")
	err := InitConfig()
	require.ErrorContains(t, err, "failed to set secrets folder")
}

func TestInitConfig_Fails_MissingRequiredVars(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("SERVICE_NAME", "")
	err := InitConfig()
	require.ErrorContains(t, err, "failed to register runaidb config: failed to validate config struct Config, error: Key: 'Config.ServiceName'")
}

func TestInitConfig_ReadReplica(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	os.WriteFile(filepath.Join(os.Getenv("SECRETS_MOUNT_PATH"), "DB_READ_USER"), []byte("read-user"), 0600)
	os.WriteFile(filepath.Join(os.Getenv("SECRETS_MOUNT_PATH"), "DB_READ_PASSWORD"), []byte("read-password"), 0600)
	t.Setenv("DB_READ_HOST", "read-host")
	t.Setenv("DB_READ_PORT", "5433")
	t.Setenv("DB_READ_SSL_MODE", "require")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.Equal(t, "read-host", conf.ReadConnection.Host)
	require.Equal(t, 5433, conf.ReadConnection.Port)
	require.Equal(t, "read-user", conf.ReadConnection.User)
	require.Equal(t, "read-password", conf.ReadConnection.Password)
	require.Equal(t, SslModeRequire, conf.ReadConnection.SslMode)
}

func TestInitConfig_ReadReplica_MissingCredentials_ShouldFallbackToPrimary(t *testing.T) {
	defer config.FlushAll()

	setupRequiredConfigVars(t)
	t.Setenv("DB_READ_HOST", "read-host")

	err := InitConfig()
	require.NoError(t, err)

	conf := config.Get[*Config]()
	require.Equal(t, "user", conf.ReadConnection.User)
	require.Equal(t, "password", conf.ReadConnection.Password)
}
