/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/enrich"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/filter"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/notify"
)

type Route interface {
	Route(ctx context.Context, event *events.Event) error
}

type HandlersRoute struct {
	ConfigProvider RouteConfigProvider
	Handler        events.EventHandler
}

func NewHandlersRoute(configProvider RouteConfigProvider, sinkProvider SinkConfigProvider) Route {
	notifyHandler := notify.NewNotifyingHandler(configProvider, sinkProvider)
	filterHandler := filter.NewFilteringEventHandler(configProvider, notifyHandler)

	enricherFactory := enrich.NewEnricherFactory()
	enrichHandler := enrich.NewEnrichingEventHandler(configProvider, enricherFactory, filterHandler)

	return &HandlersRoute{
		ConfigProvider: configProvider,
		Handler:        enrichHandler,
	}
}

func (route *HandlersRoute) Route(ctx context.Context, event *events.Event) error {
	log.Debug().Msgf("routing event %s with route %s", event.UID, route.ConfigProvider.GetRouteId())
	return route.Handler.Handle(ctx, event)
}
