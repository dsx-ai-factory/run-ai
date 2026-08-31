/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route_test

import (
	"context"
	"errors"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	em "github.com/run-ai/runai/control-plane/notifications-service/internal/events/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoute_Error(t *testing.T) {
	handler := em.NewEventHandler(t)
	configProviderMock := mocks.NewRouteConfigProvider(t)
	routeInstance := &route.HandlersRoute{
		ConfigProvider: configProviderMock,
		Handler:        handler,
	}
	configProviderMock.On("GetRouteId").Return("")
	handler.On("Handle", mock.Anything, mock.Anything).Once().Return(errors.New("stam"))

	result := routeInstance.Route(context.Background(), &events.Event{})
	assert.ErrorContains(t, result, "stam")
}

func TestRoute_Works(t *testing.T) {
	handler := em.NewEventHandler(t)
	configProviderMock := mocks.NewRouteConfigProvider(t)
	routeInstance := &route.HandlersRoute{
		ConfigProvider: configProviderMock,
		Handler:        handler,
	}
	configProviderMock.On("GetRouteId").Return("")
	handler.On("Handle", mock.Anything, mock.Anything).Once().Return(nil)

	result := routeInstance.Route(context.Background(), &events.Event{})
	assert.Nil(t, result)
}
