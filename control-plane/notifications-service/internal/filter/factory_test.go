/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter/filters"
	"github.com/stretchr/testify/assert"
)

func TestCreateFilter_CreatesPromAlerts(t *testing.T) {
	factory := NewFilterFactory()
	config := FilterConfig{
		Type:               filters.PrometheusAlertsFilterName,
		PrometheusMatchers: &filters.PrometheusAlertsFilterConfig{},
	}
	f, err := factory.CreateFilter(&config)
	assert.Nil(t, err)
	mf := &filters.PrometheusAlertsFilter{}
	assert.IsType(t, f, mf)
}

func TestCreateFilter_CreatesMatcher(t *testing.T) {
	factory := NewFilterFactory()
	config := FilterConfig{
		Type:    filters.MatcherFilterName,
		Matcher: &filters.MatcherFilterConfig{},
	}
	f, err := factory.CreateFilter(&config)
	assert.Nil(t, err)
	mf := &filters.MatcherFilter{}
	assert.IsType(t, f, mf)
}
