package resources

import (
	"context"
	"testing"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	k8sClient        client.Client
	testEnv          *envtest.Environment
	apiAccessContext context.Context
	suite            = "runai scheduler"
)

func TestResources(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, suite)
}

// Seed the package-level config singleton with runai-vocabulary values so
// production code paths under test (which now read scheduler-name / metrics
// namespace / default-nodepool-name from config.Get()) exercise the runai
// variants by default. Specs that need KAI values override per-It via
// config.SetForTest(...) + DeferCleanup (see kai_config_test.go).
var _ = BeforeSuite(func() {
	config.SetCurrent(&config.NodePoolControllerConfig{
		NodePoolNameLabel:   kaiconstants.DefaultNodePoolLabelKey,
		DefaultNodepoolName: kaiconstants.DefaultNodePoolName,
		SchedulerName:       "runai-scheduler",
		SchedulerNamespace:  "runai",
		MetricsNamespace:    "runai",
		FinalizerDomain:     "run.ai",
	})
})
