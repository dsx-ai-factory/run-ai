package postgresdao_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

func TestPostgresDao(t *testing.T) {
	RegisterFailHandler(Fail)

	conf, cleanup, err := integdb.SetupPostgresDB(&integdb.IntegdbConfig{
		Schema:           "tenants_manager",
		MigrationDirPath: "file://../migrations",
	})
	if err != nil {
		t.Fatalf("failed to set up test postgres: %v", err)
	}
	t.Cleanup(cleanup)

	if err := runaidb.Init(conf); err != nil {
		t.Fatalf("failed to init database: %v", err)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	RunSpecs(t, "postgres dao tests")
}
