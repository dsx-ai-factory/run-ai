package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/route"
)

type RoutesRegister interface {
	Register(router *chi.Mux)
	GetName() string
}

type App struct {
	config                *AppConfig
	configProvider        ConfigurationProvider
	configUpdatesListener ConfigUpdatesListener
	routesRegisters       []RoutesRegister
	cleaners              []Cleaner
	listenerFactory       listen.Factory
	routers               map[string]*route.NotificationsRouter
	ctx                   context.Context
	cancelCtx             context.CancelFunc
	appServer             *http.Server
}

func NewApp(factory ConfigFactory) *App {
	ctx, cancelCtx := context.WithCancel(context.Background())
	cp := factory.CreateConfigurationProvider()
	app := &App{
		config:                cp.GetAppConfig(),
		configProvider:        cp,
		configUpdatesListener: factory.CreateListener(),
		routesRegisters:       factory.CreateRoutesRegisters(),
		cleaners:              factory.CreateCleaners(),
		listenerFactory:       listen.NewListenerFactory(),
		routers:               map[string]*route.NotificationsRouter{},
		ctx:                   ctx,
		cancelCtx:             cancelCtx,
	}

	logLevel := getLogLevel(app.configProvider.GetAppConfig().LogLevel)
	zerolog.SetGlobalLevel(logLevel)
	log.Info().Msgf("setting log level to %v", logLevel)
	return app
}

func (a *App) InitializeEventRouters() {
	for name, listeningProvider := range a.configProvider.GetListeningProviders() {
		r, err := a.buildRouter(name, listeningProvider)
		if err == nil {
			a.routers[name] = r
		}
	}
}

func (a *App) Start() {
	for _, router := range a.routers {
		err := router.Start()
		if err != nil {
			log.Err(err).Msgf("failed to start router on listener %s", router.Listener.Id())
		}
		log.Info().Msgf("router on listener %s started successfully", router.Listener.Id())
	}

	router := chi.NewRouter()
	a.registerRoutes(router)
	a.scheduleCleaners()

	a.configUpdatesListener.Listen(a.ctx, a.handleListenerChange)

	if len(router.Routes()) == 0 {
		log.Info().Msg("no routes registered, skipping app server start")
		return
	}

	routerAddress := net.JoinHostPort(a.config.APIHost, a.config.APIPort)
	log.Info().Msgf("listening for changes on address %s", routerAddress)

	a.appServer = listenAndServe(a.ctx, routerAddress, router)
}

func (a *App) Stop() {
	for _, r := range a.routers {
		r.Stop()
	}

	if a.appServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.appServer.Shutdown(ctx); err != nil {
			log.Err(err).Msg("HTTP server shutdown failed")
		}
	}
	a.cancelCtx()
}

func (a *App) registerRoutes(router *chi.Mux) {
	for _, routesRegister := range a.routesRegisters {
		routesRegister.Register(router)
		log.Info().Msgf("routes registered for %s", routesRegister.GetName())
	}
}

func (a *App) scheduleCleaners() {
	for _, cleaner := range a.cleaners {
		cleaner.Schedule(a.ctx)
	}
}

func listenAndServe(ctx context.Context, addr string, aRouter *chi.Mux) *http.Server {
	server := &http.Server{
		Addr: addr,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      aRouter, // Pass our instance of chi.Router in.
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("failed to listen on address %s", addr)
		}
	}()

	return server
}

func (a *App) buildRouter(name string, listeningProvider listen.ListeningConfigProvider) (*route.NotificationsRouter, error) {
	listenerEventCh := make(chan *events.Event, a.configProvider.GetRouterConfig().CachedEvents)
	listener, err := a.listenerFactory.CreateListener(name, listenerEventCh, listeningProvider)
	if err != nil {
		log.Err(err).Msgf("failed to create router for listener %s", name)
		return nil, err
	}

	router := route.NewNotificationsRouter(listener, listenerEventCh, a.configProvider)
	log.Info().Msgf("router created for listener %s", name)
	return router, nil
}

func getLogLevel(logLevelName string) zerolog.Level {
	logLevel, err := zerolog.ParseLevel(logLevelName)
	if err != nil {
		log.Warn().Msgf("configured log level %s is not valid", logLevelName)
		return zerolog.InfoLevel
	}
	return logLevel
}

func (a *App) handleListenerChange(listenerId string) error {
	if a.routers[listenerId] != nil {
		a.routers[listenerId].Stop()
		delete(a.routers, listenerId)
	}

	listeningProvider := a.configProvider.GetListeningProvider(listenerId)
	if listeningProvider == nil {
		return fmt.Errorf("listener not found %s", listenerId)
	}

	router, err := a.buildRouter(listenerId, listeningProvider)
	if err != nil {
		return fmt.Errorf("failed to rebuild router on listener %s type %s", listenerId, listeningProvider.GetListenerType())
	}

	err = router.Start()
	if err != nil {
		return fmt.Errorf("failed to start router on listener %s", listenerId)
	}

	a.routers[listenerId] = router
	return nil
}
