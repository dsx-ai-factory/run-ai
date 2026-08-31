/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"net/url"

	service_config "github.com/run-ai/runai/control-plane/org-units-migrator/config"
	"github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	PostgresDriverName = "postgres"
)

var (
	sourceClient *gorm.DB
	globalClient *gorm.DB
	targetClient *gorm.DB
)

func Init() {
	sourceClient = OpenDbWithSchemaName(config.Get[*service_config.ConfigData]().SchemaSource)
	globalClient = OpenDbWithSchemaName("")
	targetClient = OpenDbWithSchemaName(config.Get[*service_config.ConfigData]().SchemaTarget)
	dao_utils.Init(config.Get[*dao_utils.DataBaseConfig]())
}

func OpenDbWithSchemaName(schema string) *gorm.DB {
	conf := config.Get[*dao_utils.DataBaseConfig]()
	logging.Infof("connecting to postgres; dbname: %v, dbschema: %v, dbhostname: %s", conf.Name, schema, conf.Host)
	flags := map[string][]string{}

	flags["timeZone"] = []string{"UTC"}
	flags["sslmode"] = []string{conf.SslMode}

	if schema != "" {
		flags["search_path"] = []string{schema}
	}
	dsn := url.URL{
		User:     url.UserPassword(conf.User, conf.Password),
		Scheme:   PostgresDriverName,
		Host:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Path:     conf.Name,
		RawQuery: url.Values(flags).Encode(),
	}

	gormConfig := &gorm.Config{}

	gormConfig.Logger = logger.Default.LogMode(logger.Error)

	openedDb, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn.String(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), gormConfig)
	if err != nil {
		logging.Fatalf("database connection failed: %v", err)
		panic(err)
	}
	return openedDb
}
