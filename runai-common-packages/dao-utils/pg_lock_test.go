package dao_utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateServiceDbUser_PGLock(t *testing.T) {
	os.Setenv(EnvServiceName, "test-service")
	SetPostgresEnv("public", "")
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
