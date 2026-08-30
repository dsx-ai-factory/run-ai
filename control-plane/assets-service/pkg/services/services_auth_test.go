package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/runai-application-packages/convutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

// crossTenantAwareAuth returns real scopes for same-tenant, empty scopes for other tenants (per-tenant auth DB).
type crossTenantAwareAuth struct {
	allowedTenantId int32
	inner           *testing.RandomsBasedAuthMockClient
}

func (m *crossTenantAwareAuth) Client() authorization.ClientInterface { return m }

func (m *crossTenantAwareAuth) CheckPermittedScopes(ctx context.Context, resourceType authClient.ResourceType, action authClient.Action) (*authClient.PermittedScopes, *authClient.Error) {
	if auth.TenantIdOf(ctx) != m.allowedTenantId {
		return &authClient.PermittedScopes{}, nil
	}
	return m.inner.CheckPermittedScopes(ctx, resourceType, action)
}

// the test runs on all assets
var _ = Describe("Test Asset Service auth logic", func() {
	var randoms *testutils.RandomValues
	var differentScope *testutils.RandomValues

	router := CreateRouter()
	entries := NewAssetKindTable(router)

	BeforeEach(func() {
		dbTesting.CleanTables()

		randoms = testutils.NewRandomValues()
		differentScope = randoms.
			CloneWithDifferentProject().Forbidden().
			CloneWithDifferentDepartment().Forbidden().
			CloneWithDifferentCluster().Forbidden()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentScope)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentScope))
		authorization.SetMock(authMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
			WithCluster(differentScope.ClusterId, testing.BuildDisplayedCluster(differentScope, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	Context("Test Get by Id authorization tests", func() {
		DescribeTable("Test get by id with a non existing id", func(testData AssetTest) {
			nonExistingId := uuid.NewString()

			getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, nonExistingId), nil)
			Expect(err).To(BeNil())
			getByIdHttpReq.Header = *headerForAdmin(randoms)

			getByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
			Expect(getByIdHttpResp.Code).To(Equal(http.StatusNotFound))
		}, entries)

		DescribeTable("Test Get/Update/Delete by Id when requester is not permitted for that project", func(testData AssetTest) {
			if testData.canRunOnScope(projectScope) && !testData.customAuthorization {
				assetMeta2, createdSpec, _ := testData.factory.createProjectBasedAsset(differentScope, headerForAdmin(randoms))

				getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, assetMeta2.Id), nil)
				Expect(err).To(BeNil())
				getByIdHttpReq.Header = *headerForNonAdmin(randoms)

				getByIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
				Expect(getByIdHttpResp.Code).To(Equal(http.StatusForbidden))

				putByIdHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, assetMeta2.Id),
					strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
				Expect(err).To(BeNil())
				putByIdHttpReq.Header = *headerForNonAdmin(randoms)

				putByIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(putByIdHttpResp, putByIdHttpReq)
				Expect(putByIdHttpResp.Code).To(Equal(http.StatusForbidden))

				deleteByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, assetMeta2.Id), nil)
				Expect(err).To(BeNil())
				deleteByIdHttpReq.Header = *headerForNonAdmin(randoms)

				deleteByIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteByIdHttpResp, deleteByIdHttpReq)
				Expect(deleteByIdHttpResp.Code).To(Equal(http.StatusForbidden))
			}
		}, entries)
	})

	Context("Test list assets authorization", func() {
		DescribeTable("test list by name with a non existing name", func(testData AssetTest) {
			testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			listByNameHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?name=%s", testData.url, "nonExistingName"), nil)
			Expect(err).To(BeNil())
			listByNameHttpReq.Header = *headerForAdmin(randoms)

			listByNameHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listByNameHttpResp, listByNameHttpReq)
			Expect(listByNameHttpResp.Code).To(Equal(http.StatusOK))

			listByMetaResp := testData.factory.unmarshalList(listByNameHttpResp.Body.Bytes())
			Expect(len(listByMetaResp)).To(Equal(0))
		}, entries)

		DescribeTable("test list with projectId", func(testData AssetTest) {
			if testData.canRunOnScope(projectScope) {
				testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&%s", testData.url, randoms.ProjectIdParam()), nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				listMetaRes := testData.factory.unmarshalList(listAllHttpResp.Body.Bytes())
				Expect(len(listMetaRes)).To(Equal(1))
			}
		}, entries)

		DescribeTable("Test Get credentials with permission only for this project", func(testData AssetTest) {
			if testData.canRunOnScope(projectScope) {
				listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
				Expect(err).To(BeNil())

				testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				listAllHttpReq.Header = *headerForNonAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				listMetaRes := testData.factory.unmarshalList(listAllHttpResp.Body.Bytes())
				Expect(len(listMetaRes)).To(Equal(2))
				Expect(*listMetaRes[0].ProjectId).To(Equal(int32(randoms.ProjectId)))
				Expect(*listMetaRes[1].ProjectId).To(Equal(int32(randoms.ProjectId)))
			}
		}, entries)

		DescribeTable("test get with limited cluster & project scope", func(testData AssetTest) {
			if testData.canRunOnScope(projectScope) {
				listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
				Expect(err).To(BeNil())

				testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				testData.factory.createProjectBasedAsset(differentScope, headerForAdmin(randoms))

				listAllHttpReq.Header = *headerForNonAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				listMetaRes := testData.factory.unmarshalList(listAllHttpResp.Body.Bytes())

				Expect(len(listMetaRes)).To(Equal(1))
				Expect(*listMetaRes[0].ProjectId).To(Equal(int32(randoms.ProjectId)))
				Expect(*listMetaRes[0].ClusterId).To(Equal(randoms.ClusterId))
			}
		}, entries)
	})

	Context("Test create assets authorization", func() {
		DescribeTable("edit create -> missing permissions", func(testData AssetTest) {
			if !testData.customAuthorization {
				creationPayload := testData.factory.marshalCreationReq(
					buildRandomCreationRequest(api.SCOPE_PROJECT, int32(randoms.ProjectId), randoms.ClusterId))

				creationHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(creationPayload))
				Expect(err).To(BeNil())
				creationHttpReq.Header = *headerWithMissingInfo(randoms)

				creationHttpResp := httptest.NewRecorder()
				router.ServeHTTP(creationHttpResp, creationHttpReq)
				Expect(creationHttpResp.Code).To(Equal(http.StatusForbidden))
			}
		}, entries)
	})

	Context("only admin can update tenant scoped assets", func() {
		DescribeTable("researcher fails to update", func(testData AssetTest) {
			// Nothing to test in project scope only assets
			if testData.customAuthorization {
				return
			}
			if !testData.canRunOnScope(tenantScope) {
				return
			}

			tenantAsset, createdSpec, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))

			updateTenantReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, tenantAsset.Id),
				strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
			Expect(err).To(BeNil())
			updateTenantReq.Header = *headerForNonAdmin(randoms)

			updateTenantResp := httptest.NewRecorder()
			router.ServeHTTP(updateTenantResp, updateTenantReq)
			Expect(updateTenantResp).ToNot(BeNil())
			Expect(updateTenantResp.Code).To(Equal(http.StatusForbidden))
		}, entries)
	})

	Context("test list with non permitted project", func() {
		DescribeTable("should return empty response", func(testData AssetTest) {
			if testData.canRunOnScope(projectScope) {
				testData.factory.createProjectBasedAsset(differentScope, headerForAdmin(randoms))

				listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&%s", testData.url, differentScope.ProjectIdParam()), nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForNonAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				listByMetaResp := testData.factory.unmarshalList(listAllHttpResp.Body.Bytes())
				Expect(len(listByMetaResp)).To(Equal(0))
			}
		}, entries)
	})

	// Test cross-tenant isolation: admin from Tenant B should NOT be able to access tenant-scoped assets from Tenant A
	Context("cross-tenant access to tenant-scoped assets should be forbidden", func() {
		DescribeTable("admin from different tenant cannot GET tenant-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(tenantScope) || testData.customAuthorization {
				return
			}
			// Create asset in Tenant A (randoms)
			tenantAsset, _, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))

			// Try to GET using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, tenantAsset.Id), nil)
			Expect(err).To(BeNil())
			getByIdHttpReq.Header = *headerForAdmin(differentTenant)

			getByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
			// Admin from Tenant B should NOT be able to access Tenant A's asset
			Expect(getByIdHttpResp.Code).To(Equal(http.StatusNotFound), "cross-tenant GET should return not found")
		}, entries)

		DescribeTable("admin from different tenant cannot UPDATE tenant-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(tenantScope) || testData.customAuthorization {
				return
			}
			// Create asset in Tenant A (randoms)
			tenantAsset, createdSpec, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))

			// Try to UPDATE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, tenantAsset.Id),
				strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
			Expect(err).To(BeNil())
			updateReq.Header = *headerForAdmin(differentTenant)

			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateReq)
			// Admin from Tenant B should NOT be able to update Tenant A's asset
			Expect(updateResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant UPDATE should return not found, got %d: %s", updateResp.Code, updateResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot DELETE tenant-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(tenantScope) || testData.customAuthorization {
				return
			}
			// Create asset in Tenant A (randoms)
			tenantAsset, _, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))

			// Try to DELETE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, tenantAsset.Id), nil)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(differentTenant)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			// Admin from Tenant B should NOT be able to delete Tenant A's asset
			Expect(deleteResp.Code).To(Equal(http.StatusNotFound), "cross-tenant DELETE should return not found")
		}, entries)
	})

	// Test cross-tenant isolation for CLUSTER-scoped assets
	Context("cross-tenant access to cluster-scoped assets should be forbidden", func() {
		DescribeTable("admin from different tenant cannot GET cluster-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(clusterScope) || testData.customAuthorization {
				return
			}
			// Create cluster-scoped asset in Tenant A's cluster
			clusterAsset, _, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))

			// Try to GET using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, clusterAsset.Id), nil)
			Expect(err).To(BeNil())
			getReq.Header = *headerForAdmin(differentTenant)

			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq)
			Expect(getResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant GET of cluster-scoped asset should return not found, got %d: %s", getResp.Code, getResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot UPDATE cluster-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(clusterScope) || testData.customAuthorization {
				return
			}
			// Create cluster-scoped asset in Tenant A's cluster
			clusterAsset, createdSpec, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))

			// Try to UPDATE using admin from Tenant B (different tenant with different cluster)
			differentTenant := randoms.CloneWithDifferentTenant()
			updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, clusterAsset.Id),
				strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
			Expect(err).To(BeNil())
			updateReq.Header = *headerForAdmin(differentTenant)

			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateReq)
			Expect(updateResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant UPDATE of cluster-scoped asset should return not found, got %d: %s", updateResp.Code, updateResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot DELETE cluster-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(clusterScope) || testData.customAuthorization {
				return
			}
			// Create cluster-scoped asset in Tenant A's cluster
			clusterAsset, _, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))

			// Try to DELETE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, clusterAsset.Id), nil)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(differentTenant)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant DELETE of cluster-scoped asset should return not found, got %d: %s", deleteResp.Code, deleteResp.Body.String()))
		}, entries)
	})

	// Test cross-tenant isolation for DEPARTMENT-scoped assets
	Context("cross-tenant access to department-scoped assets should be forbidden", func() {
		DescribeTable("admin from different tenant cannot GET department-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(departmentScope) || testData.customAuthorization {
				return
			}
			// Create department-scoped asset in Tenant A's department
			deptAsset, _, _ := testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))

			// Try to GET using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, deptAsset.Id), nil)
			Expect(err).To(BeNil())
			getReq.Header = *headerForAdmin(differentTenant)

			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq)
			Expect(getResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant GET of department-scoped asset should return not found, got %d: %s", getResp.Code, getResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot UPDATE department-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(departmentScope) || testData.customAuthorization {
				return
			}
			// Create department-scoped asset in Tenant A's department
			deptAsset, createdSpec, _ := testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))

			// Try to UPDATE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, deptAsset.Id),
				strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
			Expect(err).To(BeNil())
			updateReq.Header = *headerForAdmin(differentTenant)

			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateReq)
			Expect(updateResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant UPDATE of department-scoped asset should return not found, got %d: %s", updateResp.Code, updateResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot DELETE department-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(departmentScope) || testData.customAuthorization {
				return
			}
			// Create department-scoped asset in Tenant A's department
			deptAsset, _, _ := testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))

			// Try to DELETE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, deptAsset.Id), nil)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(differentTenant)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant DELETE of department-scoped asset should return not found, got %d: %s", deleteResp.Code, deleteResp.Body.String()))
		}, entries)
	})

	// Test cross-tenant isolation for PROJECT-scoped assets
	Context("cross-tenant access to project-scoped assets should be forbidden", func() {
		DescribeTable("admin from different tenant cannot GET project-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(projectScope) || testData.customAuthorization {
				return
			}
			// Create project-scoped asset in Tenant A's project
			projectAsset, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			// Try to GET using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			getReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, projectAsset.Id), nil)
			Expect(err).To(BeNil())
			getReq.Header = *headerForAdmin(differentTenant)

			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq)
			Expect(getResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant GET of project-scoped asset should return not found, got %d: %s", getResp.Code, getResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot UPDATE project-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(projectScope) || testData.customAuthorization {
				return
			}
			// Create project-scoped asset in Tenant A's project
			projectAsset, createdSpec, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			// Try to UPDATE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			updateReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, projectAsset.Id),
				strings.NewReader(testData.factory.marshalUpdateReq(createdSpec)))
			Expect(err).To(BeNil())
			updateReq.Header = *headerForAdmin(differentTenant)

			updateResp := httptest.NewRecorder()
			router.ServeHTTP(updateResp, updateReq)
			Expect(updateResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant UPDATE of project-scoped asset should return not found, got %d: %s", updateResp.Code, updateResp.Body.String()))
		}, entries)

		DescribeTable("admin from different tenant cannot DELETE project-scoped asset", func(testData AssetTest) {
			if !testData.canRunOnScope(projectScope) || testData.customAuthorization {
				return
			}
			// Create project-scoped asset in Tenant A's project
			projectAsset, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			// Try to DELETE using admin from Tenant B (different tenant)
			differentTenant := randoms.CloneWithDifferentTenant()
			deleteReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, projectAsset.Id), nil)
			Expect(err).To(BeNil())
			deleteReq.Header = *headerForAdmin(differentTenant)

			deleteResp := httptest.NewRecorder()
			router.ServeHTTP(deleteResp, deleteReq)
			Expect(deleteResp.Code).To(Equal(http.StatusNotFound), fmt.Sprintf("cross-tenant DELETE of project-scoped asset should return not found, got %d: %s", deleteResp.Code, deleteResp.Body.String()))
		}, entries)
	})
})

// Cross-tenant endpoint tests: crossTenantAwareAuth gives empty scopes for other tenants → 403.
var _ = Describe("Cross-tenant endpoint tests", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()
		randoms = testutils.NewRandomValues()
		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(&crossTenantAwareAuth{allowedTenantId: int32(randoms.TenantId), inner: authMock})
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	DescribeTable("returns 200 with empty results for LIST when request is from different tenant",
		func(urlFunc func() string, setup func()) {
			setup()
			req, _ := http.NewRequest("GET", urlFunc(), nil)
			req.Header = *headerForAdmin(randoms.CloneWithDifferentTenant())
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK))
		},
		Entry("LIST credentials (NGC API key)", func() string { return urlNgcApiKey }, func() { createNgcApiKey(buildRandomClusterNgcApiKey(randoms), router, headerForAdmin(randoms)) }),
		Entry("LIST docker-registry credentials", func() string { return urlDockerRegistry }, func() {
			createDockerRegistry(&randoms.ClusterId, buildRandomClusterDockerRegistry(randoms), router, headerForAdmin(randoms))
		}),
		Entry("LIST environments", func() string { return urlEnvironment }, func() {
			createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
		}),
		Entry("LIST environments with clusterId filter", func() string { return fmt.Sprintf("%s?clusterId=%s", urlEnvironment, randoms.ClusterId) }, func() {
			createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
		}),
		Entry("LIST host-path datasources", func() string { return urlHostPath }, func() { createHostPathAsset(buildRandomClusterHostPathAsset(randoms), router, headerForAdmin(randoms)) }),
	)

	DescribeTable("returns 403 for CREATE when request is from different tenant",
		func(url string, bodyFunc func() *strings.Reader) {
			req, _ := http.NewRequest("POST", url, bodyFunc())
			req.Header = *headerForAdmin(randoms.CloneWithDifferentTenant())
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusForbidden))
		},
		Entry("CREATE NGC API key", urlNgcApiKey, func() *strings.Reader { return makePayload(buildRandomClusterNgcApiKey(randoms)) }),
		Entry("CREATE docker-registry credential", urlDockerRegistry, func() *strings.Reader { return makePayload(buildRandomClusterDockerRegistry(randoms)) }),
		Entry("CREATE environment", urlEnvironment, func() *strings.Reader { return makePayload(buildRandomClusterEnvironmentAsset(randoms)) }),
		Entry("CREATE host-path datasource", urlHostPath, func() *strings.Reader { return makePayload(buildRandomClusterHostPathAsset(randoms)) }),
	)

	It("returns 200 for same-tenant admin listing credentials and environments", func() {
		createNgcApiKey(buildRandomClusterNgcApiKey(randoms), router, headerForAdmin(randoms))
		createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, headerForAdmin(randoms))
		for _, url := range []string{urlNgcApiKey, urlEnvironment} {
			req, _ := http.NewRequest("GET", url, nil)
			req.Header = *headerForAdmin(randoms)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK))
		}
	})
})

func buildRandomCreationRequest(scope api.Scope, projectId int32, clusterId string) api.AssetCreationRequest {
	var clusterIdPtr *string
	var projectIdPtr *int32

	if projectId != 0 {
		projectIdPtr = convutils.Ref(projectId)
	}
	if clusterId != "" {
		clusterIdPtr = convutils.Ref(clusterId)
	}

	return api.AssetCreationRequest{
		Name:      testutils.RandString(5),
		Scope:     scope,
		ClusterId: clusterIdPtr,
		ProjectId: projectIdPtr,
	}
}
