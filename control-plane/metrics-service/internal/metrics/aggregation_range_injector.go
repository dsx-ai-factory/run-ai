/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/services"
)

const aggregationRangePlaceholder = "RANGE"

type aggregationRangeInjector struct {
	next QueryInjector
}

func NewAggregationRangeInjector(next QueryInjector) *aggregationRangeInjector {
	return &aggregationRangeInjector{next: next}
}

func (a *aggregationRangeInjector) Inject(query string, r *services.MetricInformation) (string, error) {
	if r == nil || r.Range == nil {
		return query, nil
	}
	if !r.Range.Aggregation {
		return query, nil
	}
	timeDiff, err := getTimeDiff(r.Range)
	if err != nil {
		return "", err
	}

	if strings.Count(query, aggregationRangePlaceholder) == 0 {
		return "", fmt.Errorf("aggregation metric query must contain %s placeholder", aggregationRangePlaceholder)
	}
	injectedQuery := strings.ReplaceAll(query, aggregationRangePlaceholder, fmt.Sprintf("%vs", timeDiff))
	if a.next == nil {
		return injectedQuery, nil
	}
	return a.next.Inject(injectedQuery, r)
}

func getTimeDiff(r *services.MetricInformationRange) (int, error) {
	if r.Start.IsZero() {
		return 0, errors.New("aggregation metric must have start time")
	}
	if r.End.IsZero() {
		return 0, errors.New("aggregation metrics must have end time")
	}
	timeDiff := r.End.Sub(r.Start).Seconds()
	if timeDiff <= 0 {
		return 0, errors.New("end time must be greater than start time")
	}
	return int(timeDiff), nil
}
