package integdb

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

type IntegdbConfig struct {
	User             string `mapstructure:"integdb_user"`
	Password         string `mapstructure:"integdb_password"`
	Name             string `mapstructure:"integdb_name"`
	Schema           string `mapstructure:"integdb_schema" validate:"required"`
	MigrationDirPath string `mapstructure:"integdb_migration_dir_path"`
	ImageVersion     string `mapstructure:"integdb_image_version"`
}

func (c *IntegdbConfig) SetDefaults() {
	c.User = "postgres"
	c.Password = "postgres"
	c.Name = "backend"
	c.Schema = "test"
	c.MigrationDirPath = "file:///migrations"
	c.ImageVersion = "16"
}

func SetupPostgresDB(conf *IntegdbConfig) (*dbconfig.Config, func(), error) {
	conf = resolveConfig(conf)

	ctx := context.Background()
	container, err := tcpostgres.Run(ctx,
		"postgres:"+conf.ImageVersion,
		tcpostgres.WithDatabase(conf.Name),
		tcpostgres.WithUsername(conf.User),
		tcpostgres.WithPassword(conf.Password),
		tcpostgres.BasicWaitStrategies(),
		extendPGImage(conf.ImageVersion),
		testcontainers.WithCmd("postgres", "-c", "shared_preload_libraries=pg_cron", "-c", "cron.database_name="+conf.Name),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	host, err := container.Host(ctx)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to get container port: %w", err)
	}

	dbConf := &dbconfig.Config{}
	dbConf.SetDefaults()

	dbConf.Connection.Host = host
	dbConf.Connection.Port = int(port.Num())
	dbConf.Connection.Name = conf.Name
	dbConf.Connection.Schema = conf.Schema
	dbConf.Connection.User = conf.User
	dbConf.Connection.Password = conf.Password
	dbConf.Connection.SslMode = dbconfig.SslModeDisable

	dbConf.Migration.Files = conf.MigrationDirPath
	dbConf.Migration.Auto = ptr.Bool(conf.MigrationDirPath != "")

	dbConf.ServiceName = "test-integdb"
	dbConf.EncryptionKey = generateEncryptionKey()

	return dbConf, cleanup, nil
}

func extendPGImage(version string) testcontainers.CustomizeRequestOption {
	major := majorVersion(version)
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Image = ""
		req.FromDockerfile = testcontainers.FromDockerfile{
			Context:    packageDir(),
			Dockerfile: "Dockerfile",
			BuildArgs: map[string]*string{
				"POSTGRES_VERSION":       &version,
				"POSTGRES_VERSION_MAJOR": &major,
			},
			KeepImage: true,
		}
		return nil
	}
}

func majorVersion(tag string) string {
	if i := strings.IndexAny(tag, ".-"); i >= 0 {
		return tag[:i]
	}
	return tag
}

func resolveConfig(conf *IntegdbConfig) *IntegdbConfig {
	if conf == nil {
		_ = config.Register[*IntegdbConfig]()
		return config.Get[*IntegdbConfig]()
	}

	_ = config.SetWithDefaults(conf)
	return config.Get[*IntegdbConfig]()
}

func generateEncryptionKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func packageDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}
