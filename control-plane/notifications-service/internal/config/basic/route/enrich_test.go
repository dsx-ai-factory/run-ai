/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route_test

import (
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
)

func TestEnrichKubeMetadata_Defaults(t *testing.T) {
	configContent := `
routes: 
  route1: 
    listener:
      relevant_namespaces: some_namespace
    enrich: 
      kube_metadata: {}
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 1)
	assert.Contains(t, enrichers, "kube_metadata")
	config := provider.GetGlobalEnrichingConfig().KubeMetadataEnricherConfig
	assert.Len(t, config.Labels, 4)
	assert.Len(t, config.Annotations, 1)
}

func TestKubeMetadata_WithValues(t *testing.T) {
	configContent := `
routes: 
  route1: 
    enrich:
      kube_metadata:
        labels:
          test: test
        annotations:
          test: test
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 1)
	assert.Contains(t, enrichers, "kube_metadata")
	meta := provider.GetGlobalEnrichingConfig().KubeMetadataEnricherConfig
	assert.Len(t, meta.Labels, 1)
	assert.Len(t, meta.Annotations, 1)
	assert.Contains(t, meta.Labels, "test")
	assert.Contains(t, meta.Annotations, "test")
}

func TestEnrichAll_MixDefaults(t *testing.T) {
	configContent := `
routes: 
  route1: 
    enrich:
      kube_metadata: {}
      static: 
        env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 2)
	assert.Contains(t, enrichers, "kube_metadata")
	assert.Contains(t, enrichers, "static")
	meta := provider.GetGlobalEnrichingConfig().KubeMetadataEnricherConfig
	assert.Len(t, meta.Labels, 4)
	assert.Len(t, meta.Annotations, 1)
	static := provider.GetGlobalEnrichingConfig().StaticEnricherConfig
	assert.Len(t, *static, 1)
}

func TestEnrichAll_WithValues(t *testing.T) {
	configContent := `
routes: 
  route1: 
    enrich:
      kube_metadata:
        labels:
          test: test
        annotations:
          test: test
      static: 
        env: bla
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 2)
	assert.Contains(t, enrichers, "kube_metadata")
	meta := provider.GetGlobalEnrichingConfig().KubeMetadataEnricherConfig
	assert.Len(t, meta.Labels, 1)
	assert.Len(t, meta.Annotations, 1)
	assert.Contains(t, meta.Labels, "test")
	assert.Contains(t, meta.Annotations, "test")
	static := provider.GetGlobalEnrichingConfig().StaticEnricherConfig
	assert.Len(t, *static, 1)
}

func TestEnrichStatic_Defaults(t *testing.T) {
	configContent := `
routes: 
  route1: 
    enrich:
      static: {}
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 1)
	assert.Contains(t, enrichers, "static")
	static := provider.GetGlobalEnrichingConfig().StaticEnricherConfig
	assert.Len(t, *static, 0)
}

func TestEnrichStatic_WithValues(t *testing.T) {
	configContent := `
routes: 
  route1: 
    enrich:
      static:
        environment: test
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	enrichers := provider.GetGlobalEnrichers()
	assert.Len(t, enrichers, 1)
	assert.Contains(t, enrichers, "static")
	static := provider.GetGlobalEnrichingConfig().StaticEnricherConfig
	assert.Len(t, *static, 1)
	assert.Contains(t, *static, "environment")
}
