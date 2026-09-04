/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/sink"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	lm "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/route/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// A panic inside a sink must not take down the process, and must still release
// the concurrency slot so subsequent events keep being handled.
func TestPanickingRoute_DoesNotKillRouterAndReleasesSlot(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)
	routeCProviderMock := mocks.NewRouteConfigProvider(t)
	routeMock := mocks.NewRoute(t)

	eventsCh := make(chan *events.Event, 10)
	ctx, cancelCtx := context.WithCancel(context.Background())

	nr := route.NotificationsRouter{
		ConcurrentLimitCh: make(chan struct{}, 1),
		CachedEventsCh:    eventsCh,
		ConfigProvider:    configProviderMock,
		RouteBuilder: func(rcp route.RouteConfigProvider, sinkProvider route.SinkConfigProvider) route.Route {
			return routeMock
		},
		Listener:  listenerMock,
		Ctx:       ctx,
		CancelCtx: cancelCtx,
		Wg:        &sync.WaitGroup{},
	}

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) { <-ctx.Done() }).Return(nil)
	listenerMock.On("Id").Return("listener_name")

	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Return(
		[]route.RouteConfigProvider{routeCProviderMock})
	configProviderMock.On("GetSinkProvider", mock.Anything, "email1").Return(
		&sink.ConfigProvider{SinkConfig: map[string]*sinks.SinkConfig{"email1": {Email: &sinks.EmailSinkConfig{}}}}, nil)
	routeCProviderMock.On("GetRouteId").Return("")
	routeCProviderMock.On("GetNotifyIds").Return([]string{"email1"})

	panicked := make(chan struct{}, 1)
	recovered := make(chan struct{}, 1)
	routeMock.On("Route", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			panicked <- struct{}{}
			panic("interface conversion: interface {} is float64, not string")
		}).Return(nil)
	routeMock.On("Route", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) { recovered <- struct{}{} }).Return(nil)

	nr.Start()
	defer nr.Stop()

	eventsCh <- &events.Event{UID: "panics"}
	select {
	case <-panicked:
	case <-time.After(time.Second):
		t.Fatal("first event was never routed")
	}

	eventsCh <- &events.Event{UID: "survives"}
	select {
	case <-recovered:
	case <-time.After(time.Second):
		t.Fatal("router stopped handling events after a route panicked")
	}
}

// A malformed event property that panics during route creation is handled on
// the event loop goroutine itself, so it must be recovered there too.
func TestPanickingRouteCreation_DoesNotKillRouter(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)

	eventsCh := make(chan *events.Event, 10)
	ctx, cancelCtx := context.WithCancel(context.Background())

	nr := route.NotificationsRouter{
		ConcurrentLimitCh: make(chan struct{}, 1),
		CachedEventsCh:    eventsCh,
		ConfigProvider:    configProviderMock,
		Listener:          listenerMock,
		Ctx:               ctx,
		CancelCtx:         cancelCtx,
		Wg:                &sync.WaitGroup{},
	}

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) { <-ctx.Done() }).Return(nil)
	listenerMock.On("Id").Return("listener_name")

	handled := make(chan string, 2)
	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			handled <- "panics"
			panic("interface conversion: interface {} is nil, not string")
		}).Return(nil)
	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) { handled <- "survives" }).Return([]route.RouteConfigProvider{})

	nr.Start()
	defer nr.Stop()

	eventsCh <- &events.Event{UID: "panics"}
	eventsCh <- &events.Event{UID: "survives"}

	for i := 0; i < 2; i++ {
		select {
		case <-handled:
		case <-time.After(time.Second):
			t.Fatal("router stopped handling events after route creation panicked")
		}
	}
	assert.Empty(t, eventsCh)
}
