package test

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/pkg/logger"
	"go.uber.org/zap"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRuntimeRules(t *testing.T) {
	logger.InitLogger(int(zap.DebugLevel))
	zap.S().Infof("Started service tests")
	RegisterFailHandler(Fail)
	MockEnv(t)
	RunSpecs(t, "runtime-rules service Test Suite")
}
