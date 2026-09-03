/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	dbmetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/db"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing"
	oteldb "github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing/db"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ConnectionType string

const (
	CTAdmin  ConnectionType = "admin"
	CTWriter ConnectionType = "writer"
	CTReader ConnectionType = "reader"
)

type ConnectionFactoryer interface {
	Get(ct ConnectionType) *gorm.DB
}

type ConnectionFactory struct {
}

func (cf *ConnectionFactory) Get(ct ConnectionType, config *DataBaseConfig) *gorm.DB {
	switch ct {
	case CTAdmin:
		return OpenAdmin(config)
	case CTReader:
		return OpenReader(config)
	case CTWriter:
		return OpenWriter(config)
	}
	return nil
}

func OpenWriter(config *DataBaseConfig) *gorm.DB {
	user := getUserFromConfig(config)

	logging.Infof("connecting to postgres  db: %v, schema: %v , hostname: %s port: %s, user: %s", conf.Name, conf.Schema, conf.Host, conf.Port, user)

	dsn := createDsnNew(config.Name, user, config.Password, config.Schema, config.Host, config.Port, config.SslMode, config.StatementTimeout)

	gormConfig := getGormConfig()

	db := openDbConnection(dsn, gormConfig)

	return db
}

func OpenReader(config *DataBaseConfig) *gorm.DB {
	if conf.ReadHost == "" || conf.ReadPort == "" {
		return nil
	}

	user := getUserFromConfig(config)

	logging.Infof("connecting to postgres  db: %v, schema: %v , hostname: %s port: %s, user: %s", conf.Name, conf.Schema, conf.ReadHost, conf.ReadPort, user)

	dsn := createDsnNew(config.Name, user, config.Password, config.Schema, config.ReadHost, config.ReadPort, config.SslMode, config.StatementTimeout)

	gormConfig := getGormConfig()

	db := openDbConnection(dsn, gormConfig)

	return db
}

func OpenAdmin(config *DataBaseConfig) *gorm.DB {
	logging.Infof("connecting to postgres  db: %v, schema: %v , hostname: %s port: %s, user: %s", conf.Name, conf.Schema, conf.Host, conf.Port, conf.User)

	dsn := createDsnNew(config.Name, config.User, config.Password, config.Schema, config.Host, config.Port, config.SslMode, config.StatementTimeout)

	gormConfig := getGormConfig()

	db := openDbConnection(dsn, gormConfig)

	return db
}

func openDbConnection(dsn url.URL, gormConfig *gorm.Config) *gorm.DB {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn.String(),
	}), gormConfig)
	if err != nil {
		logging.Fatalf("database connection failed: %v", err)
		panic(err)
	}

	otelGormPlugin, ok := oteldb.DBMiddleware(tracing.GetTracerProvider())
	if ok {
		err = db.Use(otelGormPlugin)
		if err != nil {
			logging.Errorf("failed to add otel middleware to gorm: %v", err)
		}
	}

	dbMetricsPlugin := dbmetrics.DBMiddleware(func(e error) {
		logging.Errorf("database metrics middleware error: %v", e)
	})
	err = db.Use(dbMetricsPlugin)
	if err != nil {
		logging.Errorf("failed to add db metrics middleware to gorm: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logging.Fatalf("sql connection failed: %v", err)
		panic(err)
	}
	sqlDB.SetMaxOpenConns(conf.MaxConnections)
	sqlDB.SetMaxIdleConns(conf.MaxIdleConnections)
	sqlDB.SetConnMaxIdleTime(conf.MaxIdleConnectionTimeDuration)
	return db
}

func getGormConfig() *gorm.Config {
	gormConfig := &gorm.Config{
		SkipDefaultTransaction: conf.SkipDefaultTransaction,
	}
	gormConfig.Logger = logger.Default.LogMode(logger.Silent)
	if conf.FullSaveAssociations {
		gormConfig.FullSaveAssociations = true
	}
	return gormConfig
}

func createDsnNew(dbName, user, password, schema, host, port, sslMode, statementTimeout string) url.URL {
	flags := map[string][]string{}
	flags["timeZone"] = []string{"UTC"}
	flags["sslmode"] = []string{sslMode}

	if schema != "" {
		flags["search_path"] = []string{schema}
	}
	if statementTimeout != "" {
		flags["statement_timeout"] = []string{statementTimeout}
	}

	dsn := url.URL{
		User:     url.UserPassword(user, password),
		Scheme:   PostgresDriverName,
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     dbName,
		RawQuery: url.Values(flags).Encode(),
	}
	return dsn
}

func getUserFromConfig(config *DataBaseConfig) string {
	if config.ServiceName != "" && useServiceSession {
		return strings.ReplaceAll(config.ServiceName, "-", "_")
	}
	return config.User
}
