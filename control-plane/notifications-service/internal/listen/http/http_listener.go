/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/spf13/viper"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	api "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen/http/openapi/generated"
	runaimetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/http"
)

const ListenerType = "http_listener"

type HttpListenerConfig struct {
	Port int `mapstructure:"port" json:"port"`
}

func NewHttpListenerConfig(v *viper.Viper) *HttpListenerConfig {
	httpListenerDefaults(v)
	cfg := HttpListenerConfig{}
	if err := v.Unmarshal(&cfg, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed unmarshalling configuration of route %q", v.GetString("Id"))
	}

	return &cfg
}

func httpListenerDefaults(v *viper.Viper) {
	config.SetDefault(v, "port", 5000)
}

type HttpListener struct {
	id                 string
	config             *HttpListenerConfig
	eventCh            chan *events.Event
	listenAndServeFunc func() error // enables mocking the listener
	closeServer        func(ctx context.Context) error
	server             *http.Server
}

func NewHttpListener(id string, eventCh chan *events.Event, config *HttpListenerConfig) *HttpListener {
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      createRouter(id, eventCh),
	}
	return &HttpListener{
		id:                 id,
		config:             config,
		eventCh:            eventCh,
		listenAndServeFunc: server.ListenAndServe,
		closeServer:        server.Shutdown,
		server:             &server,
	}
}

func (hl *HttpListener) Id() string {
	return hl.id
}

func (hl *HttpListener) Listen(ctx context.Context) {
	addr := fmt.Sprintf(":%d", hl.config.Port)
	log.Info().Msgf("starting the Http listener on %q", addr)
	failedCh := make(chan bool)
	go hl.listenAndServe(failedCh, addr)
	hl.waitAndCloseServer(failedCh, ctx, addr)
}

func (hl *HttpListener) listenAndServe(failedCh chan bool, addr string) {
	err := hl.listenAndServeFunc()
	if err != nil {
		failedCh <- true
		log.Err(err).Msgf("listener %s failed to listen on %q", hl.id, addr)
	}
}

func (hl *HttpListener) waitAndCloseServer(failedCh chan bool, ctx context.Context, addr string) {
	select {
	case <-failedCh:
		return
	case <-ctx.Done():
		err := hl.closeServer(ctx)
		log.Info().Msgf("stopped the Http listener %s on %q", hl.id, addr)
		if err != nil {
			log.Err(err).Msgf("failed to close listener %s on %q", hl.id, addr)
		}
	}
}

func (hl *HttpListener) Stop() {
	port := hl.config.Port
	log.Info().Msgf("stopping the Http listener on %d", port)

	err := hl.server.Close()
	if err != nil {
		log.Err(err).Msgf("failed to stop listener on %d", port)
	}
	log.Info().Msgf("stopped the Http listener on %d", port)
}

func createRouter(listenerId string, eventCh chan *events.Event) chi.Router {
	service := &HttpService{
		listenerId: listenerId,
		eventCh:    eventCh,
	}

	notifyApiController := api.NewNotifyAPIController(service)

	router := chi.NewRouter()

	router.Group(func(router chi.Router) {
		router.Use(runaimetrics.HttpMiddleware)
		AddRoutes(router, notifyApiController)
	})

	return router
}
