package main

import (
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfigProvider_File(t *testing.T) {
	configContent := `
app:
  log_level: debug
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	t.Setenv(envConfPath, file.Name())
	t.Setenv(envConfigProvider, "basic")
	factory := GetConfigFactory()
	assert.Equal(t, factory.CreateConfigurationProvider().GetName(), "Basic")
}

func TestGetConfigProvider_File_EnvNotSet(t *testing.T) {
	configContent := `
app:
  log_level: debug
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	t.Setenv(envConfPath, file.Name())
	t.Setenv(envConfigProvider, "")
	factory := GetConfigFactory()
	assert.Equal(t, factory.CreateConfigurationProvider().GetName(), "Basic")
}

//func TestGetConfigProvider_Postgres(t *testing.T) {
//	configContent :=
//		`
//app:
// log_level: debug
//database:
//  host: "localhost"
//  port: 5432
//  database: "notifications_service"
//  user: "runai"
//  password: ""
//  ssl_mode: "disable"
//  http_events_port: 8085
//`
//	file := config.CreateTempConfigFile(configContent)
//	defer os.Remove(file.Name())
//
//	_ = os.Setenv(envConfPath, file.Name())
//	_ = os.Setenv(envConfigProvider, "postgres")
//	config := GetConfig()
//	assert.Equal(t, config.GetName(), "Postgres")
//}
