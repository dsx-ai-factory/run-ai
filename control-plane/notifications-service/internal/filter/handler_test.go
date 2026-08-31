/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	em "github.com/run-ai/runai/control-plane/notifications-service/internal/events/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter/filters"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter/mocks"
	"github.com/stretchr/testify/mock"
)

func TestHandler_PolicyMatchAll_Match(t *testing.T) {
	configProviderMock, factoryMock, filterMock, nextMock, handler := prepareMocks(t)

	configProviderMock.On("GetGlobalFilteringConfig").Return(&filter.FilteringConfig{
		Policy: "all",
		Filters: []*filter.FilterConfig{
			{
				Type: filters.PrometheusAlertsFilterName,
			},
			{
				Type: filters.PrometheusAlertsFilterName,
			},
		},
	}).Once()

	factoryMock.On("CreateFilter", mock.Anything).Return(filterMock, nil).Twice()

	nextMock.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()

	mockCall := filterMock.On("Filter", mock.Anything, mock.Anything)
	callCount := 0
	mockCall.Run(func(args mock.Arguments) {
		callCount++
		if callCount == 0 {
			mockCall.Return(false, nil)
		} else {
			mockCall.Return(false, nil)
		}
	})

	handler.Handle(context.Background(), &events.Event{})
}

func TestHandler_PolicyMatchAll_DontMatch(t *testing.T) {
	configProviderMock, factoryMock, filterMock, nextMock, handler := prepareMocks(t)

	configProviderMock.On("GetGlobalFilteringConfig").Return(&filter.FilteringConfig{
		Policy: "all",
		Filters: []*filter.FilterConfig{
			{
				Type: filters.PrometheusAlertsFilterName,
			},
			{
				Type: filters.PrometheusAlertsFilterName,
			},
		},
	}).Once()

	factoryMock.On("CreateFilter", mock.Anything).Return(filterMock, nil).Once()

	nextMock.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)

	mockCall := filterMock.On("Filter", mock.Anything, mock.Anything)
	callCount := 0
	mockCall.Run(func(args mock.Arguments) {
		callCount++
		if callCount == 0 {
			mockCall.Return(false, nil)
		} else {
			mockCall.Return(true, nil)
		}
	})

	handler.Handle(context.Background(), &events.Event{})
}

func TestHandler_PolicyMatchAny_Match(t *testing.T) {
	configProviderMock, factoryMock, filterMock, nextMock, handler := prepareMocks(t)

	configProviderMock.On("GetGlobalFilteringConfig").Return(&filter.FilteringConfig{
		Policy: "any",
		Filters: []*filter.FilterConfig{
			{
				Type: filters.PrometheusAlertsFilterName,
			},
			{
				Type: filters.PrometheusAlertsFilterName,
			},
		},
	}).Once()

	factoryMock.On("CreateFilter", mock.Anything).Return(filterMock, nil).Once()

	nextMock.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()

	mockCall := filterMock.On("Filter", mock.Anything, mock.Anything)
	callCount := 0
	mockCall.Run(func(args mock.Arguments) {
		callCount++
		if callCount == 0 {
			mockCall.Return(true, nil)
		} else {
			mockCall.Return(false, nil)
		}
	})

	handler.Handle(context.Background(), &events.Event{})
}

func TestHandler_PolicyMatchAny_DontMatch(t *testing.T) {
	configProviderMock, factoryMock, filterMock, nextMock, handler := prepareMocks(t)

	configProviderMock.On("GetGlobalFilteringConfig").Return(&filter.FilteringConfig{
		Policy: "any",
		Filters: []*filter.FilterConfig{
			{
				Type: filters.PrometheusAlertsFilterName,
			},
			{
				Type: filters.PrometheusAlertsFilterName,
			},
		},
	}).Once()

	factoryMock.On("CreateFilter", mock.Anything).Return(filterMock, nil).Twice()

	nextMock.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)

	mockCall := filterMock.On("Filter", mock.Anything, mock.Anything)
	callCount := 0
	mockCall.Run(func(args mock.Arguments) {
		callCount++
		if callCount == 0 {
			mockCall.Return(true, nil)
		} else {
			mockCall.Return(true, nil)
		}
	})

	handler.Handle(context.Background(), &events.Event{})
}

func TestHandler_PolicyMatchAny_NoFilters(t *testing.T) {
	configProviderMock, _, _, nextMock, handler := prepareMocks(t)

	configProviderMock.On("GetGlobalFilteringConfig").Return(&filter.FilteringConfig{
		Policy:  "any",
		Filters: []*filter.FilterConfig{},
	}).Once()

	nextMock.On("Handle", mock.Anything, mock.Anything).Return(nil).Once()

	handler.Handle(context.Background(), &events.Event{})
}

func prepareMocks(t *testing.T) (configProviderMock *mocks.FilteringConfigProvider, factoryMock *mocks.Factory, filterMock *mocks.Filter, nextMock *em.EventHandler, handler *filter.FilteringEventHandler) {
	configProviderMock = mocks.NewFilteringConfigProvider(t)
	nextMock = em.NewEventHandler(t)
	factoryMock = mocks.NewFactory(t)
	filterMock = mocks.NewFilter(t)
	handler = &filter.FilteringEventHandler{
		ConfigProvider: configProviderMock,
		Next:           nextMock,
		FilterFactory:  factoryMock,
	}

	return configProviderMock, factoryMock, filterMock, nextMock, handler
}
