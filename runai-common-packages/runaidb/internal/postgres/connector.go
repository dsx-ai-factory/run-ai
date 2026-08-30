package postgres

import (
	"fmt"
	"net/url"
	"strings"

	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/gormext"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm"

	dbmetrics "github.com/run-ai/runai/runai-common-packages/metrics/db"
	"github.com/run-ai/runai/runai-common-packages/tracing"
	oteldb "github.com/run-ai/runai/runai-common-packages/tracing/db"
	"gorm.io/gorm/schema"
)

type Connector struct {
}

func NewConnector() *Connector {
	return &Connector{}
}

func (c *Connector) ConnectDB(conf *dbconfig.Config) (orm.DB, error) {
	gormDB, dialector, err := c.open(conf, c.buildDSN(&conf.Connection))
	if err != nil {
		return nil, err
	}
	return orm.NewDB(gormDB, dialector), nil
}

func (c *Connector) ConnectReadDB(conf *dbconfig.Config) (orm.ReadDB, error) {
	gormDB, dialector, err := c.open(conf, c.buildReadDSN(conf))
	if err != nil {
		return nil, err
	}
	return orm.NewReadDB(gormDB, dialector), nil
}

func (c *Connector) open(conf *dbconfig.Config, dsn url.URL) (*gorm.DB, gorm.ErrorTranslator, error) {
	gormConfig := &gorm.Config{
		Logger:                 gormext.NewLoggerAdapter(),
		TranslateError:         true,
		SkipDefaultTransaction: conf.SkipDefaultTransaction,
		FullSaveAssociations:   conf.FullSaveAssociations,
	}

	dialector := newTranslatingDialector(dsn.String())
	gormDB, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err = c.registerPlugins(gormDB); err != nil {
		return nil, nil, err
	}
	if err = c.registerSerializers(conf); err != nil {
		return nil, nil, err
	}
	if err = c.configurePool(gormDB, conf); err != nil {
		return nil, nil, err
	}

	return gormDB, dialector, nil
}

func (c *Connector) registerPlugins(db *gorm.DB) error {
	dbMetricsPlugin := dbmetrics.DBMiddleware(func(e error) {
		logger.GetLogger().Errorf("database metrics middleware error: %v", e)
	})
	if err := db.Use(dbMetricsPlugin); err != nil {
		return fmt.Errorf("failed to add db metrics plugin to db: %w", err)
	}

	dbOtelPlugin, ok := oteldb.DBMiddleware(tracing.GetTracerProvider())
	if !ok {
		return nil
	}
	if err := db.Use(dbOtelPlugin); err != nil {
		return fmt.Errorf("failed to add otel plugin to db: %w", err)
	}

	return nil
}

func (c *Connector) registerSerializers(conf *dbconfig.Config) error {
	if conf.EncryptionKey == "" {
		return nil
	}

	encSerializer, err := gormext.NewSymmetricEncryptionSerializer(conf.EncryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create symmetric encryption serializer: %w", err)
	}

	schema.RegisterSerializer("encrypted", encSerializer)
	return nil
}

func (c *Connector) configurePool(gormDB *gorm.DB, conf *dbconfig.Config) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(conf.Pool.MaxConnections)
	sqlDB.SetMaxIdleConns(conf.Pool.MaxIdleConnections)
	sqlDB.SetConnMaxIdleTime(conf.Pool.MaxIdleConnectionTimeDuration)

	return nil
}

func (c *Connector) buildDSN(conn *dbconfig.ConnectionConfig) url.URL {
	flags := initializeFlags(conn)

	return url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(conn.User, conn.Password),
		Host:     buildHostList(conn.Host, conn.Port),
		Path:     conn.Name,
		RawQuery: flags.Encode(),
	}
}

func (c *Connector) buildReadDSN(conf *dbconfig.Config) url.URL {
	flags := initializeFlags(&conf.Connection)
	if conf.ReadConnection.SslMode != "" {
		flags.Set("sslmode", string(conf.ReadConnection.SslMode))
	}
	if conf.ReadConnection.TargetSessionAttrs != "" {
		flags.Set("target_session_attrs", conf.ReadConnection.TargetSessionAttrs)
	} else {
		// don't inherit the primary's target_session_attrs (e.g. read-write) on the read path
		flags.Del("target_session_attrs")
	}
	flags.Set("default_transaction_read_only", "true")

	return url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(conf.ReadConnection.User, conf.ReadConnection.Password),
		Host:     buildHostList(conf.ReadConnection.Host, conf.ReadConnection.Port),
		Path:     conf.Connection.Name,
		RawQuery: flags.Encode(),
	}
}

func initializeFlags(conf *dbconfig.ConnectionConfig) url.Values {
	flags := url.Values{}
	flags.Set("timeZone", "UTC")
	flags.Set("sslmode", string(conf.SslMode))
	flags.Set("statement_timeout", conf.StatementTimeout)
	if conf.Schema != "" {
		flags.Set("search_path", conf.Schema)
	}
	if conf.TargetSessionAttrs != "" {
		flags.Set("target_session_attrs", conf.TargetSessionAttrs)
	}
	return flags
}

func buildHostList(host string, port int) string {
	hosts := strings.Split(host, ",")
	for i, h := range hosts {
		h = strings.TrimSpace(h)
		if !strings.Contains(h, ":") {
			h = fmt.Sprintf("%s:%d", h, port)
		}
		hosts[i] = h
	}
	return strings.Join(hosts, ",")
}
