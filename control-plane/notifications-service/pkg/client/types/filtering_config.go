/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

const (
	PolicyMatchAll              = "all"
	PolicyMatchAny              = "any"
	PolicyDrop                  = "drop"
	FilterTypeMatcher           = "matcher"
	FilterTypePrometheusMatcher = "prometheus_matcher"
)

type FilteringConfig struct {
	Policy  string          `mapstructure:"policy,omitempty" json:"policy,omitempty"`
	Filters []*FilterConfig `mapstructure:"filters" json:"filters"`
}

type FilterConfig struct {
	Type               string                        `mapstructure:"type" json:"type"`
	PrometheusMatchers *PrometheusAlertsFilterConfig `mapstructure:"prometheus_matchers" json:"prometheus_matchers"`
	Matcher            *MatcherFilterConfig          `mapstructure:"matchers" json:"matchers"`
}

type (
	PrometheusAlertsFilterConfig MatcherFilterConfig
	MatcherFilterConfig          []*PropertyMatcher
)

type PropertyMatcher struct {
	MatchType string `mapstructure:"match_type" json:"match_type"`
	Key       string `mapstructure:"key" json:"key"`
	Value     string `mapstructure:"value" json:"value"`
}

type PrometheusAlertsFilter struct {
	*MatcherFilter
}

type MatcherFilter struct {
	MatcherFilterConfig
}
