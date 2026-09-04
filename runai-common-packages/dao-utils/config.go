/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	EnvDbUser                        = "DB_USER"
	EnvServiceName                   = "SERVICE_NAME"
	EnvDbPassword                    = "DB_PASSWORD"
	EnvDbName                        = "DB_NAME"
	EnvDbHost                        = "DB_HOST"
	EnvDbPort                        = "DB_PORT"
	EnvDbReadHost                    = "DB_READ_HOST"
	EnvDbReadPort                    = "DB_READ_PORT"
	EnvDbSchema                      = "DB_SCHEMA"
	EnvDbSslMode                     = "DB_SSL_MODE"
	EnvStatementTimeout              = "STATEMENT_TIMEOUT"
	EnvMaxConnections                = "MAX_CONNECTIONS"
	EnvMaxIdleConnectionTimeDuration = "MAX_IDLE_CONNECTIONS_TIME_DURATION"
	EnvSkipDefaultTransaction        = "SKIP_DEFAULT_TRANSACTION"
	EnvFullSaveAssociations          = "FULL_SAVE_ASSOCIATIONS"
	EnvUseServiceSession             = "DB_USE_SERVICE_SESSION"

	EnvMigrationFiles       = "MIGRATION_FILES"
	EnvMigrationCommand     = "MIGRATION_COMMAND"
	EnvMigrationLockRetries = "MIGRATION_LOCK_RETRIES"
	EnvSecretsMountPath     = "SECRETS_MOUNT_PATH"

	// Db Transformation
	EnvDbTransformationId = "DB_TRANSFORMATION_ID"
	EnvDbSchemaSource     = "DB_SCHEMA_SOURCE"
	EnvDbSchemaTarget     = "DB_SCHEMA_TARGET"
)

const (
	SslModeDisable    = "disable"
	SslModeAllow      = "allow"
	SslModePrefer     = "prefer"
	SslModeRequire    = "require"
	SslModeVerifyCA   = "verify-ca"
	SslModeVerifyFull = "verify-full"
)

type SslMode string

var SslModes = []SslMode{
	SslModeDisable,
	SslModeAllow,
	SslModePrefer,
	SslModeRequire,
	SslModeVerifyCA,
	SslModeVerifyFull,
}

type DataBaseConfig struct {
	MigrationFiles                string        `mapstructure:"migration_files"`
	MigrationCommand              string        `mapstructure:"migration_command"`
	Host                          string        `mapstructure:"db_host"`
	Port                          string        `mapstructure:"db_port"`
	ReadHost                      string        `mapstructure:"db_read_host"`
	ReadPort                      string        `mapstructure:"db_read_port"`
	Name                          string        `mapstructure:"db_name"`
	Schema                        string        `mapstructure:"db_schema"`
	ServiceName                   string        `mapstructure:"service_name"`
	User                          string        `mapstructure:"db_user" config:"secret"`
	Password                      string        `mapstructure:"db_password" config:"secret"`
	MaxConnections                int           `mapstructure:"max_connections"`
	MaxIdleConnections            int           `mapstructure:"max_idle_connections"`
	MaxIdleConnectionTimeDuration time.Duration `mapstructure:"max_idle_connections_time_duration"`
	MigrationLockRetries          int           `mapstructure:"migration_lock_retries"`
	SslMode                       string        `mapstructure:"db_ssl_mode"`
	StatementTimeout              string        `mapstructure:"statement_timeout"` // StatementTimeout ref https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-STATEMENT-TIMEOUT
	SkipDefaultTransaction        bool          `mapstructure:"skip_default_transaction"`
	FullSaveAssociations          bool          `mapstructure:"full_save_associations"`
	UseServiceSession             bool          `mapstructure:"db_use_service_session"`
	VerboseLogging                bool          `mapstructure:"db_verbose_logging"`
}

func (dbc *DataBaseConfig) SetDefaults() {
	dbc.MigrationFiles = "file:///migrations"
	dbc.Port = "5432"
	dbc.Name = "backend"
	dbc.MaxConnections = 20
	dbc.MaxIdleConnections = 2
	dbc.MaxIdleConnectionTimeDuration = time.Hour
	dbc.MigrationLockRetries = DefaultLockRetires
	dbc.SslMode = SslModeDisable
	dbc.StatementTimeout = "2min"
	dbc.UseServiceSession = true
	dbc.VerboseLogging = false
}

func ConstructDebugConfig(defaultSchema string) *DataBaseConfig {
	return &DataBaseConfig{
		Host:     "localhost",
		Port:     "5432",
		ReadHost: "localhost",
		ReadPort: "5432",
		Name:     "backend",
		Schema:   defaultSchema,
		User:     "user",
		Password: "password",
		SslMode:  "disable",
	}
}

func ConstructConfig(defaultSchema string) *DataBaseConfig {
	err := config.SetKbSecretsFolder("")
	if err != nil {
		logging.Fatalf("failed to set secrets folder, error: %v. Service is terminating", err)
	}

	dbConfig, err := config.GetR[*DataBaseConfig]()
	if err != nil {
		logging.Fatalf("failed to register config, error: %v. Service is terminating", err)
	}

	if dbConfig.Schema != "" {
		return dbConfig
	}

	dbConfig.Schema = defaultSchema
	if err = config.Set(dbConfig); err != nil {
		logging.Fatalf("failed to set config, error: %v. Service is terminating", err)
	}
	return dbConfig
}
