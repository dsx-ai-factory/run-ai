/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/app"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/listeners"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/basic/sink"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/history"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/listen"
	router "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/route"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

type Config struct {
	appConfig     *app.AppConfig
	historyConfig *history.HistoryConfig
	routerConfig  *router.RouterConfig
	listeners     map[string]listen.ListeningConfigProvider
	sinkProvider  *sink.ConfigProvider
	Viper         *viper.Viper
}

func NewConfig(configPath string) *Config {
	config := &Config{
		Viper:     initViperFromConfig(configPath),
		listeners: map[string]listen.ListeningConfigProvider{},
	}

	config.initAppConfig()
	if err := config.initHistoryConfig(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize history config")
	}
	config.initRouterConfig()
	config.initListenProviders()
	config.initSinkConfigProvider()
	return config
}

func initViperFromConfig(configPath string) *viper.Viper {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		log.Panic().Err(err).Msg("failed to read config file")
	}
	return v
}

func (slc *Config) GetAppConfig() *app.AppConfig {
	return slc.appConfig
}

func (slc *Config) GetHistoryConfig() *history.HistoryConfig {
	return slc.historyConfig
}

func (slc *Config) GetRouterConfig() *router.RouterConfig {
	return slc.routerConfig
}

func (slc *Config) GetListeningProviders() map[string]listen.ListeningConfigProvider {
	return slc.listeners
}

func (slc *Config) GetSinkProvider(_ *events.Event, id ...string) (router.SinkConfigProvider, error) {
	return sink.ExtractSinkConfigProvider(slc.sinkProvider, id...), nil
}

func (slc *Config) SinkProvider() *sink.ConfigProvider {
	return slc.sinkProvider
}

func (slc *Config) initAppConfig() {
	slc.Viper.SetDefault("port", 5000)
	v := slc.Viper.Sub("app")
	slc.appConfig = &app.AppConfig{}
	// in case app returns an empty subtree, v is nil so no app section was defined
	if v == nil {
		v = viper.New()
	}

	v.SetDefault("log_level", "info")
	v.SetDefault("api_host", "0.0.0.0")
	v.SetDefault("api_port", "8080")

	v.Unmarshal(slc.appConfig)
	log.Info().Msgf("app config: %+v", slc.appConfig)
}

// TODO: Why is it done here and not in the routerConfig file?
func (slc *Config) initRouterConfig() {
	v := slc.Viper.Sub("router")
	slc.routerConfig = &router.RouterConfig{}
	// in case app returns an empty subtree, v is nil so no router section was defined
	if v == nil {
		v = viper.New()
	}

	v.SetDefault("concurrent_limit", 100)
	v.SetDefault("cached_events", 1000)
	v.Unmarshal(slc.routerConfig)
}

func (slc *Config) initListenProviders() {
	listenerNames := maps.Keys(slc.Viper.GetStringMap("listeners"))
	for _, name := range listenerNames {
		v := slc.Viper.Sub("listeners").Sub(name)
		slc.listeners[name] = listeners.NewListeningConfigProvider(v.GetString("type"), v)
	}
}

func (slc *Config) getSinkIds() []string {
	return maps.Keys(slc.Viper.GetStringMap("sinks"))
}

func (slc *Config) initSinkConfigProvider() {
	sinkIds := slc.getSinkIds()
	if len(sinkIds) == 0 {
		log.Warn().Msg("no sink found")
	}

	sinksViper := viper.New()
	for _, id := range sinkIds {
		sub := slc.enrichHistorySink(id)
		sinksViper.Set(id, sub.AllSettings())
	}

	slc.sinkProvider = sink.NewSinkConfigProvider(sinksViper, sinkIds...)
}

func (slc *Config) initHistoryConfig() error {
	v := slc.Viper.Sub("history")
	if v == nil {
		v = viper.New()
	}
	history.Init(v)
	slc.historyConfig = &history.HistoryConfig{}
	v.Unmarshal(slc.historyConfig)

	validate := validator.New()
	return validate.Struct(slc.historyConfig)
}

func (slc *Config) enrichHistorySink(id string) *viper.Viper {
	sub := slc.Viper.Sub(fmt.Sprintf("sinks.%s", id))
	if sub == nil {
		sub = viper.New()
	}
	if sub.GetString("type") != sinks.HistorySinkType {
		return sub
	}
	sub.Set("history_enabled", slc.historyConfig.Enabled)
	return sub
}
