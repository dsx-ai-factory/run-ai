/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
)

func TestFilterPrometheusMatcher_Defaults(t *testing.T) {
	configContent := `
routes: 
  route1: 
    filtering:
      filters:
      - type: prometheus_matcher
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	p := provider.GetGlobalFilteringConfig()
	require.Equal(t, "all", p.Policy)

	config := *p.Filters[0].PrometheusMatchers
	require.Len(t, config, 1)
	require.Equal(t, config[0].Key, "alertname")
	require.Equal(t, config[0].Value, "^Runai.*$")
	require.Equal(t, string(config[0].MatchType), "=~")
}

func TestFilterPrometheusMatcher_OneMatcher(t *testing.T) {
	configContent := `
routes:
  route1:
   filtering:
      policy: all
      filters:
      - type: prometheus_matcher
        prometheus_matchers:
        - match_type: ==
          key: test
          value: test
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	config := *provider.GetGlobalFilteringConfig().Filters[0].PrometheusMatchers
	require.Len(t, config, 1)
	require.Equal(t, "test", config[0].Key)
	require.Equal(t, "test", config[0].Value)
	require.Equal(t, "==", string(config[0].MatchType))
}

func TestFilterPrometheusMatcher_MultipleMatcher(t *testing.T) {
	configContent := `
routes:
  route1:
    filtering:
      policy: all
      filters:
      - type: prometheus_matcher
        prometheus_matchers:
        - match_type: ==
          key: test
          value: test
        - match_type: ==
          key: test2
          value: test2
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	config := *provider.GetGlobalFilteringConfig().Filters[0].PrometheusMatchers
	require.Len(t, config, 2)
	require.Equal(t, "test", config[0].Key)
	require.Equal(t, "test", config[0].Value)
	require.Equal(t, "==", string(config[0].MatchType))
	require.Equal(t, "test2", config[1].Key)
	require.Equal(t, "test2", config[1].Value)
	require.Equal(t, "==", string(config[1].MatchType))
}

func TestFilterPrometheusMatcher_MultipleFilters(t *testing.T) {
	configContent := `
routes:
  route1:
   filtering:
      policy: all
      filters:
      - type: prometheus_matcher
        prometheus_matchers:
        - match_type: ==
          key: test
          value: test
      - type: prometheus_matcher
        prometheus_matchers:
        - match_type: ==
          key: test
          value: test
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	filtersConfig := provider.GetGlobalFilteringConfig().Filters
	require.Len(t, filtersConfig, 2)
	config1 := *filtersConfig[0].PrometheusMatchers
	config2 := *filtersConfig[1].PrometheusMatchers
	require.Len(t, config1, 1)
	require.Len(t, config2, 1)
	require.Equal(t, "test", config1[0].Key)
	require.Equal(t, "test", config1[0].Value)
	require.Equal(t, "==", string(config1[0].MatchType))
	require.Equal(t, "test", config2[0].Key)
	require.Equal(t, "test", config2[0].Value)
	require.Equal(t, "==", string(config2[0].MatchType))
}

func TestFilterMatcher_MultipleFilters(t *testing.T) {
	configContent := `
routes:
  route1:
   filtering:
      policy: all
      filters:
      - name: matcher
        matchers:
        - match_type: ==
          key: test
          value: test
      - name: matcher
        matchers:
        - match_type: ==
          key: test
          value: test
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	filtersConfig := provider.GetGlobalFilteringConfig().Filters
	require.Len(t, filtersConfig, 2)
	config1 := *filtersConfig[0].Matcher
	config2 := *filtersConfig[1].Matcher
	require.Len(t, config1, 1)
	require.Len(t, config2, 1)
	require.Equal(t, "test", config1[0].Key)
	require.Equal(t, "test", config1[0].Value)
	require.Equal(t, "==", string(config1[0].MatchType))
	require.Equal(t, "test", config2[0].Key)
	require.Equal(t, "test", config2[0].Value)
	require.Equal(t, "==", string(config2[0].MatchType))
}

func TestFilterMatcher_Defaults(t *testing.T) {
	configContent := `
routes: 
  route1: 
    filtering:
      filters:
      - type: matcher
`
	file := config.CreateTempConfigFile(configContent)
	defer os.Remove(file.Name())
	provider := basic.NewConfigurationProvider(file.Name()).GetRouteProviders("", &events.Event{})[0]

	config := *provider.GetGlobalFilteringConfig().Filters[0].Matcher
	require.Len(t, config, 0)
}
