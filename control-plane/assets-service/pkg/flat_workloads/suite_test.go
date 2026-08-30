package flat_workloads

import (
	"testing"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
)

func TestWorkloads_v2(t *testing.T) {
	RegisterFailHandler(Fail)

	config.MockConfig()
	config.Init()

	//
	//   we run the service test with "real" database, so need to start it prior
	//   to running any services related test
	//
	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../database/migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	// tests run against a single-user DB; use the owner role (not runaidb's limited
	// service session) so runtime AutoMigrate can create tables.
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	backend_db.CreateTables(runaidb.DB().Actual())

	RunSpecs(t, "Workloads V2 Suite")
}
