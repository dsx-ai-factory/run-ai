/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"slices"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

var sslModes = []dbconfig.SslMode{
	dbconfig.SslModeDisable,
	dbconfig.SslModeAllow,
	dbconfig.SslModePrefer,
	dbconfig.SslModeRequire,
	dbconfig.SslModeVerifyCA,
	dbconfig.SslModeVerifyFull,
}

type ConfigData struct {
	Host                  string `mapstructure:"db_host,required"`
	Port                  string `mapstructure:"db_port,required"`
	Name                  string `mapstructure:"db_name,required"`
	SchemaSourceBackend   string `mapstructure:"db_schema_source_backend,required"`
	SchemaSourceTenants   string `mapstructure:"db_schema_source_tenants,required"`
	SchemaTarget          string `mapstructure:"db_schema_target,required"`
	MigrationVersion      string `mapstructure:"db_migration_version,required"`
	User                  string `mapstructure:"db_user,required" config:"secret"`
	Password              string `mapstructure:"db_password,required" config:"secret"`
	SslMode               string `mapstructure:"db_ssl_mode,required"`
	KeycloakHost          string `mapstructure:"keycloak_host,required"`
	KeycloakAdminUser     string `mapstructure:"keycloak_admin_username,required" config:"secret"`
	KeycloakAdminPassword string `mapstructure:"keycloak_admin_password,required" config:"secret"`
}

func (c *ConfigData) SetDefaults() {
	c.Port = "5432"
	c.SchemaSourceBackend = "backend"
	c.SchemaSourceTenants = "tenants_manager"
	c.SchemaTarget = "authorization"
	c.MigrationVersion = "v1"
	c.SslMode = string(dbconfig.SslModeDisable)
	c.KeycloakHost = "http://keycloak-http:80"
}

func Init() {
	logging.Debugf("initializing config")
	config.SetKbSecretsFolder("")
	if err := config.Register[*ConfigData](); err != nil {
		logging.Fatalf("failed to register config: %v", err)
	}
	conf := config.Get[*ConfigData]()
	if !slices.Contains(sslModes, dbconfig.SslMode(conf.SslMode)) {
		conf.SslMode = string(dbconfig.SslModeDisable)
		config.Set(conf)
	}
}
