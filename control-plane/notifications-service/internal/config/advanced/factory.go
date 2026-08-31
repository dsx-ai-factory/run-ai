/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package advanced

import (
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/app"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/basic/route"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/config/service"
	history_service "github.com/run-ai/runai/control-plane/notifications-service/internal/history"
	history_api "github.com/run-ai/runai/control-plane/notifications-service/internal/history/api"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/listen"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/postgres/history"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks"
	"github.com/spf13/viper"
)

type Factory struct {
	serviceConfig       *service.Config
	dao                 *configprovider.Dao
	secretFieldsHandler encryption.SecretFieldHandlerer
	updatesChan         chan string
	eventsRepository    history_service.EventsRepository
}

func NewFactory(configPath string) *Factory {
	serviceConfig := service.NewConfig(configPath)
	dbConfig := initDBConfig(serviceConfig.Viper)
	ed := encryption.NewEncrypterDecrypter([]byte(dbConfig.EncryptDecryptKey))
	secretFieldsHandler := encryption.NewSecretFieldHandler(ed, route.BasicRouteConfiguration{}, listen.ListeningConfig{}, sinks.EmailSinkConfig{}, sinks.SlackSinkConfig{}, sinks.PostgresSinkConfig{})
	dao := configprovider.NewDao(dbConfig, secretFieldsHandler, ed)
	historyRepository, err := history.NewEventsRepository()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create history repository")
	}

	return &Factory{
		serviceConfig:       serviceConfig,
		secretFieldsHandler: secretFieldsHandler,
		dao:                 dao,
		updatesChan:         make(chan string, DefaultUpdateEventsChannelSize),
		eventsRepository:    historyRepository,
	}
}

func (f *Factory) CreateConfigurationProvider() app.ConfigurationProvider {
	provider := NewConfigurationProvider(f.serviceConfig, f.dao)
	createListeners(provider)
	createSinks(provider)
	return provider
}

func (f *Factory) CreateListener() app.ConfigUpdatesListener {
	return NewUpdatesListener(f.updatesChan)
}

func (f *Factory) CreateRoutesRegisters() []app.RoutesRegister {
	registers := []app.RoutesRegister{NewRoutesRegister(f.dao, f.secretFieldsHandler, f.updatesChan)}
	if !f.serviceConfig.GetHistoryConfig().Enabled {
		return registers
	}

	return append(registers, history_api.NewHistoryRoutesRegister(f.eventsRepository))
}

func (f *Factory) CreateCleaners() []app.Cleaner {
	if !f.serviceConfig.GetHistoryConfig().Enabled {
		return []app.Cleaner{}
	}

	return []app.Cleaner{history_service.NewEventsCleaner(f.eventsRepository, f.serviceConfig.GetHistoryConfig().CleanupConfig)}
}

func initDBConfig(vip *viper.Viper) *postgres.DataBaseConfig {
	v := vip.Sub("database")
	dbConf := &postgres.DataBaseConfig{}
	if v == nil {
		v = viper.New()
	}

	v.SetDefault("host", "localhost")
	v.SetDefault("port", "5432")
	v.SetDefault("database", "runai")
	v.SetDefault("schema", "notifications_service")
	v.SetDefault("user", "runai")
	v.SetDefault("password", "admin")
	v.SetDefault("ssl_mode", "disable")
	v.SetDefault("encrypt_decrypt_key", "$__file{/etc/notifications-service/secret/ENCRYPT_DECRYPT_KEY}")
	v.SetDefault("max_open_connections", 25)
	v.SetDefault("max_idle_connections", 5)
	v.SetDefault("connection_max_lifetime_minutes", 30)
	v.SetDefault("connection_max_idle_time_minutes", 5)

	err := v.Unmarshal(dbConf, config.CustomViperDecoderHook())
	if err != nil {
		log.Panic().Err(err).Msgf("failed to unmarshal database configuration")
	}

	return dbConf
}
