package services_crud

import (
	ginkgoTesting "testing"

	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/mocks"

	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"

	k8sObjectsClient "github.com/run-ai/runai/control-plane/workloads/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"
)

func TestCrud(t *ginkgoTesting.T) {
	RegisterFailHandler(Fail)

	dao_utils.SetPostgresEnv("", "file://../database/migrations")
	config.Init()
	conf := config.Database()
	conf.SkipDefaultTransaction = true
	dao_utils.OpenPostgresForTesting(t, config.Database())

	cache_testing.EnableCache()

	k8sObjectsClient.SetMock(mocks.NewK8SObjectTrackerMock())

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	RunSpecs(t, "Api tests")

	t.Cleanup(func() {
		testing.CleanTables()
		cache_testing.DisableCache()
	})
}
