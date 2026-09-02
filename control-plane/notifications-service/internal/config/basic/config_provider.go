/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package basic

import (
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/service"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen"
	router "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/util"
	"golang.org/x/exp/maps"
)

type ConfigurationProvider struct {
	service.Config
	routeProviders []router.RouteConfigProvider
}

func NewConfigurationProvider(configPath string) *ConfigurationProvider {
	serviceConf := service.NewConfig(configPath)
	if len(serviceConf.GetListeningProviders()) == 0 {
		log.Warn().Msg("no listeners found")
	}

	return NewConfigurationProviderFromServiceConf(serviceConf)
}

func NewConfigurationProviderFromServiceConf(serviceConf *service.Config) *ConfigurationProvider {
	provider := &ConfigurationProvider{
		Config:         *serviceConf,
		routeProviders: []router.RouteConfigProvider{},
	}

	provider.initRouteProviders()
	return provider
}

func (bcp *ConfigurationProvider) GetListeningProviders() map[string]listen.ListeningConfigProvider {
	return bcp.Config.GetListeningProviders()
}

func (bcp *ConfigurationProvider) GetListeningProvider(id string) listen.ListeningConfigProvider {
	return nil
}

func (bcp *ConfigurationProvider) GetName() string {
	return "Basic"
}

func (bcp *ConfigurationProvider) GetRouteProviders(listenerId string, event *events.Event) []router.RouteConfigProvider {
	keepPredicate := func(_ int, val router.RouteConfigProvider) bool {
		return val.GetListenerId() == listenerId
	}
	return util.FilterSlice(bcp.routeProviders, keepPredicate)
}

func (bcp *ConfigurationProvider) getRoutes() []string {
	return maps.Keys(bcp.Viper.GetStringMap("routes"))
}

func (bcp *ConfigurationProvider) initRouteProviders() {
	routes := bcp.getRoutes()
	if len(routes) == 0 {
		log.Warn().Msg("no routes were found or not fully configured")
	}

	for _, routeName := range routes {
		v := bcp.Config.Viper.Sub("routes").Sub(routeName)
		routeProvider := route.NewBasicRouteConfigurationProvider(v, routeName)
		bcp.routeProviders = append(bcp.routeProviders, routeProvider)
	}
}
