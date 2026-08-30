package services_utils

import (
	"context"
	"net/http"

	uuid2 "github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/control-plane/workloads/pkg/testing_crud"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("utils testing", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	var uuid uuid2.UUID

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})
		uuid, _ = uuid2.Parse(randoms.ClusterId)

		authClient := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authClient.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authClient.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionForOwnershipProtection, convutils.Ref(true))))

		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	Context("CheckWorkloadOwnershipProtection", func() {
		It("same user", func() {
			err := CheckWorkloadOwnershipProtection(ctx, uuid.String(), randoms.DepartmentId, testing.TestUser)
			Expect(err).To(BeNil())
		})

		It("different user - same permission scope", func() {
			err := CheckWorkloadOwnershipProtection(ctx, uuid.String(), randoms.DepartmentId, testing.AnotherUser)
			Expect(err.Code).To(Equal(int32(http.StatusForbidden)))
		})

		It("different user - with permissions on a higher scope", func() {
			ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

			err := CheckWorkloadOwnershipProtection(ctx, uuid.String(), randoms.DepartmentId, testing.AnotherUser)
			Expect(err).To(BeNil())
		})

		It("different user - no ownership protection", func() {
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestDefaultClusterVersion, convutils.Ref(false))))

			err := CheckWorkloadOwnershipProtection(ctx, uuid.String(), randoms.DepartmentId, testing.AnotherUser)
			Expect(err).To(BeNil())
		})
	})
})
