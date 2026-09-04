/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

type ExampleMetricFilters struct {
	ClusterId []string `mapstructure:"clusterId,omitempty"`
}

type ExampleMetricResponseLabels struct {
	ClusterId bool `mapstructure:"clusterId,omitempty"`
}

type ExampleMetric struct {
	Filters        ExampleMetricFilters
	ResponseLabels ExampleMetricResponseLabels
	Quantile       *float32
}

func (m ExampleMetric) GetName() string {
	return "example_metric"
}

func (m ExampleMetric) GetFilters() interface{} {
	return m.Filters
}

func (m ExampleMetric) GetResponseLabels() interface{} {
	return m.ResponseLabels
}

func (m ExampleMetric) GetQuantile() *float32 {
	return m.Quantile
}

func (m ExampleMetric) IsAggregation() bool {
	return false
}

func (m ExampleMetric) IsQuantile() bool {
	return false
}
