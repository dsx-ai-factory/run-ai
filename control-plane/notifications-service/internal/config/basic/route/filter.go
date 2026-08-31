/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route

import (
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter/filters"
	"github.com/spf13/viper"
)

func (brcp *BasicRouteConfigurationProvider) GetGlobalFilteringConfig() *filter.FilteringConfig {
	return &brcp.config.Filtering
}

func setFilteringDefaults(fc *filter.FilteringConfig, v *viper.Viper) error {
	v = v.Sub("filtering")
	if v == nil {
		return nil
	}

	v.SetDefault("policy", filter.PolicyMatchAll)
	err := v.Unmarshal(&fc)
	if err != nil {
		return err
	}

	for _, f := range fc.Filters {
		setFilterDefaults(f)
	}

	return nil
}

func setFilterDefaults(f *filter.FilterConfig) {
	switch f.Type {
	case "prometheus_matcher":
		if f.PrometheusMatchers == nil {
			prometheusMatcherDefaults(f)
		}
	case "matcher":
		if f.Matcher == nil {
			f.Matcher = &filters.MatcherFilterConfig{}
		}

	}
}

func prometheusMatcherDefaults(fc *filter.FilterConfig) {
	fc.Type = "prometheus_matcher"
	fc.PrometheusMatchers = &filters.PrometheusAlertsFilterConfig{
		{
			MatchType: filters.MtRegex,
			Key:       "alertname",
			Value:     "^Runai.*$",
		},
	}
}
