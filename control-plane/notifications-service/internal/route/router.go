/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route

import (
	"context"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen"
)

type RouteBuilder func(RouteConfigProvider, SinkConfigProvider) Route

type RouterConfig struct {
	ConcurrentLimit int `mapstructure:"concurrent_limit"`
	CachedEvents    int `mapstructure:"cached_events"`
}

func NewNotificationsRouter(listener listen.Listener, eventCh chan *events.Event, configProvider RouterConfigProvider) *NotificationsRouter {
	ctx := context.Background()
	ctx, cancelCtx := context.WithCancel(ctx)
	return &NotificationsRouter{
		ConfigProvider:    configProvider,
		ConcurrentLimitCh: make(chan struct{}, configProvider.GetRouterConfig().ConcurrentLimit),
		CachedEventsCh:    eventCh,
		RouteBuilder:      NewHandlersRoute,
		Listener:          listener,
		Ctx:               ctx,
		CancelCtx:         cancelCtx,
		Wg:                &sync.WaitGroup{},
	}
}

type NotificationsRouter struct {
	ConfigProvider    RouterConfigProvider
	ConcurrentLimitCh chan struct{}
	CachedEventsCh    chan *events.Event
	RouteBuilder      RouteBuilder
	Listener          listen.Listener
	Ctx               context.Context
	CancelCtx         context.CancelFunc
	Wg                *sync.WaitGroup
}

func (ns NotificationsRouter) Start() error {
	log.Info().Msg("notifications-service started")
	go ns.startListener()
	go ns.handleEvents()
	return nil
}

func (ns NotificationsRouter) Stop() {
	log.Debug().Msgf("router of listener %s is stopping", ns.Listener.Id())
	close(ns.CachedEventsCh)
	ns.CancelCtx()
	ns.Wg.Wait()
	log.Debug().Msgf("router of listener %s is stopped", ns.Listener.Id())
}

func (ns NotificationsRouter) startListener() {
	ns.Wg.Add(1)
	defer ns.Wg.Done()
	ns.Listener.Listen(ns.Ctx)
	log.Info().Msg("listener shut down")
}

func (ns NotificationsRouter) handleEvents() {
	ns.Wg.Add(1)
	defer ns.Wg.Done()
	for {
		select {
		case ns.ConcurrentLimitCh <- struct{}{}:
			ns.handleEvent(<-ns.CachedEventsCh)
		case <-ns.Ctx.Done():
			log.Info().Msg("stopped handling event loop, processing left overs")
			return
		}
	}
}

func (ns NotificationsRouter) handleEvent(event *events.Event) {
	if event == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("recovered from panic while creating routes for event %s: %v", event.UID, r)
			<-ns.ConcurrentLimitCh
		}
	}()

	routes, routeIds := ns.createRoutes(event)
	e := *event
	log.Info().Msgf("listener id: %s: found routes: %s for event %s", e.ListenerId, strings.Join(routeIds, ","), e.UID)

	ns.Wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error().Msgf("recovered from panic while routing event %s: %v", e.UID, r)
			}
			<-ns.ConcurrentLimitCh
			ns.Wg.Done()
		}()

		for _, route := range routes {
			eCopy := e
			err := route.Route(ns.Ctx, &eCopy)
			if err != nil {
				log.Err(err).Msgf("failed to handle event %s", eCopy.UID)
			}
		}
	}()
}

func (ns NotificationsRouter) createRoutes(event *events.Event) (routes []Route, routeIds []string) {
	providers := ns.ConfigProvider.GetRouteProviders(ns.Listener.Id(), event)

	log.Debug().Msgf("got %d available routes, attempting to send notifications, event: %v", len(providers), event.UID)
	for _, provider := range providers {
		sinkProvider, err := ns.ConfigProvider.GetSinkProvider(event, provider.GetNotifyIds()...)
		if err != nil {
			log.Err(err).Msg("failed getting sink provider!")
		} else {
			routes = append(routes, ns.RouteBuilder(provider, sinkProvider))
			routeIds = append(routeIds, provider.GetRouteId())
		}
	}
	return routes, routeIds
}
