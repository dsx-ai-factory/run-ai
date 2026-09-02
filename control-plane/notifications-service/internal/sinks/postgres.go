/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	postgres "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/pgsink"
	"github.com/spf13/viper"
)

const PostgresSinkType = "postgres"

type EventDaoer interface {
	Store(event events.Event) error
	IsHealthy() (bool, error)
}

type PostgresSinkConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Schema   string `mapstructure:"schema"`
	Table    string `mapstructure:"table"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password" secret:"true"`
	SslMode  string `mapstructure:"ssl_mode"`

	MaxOpenConns          int `mapstructure:"max_open_connections"`
	MaxIdleConns          int `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime int `mapstructure:"connection_max_lifetime_minutes"`
	ConnectionMaxIdleTime int `mapstructure:"connection_max_idle_time_minutes"`
}

func SetPostgresSinkConfigDefaults(viper *viper.Viper) {
	viper.SetDefault("max_open_connections", 25)
	viper.SetDefault("max_idle_connections", 5)
	viper.SetDefault("connection_max_lifetime_minutes", 30)
	viper.SetDefault("connection_max_idle_time_minutes", 5)
}

func NewPostgresSinkConfig(viper *viper.Viper, id string) (*PostgresSinkConfig, error) {
	SetPostgresSinkConfigDefaults(viper)
	psc := &PostgresSinkConfig{}
	if err := viper.Unmarshal(psc, config.CustomViperDecoderHook()); err != nil {
		log.Err(err).Msgf("failed to unmarshal configuration of postgres sink %q", id)
		return nil, err
	}
	return psc, nil
}

type PostgresSink struct {
	typ      string
	sinkId   string
	eventDao EventDaoer
}

func NewPostgresSink(sinkId string, sc *PostgresSinkConfig) *PostgresSink {
	eventDao, err := pgsink.NewEventDao(&postgres.DataBaseConfig{
		Host:                  sc.Host,
		Port:                  sc.Port,
		Database:              sc.Database,
		Schema:                sc.Schema,
		User:                  sc.User,
		Password:              sc.Password,
		SslMode:               sc.SslMode,
		MaxOpenConns:          sc.MaxOpenConns,
		MaxIdleConns:          sc.MaxIdleConns,
		ConnectionMaxLifetime: sc.ConnectionMaxLifetime,
		ConnectionMaxIdleTime: sc.ConnectionMaxIdleTime,
	}, sc.Table)
	if err != nil {
		return nil
	}

	return &PostgresSink{typ: PostgresSinkType, sinkId: sinkId, eventDao: eventDao}
}

func (ps *PostgresSink) Type() string {
	return ps.typ
}

func (ps *PostgresSink) Verify() error {
	ok, err := ps.eventDao.IsHealthy()
	if err != nil {
		return fmt.Errorf("%s sink %s failed to verify event: %w", ps.Type(), ps.sinkId, err)
	}
	if !ok {
		return fmt.Errorf("%s sink %s dao is not healthy", ps.Type(), ps.sinkId)
	}
	return nil
}

func (ps *PostgresSink) Sink(event *events.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}
	if err := ps.eventDao.Store(*event); err != nil {
		return fmt.Errorf("%s sink %s failed to sink event: %w", ps.Type(), ps.sinkId, err)
	}
	return nil
}
