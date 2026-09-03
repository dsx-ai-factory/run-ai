/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_test

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricInformationFromApi(t *testing.T) {
	source := &api.MetricInformation{
		Name: "metric",
		Filters: map[string][]string{
			"filter": {"a", "b"},
		},
		ResponseLabels: []string{},
		Range:          nil,
	}
	expected := &services.MetricInformation{
		Name: "metric",
		Filters: map[string][]string{
			"filter": {"a", "b"},
		},
		ResponseLabels: []string{},
		Range:          nil,
	}
	assert.Equal(t, expected, services.NewMetricInformationFromApi(source))
}

func TestNewMetricInformationRangeFromApi(t *testing.T) {
	end := time.Now()
	start := end.Add(time.Second * -60)
	source := &api.MetricInformationRange{
		Start:        start,
		End:          end,
		NumOfSamples: 11,
	}
	expected := &services.MetricInformationRange{
		Start:        start,
		End:          end,
		NumOfSamples: 11,
	}
	assert.Equal(t, expected, services.NewMetricInformationRangeFromApi(source))
}
