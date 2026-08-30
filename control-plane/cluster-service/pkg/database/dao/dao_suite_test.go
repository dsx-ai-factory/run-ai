package dao

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/testutils"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	defer runai_config.FlushAll()

	testutils.RegisterPostgresSuite(config.Schema, "file://../../database/migrations")

	RunSpecs(t, "DAO Suite")
}
