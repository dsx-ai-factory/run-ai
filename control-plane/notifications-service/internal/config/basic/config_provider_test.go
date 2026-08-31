/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basic

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
)

func TestAppConfig_WithLogLevel(t *testing.T) {
	configContent := `
app:
  log_level: debug
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	appConfig := NewConfigurationProvider(file.Name()).GetAppConfig()

	require.Equal(t, "debug", appConfig.LogLevel)
}

func TestAppConfig_GetDefaultLogLevel(t *testing.T) {
	configContent := `
routes: 
  route1: 
    listener:
`

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	appConfig := NewConfigurationProvider(file.Name()).GetAppConfig()

	require.Equal(t, "info", appConfig.LogLevel)
}

func TestGetRouterConfig_GetDefaults(t *testing.T) {
	configContent := `
router:
`

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	routerConfig := NewConfigurationProvider(file.Name()).GetRouterConfig()

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
	routerConfig := NewConfigurationProvider(file.Name()).GetRouterConfig()

	require.Equal(t, 120, routerConfig.ConcurrentLimit)
	require.Equal(t, 1000, routerConfig.CachedEvents)
}

func TestGetRouteProviders_OfTheRightListener(t *testing.T) {
	configContent := `
listeners:
  listener1:
    kube_events: {}
  listener2:
    kube_events: {}

routes:
  route_a_1:
    listener_id: listener1
  route_b_1:
    listener_id: listener1
  route_c_2:
    listener_id: listener2
`

	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	routeProviders1 := NewConfigurationProvider(file.Name()).GetRouteProviders("listener1", &events.Event{})
	routeProviders2 := NewConfigurationProvider(file.Name()).GetRouteProviders("listener2", &events.Event{})

	require.Len(t, routeProviders1, 2)
	require.Equal(t, "listener1", routeProviders1[0].GetListenerId())
	require.Equal(t, "listener1", routeProviders1[1].GetListenerId())
	require.Len(t, routeProviders2, 1)
	require.Equal(t, "listener2", routeProviders2[0].GetListenerId())
}

func TestAppConfigWithSecrets(t *testing.T) {
	configContent := `
sinks:
  email1:
    type: email
    password: ${SMTP_PASSWORD}

routes:
  email_k8s_notifications:
    listener_id: listener1
    notify:
      sink_id: email1
`
	file := createTempConfigFile(configContent)
	defer os.Remove(file.Name())

	os.Setenv("SMTP_PASSWORD", "my_resolved_secret")

	provider, err := NewConfigurationProvider(file.Name()).GetSinkProvider(nil, "email1")
	require.NoError(t, err)

	sinkConfigs, err := provider.GetSinks()
	require.NoError(t, err)

	sinkConfig := sinkConfigs["email1"]
	require.NotNil(t, sinkConfig)
	require.Equal(t, "my_resolved_secret", sinkConfig.Email.Smtp.Password)
}

func TestAppConfigWithSinkHistory_HonorsEnableHistoryFlag(t *testing.T) {
	configContent := `
history:
  enabled: true
listeners:
  listener1:
    kube_events: {}
sinks:
  my_history_sink:
    type: history

routes:
  email_k8s_notifications:
    listener_id: listener1
    notify:
      sink_id: my_history_sink
`
	file := createTempConfigFile(configContent)
	defer os.Remove(file.Name())

	provider, err := NewConfigurationProvider(file.Name()).GetSinkProvider(nil, "my_history_sink")
	require.NoError(t, err)

	sinkConfigs, err := provider.GetSinks()
	require.NoError(t, err)

	sinkConfig := sinkConfigs["my_history_sink"]
	require.NotNil(t, sinkConfig)
	require.NotNil(t, sinkConfig.History)
	require.True(t, sinkConfig.History.HistoryEnabled)
}

func createTempConfigFile(content string) *os.File {
	file, err := os.CreateTemp("", "service-config*.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create temp file for configuration")
	}
	file.WriteString(content)
	file.Close()
	return file
}
