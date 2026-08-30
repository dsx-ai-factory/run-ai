package test

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/logger"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClusterSync(t *testing.T) {
	logger.InitLogger()
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	MockEnv(t)
	RunSpecs(t, "cluster-sync service Test Suite")
}
