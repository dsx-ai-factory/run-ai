/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"time"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
)

type MetricInformation struct {
	Name           string
	Filters        map[string][]string
	ResponseLabels []string
	Range          *MetricInformationRange
	Quantile       *float32
}

func NewMetricInformationFromApi(source *api.MetricInformation) *MetricInformation {
	if source == nil {
		return nil
	}
	return &MetricInformation{
		Name:           source.Name,
		Filters:        source.Filters,
		ResponseLabels: source.ResponseLabels,
		Quantile:       source.Quantile,
		Range:          NewMetricInformationRangeFromApi(source.Range),
	}
}

type MetricInformationRange struct {
	Start        time.Time
	End          time.Time
	NumOfSamples int32
	Aggregation  bool
}

func NewMetricInformationRangeFromApi(source *api.MetricInformationRange) *MetricInformationRange {
	if source == nil {
		return nil
	}
	return &MetricInformationRange{
		Start:        source.Start,
		End:          source.End,
		NumOfSamples: source.NumOfSamples,
		Aggregation:  source.Aggregation,
	}
}
