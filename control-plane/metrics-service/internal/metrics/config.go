/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/hashicorp/go-version"
)

type MetricConfiguration struct {
	CrossVersionQuery string            `mapstructure:"cross_version_query" validate:"required"`
	Queries           map[string]string `mapstructure:"queries" validate:"required,dive,keys,semver,endkeys"`
}

type ProcessedMetricConfiguration struct {
	CrossVersionQuery string
	VersionedQueries  []*VersionedQuery
}

type VersionedQuery struct {
	Version *version.Version
	Query   string
}

type MetricsConfiguration struct {
	Metrics map[string]*MetricConfiguration `mapstructure:"metrics" validate:"required,dive"`
}

func (mc *MetricsConfiguration) SetDefaults() {
	mc.Metrics = make(map[string]*MetricConfiguration)
}
