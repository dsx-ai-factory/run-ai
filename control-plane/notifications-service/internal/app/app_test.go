package app

import (
	"context"
	"fmt"
	nethttp "net/http"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	lm "github.com/run-ai/runai/control-plane/notifications-service/internal/listen/mocks"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMultipleRouters_Starts(t *testing.T) {
	kubeListener := lm.NewListeningConfigProvider(t)
	httpListener := lm.NewListeningConfigProvider(t)

	provider := NewConfigurationProviderMock(t)
	provider.On("GetListeningProviders").Return(map[string]listen.ListeningConfigProvider{"kube": kubeListener, "http": httpListener})
	provider.On("GetRouterConfig").Return(&route.RouterConfig{CachedEvents: 1})

	listenerFactory := lm.NewFactory(t)
	listenerFactory.On("CreateListener", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	appInstance := getApp(provider, listenerFactory)
	appInstance.InitializeEventRouters()
	assert.Len(t, appInstance.routers, 2)
}

func TestListenerNotExists_Skipped(t *testing.T) {
	provider := NewConfigurationProviderMock(t)
	listener := lm.NewListeningConfigProvider(t)
	provider.On("GetListeningProviders").Return(map[string]listen.ListeningConfigProvider{"http": listener})
	provider.On("GetRouterConfig").Return(&route.RouterConfig{CachedEvents: 1})

	listenerFactory := lm.NewFactory(t)
	listenerFactory.On("CreateListener", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

	appInstance := getApp(provider, listenerFactory)
	appInstance.InitializeEventRouters()
	assert.Len(t, appInstance.routers, 1)
}

func TestHandleListenerChange_Succeeded(t *testing.T) {
	config := NewConfigurationProviderMock(t)
	listenerFactory := lm.NewFactory(t)
	appInstance := getApp(config, listenerFactory)

	httpListener := lm.NewListeningConfigProvider(t)
	createdListener := lm.NewListener(t)
	createdListener.On("Listen", mock.Anything).Return().Maybe()
	createdListener.On("Id").Return("http-runai").Maybe()

	config.On("GetListeningProvider", "http-runai").Return(httpListener)
	config.On("GetRouterConfig").Return(&route.RouterConfig{CachedEvents: 1})
	listenerFactory.On("CreateListener", "http-runai", mock.Anything, httpListener).Return(createdListener, nil)

	err := appInstance.handleListenerChange("http-runai")
	assert.Nil(t, err)
	t.Cleanup(appInstance.routers["http-runai"].Stop)
}

func TestHandleListenerChange_ListenerAlreadyExistStopOld(t *testing.T) {
	config := NewConfigurationProviderMock(t)
	listenerFactory := lm.NewFactory(t)
	appInstance := getApp(config, listenerFactory)
	var wg sync.WaitGroup
	called := make(chan bool, 1)

	httpListener := lm.NewListeningConfigProvider(t)
	createdListener := lm.NewListener(t)
	createdListener.On("Listen", mock.Anything).Return().Maybe()
	createdListener.On("Id").Return("http").Maybe()

	config.On("GetListeningProvider", "http").Return(httpListener)
	config.On("GetRouterConfig").Return(&route.RouterConfig{CachedEvents: 1})
	listenerFactory.On("CreateListener", "http", mock.Anything, httpListener).Return(createdListener, nil)

	listener := lm.NewListener(t)
	listener.On("Id").Return("http")

	appInstance.routers["http"] = &route.NotificationsRouter{
		Listener: listener,
		CancelCtx: func() {
			called <- true
		},
		CachedEventsCh: make(chan *events.Event),
		Wg:             &wg,
	}

	err := appInstance.handleListenerChange("http")
	<-called
	assert.Nil(t, err)
	t.Cleanup(appInstance.routers["http"].Stop)
}

func TestHandleListenerChange_NoRouteInConfig(t *testing.T) {
	config := NewConfigurationProviderMock(t)
	appInstance := getApp(config, lm.NewFactory(t))

	config.On("GetListeningProvider", "http").Return(nil)

	err := appInstance.handleListenerChange("http")
	assert.Equal(t, "listener not found http", err.Error())
}

func TestHandleListenerChange_FailedToCreateListener(t *testing.T) {
	config := NewConfigurationProviderMock(t)
	listenerFactory := lm.NewFactory(t)
	appInstance := getApp(config, listenerFactory)

	badListener := lm.NewListeningConfigProvider(t)
	badListener.On("GetListenerType").Return("bad_type")

	config.On("GetListeningProvider", "http").Return(badListener)
	config.On("GetRouterConfig").Return(&route.RouterConfig{CachedEvents: 1})
	listenerFactory.On("CreateListener", "http", mock.Anything, badListener).Return(nil, fmt.Errorf("listener bad_type not found"))

	err := appInstance.handleListenerChange("http")
	assert.Equal(t, "failed to rebuild router on listener http type bad_type", err.Error())
}

func getApp(config ConfigurationProvider, listenerFactory listen.Factory) *App {
	return &App{
		configProvider:  config,
		routers:         map[string]*route.NotificationsRouter{},
		listenerFactory: listenerFactory,
	}
}

func TestStart_RegistersAndListens(t *testing.T) {
	routesRegister := NewRoutesRegisterMock(t)
	cleaner := NewCleanerMock(t)
	cleaner.On("Schedule", mock.Anything).Return().Once()
	routesRegister.On("Register", mock.AnythingOfType("*chi.Mux")).
		Return().
		Run(func(args mock.Arguments) {
			args.Get(0).(*chi.Mux).Get("/some_route", func(w nethttp.ResponseWriter, _ *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			})
		}).
		Once()
	routesRegister.On("GetName").Return("test-routes")
	updatesListener := NewConfigUpdatesListenerMock(t)
	listenCalled := make(chan struct{}, 1)
	updatesListener.On("Listen", mock.Anything, mock.AnythingOfType("func(string) error")).
		Return().
		Run(func(args mock.Arguments) { listenCalled <- struct{}{} }).
		Once()

	ctx, cancel := context.WithCancel(context.Background())

	appInstance := &App{
		config:                &AppConfig{APIHost: "127.0.0.1", APIPort: "0"},
		routesRegisters:       []RoutesRegister{routesRegister},
		cleaners:              []Cleaner{cleaner},
		configUpdatesListener: updatesListener,
		routers:               map[string]*route.NotificationsRouter{},
		ctx:                   ctx,
		cancelCtx:             cancel,
	}

	appInstance.Start()
	t.Cleanup(appInstance.Stop)

	assert.NotNil(t, appInstance.appServer)
	<-listenCalled
}

func TestStart_NoRoutesRegistered(t *testing.T) {
	routesRegister := NewRoutesRegisterMock(t)
	updatesListener := NewConfigUpdatesListenerMock(t)
	listenCalled := make(chan struct{}, 1)
	routesRegister.On("Register", mock.AnythingOfType("*chi.Mux")).Return().Once()
	updatesListener.On("Listen", mock.Anything, mock.AnythingOfType("func(string) error")).
		Return().
		Run(func(args mock.Arguments) { listenCalled <- struct{}{} }).
		Once()
	routesRegister.On("GetName").Return("noop-routes")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appInstance := &App{
		config:                &AppConfig{APIHost: "127.0.0.1", APIPort: "0"},
		routesRegisters:       []RoutesRegister{routesRegister},
		cleaners:              []Cleaner{},
		configUpdatesListener: updatesListener,
		routers:               map[string]*route.NotificationsRouter{},
		ctx:                   ctx,
		cancelCtx:             cancel,
	}

	appInstance.Start()

	assert.Nil(t, appInstance.appServer)
	<-listenCalled
}

func TestRegisterRoutes_HistoryDisabled_OnlyMainRoutesRegistered(t *testing.T) {
	appInstance := &App{
		config:          &AppConfig{},
		routesRegisters: []RoutesRegister{getRoutesRegisterMock(t, "")},
	}

	router := chi.NewRouter()
	appInstance.registerRoutes(router)

	assert.Len(t, router.Routes(), 1, "only the main route should be registered when history is disabled")
}

func TestRegisterRoutes_HistoryEnabledButServiceMissing_OnlyMainRoutesRegistered(t *testing.T) {
	appInstance := &App{
		config:          &AppConfig{},
		routesRegisters: []RoutesRegister{getRoutesRegisterMock(t, "")},
	}

	router := chi.NewRouter()
	appInstance.registerRoutes(router)

	assert.Len(t, router.Routes(), 1, "history routes must not be registered when historyService is nil even if History.Enabled is true")
}

func TestRegisterRoutes_HistoryEnabledWithService_RegistersHistoryRoutes(t *testing.T) {
	routesRegisters := []RoutesRegister{getRoutesRegisterMock(t, ""), getRoutesRegisterMock(t, "_history")}
	cleaner := NewCleanerMock(t)

	appInstance := &App{
		config:          &AppConfig{},
		routesRegisters: routesRegisters,
		cleaners:        []Cleaner{cleaner},
	}

	router := chi.NewRouter()
	appInstance.registerRoutes(router)

	assert.Greater(t, len(router.Routes()), 1, "history routes should be registered in addition to the main route")

	cleaner.On("Schedule", mock.Anything).Return().Once()
	appInstance.scheduleCleaners()
}

func getRoutesRegisterMock(t *testing.T, routeSuffix string) *RoutesRegisterMock {
	rr := NewRoutesRegisterMock(t)
	rr.On("GetName").Return("test-main-routes" + routeSuffix)
	rr.On("Register", mock.AnythingOfType("*chi.Mux")).
		Return().
		Run(func(args mock.Arguments) {
			args.Get(0).(*chi.Mux).Get("/main"+routeSuffix, func(w nethttp.ResponseWriter, _ *nethttp.Request) {
				w.WriteHeader(nethttp.StatusOK)
			})
		}).
		Once()
	return rr
}
