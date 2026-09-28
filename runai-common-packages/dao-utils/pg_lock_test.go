/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateServiceDbUser_PGLock(t *testing.T) {
	os.Setenv(EnvServiceName, "test-service")
	// Init runs migrations and exits the process if they cannot be loaded, so
	// give it a no-op migration; this test only needs the connection.
	migrationsDir := t.TempDir()
	assert.Nil(t, os.Mkdir(filepath.Join(migrationsDir, "up"), 0700))
	err := os.WriteFile(filepath.Join(migrationsDir, "up", "1_noop.up.sql"), []byte("SELECT 1;"), 0600)
	assert.Nil(t, err)
	SetPostgresEnv("public", "file://"+migrationsDir)
	config := ConstructConfig("")
	OpenPostgresForTesting(t, config)
	pgLock := NewPGLock(config.Schema)

	globalWriterDb = Init(config)

	firstTx := globalWriterDb.Begin()
	acquired, err := pgLock.AcquireSchemaLock(firstTx)
	assert.Nil(t, err)
	assert.True(t, acquired, "The lock should be acquired")

	anotherTx := globalWriterDb.Begin()
	acquired, err = pgLock.AcquireSchemaLock(anotherTx)
	assert.Nil(t, err)
	assert.False(t, acquired, "The lock should not be acquired")

	released, err := pgLock.ReleaseSchemaLock(firstTx)
	assert.Nil(t, err)
	assert.True(t, released, "The lock should be released")

	acquired, err = pgLock.AcquireSchemaLock(anotherTx)
	assert.Nil(t, err)
	assert.True(t, acquired, "The lock should be acquired")

	pgLock.ReleaseSchemaLock(anotherTx)

	firstTx.Rollback()
	anotherTx.Rollback()
}
