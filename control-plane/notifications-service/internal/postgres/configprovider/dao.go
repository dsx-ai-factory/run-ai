/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configprovider

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config/advanced/http_listener/encryption"
	postgres "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/postgres/models"
	dbmetrics "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/db"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"gorm.io/datatypes"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ValidationText = "encryption key is valid"

	once              sync.Once
	postgresDaoDB     *gorm.DB
	postgresDaoSchema string
)

type Dao struct {
	db *gorm.DB
	*listenerDao
	*tenantDao
	*userDao
	*sinksDao
	*notificationDao
	secretDataHandler encryption.SecretFieldHandlerer
}

func NewDao(dbConfig *postgres.DataBaseConfig, secretDataHandler encryption.SecretFieldHandlerer, ed *encryption.EncrypterDecrypter) *Dao {
	db := initDatabase(dbConfig)
	dao := &Dao{
		db:                db,
		listenerDao:       newListenerDao(db, dbConfig.Schema),
		tenantDao:         newTenantDao(db, dbConfig.Schema),
		userDao:           newUserDao(db, dbConfig.Schema),
		sinksDao:          newSinksDao(db, dbConfig.Schema),
		notificationDao:   newNotificationDao(db, dbConfig.Schema),
		secretDataHandler: secretDataHandler,
	}

	if err := dao.init(&postgresValidationDao{db: dao.db, schema: dbConfig.Schema}, ed); err != nil {
		log.Err(err).Msgf("failed to validate encrypt decrypt key")
	}
	return dao
}

func (dao *Dao) init(pvd ValidationDao, ed encryption.EncrypterDecrypterer) error {
	rec, err := insertValidationIfNotExist(pvd, ed)
	if err != nil {
		return err
	}

	return validateEncryptDecryptKeyNotChanged(rec, ed)
}

func (dao *Dao) FetchListeners() ([]*models.ListenersConfig, error) {
	listenerRows, err := dao.listenerDao.FetchListeners()
	if err != nil {
		return nil, err
	}

	err = decryptConfigurations(dao.secretDataHandler, listenerRows...)
	if err != nil {
		return nil, err
	}

	return listenerRows, nil
}

func (dao *Dao) FetchListenersByPort(port int) ([]*models.ListenersConfig, error) {
	listenerRows, err := dao.listenerDao.FetchListenersByPort(port)
	if err != nil {
		return listenerRows, err
	}

	err = decryptConfigurations(dao.secretDataHandler, listenerRows...)
	return listenerRows, err
}

func (dao *Dao) FetchListener(id string) (*models.ListenersConfig, error) {
	listener, err := dao.listenerDao.FetchListener(id)
	if err != nil || listener == nil {
		return nil, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, listener); err != nil {
		return nil, err
	}

	return listener, nil
}

func (dao *Dao) FetchSink(sinkId string) (*models.SinkConfig, error) {
	sink, err := dao.sinksDao.FetchSink(sinkId)
	if err != nil || sink == nil {
		return nil, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, sink); err != nil {
		return nil, err
	}

	return sink, nil
}

func (dao *Dao) FetchSinks(sinkId ...string) ([]*models.SinkConfig, error) {
	sinks, err := dao.sinksDao.FetchSinks(sinkId...)
	if err != nil {
		return sinks, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, sinks...); err != nil {
		return nil, err
	}

	return sinks, nil
}

func (dao *Dao) FetchAllSinks() ([]*models.SinkConfig, error) {
	sinks, err := dao.sinksDao.FetchAllSinks()
	if err != nil {
		return sinks, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, sinks...); err != nil {
		return nil, err
	}

	return sinks, nil
}

func (dao *Dao) CreateSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(sink)
	if err != nil {
		return fmt.Errorf("failed encrypting sink config on DB create: %w", err)
	}

	return dao.sinksDao.CreateSink(sinkId, sinkType, encrypted)
}

func (dao *Dao) UpdateSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(sink)
	if err != nil {
		return fmt.Errorf("failed encrypting sink config on DB update: %w", err)
	}
	return dao.sinksDao.UpdateSink(sinkId, sinkType, encrypted)
}

func (dao *Dao) UpsertSink(sinkId string, sinkType string, sink datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(sink)
	if err != nil {
		return fmt.Errorf("failed encrypting sink config on DB upsert: %w", err)
	}
	return dao.sinksDao.UpsertSink(sinkId, sinkType, encrypted)
}

func (dao *Dao) DeleteSink(sinkId string) error {
	return dao.sinksDao.DeleteSink(sinkId)
}

func (dao *Dao) FetchRoutes(listenerId string, tenantId string) ([]*models.RouteGlobalConfig, error) {
	routes, err := dao.tenantDao.FetchRoutes(listenerId, tenantId)
	if err != nil {
		return routes, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, routes...); err != nil {
		return nil, err
	}

	return routes, nil
}

func (dao *Dao) FetchRoute(listenerId string, tenantId string, routeId string, defaultRoute ...*models.RouteGlobalConfig) (*models.RouteGlobalConfig, error) {
	route, err := dao.tenantDao.FetchRoute(listenerId, tenantId, routeId, defaultRoute...)
	if err != nil || route == nil {
		return nil, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, route); err != nil {
		return nil, err
	}

	return route, nil
}

func (dao *Dao) CreateRoute(tenantId string, routeId string, listenerId string, route datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(route)
	if err != nil {
		return fmt.Errorf("failed encrypting route secrets: %w", err)
	}

	return dao.tenantDao.CreateRoute(tenantId, routeId, listenerId, encrypted)
}

func (dao *Dao) UpdateRoute(tenantId string, routeId string, listenerId string, route datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(route)
	if err != nil {
		return fmt.Errorf("failed encrypting route secrets: %w", err)
	}

	return dao.tenantDao.UpdateRoute(tenantId, routeId, listenerId, encrypted)
}

func (dao *Dao) FetchUserRoutes(listenerId string, tenantId string, userId string) ([]*models.RouteGlobalConfig, error) {
	routes, err := dao.userDao.FetchUserRoutes(listenerId, tenantId, userId)
	if err != nil {
		return routes, err
	}

	if err = decryptConfigurations(dao.secretDataHandler, routes...); err != nil {
		return nil, err
	}

	return routes, nil
}

func (dao *Dao) FetchUserRoute(listenerId string, tenantId string, userId string, routeId string) (*models.RouteGlobalConfig, error) {
	route, err := dao.userDao.FetchUserRoute(listenerId, tenantId, userId, routeId)
	if err != nil {
		return nil, err
	}

	if route == nil {
		return nil, nil
	}

	if err = decryptConfigurations(dao.secretDataHandler, route); err != nil {
		return nil, err
	}

	return route, nil
}

func (dao *Dao) UpdateUserRoute(tenantId string, userId string, routeId string, listenerId string, route datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(route)
	if err != nil {
		return err
	}

	return dao.userDao.UpdateUserRoute(tenantId, userId, routeId, listenerId, encrypted)
}

func (dao *Dao) DeleteUserRoute(tenantId string, userId string, routeId string) error {
	return dao.userDao.DeleteUserRoute(tenantId, userId, routeId)
}

func (dao *Dao) DeleteAllTenantsUserRoutes(tenantId string) error {
	return dao.userDao.DeleteAllTenantsUserRoutes(tenantId)
}

func (dao *Dao) DeleteNotifyIdFromAllUserRoutes(tenantId string, notifyId string) error {
	return dao.userDao.DeleteNotifyIdFromAllUserRoutes(tenantId, notifyId)
}

func (dao *Dao) CreateListener(listenerId string, listenerType string, config datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(config)
	if err != nil {
		return err
	}

	return dao.listenerDao.CreateListener(listenerId, listenerType, encrypted)
}

func (dao *Dao) UpdateListener(listenerId string, listenerType string, config datatypes.JSON) error {
	encrypted, err := dao.secretDataHandler.EncryptSecrets(config)
	if err != nil {
		return err
	}

	return dao.listenerDao.UpdateListener(listenerId, listenerType, encrypted)
}

func (dao *Dao) IsHealthy() (bool, error) {
	res, err := dao.tenantDao.IsHealthy()
	if err != nil {
		return res, err
	}

	res, err = dao.userDao.IsHealthy()
	if err != nil {
		return res, err
	}

	return dao.sinksDao.IsHealthy()
}

func GetSchema() string {
	return postgresDaoSchema
}

func GetDB() *gorm.DB {
	return postgresDaoDB
}

func decryptConfigurations[T models.ConfigGetterSetter](secretDataHandler encryption.SecretFieldHandlerer, items ...T) error {
	for _, item := range items {
		config, err := secretDataHandler.DecryptSecrets(item.GetConfig())
		if err != nil {
			return err
		}

		item.SetConfig(config)
	}
	return nil
}

func filterRoutesByListenerId(listenerId string, routes []*models.RouteGlobalConfig) []*models.RouteGlobalConfig {
	if listenerId == "" {
		return routes
	}

	finalRoutes := []*models.RouteGlobalConfig{}
	for _, route := range routes {
		if match(listenerId, route.ListenerId) {
			finalRoutes = append(finalRoutes, route)
		}
	}

	return finalRoutes
}

func match(listenerId string, listenerIdPattern string) bool {
	m, err := filepath.Match(listenerIdPattern, listenerId)
	if err != nil {
		log.Err(err).Msgf("error matching listener pattern: %s with listener id %s", listenerIdPattern, listenerId)
		return false
	}

	return m
}

func initDatabase(conf *postgres.DataBaseConfig) *gorm.DB {
	once.Do(func() {
		db, err := Open(conf)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to connect to DB")
		}
		migrate(db, conf.Schema)
		postgresDaoDB = db
		postgresDaoSchema = conf.Schema
	})
	return postgresDaoDB
}

func insertValidationIfNotExist(pvd ValidationDao, ed encryption.EncrypterDecrypterer) (*models.Validation, error) {
	encrypted, err := ed.Encrypt(ValidationText)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt validation text: %w", err)
	}

	return pvd.InsertValidationIfNotExist(encrypted)
}

func validateEncryptDecryptKeyNotChanged(rec *models.Validation, ed encryption.EncrypterDecrypterer) error {
	decrypted, err := ed.Decrypt(rec.Data)
	if err != nil {
		return fmt.Errorf("failed to decrypt validation text: %w", err)
	}

	if decrypted != ValidationText {
		return fmt.Errorf("encrypt decrypt key changed, secret data will not decrypted right\ndecrypted: %s", decrypted)
	}

	return nil
}

func Open(conf *postgres.DataBaseConfig) (*gorm.DB, error) {
	dsn := CreateDsn(conf)
	log.Info().Msgf("connecting to postgres  dbname: %s, dbschema: %s, dbhost: %s dbport %s user %s, sslMode: %s", conf.Database, conf.Schema, conf.Host, conf.Port, conf.User, conf.SslMode)
	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{
		DSN:                  dsn.String(),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	dbMetricsPlugin := dbmetrics.DBMiddleware(func(e error) {
		log.Error().Msgf("database metrics middleware error: %v", e)
	})

	if err := db.Use(dbMetricsPlugin); err != nil {
		log.Error().Msgf("failed to add db metrics middleware to gorm: %v", err)
		return nil, err
	}

	return db, nil
}

func CreateDsn(database *postgres.DataBaseConfig) url.URL {
	flags := map[string][]string{}
	flags["timeZone"] = []string{"UTC"}
	flags["sslmode"] = []string{database.SslMode}
	dsn := url.URL{
		User:     url.UserPassword(database.User, database.Password),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", database.Host, database.Port),
		Path:     database.Database,
		RawQuery: url.Values(flags).Encode(),
	}
	return dsn
}

const sinkIdUniqueIndex = "idx_sink_config_unique_sink_id"

func migrate(db *gorm.DB, schema string) {
	res := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schema)))
	if res.Error != nil {
		log.Fatal().Err(res.Error).Msgf("failed CREATE SCHEMA IF NOT EXISTS %s", pq.QuoteIdentifier(schema))
	}
	res = db.Exec(fmt.Sprintf("SET search_path TO %s", pq.QuoteIdentifier(schema)))
	if res.Error != nil {
		log.Fatal().Err(res.Error).Msgf("SET search_path TO %s", pq.QuoteIdentifier(schema))
	}

	migrateModel(db, schema, models.RouteGlobalConfig{}, models.RouteGlobalConfigTable)
	migrateModel(db, schema, models.RouteUserConfig{}, models.RouteUserConfigTable)
	migrateModel(db, schema, models.ListenersConfig{}, models.ListenersConfigsTable)
	migrateModel(db, schema, models.SinkConfig{}, models.SinkTable)
	migrateModel(db, schema, models.Notification{}, models.NotificationsTable)
	migrateModel(db, schema, models.Validation{}, models.ValidationTable)
	migrateModel(db, schema, models.Event{}, models.EventsTable)
	softCreateUniqueIndex(db, fmt.Sprintf("%s.%s", pq.QuoteIdentifier(schema), pq.QuoteIdentifier(models.SinkTable)), "sink_id", sinkIdUniqueIndex, &models.SinkConfig{})
	log.Info().Msg("migration finished successfully")
}

func migrateModel[T interface{}](serviceDb *gorm.DB, schema string, m T, table string) {
	err := serviceDb.AutoMigrate(&m)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to auto-migrate %s", table)
	}

	var exists bool
	err = serviceDb.Raw(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = ? AND table_name = ?)",
		schema, table,
	).Scan(&exists).Error
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to auto-migrate %s.%s %s", schema, table, strconv.FormatBool(serviceDb.Migrator().HasTable(&m)))
	}
	if !exists {
		log.Fatal().Msgf("failed to auto-migrate, table not exist %s.%s %s", schema, table, strconv.FormatBool(serviceDb.Migrator().HasTable(&m)))
	}
}

func softCreateUniqueIndex(db *gorm.DB, table, column, indexName string, model interface{}) {
	if db.Migrator().HasIndex(model, indexName) {
		return
	}

	createIndexSQL := fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS %s ON %s (%s) WHERE deleted_at IS NULL", pq.QuoteIdentifier(indexName), table, column)
	if err := db.Exec(createIndexSQL).Error; err != nil {
		log.Err(err).Msgf("failed creating unique index %s on %s(%s)", indexName, table, column)
		return
	}
	log.Info().Msgf("created unique index %s on %s(%s)", indexName, table, column)
}
