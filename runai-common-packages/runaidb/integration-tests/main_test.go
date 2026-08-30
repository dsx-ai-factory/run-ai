package integrationtests

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

const (
	readUser     = "read-integdb"
	readPassword = "read-password"
)

func TestMain(m *testing.M) {
	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "test_schema",
		MigrationDirPath: "file://" + getMigrationDir(),
	})
	if err != nil {
		fmt.Printf("Failed to setup postgres container: %v\n", err)
		os.Exit(1)
	}

	conf.ReadConnection.Host = conf.Connection.Host
	conf.ReadConnection.Port = conf.Connection.Port
	conf.ReadConnection.User = readUser
	conf.ReadConnection.Password = readPassword
	conf.ReadConnection.SslMode = dbconfig.SslModeDisable

	if err := runaidb.Init(conf); err != nil {
		cleanup()
		fmt.Printf("Failed to initialize runaidb: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	_ = runaidb.Close()
	cleanup()
	os.Exit(code)
}

func getMigrationDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "migrations")
}
