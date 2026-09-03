/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package listeners_test

import (
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"

	"github.com/stretchr/testify/assert"
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
	providers := config.GetRouteProviders("listener1", &events.Event{})
	provider := providers[0]

	assert.Equal(t, "listener1", provider.GetListenerId())
	assert.Equal(t, "kube_events", config.GetListeningProviders()["listener1"].GetListenerType())
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
	assert.Nil(t, keConfig)
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
	assert.Contains(t, keConfig.RelevantNamespaces, "runai-.*")
	assert.Len(t, keConfig.RelevantObjects, 1)
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

	assert.Equal(t, "", provider.GetListenerType())
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

	assert.Equal(t, "prometheus_alerts", provider.GetListenerType())
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

	assert.Equal(t, 9093, provider.GetListeningConfig().PrometheusListener.Port)
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

	assert.Equal(t, 811, provider.GetListeningConfig().PrometheusListener.Port)
}
