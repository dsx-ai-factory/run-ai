/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/metrics/mocks"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"
	"github.com/stretchr/testify/assert"
)

func initSimpleFilterInjectorForTesting() *simpleFilterInjector {
	return NewSimpleFilterInjector(nil)
}

func TestInjectSpecificFilters_EmptyQuery(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"value3"},
	}
	query, err := filterInjector.injectSpecificFilters("", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "", query)
}

func TestInjectSpecificFilters_FilterNotSatisfied(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
	}
	query, err := filterInjector.injectSpecificFilters("abc{FILTER::key1::FILTER} or on() def{FILTER::key2::FILTER}", filtersMap)
	assert.Error(t, err)
	assert.Equal(t, "", query)
}

func TestInjectSpecificFilters_Success(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"value3"},
	}
	query, err := filterInjector.injectSpecificFilters("abc{FILTER::key1::FILTER} or on() def{FILTER::key2::FILTER}", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "abc{key1=~'value1|value2'} or on() def{key2=~'value3'}", query)
}

func TestInjectGeneralFilters_EmptyQuery(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"value3"},
	}
	query, err := filterInjector.injectGeneralFilters("", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "", query)
}

func TestInjectGeneralFilters_EmptyFilters(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{}
	query, err := filterInjector.injectGeneralFilters("abc{FILTERS} or on() def{FILTERS}", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "abc{} or on() def{}", query)
}

func TestInjectGeneralFilters_EmptyFiltersWithKeys(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {},
	}
	query, err := filterInjector.injectGeneralFilters("abc{FILTERS} or on() def{FILTERS}", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "abc{key1=~'value1|value2'} or on() def{key1=~'value1|value2'}", query)
}

func TestInjectGeneralFilters_NoFiltersPlaceholder(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key2": {"value3"},
	}
	query, err := filterInjector.injectGeneralFilters("abc{} or on() def{FILTERS}", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, "abc{} or on() def{key2=~'value3'}", query)
}

func TestInjectGeneralFilters_Success(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1", "value2"},
		"key2": {"value3"},
	}
	query, err := filterInjector.injectGeneralFilters("abc{FILTERS} or on() def{FILTERS}", filtersMap)
	assert.NoError(t, err)
	variation1 := "abc{key1=~'value1|value2',key2=~'value3'} or on() def{key1=~'value1|value2',key2=~'value3'}"
	variation2 := "abc{key2=~'value3',key1=~'value1|value2'} or on() def{key2=~'value3',key1=~'value1|value2'}"
	assert.True(t, query == variation1 || query == variation2)
}

func TestInjectGeneralFilters_InvalidKey(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"bad}key": {"value1"},
	}
	_, err := filterInjector.injectGeneralFilters("abc{FILTERS}", filtersMap)
	assert.Error(t, err)
}

func TestBuildFiltersStr_InvalidKey(t *testing.T) {
	filtersMap := map[string][]string{
		"clusterId}[5m]))": {"cluster1"},
	}
	_, err := buildFiltersStr(filtersMap)
	assert.Error(t, err)
}

func TestBuildFiltersStr_SingleQuoteInValue(t *testing.T) {
	filtersMap := map[string][]string{
		"clusterId": {"cluster1', injectedLabel=~'.*"},
	}
	result, err := buildFiltersStr(filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, `clusterId=~'cluster1\', injectedLabel=~\'.*'`, result)
}

func TestBuildFiltersStr_BackslashInValue(t *testing.T) {
	filtersMap := map[string][]string{
		"clusterId": {`cluster\1`},
	}
	result, err := buildFiltersStr(filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, `clusterId=~'cluster\\1'`, result)
}

func TestInjectSpecificFilters_InvalidKey(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"bad}key": {"value1"},
	}
	query, err := filterInjector.injectSpecificFilters("abc{FILTER::bad}key::FILTER}", filtersMap)
	assert.Error(t, err)
	assert.Equal(t, "", query)
}

func TestInjectSpecificFilters_SingleQuoteInValue(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	filtersMap := map[string][]string{
		"key1": {"value1', injectedLabel=~'.*"},
	}
	query, err := filterInjector.injectSpecificFilters("abc{FILTER::key1::FILTER}", filtersMap)
	assert.NoError(t, err)
	assert.Equal(t, `abc{key1=~'value1\', injectedLabel=~\'.*'}`, query)
}

func TestInjectNextIsCalled_Filter(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	next := mocks.NewQueryInjector(t)
	filterInjector.next = next
	metricInformation := &services.MetricInformation{
		Filters: map[string][]string{
			"key2": {"value3"},
		},
	}
	expectedQuery := "abc{} or on() def{key2=~'value3'}"
	next.On("Inject", expectedQuery, metricInformation).Return("final query", nil)
	query, err := filterInjector.Inject("abc{} or on() def{FILTERS}", metricInformation)
	assert.NoError(t, err)
	assert.Equal(t, "final query", query)
}

func TestInjectNextReturnsError_Filter(t *testing.T) {
	filterInjector := initSimpleFilterInjectorForTesting()
	next := mocks.NewQueryInjector(t)
	filterInjector.next = next
	metricInformation := &services.MetricInformation{
		Filters: map[string][]string{
			"key2": {"value3"},
		},
	}
	expectedQuery := "abc{} or on() def{key2=~'value3'}"
	next.On("Inject", expectedQuery, metricInformation).Return("", errors.New("next error"))
	_, err := filterInjector.Inject("abc{} or on() def{FILTERS}", metricInformation)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "next error")
}
