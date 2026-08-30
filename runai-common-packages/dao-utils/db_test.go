package dao_utils

// import (
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestCreateServiceDbUser_CreatesNewUser(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)
// 	config := ConstructConfig("")
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	formattedServiceName := strings.Replace(config.ServiceName, "-", "_", -1)
// 	var userExists bool
// 	err = globalWriterDb.Raw("SELECT 1 FROM pg_roles WHERE rolname = ?", formattedServiceName).Scan(&userExists).Error
// 	assert.Nil(t, err)
// 	assert.True(t, userExists, "The user should exist")
// }

// func TestCreateServiceDbUser_GetsAdminSessionWhenNoServiceName(t *testing.T) {
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)
// 	config := ConstructConfig("")
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	var currentUser string
// 	err = globalWriterDb.Raw("SELECT CURRENT_USER").Scan(&currentUser).Error
// 	assert.Nil(t, err)
// 	assert.Equal(t, "postgres", currentUser, "The current user should be 'postgres'")
// }

// func TestCreateServiceDbUser_GetsAdminSessionWhenUseServiceSessionFalse(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)
// 	config := ConstructConfig("")
// 	config.UseServiceSession = false
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	var currentUser string
// 	err = globalWriterDb.Raw("SELECT CURRENT_USER").Scan(&currentUser).Error
// 	assert.Nil(t, err)
// 	assert.Equal(t, "postgres", currentUser, "The current user should be 'postgres'")
// }

// func TestCreateServiceDbUser_GetsAdminSessionReadWhenUseServiceSessionFalse(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)
// 	config := ConstructConfig("")
// 	config.UseServiceSession = false
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	var currentUser string
// 	err = DBRead().Raw("SELECT CURRENT_USER").Scan(&currentUser).Error
// 	assert.Nil(t, err)
// 	assert.Equal(t, "postgres", currentUser, "The current user should be 'postgres'")
// }

// func TestCreateServiceDbUser_GetsAdminSessionReadWhenUseServiceSessionFalse_AndReadConfigured(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)

// 	config := ConstructConfig("")
// 	config.UseServiceSession = false
// 	config.ReadHost = "sdfasdf"
// 	config.ReadPort = "123"
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	var currentUser string
// 	err = DBRead().Raw("SELECT CURRENT_USER").Scan(&currentUser).Error
// 	assert.Nil(t, err)
// 	assert.Equal(t, "postgres", currentUser, "The current user should be 'postgres'")
// }

// func TestCreateServiceDbUser_GetsServiceSessionWhenServiceName(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)

// 	config := ConstructConfig("")
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)

// 	formattedServiceName := strings.Replace(config.ServiceName, "-", "_", -1)
// 	var currentUser string
// 	err = globalWriterDb.Raw("SELECT CURRENT_USER").Scan(&currentUser).Error
// 	assert.Nil(t, err)
// 	assert.Equal(t, formattedServiceName, currentUser, "The current user should be test_service")
// }

// func TestCreateServiceDbUser_UserAlreadyExists(t *testing.T) {
// 	os.Setenv(EnvServiceName, "test-service")
// 	err := SetPostgresEnv("public", "")
// 	assert.Nil(t, err)

// 	config := ConstructConfig("")
// 	OpenPostgresForTesting(t, config)

// 	globalWriterDb = Init(config)
// 	err = createServiceDbUser(config)
// 	assert.Nil(t, err)

// 	formattedServiceName := strings.Replace(config.ServiceName, "-", "_", -1)
// 	var userExists bool
// 	err = globalWriterDb.Raw("SELECT 1 FROM pg_roles WHERE rolname = ?", formattedServiceName).Scan(&userExists).Error
// 	assert.Nil(t, err)
// 	assert.True(t, userExists, "The user should exist")
// }
