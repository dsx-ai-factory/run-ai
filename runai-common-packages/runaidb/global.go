package runaidb

import (
	"errors"
	"fmt"
	"sync"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/dbinit"
	"github.com/run-ai/runai/runai-common-packages/runaidb/internal/logger"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
	Debugf(format string, args ...any)
}

type DBInitializer interface {
	InitializeDB(conf *dbconfig.Config) (orm.DB, error)
	InitializeReadDB(conf *dbconfig.Config) (orm.ReadDB, error)
}

type MigrationParams struct {
	Direction     dbconfig.MigrationDirection
	TargetVersion *string
}

var (
	primaryDB orm.DB
	readerDB  orm.ReadDB
	schemaDBs sync.Map
	once      sync.Once
	dbInit    DBInitializer = dbinit.NewInitializer()
)

func Init(conf *dbconfig.Config) error {
	var err error
	once.Do(func() {
		if err = initConfig(conf); err != nil {
			return
		}

		dbConf := config.Get[*dbconfig.Config]()
		primaryDB, err = dbInit.InitializeDB(dbConf)
		if err != nil {
			return
		}

		if dbConf.HasReadReplica() {
			readerDB, err = dbInit.InitializeReadDB(dbConf)
		}
	})
	return err
}

func DB() orm.DB {
	return primaryDB
}

func ReadDB() (orm.ReadDB, error) {
	if readerDB != nil {
		return readerDB, nil
	}
	return nil, fmt.Errorf("read DB is not initialized")
}

func NewDB(schemaName string) (orm.DB, error) {
	if db, ok := schemaDBs.Load(schemaName); ok {
		return db.(orm.DB), nil
	}

	dbConf := config.Get[*dbconfig.Config]()
	if primaryDB != nil && schemaName == dbConf.Connection.Schema {
		return DB(), nil
	}

	schemaConf := *dbConf
	schemaConf.Migration.Auto = ptr.Bool(false)
	schemaConf.Connection.Schema = schemaName

	db, err := dbInit.InitializeDB(&schemaConf)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection for schema %s: %w", schemaName, err)
	}

	schemaDBs.Store(schemaName, db)
	return db, nil
}

func Migrate(params *MigrationParams) error {
	if primaryDB == nil {
		return fmt.Errorf("DB is not initialized")
	}
	if params == nil || params.Direction == "" {
		return fmt.Errorf("migration parameters are not valid")
	}

	dbConf := config.Get[*dbconfig.Config]()
	migConf := *dbConf
	migConf.Migration.Auto = ptr.Bool(true)
	migConf.Migration.Direction = params.Direction
	migConf.Migration.TargetVersion = params.TargetVersion
	migConf.UseServiceSession = ptr.Bool(false)

	db, err := dbInit.InitializeDB(&migConf)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	_ = db.Close()
	return nil
}

func Close() error {
	once = sync.Once{}
	if primaryDB == nil {
		return fmt.Errorf("DB is not initialized")
	}

	var errs []error

	if err := primaryDB.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close primary DB: %w", err))
	}

	if readerDB != nil {
		if err := readerDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close read DB: %w", err))
		}
	}

	schemaDBs.Range(func(key, value any) bool {
		if err := value.(orm.DB).Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close schema DB %s: %w", key, err))
		}
		return true
	})

	primaryDB = nil
	readerDB = nil
	schemaDBs = sync.Map{}

	return errors.Join(errs...)
}

func SetMock(db orm.DB) {
	primaryDB = db
}

func SetLogger(l Logger) {
	logger.SetLogger(l)
}

func initConfig(cfg *dbconfig.Config) error {
	if cfg != nil {
		return config.SetWithDefaults(cfg)
	}
	return dbconfig.InitConfig()
}
