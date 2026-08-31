/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sink_test

import (
	"os"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/sink"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
)

func TestGetListenerName_Returns(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: kube_events
    relevant_namespaces: 
    - runai-.*
routes: 
  route1: 
    listener_id: listener1
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	config := basic.NewConfigurationProvider(file.Name())
	provider := config.GetRouteProviders("listener1", &events.Event{})[0]

	require.Equal(t, "listener1", provider.GetListenerId())
	require.Equal(t, "kube_events", config.GetListeningProviders()["listener1"].GetListenerType())
}

func TestKubeEventsMissing_ReturnsNil(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: http_events
    relevant_namespaces: some_namespace
routes: 
  route1: 
    listener_id: listener1
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	keConfig := provider.GetListeningConfig().KubeListener
	require.Nil(t, keConfig)
}

func TestKubeEvents_WithValuesStillGetDefaults(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: kube_events
    relevant_namespaces: 
      - runai-.*
routes: 
  route1: 
    listener_id: listener1
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	keConfig := provider.GetListeningConfig().KubeListener
	require.Contains(t, keConfig.RelevantNamespaces, "runai-.*")
	require.Len(t, keConfig.RelevantObjects, 1)
}

func TestGetListenerName_ReturnsEmpty(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: kube_event
    relevant_namespaces: some_namespace
routes: 
  route1: 
    enrich:
        kube_metadata: {}
        static: 
          env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	require.Equal(t, "", provider.GetListenerType())
}

func TestGetListenerName_ReturnsPrometheus(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: prometheus_alerts
    relevant_namespaces: some_namespace
routes: 
  route1: 
    listener_id: listener1
    enrich:
        kube_metadata: {}
        static: 
          env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	require.Equal(t, "prometheus_alerts", provider.GetListenerType())
}

func TestGetListenerName_ReturnsPrometheusDefaults(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: prometheus_alerts
    relevant_namespaces: some_namespace
routes: 
  route1: 
    listener_id: listener1
    enrich:
        kube_metadata: {}
        static: 
          env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	require.Equal(t, 9093, provider.GetListeningConfig().PrometheusListener.Port)
}

func TestGetListenerName_ReturnsPrometheusValues(t *testing.T) {
	configContent := `
listeners:
  listener1:
    type: prometheus_alerts
    port: 811
    relevant_namespaces: some_namespace
routes: 
  route1: 
    listener_id: listener1
    enrich:
        kube_metadata: {}
        static: 
          env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetListeningProviders()["listener1"]

	require.Equal(t, 811, provider.GetListeningConfig().PrometheusListener.Port)
}

func TestNewSinkConfigProvider(t *testing.T) {
	configContent := `
sinks:
  email1:
    type: email
    user: smtp_user
    password: password
    smtp_host: smtp.mail.com
    smtp_port: 587
  email2:
    type: email
    user: smtp_user
    password: password
    smtp_host: smtp.mail.com
    smtp_port: 587
`

	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(strings.NewReader(configContent))
	provider := sink.NewSinkConfigProvider(v.Sub("sinks"), "email1", "email2")
	sinksMap, err := provider.GetSinks()
	require.NoError(t, err)
	require.Len(t, sinksMap, 2)
}

func TestNewSinkConfig_Postgres(t *testing.T) {
	configContent := `
type: postgres
host: db.example.com
port: "5432"
database: notifications
schema: audit
table: notification_events
user: app_user
password: s3cret
ssl_mode: disable
`
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(configContent)))

	cfg, err := sink.NewSinkConfig(v, "pg-sink", sinks.PostgresSinkType)
	require.NoError(t, err)
	require.NotNil(t, cfg.Postgres)
	require.Nil(t, cfg.Email)
	require.Nil(t, cfg.Slack)

	require.Equal(t, "db.example.com", cfg.Postgres.Host)
	require.Equal(t, "5432", cfg.Postgres.Port)
	require.Equal(t, "notifications", cfg.Postgres.Database)
	require.Equal(t, "audit", cfg.Postgres.Schema)
	require.Equal(t, "notification_events", cfg.Postgres.Table)
	require.Equal(t, "app_user", cfg.Postgres.User)
	require.Equal(t, "s3cret", cfg.Postgres.Password)
	require.Equal(t, "disable", cfg.Postgres.SslMode)
}

func TestNewSinkConfigProvider_PostgresSink(t *testing.T) {
	configContent := `
sinks:
  postgres:
    type: postgres
    host: localhost
    port: "5432"
    database: db
    schema: public
    table: events
    user: u
    password: p
    ssl_mode: require
`
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(configContent)))

	provider := sink.NewSinkConfigProvider(v.Sub("sinks"), "postgres")
	sinksMap, err := provider.GetSinks()
	require.NoError(t, err)
	require.Len(t, sinksMap, 1)

	got := sinksMap["postgres"]
	require.NotNil(t, got)
	require.NotNil(t, got.Postgres)
	require.Equal(t, "localhost", got.Postgres.Host)
	require.Equal(t, "require", got.Postgres.SslMode)
}

func TestNewSinkConfigProvider_HistoryDisabledCarriesDisabledFlag(t *testing.T) {
	configContent := `
sinks:
  my_history:
    type: history
    history_enabled: false
`
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(configContent)))

	provider := sink.NewSinkConfigProvider(v.Sub("sinks"), "my_history")
	got, err := provider.GetSinks()
	require.NoError(t, err)

	cfg := got["my_history"]
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.History)
	require.False(t, cfg.History.HistoryEnabled)
}

func TestNewSinkConfigProvider_HistoryEnabledCarriesEnabledFlag(t *testing.T) {
	configContent := `
sinks:
  my_history:
    type: history
    history_enabled: true
`
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(configContent)))

	provider := sink.NewSinkConfigProvider(v.Sub("sinks"), "my_history")
	got, err := provider.GetSinks()
	require.NoError(t, err)

	cfg := got["my_history"]
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.History)
	require.True(t, cfg.History.HistoryEnabled)
}

func TestExtractSinkConfigProvider_WithHistorySink(t *testing.T) {
	configContent := `
sinks:
  history:
    type: history
  email1:
    type: email
    user: smtp_user
    password: password
    smtp_host: smtp.mail.com
    smtp_port: 587
`
	v := viper.New()
	v.SetConfigType("yaml")
	require.NoError(t, v.ReadConfig(strings.NewReader(configContent)))

	configProvider := sink.NewSinkConfigProvider(v.Sub("sinks"), "email1")

	extracted := sink.ExtractSinkConfigProvider(configProvider, "email1", "history")
	require.NotNil(t, extracted)

	sinks, err := extracted.GetSinks()
	require.NoError(t, err)
	require.Len(t, sinks, 1)
	require.Contains(t, sinks, "email1")
	require.NotContains(t, sinks, "history")
}
