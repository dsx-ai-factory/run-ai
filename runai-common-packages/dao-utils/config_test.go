/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"os"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

func TestConstructConfig_SetsDefaultsAsExpected(t *testing.T) {
	defer config.FlushAll()
	conf := ConstructConfig("")
	assert.Equal(t, "file:///migrations", conf.MigrationFiles)
	assert.Equal(t, "5432", conf.Port)
	assert.Equal(t, "backend", conf.Name)
	assert.Equal(t, 20, conf.MaxConnections)
	assert.Equal(t, DefaultLockRetires, conf.MigrationLockRetries)
	assert.Equal(t, SslModeDisable, conf.SslMode)
	assert.Equal(t, "2min", conf.StatementTimeout)
	assert.Equal(t, true, conf.UseServiceSession)
}

func TestConstructConfig_NonSecretValuesShouldBeSet_ThroughEnv(t *testing.T) {
	defer config.FlushAll()

	t.Setenv(EnvServiceName, "test-service")
	t.Setenv(EnvDbName, "myDb")
	t.Setenv(EnvDbHost, "myHost")
	t.Setenv(EnvDbPort, "1234")
	t.Setenv(EnvDbReadHost, "myReadHost")
	t.Setenv(EnvDbReadPort, "2345")
	t.Setenv(EnvDbSchema, "mySchema")
	t.Setenv(EnvDbSslMode, "require")
	t.Setenv(EnvStatementTimeout, "3sec")
	t.Setenv(EnvMaxConnections, "15")
	t.Setenv(EnvSkipDefaultTransaction, "true")
	t.Setenv(EnvFullSaveAssociations, "true")
	t.Setenv(EnvUseServiceSession, "true")
	t.Setenv(EnvMigrationFiles, "./migrations")
	t.Setenv(EnvMigrationCommand, "cmd")
	t.Setenv(EnvMigrationLockRetries, "10")
	t.Setenv(EnvMaxIdleConnectionTimeDuration, "10m")

	conf := ConstructConfig("")

	assert.Equal(t, "test-service", conf.ServiceName)
	assert.Equal(t, "myDb", conf.Name)
	assert.Equal(t, "myHost", conf.Host)
	assert.Equal(t, "1234", conf.Port)
	assert.Equal(t, "myReadHost", conf.ReadHost)
	assert.Equal(t, "2345", conf.ReadPort)
	assert.Equal(t, "mySchema", conf.Schema)
	assert.Equal(t, "require", conf.SslMode)
	assert.Equal(t, "3sec", conf.StatementTimeout)
	assert.Equal(t, 15, conf.MaxConnections)
	assert.Equal(t, true, conf.SkipDefaultTransaction)
	assert.Equal(t, true, conf.FullSaveAssociations)
	assert.Equal(t, true, conf.UseServiceSession)
	assert.Equal(t, "./migrations", conf.MigrationFiles)
	assert.Equal(t, "cmd", conf.MigrationCommand)
	assert.Equal(t, 10, conf.MigrationLockRetries)
	assert.Equal(t, 10*time.Minute, conf.MaxIdleConnectionTimeDuration)
}

func TestConstructConfig_SecretValuesShouldNotBeSet_ThroughEnv(t *testing.T) {
	defer config.FlushAll()

	t.Setenv(EnvDbUser, "myUser")
	t.Setenv(EnvDbPassword, "myPassword")

	conf := ConstructConfig("")

	assert.Equal(t, "", conf.User)
	assert.Equal(t, "", conf.Password)
}

func TestConstructConfig_SecretValuesShouldBeSet_ThroughSecretsFolder(t *testing.T) {
	defer config.FlushAll()

	secretsDir, err := MountSecretsToDir("postgres", "admin")
	assert.Nil(t, err)
	defer os.RemoveAll(secretsDir)

	t.Setenv(EnvSecretsMountPath, secretsDir)

	conf := ConstructConfig("")

	assert.Equal(t, "postgres", conf.User)
	assert.Equal(t, "admin", conf.Password)
}

func TestConstructConfig_WithPassedDbSchemaArgument(t *testing.T) {
	cases := []struct {
		name           string
		schemaArgument string
		expected       string
	}{
		{"Empty arg", "", ""},
		{"Non-empty arg", "test", "test"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer config.FlushAll()
			conf := ConstructConfig(c.schemaArgument)
			assert.Equal(t, c.expected, conf.Schema)
		})

	}
}
