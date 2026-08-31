/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package database

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/run-ai/runai/control-plane/policy-service/pkg/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	DefaultPostgresPort = "5432"
	PostgresDriverName  = "postgres"
	PostgresFlagSslMode = "sslmode"
	PostgresSearchPath  = "search_path"
)

var once sync.Once

var db *gorm.DB

// DB initializes (once) a database object base on env variables
// and returns a pointer to the DB object
func DB() *gorm.DB {
	once.Do(OpenDbByConfig)
	return db
}

func OpenDbByConfig() {
	if db != nil {
		return
	}

	dsn := createDsn(config.Database().Schema)

	var err error

	var gormLogger logger.Interface

	gormConfig := &gorm.Config{
		Logger: gormLogger,
	}

	gormConfig.Logger = logger.Default.LogMode(logger.Error)

	db, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn.String(),
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), gormConfig)
	if err != nil {
		logging.Fatalf("database connection failed: %v", err)
	}
}

// FinalizeTransaction closes the transaction, rollback it in case of error or panic
func FinalizeTransaction(tx *gorm.DB, committed *bool) {
	if !*committed {
		tx.Rollback()
	} else if r := recover(); r != nil {
		tx.Rollback()
	}
}

func createDsn(schema string) url.URL {
	conf := config.Database()

	flags := map[string][]string{}
	flags["timeZone"] = []string{"UTC"}
	flags[PostgresFlagSslMode] = []string{conf.SslMode}

	if schema != "" {
		flags[PostgresSearchPath] = []string{schema}
	}

	dsn := url.URL{
		User:     url.UserPassword(conf.User, conf.Password),
		Scheme:   PostgresDriverName,
		Host:     fmt.Sprintf("%s:%s", conf.Host, conf.Port),
		Path:     conf.Name,
		RawQuery: url.Values(flags).Encode(),
	}
	return dsn
}
