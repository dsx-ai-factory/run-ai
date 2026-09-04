/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pgsink

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	postgres "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres"
	configprovider "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/configprovider"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

var sharedDbs sync.Map

type EventDao struct {
	db    *gorm.DB
	table string
}

func NewEventDao(dbConfig *postgres.DataBaseConfig, table string) (*EventDao, error) {
	db, err := getSharedDB(dbConfig)
	if err != nil {
		return nil, err
	}

	tableName := fmt.Sprintf("%s.%s", pq.QuoteIdentifier(dbConfig.Schema), pq.QuoteIdentifier(table))
	return &EventDao{db: db, table: tableName}, nil
}

func NewDefaultEventDao() (*EventDao, error) {
	db := configprovider.GetDB()
	if db == nil {
		return nil, fmt.Errorf("default event dao is not initialized")
	}

	tableName := fmt.Sprintf("%s.%s", pq.QuoteIdentifier(configprovider.GetSchema()), pq.QuoteIdentifier(models.EventsTable))
	return &EventDao{db: db, table: tableName}, nil
}

func (ed *EventDao) Store(event events.Event) error {
	tenantId, ok := event.Properties["tenant_id"].(string)
	if !ok {
		log.Warn().Msgf("skipping postgres sink store: missing tenant_id for event %s", event.UID)
		return nil
	}
	userId, ok := event.Properties["user_id"].(string)
	if !ok {
		log.Warn().Msgf("skipping postgres sink store: missing user_id for event %s", event.UID)
		return nil
	}
	description, ok := event.Properties["description"].(string)
	if !ok {
		log.Warn().Msgf("skipping postgres sink store: missing description for event %s", event.UID)
		return nil
	}

	return ed.db.Table(ed.table).Create(&models.Event{
		Id:          string(event.UID),
		TenantId:    tenantId,
		UserId:      userId,
		ListenerId:  event.ListenerId,
		Description: description,
		Properties:  event.Properties,
	}).Error
}

func (ed *EventDao) IsHealthy() (bool, error) {
	res := ed.db.Exec(fmt.Sprintf("select count(*) from %s", ed.table))
	return res.Error == nil, res.Error
}

func getSharedDB(conf *postgres.DataBaseConfig) (*gorm.DB, error) {
	dsn := configprovider.CreateDsn(conf)
	if db, ok := sharedDbs.Load(dsn.String()); ok {
		return db.(*gorm.DB), nil
	}

	db, err := configprovider.Open(conf)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.ConnectionMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnectionMaxIdleTime) * time.Minute)

	sharedDbs.Store(dsn.String(), db)
	return db, nil
}
