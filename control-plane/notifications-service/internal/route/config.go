/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/enrich"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/notify"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
)

type RouteConfigProvider interface {
	GetListenerId() string
	GetRouteId() string
	enrich.EnrichingConfigProvider
	filter.FilteringConfigProvider
	notify.NotifyingConfigProvider
}

type SinkConfigProvider interface {
	GetSinks(id ...string) (map[string]*sinks.SinkConfig, error)
}

type RouterConfigProvider interface {
	GetRouterConfig() *RouterConfig
	// GetRouteProviders should return all the relevant configuration providers according to the event id
	GetRouteProviders(listenerId string, event *events.Event) []RouteConfigProvider

	GetSinkProvider(event *events.Event, id ...string) (SinkConfigProvider, error)
}
