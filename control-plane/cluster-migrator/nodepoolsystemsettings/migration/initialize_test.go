package migration

import (
	ginkgoTesting "testing"

	migrator_config "github.com/run-ai/runai/control-plane/cluster-migrator/config"
	"github.com/run-ai/runai/control-plane/cluster-migrator/database"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApi(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	migrator_config.Init()
	conf := config.Get[*migrator_config.ConfigData]()
	dbConf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           conf.SchemaTarget,
		MigrationDirPath: "file://../../../cluster-service/pkg/database/migrations",
	})
	require.NoError(t, err)
	t.Cleanup(cleanup)
	dbConf.UseServiceSession = ptr.To(false)

	runaidb.SetLogger(logging.Actual())
	require.NoError(t, runaidb.Init(dbConf))
	t.Cleanup(func() { _ = runaidb.Close() })

	conf.ShouldMarkCompleted = true
	_, _, _, err = database.Init(conf)
	require.NoError(t, err)
	database.SetupTables(conf, true)

	RunSpecs(t, "Migration tests")
}
