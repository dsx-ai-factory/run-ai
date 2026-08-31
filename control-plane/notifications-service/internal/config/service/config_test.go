/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAppConfig_WithLogLevel(t *testing.T) {
	configContent := `
app:
  log_level: debug
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	config := NewConfig(file.Name())

	require.Equal(t, "debug", config.GetAppConfig().LogLevel)
	require.False(t, config.historyConfig.Enabled)
	require.Equal(t, time.Hour, config.historyConfig.Interval)
	require.Equal(t, 10_000, config.historyConfig.MaxEventsPerUser)
	require.Equal(t, 30*24*time.Hour, config.historyConfig.MaxAge)
}

func TestAppConfig_HistoryEnabled(t *testing.T) {
	configContent := `
app:
  log_level: debug
history:
  enabled: true
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	config := NewConfig(file.Name())

	require.Equal(t, "debug", config.GetAppConfig().LogLevel)
	require.True(t, config.historyConfig.Enabled)
}

func TestAppConfig_HistoryDisabled(t *testing.T) {
	configContent := `
history:
  enabled: false
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	config := NewConfig(file.Name())

	require.Equal(t, "info", config.GetAppConfig().LogLevel)
	require.False(t, config.historyConfig.Enabled)
}

func TestAppConfig_HistoryDefaults_WhenAppSectionMissing(t *testing.T) {
	file := config.CreateTempConfigFile(``)
	defer os.Remove(file.Name())

	config := NewConfig(file.Name())

	require.False(t, config.historyConfig.Enabled)
	require.Equal(t, time.Hour, config.historyConfig.Interval)
	require.Equal(t, 10_000, config.historyConfig.MaxEventsPerUser)
	require.Equal(t, 30*24*time.Hour, config.historyConfig.MaxAge)
}

func TestAppConfig_HistoryDefaults_WhenHistorySectionMissing(t *testing.T) {
	configContent := `
app:
  log_level: debug
  api_port: 12345
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	config := NewConfig(file.Name())

	require.False(t, config.historyConfig.Enabled)
	require.Equal(t, time.Hour, config.historyConfig.Interval)
	require.Equal(t, 10_000, config.historyConfig.MaxEventsPerUser)
	require.Equal(t, 30*24*time.Hour, config.historyConfig.MaxAge)
	require.Equal(t, "12345", config.appConfig.APIPort)
	require.Equal(t, "0.0.0.0", config.appConfig.APIHost)
}

func TestAppConfig_HistoryCustomValues_OverrideDefaults(t *testing.T) {
	configContent := `
history:
  enabled: true
  cleanup:
    interval_duration: 15m
    max_events_per_user: 500
    max_age_duration: 168h
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	appConfig := NewConfig(file.Name())

	require.True(t, appConfig.historyConfig.Enabled)
	require.Equal(t, 15*time.Minute, appConfig.historyConfig.Interval)
	require.Equal(t, 500, appConfig.historyConfig.MaxEventsPerUser)
	require.Equal(t, 168*time.Hour, appConfig.historyConfig.MaxAge)
}

func TestInitAppConfig_ValidDefaults_ReturnsNil(t *testing.T) {
	cfg := newConfigForInitTest(t, ``)

	cfg.initHistoryConfig()
	require.Equal(t, cfg.historyConfig.Interval, time.Hour)
	require.Equal(t, cfg.historyConfig.MaxEventsPerUser, 10_000)
	require.Equal(t, cfg.historyConfig.MaxAge, 30*24*time.Hour)
}

func TestInitAppConfig_InvalidHistoryInterval_ReturnsValidationError(t *testing.T) {
	cfg := newConfigForInitTest(t, `
history:
  enabled: true
  cleanup:
    interval_duration: 0s
`)

	err := cfg.initHistoryConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Field validation for 'Interval' failed on the 'gt' tag")
}

func TestInitAppConfig_InvalidMaxEventsPerUser_ReturnsValidationError(t *testing.T) {
	cfg := newConfigForInitTest(t, `
history:
  enabled: true
  cleanup:
    max_events_per_user: 0
`)

	err := cfg.initHistoryConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Field validation for 'MaxEventsPerUser' failed on the 'gt' tag")
}

func TestInitAppConfig_InvalidMaxAge_ReturnsValidationError(t *testing.T) {
	cfg := newConfigForInitTest(t, `
history:
  enabled: true
  cleanup:
    max_age_duration: 0s
`)

	err := cfg.initHistoryConfig()
	require.Error(t, err)
	require.Contains(t, err.Error(), "Field validation for 'MaxAge' failed on the 'gt' tag")
}

func TestInitAppConfig_AllHistoryFieldsSet_PopulatesValues(t *testing.T) {
	cfg := newConfigForInitTest(t, `
history:
  enabled: true
  cleanup:
    max_age_duration: 720h
    interval_duration: 30m
    max_events_per_user: 10000
`)

	err := cfg.initHistoryConfig()
	require.NoError(t, err)
	require.Equal(t, 30*time.Minute, cfg.historyConfig.Interval)
	require.Equal(t, 10_000, cfg.historyConfig.MaxEventsPerUser)
	require.Equal(t, 720*time.Hour, cfg.historyConfig.MaxAge)
}

func TestAppConfig_HistoryPartialOverride_FillsMissingFromDefaults(t *testing.T) {
	configContent := `
history:
  enabled: true
  cleanup:
    max_age_duration: 2160h
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	config := NewConfig(file.Name())

	require.True(t, config.historyConfig.Enabled)
	require.Equal(t, time.Hour, config.historyConfig.Interval, "interval should fall back to the 1h default")
	require.Equal(t, 10_000, config.historyConfig.MaxEventsPerUser, "max_events_per_user should fall back to the default")
	require.Equal(t, 2160*time.Hour, config.historyConfig.MaxAge)
}

func TestAppConfig_GetDefaultLogLevel(t *testing.T) {
	configContent := ``

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	appConfig := NewConfig(file.Name()).GetAppConfig()

	require.Equal(t, "info", appConfig.LogLevel)
}

func TestGetRouterConfig_GetDefaults(t *testing.T) {
	configContent := ``

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	routerConfig := NewConfig(file.Name()).GetRouterConfig()

	require.Equal(t, 100, routerConfig.ConcurrentLimit)
	require.Equal(t, 1000, routerConfig.CachedEvents)
}

func TestGetRouterConfig_GetValues(t *testing.T) {
	configContent := `
router:
  concurrent_limit: 120
`

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	routerConfig := NewConfig(file.Name()).GetRouterConfig()

	require.Equal(t, 120, routerConfig.ConcurrentLimit)
	require.Equal(t, 1000, routerConfig.CachedEvents)
}

func TestServiceLevelConfig_WithListeners(t *testing.T) {
	configContent := `
listeners:
  http:
    type: http_listener
    port: 5000
`

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	listeners := NewConfig(file.Name()).GetListeningProviders()

	require.Len(t, listeners, 1)
	require.Equal(t, "http_listener", listeners["http"].GetListenerType())
	require.Equal(t, 5000, listeners["http"].GetListeningConfig().HttpListenerConfig.Port)
}

func TestServiceLevelConfig_WithSink(t *testing.T) {
	configContent := `
sinks:
  email1:
    type: email
    password: ${SMTP_PASSWORD}
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())

	os.Setenv("SMTP_PASSWORD", "my_resolved_secret")

	provider, err := NewConfig(file.Name()).GetSinkProvider(nil, "email1")
	require.NoError(t, err)

	sinkConfigs, err := provider.GetSinks()
	require.NoError(t, err)

	sinkConfig := sinkConfigs["email1"]
	require.NotNil(t, sinkConfig)
	require.Equal(t, "my_resolved_secret", sinkConfig.Email.Smtp.Password)
}

func newConfigForInitTest(t *testing.T, yaml string) *Config {
	t.Helper()
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(yaml)))
	return &Config{
		Viper:     v,
		listeners: map[string]listen.ListeningConfigProvider{},
	}
}
