/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/filter/filters"
)

func NewFilterFactory() Factory {
	return &FilterFactory{}
}

type FilterFactory struct{}

func (ff *FilterFactory) CreateFilter(fc *FilterConfig) (Filter, error) {
	switch fc.Type {
	case filters.PrometheusAlertsFilterName:
		return filters.NewPrometheusAlertsFilter(fc.PrometheusMatchers), nil
	case filters.MatcherFilterName:
		return filters.NewMatcherFilter(fc.Matcher), nil
	}

	return nil, fmt.Errorf("filter %s not found in factory", fc.Type)
}
