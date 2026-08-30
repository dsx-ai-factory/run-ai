package route_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/sink"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	lm "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRateLimit - this test will fail if the service is processing more than 1 event when limited by rate limiting and fails to stop mid handling
func TestRateLimit(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)
	routeCProviderMock := mocks.NewRouteConfigProvider(t)
	routeMock := mocks.NewRoute(t)

	routeCProviderMock.On("GetRouteId").Return("")
	eventsCh := make(chan *events.Event, 10)
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)

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

	called := make(chan struct{})

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)
	listenerMock.On("Id").Times(3).Return("listener_name")

	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Once().Return(
		[]route.RouteConfigProvider{routeCProviderMock})

	configProviderMock.On("GetSinkProvider", mock.Anything, "email1").Return(
		&sink.ConfigProvider{SinkConfig: map[string]*sinks.SinkConfig{"email1": {Email: &sinks.EmailSinkConfig{}}}}, nil)

	routeCProviderMock.On("GetNotifyIds").Return([]string{"email1"})
	routeMock.On("Route", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)

	nr.Start()
	eventsCh <- &events.Event{}
	eventsCh <- &events.Event{}
	<-called
	<-called
	nr.Stop()
	assert.Equal(t, 1, len(eventsCh))
}

func TestRateLimit_ProcessingOneAtATime(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)
	routeCProviderMock := mocks.NewRouteConfigProvider(t)
	routeMock := mocks.NewRoute(t)

	routeCProviderMock.On("GetRouteId").Return("")
	eventsCh := make(chan *events.Event, 10)
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)

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

	calledListen := make(chan struct{}, 1)
	calledRoute := make(chan struct{}, 1)
	ctxFirstRoute := context.Background()
	ctxFirstRoute, cancelCtxFirstRoute := context.WithCancel(ctxFirstRoute)

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			calledListen <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)
	listenerMock.On("Id").Return("listener_name")

	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Twice().Return(
		[]route.RouteConfigProvider{routeCProviderMock})

	configProviderMock.On("GetSinkProvider", mock.Anything, "email1").Twice().Return(
		&sink.ConfigProvider{SinkConfig: map[string]*sinks.SinkConfig{"email1": {Email: &sinks.EmailSinkConfig{}}}}, nil)

	routeCProviderMock.On("GetNotifyIds").Twice().Return([]string{"email1"})

	routeMock.On("Route", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			calledRoute <- struct{}{}
			<-ctxFirstRoute.Done()
		},
	).Return(nil)
	routeMock.On("Route", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			calledRoute <- struct{}{}
		},
	).Return(nil)

	nr.Start()
	eventsCh <- &events.Event{}
	eventsCh <- &events.Event{}
	<-calledListen
	<-calledRoute
	assert.Equal(t, 1, len(eventsCh))
	cancelCtxFirstRoute()

	<-calledRoute
	assert.Equal(t, 0, len(eventsCh))

	nr.Stop()
}

// TestWehnNoEventsExists - This test will fail for example in case where we don't close CachedEventsCh at the service Stop function
func TestWhenNoEventsExist(t *testing.T) {
	listenerMock := lm.NewListener(t)
	eventsCh := make(chan *events.Event, 10)
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	ns := route.NotificationsRouter{
		ConcurrentLimitCh: make(chan struct{}, 1),
		CachedEventsCh:    eventsCh,
		Listener:          listenerMock,
		Ctx:               ctx,
		CancelCtx:         cancelCtx,
		Wg:                &sync.WaitGroup{},
	}

	called := make(chan struct{})
	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)
	listenerMock.On("Id").Twice().Return("listener_name")

	ns.Start()
	stopped := make(chan struct{})
	<-called
	go func() {
		ns.Stop()
		stopped <- struct{}{}
	}()
	select {
	case <-stopped:
		return
	case <-time.After(500 * time.Millisecond):
		t.Error("route failed to stop when events channel is empty")
	}
}

func TestTriggers2Routes_When2RouteProviders(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)
	routeCProviderMock := mocks.NewRouteConfigProvider(t)
	routeMock := mocks.NewRoute(t)

	eventsCh := make(chan *events.Event, 10)
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)

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

	called := make(chan struct{})

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)
	listenerMock.On("Id").Times(3).Return("listener_name")

	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Once().Return(
		[]route.RouteConfigProvider{
			routeCProviderMock,
			routeCProviderMock,
		})

	configProviderMock.On("GetSinkProvider", mock.Anything, "email1").Return(
		&sink.ConfigProvider{SinkConfig: map[string]*sinks.SinkConfig{"email1": {Email: &sinks.EmailSinkConfig{}}}}, nil)

	routeCProviderMock.On("GetNotifyIds").Return([]string{"email1"})
	routeCProviderMock.On("GetRouteId").Return("")

	routeMock.On("Route", mock.Anything, mock.Anything).Twice().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
		},
	).Return(nil)

	nr.Start()
	eventsCh <- &events.Event{}
	<-called
	<-called
	<-called
	nr.Stop()
	assert.Equal(t, 0, len(eventsCh))
}

func TestHandleEvent_SendsCopies(t *testing.T) {
	listenerMock := lm.NewListener(t)
	configProviderMock := mocks.NewRouterConfigProvider(t)
	routeCProviderMock := mocks.NewRouteConfigProvider(t)
	routeMock := mocks.NewRoute(t)

	eventsCh := make(chan *events.Event, 10)
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)

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

	called := make(chan struct{})

	listenerMock.On("Listen", mock.Anything, mock.Anything).Once().Run(
		func(args mock.Arguments) {
			called <- struct{}{}
			<-ctx.Done()
		},
	).Return(nil)
	listenerMock.On("Id").Times(3).Return("listener_name")

	configProviderMock.On("GetRouteProviders", mock.Anything, mock.Anything).Once().Return(
		[]route.RouteConfigProvider{
			routeCProviderMock,
			routeCProviderMock,
		})

	configProviderMock.On("GetSinkProvider", mock.Anything, "email1").Return(
		&sink.ConfigProvider{SinkConfig: map[string]*sinks.SinkConfig{"email1": {Email: &sinks.EmailSinkConfig{}}}}, nil)

	routeCProviderMock.On("GetNotifyIds").Return([]string{"email1"}).Twice()
	routeCProviderMock.On("GetRouteId").Return("").Twice()

	e := &events.Event{}
	routeMock.On("Route", mock.Anything, mock.Anything).Twice().Run(
		func(args mock.Arguments) {
			eCopy := args[1].(*events.Event)
			if eCopy == e {
				t.Errorf("events passed to route are not duplicated")
			}
			called <- struct{}{}
		},
	).Return(nil)

	nr.Start()

	eventsCh <- e
	<-called
	<-called
	<-called
	nr.Stop()
}
