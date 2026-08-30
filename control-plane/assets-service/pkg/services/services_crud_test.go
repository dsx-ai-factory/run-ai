package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbcleaner"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"gorm.io/gorm"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
)

const (
	urlCountAssets = "/api/v1/asset/count"
)

type GenericGetResponse struct {
	Meta   api.AssetMeta
	Status *api.AssetClusterStatusInfo `json:"status,omitempty"`
}

type LastUsageKey struct {
	WorkloadType string
	ProjectId    int32
	AssetId      string
}

func NewLastUsageKey(workloadType string, projectId int32, assetId string) LastUsageKey {
	return LastUsageKey{
		WorkloadType: workloadType,
		ProjectId:    projectId,
		AssetId:      assetId,
	}
}

// the test runs on all assets
var _ = Describe("Test Asset Service CRUD logic", func() {
	var randoms *testutils.RandomValues
	var differentProject *testutils.RandomValues
	var differentCluster *testutils.RandomValues
	var policyMockUpdateTime time.Time
	var origGraceTime int

	router := CreateRouter()
	entries := NewAssetKindTable(router)

	BeforeEach(func() {
		dbTesting.CleanTables()

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		randoms = testutils.NewRandomValues()
		differentProject = randoms.CloneWithDifferentProject()
		differentCluster = randoms.CloneWithDifferentCluster()

		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithConfigMap(randoms.ClusterId, &api.ConfigMapCreationRequest{
			Spec: api.ConfigMapAssetSpec{
				ConfigMap: convutils.Ref(strings.ToLower(testutils.RandString(8))),
			},
		}, nil))

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)).
			WithCluster(differentCluster.ClusterId, testing.BuildDisplayedCluster(differentCluster, testing.UnitTestDefaultClusterVersion, false)))

		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject, differentCluster)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject, differentCluster))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		auth.TestingAllowCreateSystemScopedAssets = true
	})

	Context("create system, tenant, cluster and project scoped assets -> list all", func() {
		DescribeTable("all assets included in the list", func(testData AssetTest) {
			var specs [][]byte

			if testData.canRunOnScope(api.SCOPE_SYSTEM) {
				_, spec, _ := testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
				specs = append(specs, spec)
			}

			if testData.canRunOnScope(api.SCOPE_TENANT) {
				_, spec, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))
				specs = append(specs, spec)
			}
			if testData.canRunOnScope(api.SCOPE_CLUSTER) {
				_, spec, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))
				specs = append(specs, spec)
			}

			_, spec4, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			specs = append(specs, spec4)

			//  list all assets of the type being tested
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains both of the created assets
			testData.factory.validateListContent(listAllHttpReq, nil, specs...)
		}, entries)
	})

	Context("create system and tenant scoped assets for two different tenants -> list all", func() {
		DescribeTable("only the requesting tenant's assets are included in the list", func(testData AssetTest) {
			if !testData.canRunOnScope(api.SCOPE_SYSTEM) || !testData.canRunOnScope(api.SCOPE_TENANT) {
				return
			}
			randomsTenant2 := randoms.CloneWithDifferentTenant()
			_, specSystem, _ := testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			_, specTenant1, _ := testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))
			_, specTenant2, _ := testData.factory.createTenantBasedAsset(randomsTenant2, headerForAdmin(randomsTenant2))

			//  list all assets of the type being tested - with tenant1
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			testData.factory.validateListContent(listAllHttpReq, nil, specSystem, specTenant1)

			//  list all assets of the type being tested - with tenant2
			listAllHttpReq, err = http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randomsTenant2)

			testData.factory.validateListContent(listAllHttpReq, nil, specSystem, specTenant2)
		}, entries)
	})

	Context("create two assets -> count by name", func() {
		DescribeTable("both assets included in the list", func(testData AssetTest) {
			meta1, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//  count by name
			countHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?kind=%s&name=%s", urlCountAssets, testData.kind, meta1.Name), nil)
			Expect(err).To(BeNil())
			countHttpReq.Header = *headerForAdmin(randoms)

			countHttpResp := httptest.NewRecorder()
			router.ServeHTTP(countHttpResp, countHttpReq)
			Expect(countHttpResp.Code).To(Equal(http.StatusOK))

			resp := api.AssetCountResponse{}
			respJsonErr := json.Unmarshal(countHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(resp.Result).To(Equal(int64(1)))
		}, entries)
	})

	Context("create two assets -> list by name", func() {
		DescribeTable("only the requested asset is listed", func(testData AssetTest) {
			meta1, spec1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//  list assets by name providing the name of the first asset
			listByNameHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?name=%s", testData.url, meta1.Name), nil)
			Expect(err).To(BeNil())
			listByNameHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains only the first asset
			testData.factory.validateListContent(listByNameHttpReq, nil, spec1)
		}, entries)
	})

	Context("create system, tenant, cluster, department and project scoped assets -> list by scope", func() {
		DescribeTable("only the relevant asset is listed", func(testData AssetTest) {
			var specSystem []byte
			var specTenant []byte
			var specCluster []byte
			var specDept []byte
			if testData.canRunOnScope(api.SCOPE_SYSTEM) {
				_, specSystem, _ = testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_TENANT) {
				_, specTenant, _ = testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_CLUSTER) {
				_, specCluster, _ = testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_DEPARTMENT) {
				_, specDept, _ = testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))
			}
			_, specProject, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//  list system scope assets
			listSystemHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_SYSTEM), nil)
			Expect(err).To(BeNil())
			listSystemHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listSystemHttpReq, nil, specSystem)

			//  list tenant scope assets
			listTenantHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_TENANT), nil)
			Expect(err).To(BeNil())
			listTenantHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listTenantHttpReq, nil, specSystem, specTenant)

			//  list cluster scope assets
			listClusterHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_CLUSTER), nil)
			Expect(err).To(BeNil())
			listClusterHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listClusterHttpReq, nil, specCluster)

			//  list project scope assets
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_PROJECT), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProject)

			//  list department scope assets
			listDepartmentHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_DEPARTMENT), nil)
			Expect(err).To(BeNil())
			listDepartmentHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listDepartmentHttpReq, nil, specDept)
		}, entries)
	})

	Context("create system, tenant, cluster department and project scoped assets -> list by department id and project id", func() {
		DescribeTable("only the relevant assets are listed", func(testData AssetTest) {
			var specSystem []byte
			var specTenant []byte
			var specCluster []byte
			var specDept []byte
			if testData.canRunOnScope(api.SCOPE_SYSTEM) {
				_, specSystem, _ = testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_TENANT) {
				_, specTenant, _ = testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_CLUSTER) {
				_, specCluster, _ = testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))
			}
			if testData.canRunOnScope(api.SCOPE_DEPARTMENT) {
				_, specDept, _ = testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))
			}
			_, specProject, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//  list by project id
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", testData.url, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specSystem, specTenant, specCluster, specDept, specProject)

			if testData.canRunOnScope(departmentScope) {
				//  list by department if
				listDepartmentHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?departmentId=%s", testData.url, randoms.DepartmentId), nil)
				Expect(err).To(BeNil())
				listDepartmentHttpReq.Header = *headerForAdmin(randoms)
				testData.factory.validateListContent(listDepartmentHttpReq, nil, specSystem, specTenant, specCluster, specDept)
			}
		}, entries)
	})

	Context("create two project scoped assets -> list by project id", func() {
		DescribeTable("only the relevant asset is listed", func(testData AssetTest) {
			_, specProj1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			testData.factory.createProjectBasedAsset(differentProject, headerForAdmin(randoms))

			//  list system scope assets shows nothing
			listSystemHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_SYSTEM), nil)
			Expect(err).To(BeNil())
			listSystemHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listSystemHttpReq, nil)

			//  list tenant scope assets shows nothing
			listTenantHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", testData.url, api.SCOPE_TENANT), nil)
			Expect(err).To(BeNil())
			listTenantHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listTenantHttpReq, nil)

			//  list by project shows only the correct asset
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProj1)
		}, entries)
	})

	Context("create an asset -> get it by its id", func() {
		DescribeTable("asset content should change", func(testData AssetTest) {
			meta1, spec1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, meta1.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			Expect(getHttpResp.Body.String()).To(Equal(string(spec1)))
		}, entries)
	})

	Context("create two assets -> delete one of them", func() {
		DescribeTable("only the non deleted asset is listed", func(testData AssetTest) {
			meta1, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			_, spec2, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//   delete the 2nd asset
			delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, meta1.Id), nil)
			Expect(err).To(BeNil())
			delByIdHttpReq.Header = *headerForAdmin(randoms)

			delByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
			if testData.assetDeletionResponseCode != 0 {
				Expect(delByIdHttpResp.Code).To(Equal(testData.assetDeletionResponseCode))
			} else {
				Expect(delByIdHttpResp.Code).To(Equal(http.StatusAccepted))
			}

			//   list all assets
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//  the list should include only the non deleted element
			testData.factory.validateListContent(listAllHttpReq, nil, spec2)
		}, entries)
	})

	Context("create an asset -> delete it", func() {
		DescribeTable("get/update should fail for the deleted item", func(testData AssetTest) {
			meta1, createdSpec, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			updateRequest := testData.factory.marshalUpdateReq(createdSpec) // we need it only to prevent update from returning 400 error

			//   delete the asset
			delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, meta1.Id), nil)
			Expect(err).To(BeNil())
			delByIdHttpReq.Header = *headerForAdmin(randoms)

			delByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
			if testData.assetDeletionResponseCode != 0 {
				Expect(delByIdHttpResp.Code).To(Equal(testData.assetDeletionResponseCode))
			} else {
				Expect(delByIdHttpResp.Code).To(Equal(http.StatusAccepted))
			}

			for _, action := range []string{"GET", "PUT"} {
				//    get/delete/update by ID should fail
				byIdHttpReq, err := http.NewRequest(action, fmt.Sprintf("%s/%s", testData.url, meta1.Id), strings.NewReader(updateRequest))
				Expect(err).To(BeNil())
				byIdHttpReq.Header = *headerForAdmin(randoms)

				byIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(byIdHttpResp, byIdHttpReq)
				Expect(byIdHttpResp.Code).To(Equal(http.StatusNotFound))
			}
		}, entries)
	})

	Context("create two assets -> update one of them", func() {
		DescribeTable("asset content should change", func(testData AssetTest) {
			meta1, spec1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			_, spec2, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			if testData.renameNotAllowedDuringSync {
				//
				//   simulate sync of the asset, if needed
				//
				var metaForUpdate dbmodel.AssetMetaRec
				copier.Copy(&metaForUpdate.AssetMeta, &meta1)
				metaForUpdate.AssetMeta.DeletedAt.Valid = false

				syncErr := dao.UpdateSyncStatusSendToAgent(context.Background(), *meta1.TenantId, *meta1.ClusterId, []dbmodel.AssetsToSync{
					{
						AssetMetaRec:        metaForUpdate,
						AssetsGenericSecret: &dbmodel.GenericSecretRecord{},
					},
				})
				Expect(syncErr).To(BeNil())

				dao.HandleAgentSyncUpdates(context.Background(), *meta1.TenantId, *meta1.ClusterId, []api.ClusterSyncReportUpdate{
					{
						Id:          meta1.Id,
						UpdateCount: meta1.UpdateCount,
					},
				})
			}

			updateReqPayload := testData.factory.marshalUpdateReq(spec1)

			updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, meta1.Id), strings.NewReader(updateReqPayload))
			Expect(err).To(BeNil())
			updateHttpReq.Header = *headerForAdmin(randoms)

			updateHttpResp := httptest.NewRecorder()
			router.ServeHTTP(updateHttpResp, updateHttpReq)
			if testData.assetUpdateResponseCode != 0 {
				Expect(updateHttpResp.Code).To(Equal(testData.assetUpdateResponseCode))
			} else {
				Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))
			}

			//   list all assets
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//  the list should include only the non deleted element
			testData.factory.validateListContent(listAllHttpReq, nil, updateHttpResp.Body.Bytes(), spec2)
		}, entries)
	})

	Context("create asset with invalid scope", func() {
		DescribeTable("request should fail with bad request error", func(testData AssetTest) {
			assetCreationRequest := api.AssetCreationRequest{
				Name:  "whatever",
				Scope: "scope that does not exist",
			}
			postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

			createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		}, entries)
	})

	Context("create asset with system scope (not in testing mode)", func() {
		DescribeTable("request should fail with forbidden error", func(testData AssetTest) {
			// not allowing system scope, as in real mode
			auth.TestingAllowCreateSystemScopedAssets = false

			assetCreationRequest := api.AssetCreationRequest{
				Name:  "system-scoped asset",
				Scope: api.SCOPE_SYSTEM,
			}

			if testData.supportsWorkloadType {
				assetCreationRequest.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
					Workspace: convutils.Ref(true),
				}
			}

			postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

			createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			expectedStatusCode := http.StatusForbidden
			if testData.canRunOnlyOnScope(tenantScope) {
				expectedStatusCode = http.StatusBadRequest
			}
			Expect(createHttpResp.Code).To(Equal(expectedStatusCode))
		}, entries)
	})

	Context("delete asset with system scope", func() {
		DescribeTable("request should fail with forbidden error", func(testData AssetTest) {
			if !testData.canRunOnScope(api.SCOPE_SYSTEM) {
				return
			}

			systemAssetMeta, systemAssetSpec, _ := testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))

			// skip registry assets which cannot be created for system scope
			if testData.kind == "registry" {
				return
			}

			// not allowing system scope, as in real mode
			auth.TestingAllowCreateSystemScopedAssets = false

			//   delete the 2nd asset
			delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, systemAssetMeta.Id), nil)
			Expect(err).To(BeNil())
			delByIdHttpReq.Header = *headerForAdmin(randoms)

			delByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
			Expect(delByIdHttpResp.Code).To(Equal(http.StatusForbidden))

			//   list all assets
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//  the list should include the non deleted element
			testData.factory.validateListContent(listAllHttpReq, nil, systemAssetSpec)
		}, entries)
	})

	Context("delete asset with tenant, cluster, department and project scope", func() {
		DescribeTable("request should succeed with status approved", func(testData AssetTest) {
			if testData.kind == "registry" {
				return
			}
			var tenantAssetMeta api.AssetMeta
			var clusterAssetMeta api.AssetMeta
			var departmentAssetMeta api.AssetMeta
			var assetMetas []api.AssetMeta
			if testData.canRunOnScope(tenantScope) {
				tenantAssetMeta, _, _ = testData.factory.createTenantBasedAsset(randoms, headerForAdmin(randoms))
				assetMetas = append(assetMetas, tenantAssetMeta)
			}
			if testData.canRunOnScope(clusterScope) {
				clusterAssetMeta, _, _ = testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))
				assetMetas = append(assetMetas, clusterAssetMeta)
			}
			if testData.canRunOnScope(departmentScope) {
				departmentAssetMeta, _, _ = testData.factory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))
				assetMetas = append(assetMetas, departmentAssetMeta)
			}

			projectAssetMeta, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			assetMetas = append(assetMetas, projectAssetMeta)

			for _, assetMeta := range assetMetas {
				delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, assetMeta.Id), nil)
				Expect(err).To(BeNil())
				delByIdHttpReq.Header = *headerForAdmin(randoms)

				delByIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
				Expect(delByIdHttpResp.Code).To(Equal(http.StatusAccepted))
			}

			//   list all assets
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//  the list should not include the deleted elements
			testData.factory.validateListContent(listAllHttpReq, nil)
		}, entries)
	})

	Context("delete an asset which is being used by a workload", func() {
		DescribeTable("delete fail for the deleted item with precondition error", func(testData AssetTest) {
			if testData.notSupportedPre217 || testData.notSupportedPre220 {
				return
			}

			cannotDeleteAssetBeingUsed(testData, auth.WorkloadTypeTraining, randoms, router)
			cannotDeleteAssetBeingUsed(testData, auth.WorkloadTypeWorkspace, randoms, router)
		}, entries)
	})

	Context("test usage info - workspaces, trainings and templates", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if testData.testUsageInfo && !testData.notSupportedPre217 && !testData.notSupportedPre220 {
				//   environment asset we always need
				envAsset, envSpec, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				var testAsset api.AssetMeta
				var testSpec []byte

				if testData.kind == api.ASSETKIND_ENVIRONMENT {
					testAsset = envAsset.Meta
					testSpec = envSpec
				} else {
					testAsset, testSpec, _ = testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				}

				expectedUsage := datastructs.Set[string]{} // expected set of workloads which are using our asset

				//  list with usage, currently not used by any workspace/template
				listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				//   make sure the list contains the created asset with no usage info
				testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

				// for this test we prefer old clusters, because we do not want workloads to be rejected for pvcs
				// or secrets missing in the cluster (with new clusters we have to set k8s objects tracker mock to
				// hold those assets, otherwise workload creation will fail
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

				//  now the environment will be used by two workspaces and one training
				expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeWorkspace, randoms, envAsset, testAsset))
				expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeWorkspace, randoms, envAsset, testAsset))
				expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeTraining, randoms, envAsset, testAsset))
				//  ...and by one template
				template, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
				expectedUsage.Add(template.Meta.Id)

				//  list with usage, now used by the three workloads + one template
				listAllHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", testData.url), nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				//   make sure the list contains the created assets and the usage info now contains the three workloads
				testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

				//  specific "Get" with usage, again used by three workloads + one template
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", testData.url, testAsset.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				validateUsageInfoOfAsset(router, getHttpReq, expectedUsage)
			}
		}, entries)
	})

	Context("test usage info with system scope assets - workspaces, trainings and templates", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if !testData.testUsageInfo || !testData.canRunOnScope(api.SCOPE_SYSTEM) {
				return
			}

			// for this test we prefer old clusters, because we do not want workloads to be rejected for pvcs
			// or secrets missing in the cluster (with new clusters we have to set k8s objects tracker mock to
			// hold those assets, otherwise workload creation will fail
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

			//   environment asset we always need
			envAsset, envSpec, _ := createEnvironmentAsset(buildRandomSystemEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			var testAsset api.AssetMeta
			var testSpec []byte

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				testAsset = envAsset.Meta
				testSpec = envSpec
			} else {
				testAsset, testSpec, _ = testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			}

			expectedUsage := datastructs.Set[string]{} // expected set of workloads which are using our asset

			//  list with usage, currently not used by any workspace/template
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains the created asset with no usage info
			testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

			//  now the environment will be used by two workspaces and one training
			expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeWorkspace, randoms, envAsset, testAsset))
			expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeWorkspace, randoms, envAsset, testAsset))
			expectedUsage.Add(createTestWorkload(router, auth.WorkloadTypeTraining, randoms, envAsset, testAsset))
			//  ...and by one template
			template, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
			expectedUsage.Add(template.Meta.Id)

			//  list with usage, now used by the three workloads + one template
			listAllHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", testData.url), nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains the created assets and the usage info now contains the three workloads
			testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

			//  specific "Get" with usage, again used by three workloads + one template
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", testData.url, testAsset.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			validateUsageInfoOfAsset(router, getHttpReq, expectedUsage)
		}, entries)
	})

	Context("test usage times - workspaces and trainings", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if testData.testComplianceInfo {
				testAsset1, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				testAsset2, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				january := time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)
				february := time.Date(2023, time.February, 1, 0, 0, 0, 0, time.UTC)
				march := time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
				april := time.Date(2023, time.April, 1, 0, 0, 0, 0, time.UTC)

				ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)

				//
				//   build the following table
				//                                        asset1         asset2
				//        training     main-project       january        april
				//           "         different-proj     february       march
				//        workspace    main-project       march          february
				//           "         different-proj     april          january
				//
				expected := map[LastUsageKey]time.Time{
					NewLastUsageKey(auth.WorkloadTypeTraining, int32(randoms.ProjectId), testAsset1.Id):           january,
					NewLastUsageKey(auth.WorkloadTypeTraining, int32(differentProject.ProjectId), testAsset1.Id):  february,
					NewLastUsageKey(auth.WorkloadTypeWorkspace, int32(randoms.ProjectId), testAsset1.Id):          march,
					NewLastUsageKey(auth.WorkloadTypeWorkspace, int32(differentProject.ProjectId), testAsset1.Id): april,

					NewLastUsageKey(auth.WorkloadTypeTraining, int32(randoms.ProjectId), testAsset2.Id):           april,
					NewLastUsageKey(auth.WorkloadTypeTraining, int32(differentProject.ProjectId), testAsset2.Id):  march,
					NewLastUsageKey(auth.WorkloadTypeWorkspace, int32(randoms.ProjectId), testAsset2.Id):          february,
					NewLastUsageKey(auth.WorkloadTypeWorkspace, int32(differentProject.ProjectId), testAsset2.Id): january,
				}

				for _, assetId := range []string{testAsset1.Id, testAsset2.Id} {
					for _, workloadType := range []string{auth.WorkloadTypeTraining, auth.WorkloadTypeWorkspace} {
						for _, projectId := range []int32{int32(randoms.ProjectId), int32(differentProject.ProjectId)} {
							err := dao.UpsertAssetLastUsedByWorkload(ctx, assetId, workloadType, uuid.NewString(), projectId,
								expected[NewLastUsageKey(workloadType, projectId, assetId)])
							Expect(err).To(BeNil())
						}
					}
				}

				//
				//  now we do some combinations of listing and make sure they come out ok
				//
				for _, workloadType := range []string{auth.WorkloadTypeTraining, auth.WorkloadTypeWorkspace} {
					for _, projectId := range []int32{int32(randoms.ProjectId), int32(differentProject.ProjectId)} {
						listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?complyToProject=%d&complyToWorkloadType=%s", testData.url, projectId, workloadType), nil)
						Expect(err).To(BeNil())
						listHttpReq.Header = *headerForAdmin(randoms)

						listHttpResp := httptest.NewRecorder()
						router.ServeHTTP(listHttpResp, listHttpReq)
						Expect(listHttpResp.Code).To(Equal(http.StatusOK))

						var listResponse MetaAndUsageEntries
						err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResponse)
						Expect(err).To(BeNil())
						Expect(len(listResponse.Entries)).To(Equal(2)) // one per asset

						for _, entry := range listResponse.Entries {
							actualResult := entry.UsageTimes.LastUsedByWorkload.UTC()
							expectedResult := expected[NewLastUsageKey(workloadType, projectId, entry.Meta.Id)].UTC()
							Expect(actualResult).To(Equal(expectedResult), fmt.Sprintf("usage time of asset %s workloadType %s projectId %d", entry.Meta.Id, workloadType, projectId))
						}
					}
				}
			}
		}, entries)
	})

	Context("test usage info - templates only", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if testData.testUsageInfo {
				//   environment asset we always need
				envAsset, envSpec, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				var testAsset api.AssetMeta
				var testSpec []byte

				if testData.kind == api.ASSETKIND_ENVIRONMENT {
					testAsset = envAsset.Meta
					testSpec = envSpec
				} else {
					testAsset, testSpec, _ = testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				}

				expectedUsage := datastructs.Set[string]{} // expected set of templates which are using our asset

				//  now the asset will be used by two templates
				template1, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
				expectedUsage.Add(template1.Meta.Id)
				template2, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
				expectedUsage.Add(template2.Meta.Id)

				//  list with usage, now used by the templates
				listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", testData.url), nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				//   make sure the list contains the created assets and the usage info now contains the two workspaces
				testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

				//  specific "Get" with usage, again used by two workspaces
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", testData.url, testAsset.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				validateUsageInfoOfAsset(router, getHttpReq, expectedUsage)
			}
		}, entries)
	})

	Context("test usage info with system scope assets - templates only", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if !testData.testUsageInfo || !testData.canRunOnScope(api.SCOPE_SYSTEM) {
				return
			}
			//   environment asset we always need
			envAsset, envSpec, _ := createEnvironmentAsset(buildRandomSystemEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			var testAsset api.AssetMeta
			var testSpec []byte

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				testAsset = envAsset.Meta
				testSpec = envSpec
			} else {
				testAsset, testSpec, _ = testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			}

			expectedUsage := datastructs.Set[string]{} // expected set of templates which are using our asset

			//  now the asset will be used by two templates
			template1, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
			expectedUsage.Add(template1.Meta.Id)
			template2, _, _ := createTemplateAsset(buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
			expectedUsage.Add(template2.Meta.Id)

			//  list with usage, now used by the templates
			listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", testData.url), nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains the created assets and the usage info now contains the two workspaces
			testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

			//  specific "Get" with usage, again used by two workspaces
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", testData.url, testAsset.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			validateUsageInfoOfAsset(router, getHttpReq, expectedUsage)
		}, entries)
	})

	Context("test usage info with system scope assets - for templates of system scope", func() {
		DescribeTable("request should return usage info of the asset", func(testData AssetTest) {
			if !testData.testUsageInfo || !testData.canRunOnScope(api.SCOPE_SYSTEM) {
				return
			}

			//   environment asset we always need
			envAsset, envSpec, _ := createEnvironmentAsset(buildRandomSystemEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			var testAsset api.AssetMeta
			var testSpec []byte

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				testAsset = envAsset.Meta
				testSpec = envSpec
			} else {
				testAsset, testSpec, _ = testData.factory.createSystemBasedAsset(randoms, headerForAdmin(randoms))
			}

			expectedUsage := datastructs.Set[string]{} // expected set of templates which are using our asset

			//  now the asset will be used by two templates
			template1, _, _ := createTemplateAsset(buildRandomSystemTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
			expectedUsage.Add(template1.Meta.Id)
			template2, _, _ := createTemplateAsset(buildRandomSystemTemplate(randoms, envAsset.Meta.Id, testAsset), router, headerForAdmin(randoms))
			expectedUsage.Add(template2.Meta.Id)

			//  list with usage, now used by the templates
			listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?usageInfo=true", testData.url), nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			//   make sure the list contains the created assets and the usage info now contains the two workspaces
			testData.factory.validateListContent(listAllHttpReq, expectedUsage, testSpec)

			//  specific "Get" with usage, again used by two workspaces
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?usageInfo=true", testData.url, testAsset.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			validateUsageInfoOfAsset(router, getHttpReq, expectedUsage)
		}, entries)
	})

	Context("asset not supporting auto deletion", func() {
		DescribeTable("request should not fail but only for pvc the auto delete is accepted", func(testData AssetTest) {
			if !testData.canRunOnScope(api.SCOPE_TENANT) {
				return
			}

			assetCreationRequest := api.AssetCreationRequest{
				Name:       "whatever",
				Scope:      api.SCOPE_TENANT,
				AutoDelete: convutils.Ref(true),
			}
			// By default, test using tenant scope, as some assets required tenant scope (registry)
			// however in case of project only scope asset, test using project scope
			if testData.canRunOnScope(api.SCOPE_PROJECT) {
				assetCreationRequest = api.AssetCreationRequest{
					Name:       "whatever",
					Scope:      api.SCOPE_PROJECT,
					ProjectId:  convutils.Ref(int32(randoms.ProjectId)),
					AutoDelete: convutils.Ref(true),
				}
			}

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				assetCreationRequest.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
					Workspace:     convutils.Ref(true),
					Training:      convutils.Ref(false),
					Inference:     convutils.Ref(false),
					Distributed:   convutils.Ref(false),
					DistFramework: nil,
				}
			}

			var meta api.AssetMeta

			switch testData.kind {
			case api.ASSETKIND_WORKLOAD_TEMPLATE:
				// Creation of template needs to be more precise
				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				resp, _, _ := createTemplateAsset(buildRandomTenantTemplate(randoms, envAsset.Meta.Id), router, headerForAdmin(randoms))
				meta = resp.Meta
			case api.ASSETKIND_SECRET_VOLUME:
				// Creation of secret storage needs to create a credential as well
				secretReq := testing.BuildRandomProjectSecretVolumeAsset(randoms, nil)
				secretReq.Meta.AutoDelete = convutils.Ref(true)
				secret, _, _ := createSecretVolumeAsset(&randoms.ClusterId, secretReq, router, headerForAdmin(randoms))
				meta = secret.Meta
			default:
				postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

				createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)

				Expect(createHttpResp.Code).To(Or(Equal(http.StatusAccepted), Equal(http.StatusCreated)))

				var createResp GenericGetResponse
				err = json.Unmarshal(createHttpResp.Body.Bytes(), &createResp)
				Expect(err).To(BeNil())

				meta = createResp.Meta
			}

			Expect(convutils.Deref(meta.AutoDelete, dao.DefaultAutoDelete)).To(Equal(testData.supportsAutoDeletion))
		}, entries)
	})

	Context("only auto delete PVC assets can be used as workloadVolumes", func() {
		DescribeTable("any other combination should fail", func(testData AssetTest) {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			var testAsset api.AssetMeta

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				testAsset = envAsset.Meta
			} else {
				testAsset, _, _ = testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			}

			templatePayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id, testAsset)
			templatePayload.Spec.Assets.WorkloadVolumes = []string{testAsset.Id}

			createHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(templatePayload))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))

			if !testData.supportsAutoDeletion {
				Expect(createHttpResp.Body.String()).To(Or(
					ContainSubstring("is incompatible with workloadVolume (asset kind should be PVC)"),
					ContainSubstring("rather than")))
			} else {
				Expect(createHttpResp.Body.String()).To(ContainSubstring("is not an autoDelete asset"))
			}
		}, entries)
	})

	Context("auto delete cannot be datasource", func() {
		DescribeTable("request should fail", func(testData AssetTest) {
			if testData.supportsAutoDeletion {
				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				assetCreationRequest := api.AssetCreationRequest{
					Name:       "whatever",
					Scope:      api.SCOPE_CLUSTER,
					ClusterId:  convutils.Ref(randoms.ClusterId),
					AutoDelete: convutils.Ref(true),
				}
				assetMeta, _ := createAutoDeleteAsset(testData, router, randoms, &assetCreationRequest)

				templatePayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
				templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
					{
						Id:   assetMeta.Id,
						Kind: testData.kind,
					},
				}

				createTempHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(templatePayload))
				createTempHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createTempHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createTempHttpResp, createTempHttpReq)
				Expect(createTempHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createTempHttpResp.Body.String()).To(ContainSubstring("should not be used as datasource"))
			}
		}, entries)
	})

	Context("autoDelete list", func() {
		DescribeTable("does not appear in list, only in get by id", func(testData AssetTest) {
			if testData.supportsAutoDeletion {
				regularAssetMeta, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				assetCreationRequest := api.AssetCreationRequest{
					Name:       "whatever",
					Scope:      api.SCOPE_CLUSTER,
					ClusterId:  convutils.Ref(randoms.ClusterId),
					AutoDelete: convutils.Ref(true),
				}
				autoDeleteMeta, _ := createAutoDeleteAsset(testData, router, randoms, &assetCreationRequest)

				//
				//  list should include only the regular asset
				//
				listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				var listResponse GenericListResponse
				err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
				Expect(err).To(BeNil())

				Expect(len(listResponse.Entries)).To(Equal(1))
				Expect(listResponse.Entries[0].Meta.Id).To(Equal(regularAssetMeta.Id))

				//
				//  get is possible for auto delete asset
				//
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, autoDeleteMeta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResponse GenericGetResponse
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResponse)
				Expect(err).To(BeNil())

				Expect(convutils.Deref(getResponse.Meta.AutoDelete, dbmodel.DefaultForAutoDelete)).To(Equal(true))
			}
		}, entries)
	})

	Context("auto delete + template", func() {
		BeforeEach(func() {
			//   disable grace time for the test, return it after test completed
			origGraceTime = dao.AutoDeleteGraceTimeSec
			setAutoDeleteGrateTime(0)
		})

		AfterEach(func() {
			setAutoDeleteGrateTime(origGraceTime)
		})

		DescribeTable("should be removed when template is gone", func(testData AssetTest) {
			if testData.supportsAutoDeletion {

				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				assetCreationRequest := api.AssetCreationRequest{
					Name:       "whatever",
					Scope:      api.SCOPE_CLUSTER,
					ClusterId:  convutils.Ref(randoms.ClusterId),
					AutoDelete: convutils.Ref(true),
				}
				autoDeleteMeta, _ := createAutoDeleteAsset(testData, router, randoms, &assetCreationRequest)

				templatePayload1 := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
				templatePayload1.Spec.Assets.WorkloadVolumes = []string{autoDeleteMeta.Id}
				templatePayload2 := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
				templatePayload2.Spec.Assets.WorkloadVolumes = []string{autoDeleteMeta.Id}

				templateIds := make([]string, 0)

				for _, templatePayload := range []api.WorkloadTemplateCreationRequest{templatePayload1, templatePayload2} {
					createTempHttpReq, err := http.NewRequest("POST", urlTemplate, makePayload(templatePayload))
					createTempHttpReq.Header = *headerForAdmin(randoms)
					Expect(err).To(BeNil())

					createTempHttpResp := httptest.NewRecorder()
					router.ServeHTTP(createTempHttpResp, createTempHttpReq)
					Expect(createTempHttpResp.Code).To(Equal(http.StatusAccepted))

					var createResp GenericGetResponse
					err = json.Unmarshal(createTempHttpResp.Body.Bytes(), &createResp)
					Expect(err).To(BeNil())

					templateIds = append(templateIds, createResp.Meta.Id)
				}

				//
				//   now delete the templates
				//
				deleted := 0

				for _, templateId := range templateIds {
					delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlTemplate, templateId), nil)
					Expect(err).To(BeNil())
					delByIdHttpReq.Header = *headerForAdmin(randoms)

					delByIdHttpResp := httptest.NewRecorder()
					router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
					Expect(delByIdHttpResp.Code).To(Equal(http.StatusAccepted))

					deleted += 1

					// now see if the auto delete asset is there
					getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, autoDeleteMeta.Id), nil)
					Expect(err).To(BeNil())
					getByIdHttpReq.Header = *headerForAdmin(randoms)

					getByIdHttpResp := httptest.NewRecorder()
					router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
					if deleted < len(templateIds) {
						Expect(getByIdHttpResp.Code).To(Equal(http.StatusOK))
					} else {
						Expect(getByIdHttpResp.Code).To(Equal(http.StatusNotFound))
					}
				}
			}
		}, entries)
	})

	Context("auto delete + workspace", func() {
		BeforeEach(func() {
			//   disable grace time for the test, return it after test completed
			origGraceTime = dao.AutoDeleteGraceTimeSec
			setAutoDeleteGrateTime(0)
		})

		AfterEach(func() {
			setAutoDeleteGrateTime(origGraceTime)
		})

		DescribeTable("should be removed when workspace is gone", func(testData AssetTest) {
			if testData.supportsAutoDeletion {
				// for this test we prefer old clusters, because we do not want workloads to be rejected for pvcs
				// or secrets missing in the cluster (with new clusters we have to set k8s objects tracker mock to
				// hold those assets, otherwise workload creation will fail
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				assetCreationRequest := api.AssetCreationRequest{
					Name:       "whatever",
					Scope:      api.SCOPE_CLUSTER,
					ClusterId:  convutils.Ref(randoms.ClusterId),
					AutoDelete: convutils.Ref(true),
				}
				autoDeleteMeta, _ := createAutoDeleteAsset(testData, router, randoms, &assetCreationRequest)

				reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
				reqPayload.Assets.WorkloadVolumes = []string{autoDeleteMeta.Id}

				createHttpReq, err := http.NewRequest("POST", urlTraining, makePayload(reqPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusOK))

				// we have to manually remove the workload...
				dbTesting.CleanWorkloadRelatedTables()

				// create another workload not using the workload volume
				reqPayload2 := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)

				createHttpReq2, err := http.NewRequest("POST", urlTraining, makePayload(reqPayload2))
				Expect(err).To(BeNil())
				createHttpReq2.Header = *headerForAdmin(randoms)

				createHttpResp2 := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp2, createHttpReq2)
				Expect(createHttpResp2.Code).To(Equal(http.StatusOK))

				// now it should be gone
				getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, autoDeleteMeta.Id), nil)
				Expect(err).To(BeNil())
				getByIdHttpReq.Header = *headerForAdmin(randoms)

				getByIdHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
				Expect(getByIdHttpResp.Code).To(Equal(http.StatusNotFound))
			}
		}, entries)
	})

	Context("asset description", func() {
		DescribeTable("Should create asset with description", func(testData AssetTest) {
			if testData.kind == "registry" {
				return
			}

			assetDescription := convutils.Ref("asset description")

			assetCreationRequest := api.AssetCreationRequest{
				Name:        "whatever",
				Scope:       api.SCOPE_PROJECT,
				ProjectId:   convutils.Ref(int32(randoms.ProjectId)),
				Description: assetDescription,
			}

			if testData.kind == api.ASSETKIND_ENVIRONMENT {
				assetCreationRequest.WorkloadSupportedTypes = &api.WorkloadSupportedTypes{
					Workspace: convutils.Ref(true),
				}
			}

			var meta api.AssetMeta
			switch testData.kind {
			case api.ASSETKIND_WORKLOAD_TEMPLATE:
				// Creation of template needs to be more precise
				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				request := api.WorkloadTemplateCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:        testutils.RandString(5),
						Description: assetDescription,
						Scope:       api.SCOPE_PROJECT,
						ProjectId:   convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.SpecificRunCreationFields{
						Assets: api.AssetsIds{
							Environment: envAsset.Meta.Id,
						},
						SpecificEnv: randomSpecificEnv(),
					},
				}
				resp, _, _ := createTemplateAsset(request, router, headerForAdmin(randoms))
				meta = resp.Meta
			case api.ASSETKIND_SECRET_VOLUME:
				// Creation of secret storage needs to create a credential as well
				secretReq := testing.BuildRandomProjectSecretVolumeAsset(randoms, nil)
				secretReq.Meta.Description = assetDescription
				secret, _, _ := createSecretVolumeAsset(&randoms.ClusterId, secretReq, router, headerForAdmin(randoms))
				meta = secret.Meta
			case api.ASSETKIND_CONFIG_MAP:
				configMapReq := testing.BuildRandomProjectConfigMapAsset(randoms)
				configMapReq.Meta.Description = assetDescription
				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithConfigMap(randoms.ClusterId, &configMapReq, nil))
				configMap, _, _ := createConfigMapAsset(configMapReq, router, headerForAdmin(randoms), http.StatusAccepted)
				meta = configMap.Meta
			case api.ASSETKIND_GENERIC_SECRET:
				genericSecretReq := buildRandomProjectGenericSecret(randoms)
				genericSecretReq.Meta.Description = assetDescription
				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithGenericSecret(randoms.ClusterId, &genericSecretReq, nil))
				genericSecret, _, _ := createGenericSecret(&randoms.ClusterId, genericSecretReq, router, headerForAdmin(randoms))
				meta = genericSecret.Meta
			default:

				postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

				createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Or(Equal(http.StatusAccepted), Equal(http.StatusOK)))

				var createResp GenericGetResponse
				err = json.Unmarshal(createHttpResp.Body.Bytes(), &createResp)
				Expect(err).To(BeNil())
				meta = createResp.Meta
			}

			Expect(*meta.Description).To(Equal(*assetDescription))
			getByIdHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", testData.url, meta.Id), nil)
			Expect(err).To(BeNil())
			getByIdHttpReq.Header = *headerForAdmin(randoms)

			getByIdHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getByIdHttpResp, getByIdHttpReq)
			Expect(getByIdHttpResp.Code).To(Equal(http.StatusOK))

			var getResponse GenericGetResponse
			err = json.Unmarshal(getByIdHttpResp.Body.Bytes(), &getResponse)
			Expect(err).To(BeNil())
			Expect(*getResponse.Meta.Description).To(Equal(*assetDescription))

			//	List the assets and check the description
			listAllHttpReq, err := http.NewRequest("GET", testData.url, nil)
			Expect(err).To(BeNil())
			listAllHttpReq.Header = *headerForAdmin(randoms)

			listAllHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listAllHttpResp, listAllHttpReq)
			Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

			var listResponse GenericListResponse
			err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
			Expect(err).To(BeNil())
			Expect(len(listResponse.Entries)).To(Equal(1))
			Expect(*listResponse.Entries[0].Meta.Description).To(Equal(*assetDescription))
		}, entries)
	})

	// Context("preset assets", func() {
	// 	DescribeTable("list for assets supporting cloning includes preset assets", func(testData AssetTest) {
	// 		if testData.supportsPresets {

	// 			//   create set of assets with tenant scope
	// 			createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
	// 			createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
	// 			createHostPathAsset(buildRandomTenantHostPathAsset(), router, headerForAdmin(randoms))
	// 			createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))
	// 			createGitAsset(buildRandomTenantGit(nil), router, headerForAdmin(randoms))
	// 			createS3Asset(buildRandomTenantS3(nil), router, headerForAdmin(randoms))

	// 			observedLogs := logtest.NewObservedLogs()

	// 			// change the tenantId of all those assets to the "fictive" tenant id that we use for presets
	// 			dbErr := testing.ChangeTenantIdForAssets(randoms.TenantId, dao.FictiveTenantId)
	// 			Expect(dbErr.Error).To(BeNil())

	// 			dao.ForceTenantCloneForTesting(randoms.TenantId)

	// 			listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", testData.url, randoms.ClusterIdParam()), nil)
	// 			Expect(err).To(BeNil())
	// 			listAllHttpReq.Header = *headerForAdmin(randoms)

	// 			listAllHttpResp := httptest.NewRecorder()
	// 			router.ServeHTTP(listAllHttpResp, listAllHttpReq)
	// 			Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

	// 			var listResponse GenericListResponse
	// 			err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
	// 			Expect(err).To(BeNil())

	// 			// We expect to see a single "preset" asset for each supported asset in turn
	// 			Expect(len(listResponse.Entries)).To(Equal(1))
	// 			// We expect to see at least one spec field
	// 			Expect(len(listResponse.Entries[0].Spec)).NotTo(Equal(0))

	// 			Expect(observedLogs.Len()).NotTo(Equal(0))
	// 			Expect(observedLogs.TakeAll()[0].Message).To(ContainSubstring("successfully cloned preset assets"))

	// 			//  2nd time we call list no preset is done

	// 			observedLogs = logtest.NewObservedLogs()

	// 			listAllHttpResp = httptest.NewRecorder()
	// 			router.ServeHTTP(listAllHttpResp, listAllHttpReq)
	// 			Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

	// 			Expect(observedLogs.Len()).To(Equal(0))
	// 		}
	// 	}, entries)
	// })

	Context("supports workload type", func() {
		DescribeTable("create should fail if there is no supportsWorkloadType for environment", func(testData AssetTest) {
			if testData.supportsWorkloadType {
				assetCreationRequest := api.AssetCreationRequest{
					Name:  "whatever",
					Scope: api.SCOPE_TENANT,
				}
				postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

				createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)

				Expect(createHttpResp.Code).To(Or(Equal(http.StatusAccepted), Equal(http.StatusBadRequest)))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("mandatory parameter workloadSupportedTypes is not provided"))

			}
		}, entries)
		DescribeTable("create asset with Workspace type", func(testData AssetTest) {
			if testData.supportsWorkloadType {
				assetCreationRequest := api.AssetCreationRequest{
					Name:  "whatever",
					Scope: api.SCOPE_TENANT,
					WorkloadSupportedTypes: &api.WorkloadSupportedTypes{
						Workspace:     convutils.Ref(true),
						Training:      convutils.Ref(false),
						Distributed:   convutils.Ref(false),
						Inference:     convutils.Ref(false),
						DistFramework: nil,
					},
				}
				postPayload := testData.factory.marshalCreationReq(assetCreationRequest)

				createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)

				Expect(createHttpResp.Code).To(Or(Equal(http.StatusAccepted), Equal(http.StatusCreated)))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("\"workspace\":true"))
			}
		}, entries)
	})

	Context("test compliance info sanity", func() {
		DescribeTable("assets supporting compliance", func(testData AssetTest) {
			if testData.testComplianceInfo {
				// for this test we prefer old clusters, because we do not want workloads to be rejected for pvcs
				// or secrets missing in the cluster (with new clusters we have to set k8s objects tracker mock to
				// hold those assets, otherwise workload creation will fail
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

				testAsset, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				anotherProject := randoms.CloneWithDifferentProject()

				policyInfo := dao.PolicyInfo{
					TenantId:     convutils.Deref(testAsset.TenantId, 0),
					WorkloadType: auth.WorkloadTypeWorkspace,
					ProjectId:    int32(randoms.ProjectId),
					PolicyTime:   policyMockUpdateTime,
				}

				//
				//  prepare good/erroneous expected string
				//
				expectedGood := api.ComplianceInfo{
					Compliance: true,
				}
				expectedGoodStr, jsonErr := json.Marshal(expectedGood)
				Expect(jsonErr).To(BeNil())

				expectedError := api.ComplianceInfo{
					Reason: []api.ComplianceInfoReason{
						{
							Field:   convutils.Ref("field"),
							Details: "error",
						},
					},
				}
				expectedErrorStr, jsonErr := json.Marshal(expectedError)
				Expect(jsonErr).To(BeNil())

				// mark it as compliant for one project
				dbErr := dao.UpsertPolicyCompliance(context.Background(), testAsset.Id, policyInfo, expectedGood)
				Expect(dbErr).To(BeNil())

				// and non-compliant for another project
				policyInfo.ProjectId = int32(anotherProject.ProjectId)

				dbErr = dao.UpsertPolicyCompliance(context.Background(), testAsset.Id, policyInfo, expectedError)
				Expect(dbErr).To(BeNil())

				for _, projectId := range []int32{int32(randoms.ProjectId), int32(anotherProject.ProjectId)} {

					var expectedCompliance string
					if projectId == int32(randoms.ProjectId) {
						expectedCompliance = string(expectedGoodStr)
					} else {
						expectedCompliance = string(expectedErrorStr)
					}

					//////////////////////////////////////////////////////////////////
					//  Test list compliance info in list
					//////////////////////////////////////////////////////////////////

					listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf(
						"%s?complyToProject=%d&complyToWorkloadType=Workspace", testData.url, projectId), nil)
					Expect(err).To(BeNil())
					listAllHttpReq.Header = *headerForAdmin(randoms)

					listAllHttpResp := httptest.NewRecorder()
					router.ServeHTTP(listAllHttpResp, listAllHttpReq)
					Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

					var listResponse GenericListResponse
					err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
					Expect(err).To(BeNil())

					// We expect to see a single asset which complies to the policy. we do more fine-tuned testing
					// in the compliance directory. here this is really the very basic sanity
					Expect(len(listResponse.Entries)).To(Equal(1))

					Expect(listAllHttpResp.Body.String()).To(ContainSubstring(expectedCompliance), expectedCompliance)

					//////////////////////////////////////////////////////////////////
					//  Test get compliance info
					//////////////////////////////////////////////////////////////////

					getHttpReq, err := http.NewRequest("GET", fmt.Sprintf(
						"%s/%s?complyToProject=%d&complyToWorkloadType=Workspace", testData.url, testAsset.Id, projectId), nil)
					Expect(err).To(BeNil())
					getHttpReq.Header = *headerForAdmin(randoms)

					getHttpResp := httptest.NewRecorder()
					router.ServeHTTP(getHttpResp, getHttpReq)
					Expect(getHttpResp.Code).To(Equal(http.StatusOK))

					var getResp GenericGetResponse
					err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
					Expect(err).To(BeNil())

					Expect(getHttpResp.Body.String()).To(ContainSubstring(expectedCompliance), expectedCompliance)
				}
			}
		}, entries)
	})

	Context("test status info sanity", func() {
		DescribeTable("assets supporting status info", func(testData AssetTest) {
			if testData.testStatusInfo {
				goodAsset, goodAssetPayload, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				badAsset, badAssetPayload, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				issues := []openapi.ClusterObjectIssue{
					{
						ScopeType: string(api.SCOPETYPE_PROJECT),
						ScopeId:   fmt.Sprintf("%d", goodAsset.ProjectId),
						Issue:     openapi.CLUSTEROBJECTISSUETYPE_REPLICATION_ERROR,
					},
				}

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().
					WithAsset(randoms.ClusterId, &badAsset, badAssetPayload, &issues).
					WithAsset(randoms.ClusterId, &goodAsset, goodAssetPayload, nil))

				//////////////////////////////////////////////////////////////////
				//  Test list compliance info in list
				//////////////////////////////////////////////////////////////////

				listAllHttpReq, err := http.NewRequest("GET", fmt.Sprintf(
					"%s?statusInfo=true", testData.url), nil)
				Expect(err).To(BeNil())
				listAllHttpReq.Header = *headerForAdmin(randoms)

				listAllHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listAllHttpResp, listAllHttpReq)
				Expect(listAllHttpResp.Code).To(Equal(http.StatusOK))

				var listResponse GenericListResponse
				err = json.Unmarshal(listAllHttpResp.Body.Bytes(), &listResponse)
				Expect(err).To(BeNil())

				Expect(len(listResponse.Entries)).To(Equal(2))

				if listResponse.Entries[0].Meta.Id == goodAsset.Id {
					Expect(listResponse.Entries[0].Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND))
					Expect(listResponse.Entries[1].Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_ISSUES_FOUND))
				} else {
					Expect(listResponse.Entries[1].Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND))
					Expect(listResponse.Entries[0].Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_ISSUES_FOUND))
				}

				//////////////////////////////////////////////////////////////////
				//  Test get compliance info - good asset
				//////////////////////////////////////////////////////////////////

				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s?statusInfo=true", testData.url, goodAsset.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp GenericGetResponse
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(getResp.Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND))

				//////////////////////////////////////////////////////////////////
				//  Test get compliance info - bad asset
				//////////////////////////////////////////////////////////////////

				getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s/%s?statusInfo=true", testData.url, badAsset.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp = httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(getResp.Status.Status).To(Equal(api.ASSETSYNCSTATUSENUM_ISSUES_FOUND))
			}
		}, entries)
	})

	Context("get asset as flat", func() {
		DescribeTable("asset supports the request", func(testData AssetTest) {
			if !testData.cannotAppearInWorkload {
				meta1, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
				testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

				payload := api.AssetFlatRequest{
					WorkloadType: api.WORKLOADTYPE_WORKSPACE,
					ProjectId:    randoms.ProjectIdStr,
					ClusterId:    randoms.ClusterId,
					AssetId:      meta1.Id,
				}

				getHttpReq, err := http.NewRequest("POST", urlToFlat, makePayload(payload))
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				// we only validate minimal support, not the content of the response
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))
			}
		}, entries)
	})

	Context("test project deletion", func() {
		DescribeTable("base flow", func(testData AssetTest) {
			_, specProj1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//  list by project shows only the correct asset
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProj1)

			_, specProj2, _ := testData.factory.createProjectBasedAsset(differentProject, headerForAdmin(randoms))
			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, differentProject.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProj2)

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err = HandleProjectDeleteRequest(ctx, &events.ProjectDeleteRequest{
				ProjectId: randoms.ProjectIdStr,
			})
			Expect(err).To(BeNil())

			var expectedProj1 []byte // expected to be nil
			var expectedProj2 = specProj2

			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, expectedProj1)
			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, differentProject.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, expectedProj2)

		}, entries)

	})

	Context("test cluster deletion", func() {
		DescribeTable("base flow", func(testData AssetTest) {
			/*
				if testData.skipClusterDeleteEventTest {
					return
				}*/

			_, specProj1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProj1)

			_, specProj2, _ := testData.factory.createProjectBasedAsset(differentProject, headerForAdmin(differentProject))
			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, differentProject.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specProj2)

			var specCluster2 []byte
			if !testData.skipClusterDeleteEventTest {
				_, specCluster1, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))
				listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&clusterId=%s", testData.url, api.SCOPE_CLUSTER, randoms.ClusterId), nil)
				Expect(err).To(BeNil())
				listProjectHttpReq.Header = *headerForAdmin(randoms)
				testData.factory.validateListContent(listProjectHttpReq, nil, specCluster1)

				_, specCluster2, _ = testData.factory.createClusterBasedAsset(differentCluster, headerForAdmin(differentCluster))
				listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&clusterId=%s", testData.url, api.SCOPE_CLUSTER, differentCluster.ClusterId), nil)
				Expect(err).To(BeNil())
				listProjectHttpReq.Header = *headerForAdmin(differentCluster)
				testData.factory.validateListContent(listProjectHttpReq, nil, specCluster2)
			}

			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err = HandleClusterDeleteEvent(ctx, &events.ClusterDelete{
				ClusterId: randoms.ClusterId,
			})
			Expect(err).To(BeNil())

			var expectedProj1 []byte    // expected to be nil
			var expectedProj2 []byte    // expected to be nil
			var expectedCluster1 []byte // expected to be nil
			var expectedCluster2 = specCluster2

			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, expectedProj1)
			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, differentProject.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, expectedProj2)

			if !testData.skipClusterDeleteEventTest {
				listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&clusterId=%s", testData.url, api.SCOPE_CLUSTER, randoms.ClusterId), nil)
				Expect(err).To(BeNil())
				listProjectHttpReq.Header = *headerForAdmin(randoms)
				testData.factory.validateListContent(listProjectHttpReq, nil, expectedCluster1)
				listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&clusterId=%s", testData.url, api.SCOPE_CLUSTER, differentCluster.ClusterId), nil)
				Expect(err).To(BeNil())
				listProjectHttpReq.Header = *headerForAdmin(differentCluster)
				testData.factory.validateListContent(listProjectHttpReq, nil, expectedCluster2)
			}

		}, entries)

	})

	Context("Test tenant deletion", func() {
		DescribeTable("Base flow", func(testData AssetTest) {
			// Ensure we only include assets that support cluster scope
			if testData.allowedScopes == nil || testData.allowedScopes == &projectScopeOnly {
				Skip(fmt.Sprintf("Asset type %s does not support cluster scope, skipping this test.", testData.kind))
			}

			// Create a cluster scoped asset
			_, specTenant1, _ := testData.factory.createClusterBasedAsset(randoms, headerForAdmin(randoms))

			// List assets by tenant to ensure it's correctly listed
			listTenantHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_CLUSTER, randoms.TenantIdStr), nil)
			Expect(err).To(BeNil())
			listTenantHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listTenantHttpReq, nil, specTenant1)

			// Delete the tenant
			ctx := context.WithValue(context.Background(), runaictx.CfTenantId, randoms.TenantIdStr)
			err = HandleTenantDeleteEvent(ctx, &events.TenantDelete{
				TenantId: randoms.TenantIdStr,
			})
			Expect(err).To(BeNil())

			var expectedTenant []byte // Tenant should be deleted (nil)

			//List assets in the cluster again to verify it's deleted
			listTenantHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_CLUSTER, randoms.TenantIdStr), nil)
			Expect(err).To(BeNil())
			listTenantHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listTenantHttpReq, nil, expectedTenant)
		}, entries)
	})

	Context("test cleanup", func() {
		DescribeTable("base flow", func(testData AssetTest) {
			if testData.supportsCleanup {
				return
			}

			_, specToRemain, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
			metaObsolete, specToObsolete, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//
			//  expecting to see two assets
			//
			listProjectHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specToObsolete, specToRemain)

			//
			//  simulate short time deletion
			//
			db := runaidb.DB().Actual()
			meta := dbmodel.AssetMetaRec{
				AssetMeta: dbmodel.AssetMeta{
					Id: metaObsolete.Id,
				},
			}
			result := db.Model(&meta).Where(&meta).Scan(&meta)
			Expect(result.Error).To(BeNil())

			meta.DeletedAt = gorm.DeletedAt{
				Time:  time.Now().AddDate(0, 0, -15),
				Valid: true,
			}

			result = db.Unscoped().Save(&meta)
			Expect(result.Error).To(BeNil())

			deleted, errors := dbcleaner.DoCleanup(context.Background(), runaidb.DB(), config.ServerConfig{
				CleanupKeepAssetsDays: 20,
			})
			Expect(len(errors)).To(Equal(0))
			Expect(deleted).To(Equal(int64(0)))

			deleted, errors = dbcleaner.DoCleanup(context.Background(), runaidb.DB(), config.ServerConfig{
				CleanupKeepAssetsDays: 15,
			})

			Expect(len(errors)).To(Equal(0))
			Expect(deleted).To(BeNumerically(">=", 0)) // meta + spec + (for environment also the supported types)

			//
			//  expecting to find one asset
			//
			listProjectHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&%s", testData.url, api.SCOPE_PROJECT, randoms.ProjectIdParam()), nil)
			Expect(err).To(BeNil())
			listProjectHttpReq.Header = *headerForAdmin(randoms)
			testData.factory.validateListContent(listProjectHttpReq, nil, specToRemain)
		}, entries)

	})

	XContext("attempt rename during sync", func() {
		DescribeTable("should fail for synced credentials", func(testData AssetTest) {
			//
			//   create the asset
			//
			meta1, spec1, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			//
			//   prepare rename request
			//
			updateReqPayload1 := testData.factory.marshalUpdateReq(spec1)

			updateHttpReq1, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, meta1.Id), strings.NewReader(
				updateReqPayload1))
			Expect(err).To(BeNil())
			updateHttpReq1.Header = *headerForAdmin(randoms)

			updateHttpResp1 := httptest.NewRecorder()

			//
			//   rename the asset - should fail for synced credentials due to sync not even started
			//
			router.ServeHTTP(updateHttpResp1, updateHttpReq1)

			if testData.renameNotAllowedDuringSync {
				Expect(updateHttpResp1.Code).To(Equal(http.StatusPreconditionFailed))
				Expect(updateHttpResp1.Body.String()).To(ContainSubstring("rename is not allowed while the asset is pending to be synchronized to the cluster"))
			} else {
				Expect(updateHttpResp1.Code).To(Or(Equal(http.StatusAccepted), Equal(http.StatusOK)))
			}

			//
			//   the rest of the test is only for synced credentials
			//
			if !testData.renameNotAllowedDuringSync {
				return
			}

			//
			//   add sync record to the sync info table (we simulate sync delete to avoid messing with the spec)
			//
			var metaForUpdate dbmodel.AssetMetaRec
			copier.Copy(&metaForUpdate.AssetMeta, &meta1)
			metaForUpdate.AssetMeta.DeletedAt.Valid = false

			syncErr := dao.UpdateSyncStatusSendToAgent(context.Background(), *meta1.TenantId, *meta1.ClusterId, []dbmodel.AssetsToSync{
				{
					AssetMetaRec: metaForUpdate,

					AssetsAccessKey:      &dbmodel.AccessKeyRecord{},
					AssetsPassword:       &dbmodel.PasswordRecord{},
					AssetsGenericSecret:  &dbmodel.GenericSecretRecord{},
					AssetsDockerRegistry: &dbmodel.DockerRegistryRecord{},
				},
			})
			Expect(syncErr).To(BeNil())

			//
			//   rename the asset - should fail for synced credentials because sync did not complete
			//
			updateReqPayload2 := testData.factory.marshalUpdateReq(spec1)

			updateHttpReq2, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, meta1.Id), strings.NewReader(
				updateReqPayload2))
			Expect(err).To(BeNil())
			updateHttpReq2.Header = *headerForAdmin(randoms)

			updateHttpResp2 := httptest.NewRecorder()

			router.ServeHTTP(updateHttpResp2, updateHttpReq2)

			Expect(updateHttpResp2.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(updateHttpResp2.Body.String()).To(ContainSubstring("allowed while the asset is being synchronized"))

			//
			//   finally sync completed
			//
			dao.HandleAgentSyncUpdates(context.Background(), *meta1.TenantId, *meta1.ClusterId, []api.ClusterSyncReportUpdate{
				{
					Id:          meta1.Id,
					UpdateCount: meta1.UpdateCount,
				},
			})

			//
			//   rename the asset - should now be ok
			//
			updateReqPayload3 := testData.factory.marshalUpdateReq(spec1)

			updateHttpReq3, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", testData.url, meta1.Id), strings.NewReader(
				updateReqPayload3))
			Expect(err).To(BeNil())
			updateHttpReq3.Header = *headerForAdmin(randoms)

			updateHttpResp3 := httptest.NewRecorder()

			router.ServeHTTP(updateHttpResp3, updateHttpReq3)

			Expect(updateHttpResp3.Code).To(Equal(http.StatusAccepted))

		}, entries)
	})

})

// No nested tables in ginkgo but this test needs to repeat for both training and workspaces
func cannotDeleteAssetBeingUsed(testData AssetTest, workloadType string, randoms *testutils.RandomValues, router chi.Router) {
	// for this test we prefer old clusters, because we do not want workloads to be rejected for pvcs
	// or secrets missing in the cluster (with new clusters we have to set k8s objects tracker mock to
	// hold those assets, otherwise workload creation will fail
	checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
		WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.OldClustersPreK8sObjectsTracker, false)))

	//   environment asset we always need
	envAsset := buildRandomTenantEnvironmentAsset(randoms)
	envAsset.Spec.UidGidSource = convutils.Ref(flat_workloads.CustomUidGid)

	env1, _, _ := createEnvironmentAsset(envAsset, router, headerForAdmin(randoms))

	createRequest := api.WorkloadCreationRequest{
		Name:      fmt.Sprintf("test_%s", workloadType),
		ClusterId: randoms.ClusterId,
		ProjectId: int32(randoms.ProjectId),
		Namespace: fmt.Sprintf("runai-proj-%d", randoms.ProjectId),
		Assets: api.AssetsIds{
			Environment: env1.Meta.Id,
		},
	}

	// now create the relevant asset that we want to add to the workload
	meta1, _, _ := testData.factory.createProjectBasedAsset(randoms, headerForAdmin(randoms))
	addedToWorkload := testData.factory.addToWorkloadCreationReq(&createRequest, &meta1)

	if addedToWorkload {
		createWsPayload, err := json.Marshal(createRequest)
		Expect(err).To(BeNil())

		// create the workload
		createWsHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], strings.NewReader(string(createWsPayload)))
		Expect(err).To(BeNil())
		createWsHttpReq.Header = *headerForAdmin(randoms)

		createWsHttpResp := httptest.NewRecorder()
		router.ServeHTTP(createWsHttpResp, createWsHttpReq)
		Expect(createWsHttpResp.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusAccepted)))

		//   delete the asset
		delByIdHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", testData.url, meta1.Id), nil)
		Expect(err).To(BeNil())
		delByIdHttpReq.Header = *headerForAdmin(randoms)

		delByIdHttpResp := httptest.NewRecorder()
		router.ServeHTTP(delByIdHttpResp, delByIdHttpReq)
		Expect(delByIdHttpResp.Code).To(Equal(http.StatusAccepted))
	}
}

func createTestWorkload(router chi.Router, workloadType string, randoms *testutils.RandomValues, envAsset api.EnvironmentAsset, assets ...api.AssetMeta) string {
	reqPayload := randomWorkloadCreationRequest(randoms, envAsset.Meta.Id)
	reqPayload.Assets.Datasources = []api.AssetIdAndKind{}

	for _, asset := range assets {
		if asset.Kind == api.ASSETKIND_ENVIRONMENT {
			reqPayload.Assets.Environment = asset.Id
		} else if asset.Kind == api.ASSETKIND_COMPUTE {
			reqPayload.Assets.Compute = convutils.Ref(asset.Id)
		} else {
			reqPayload.Assets.Datasources = append(reqPayload.Assets.Datasources, api.AssetIdAndKind{
				Kind: asset.Kind,
				Id:   asset.Id,
			})
		}
	}

	if !IsUidGidFromTheImage(envAsset.Spec.UidGidSource) {
		reqPayload.SpecificEnv = &api.SpecificRunParams{
			RunAsUid: convutils.Ref(int64(500)),
			RunAsGid: convutils.Ref(int64(30)),
		}
	}

	createHttpReq, err := http.NewRequest("POST", urlForCreatingWorkload[workloadType], makePayload(reqPayload))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(randoms)

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusOK))

	var workload dbmodel.GenericWorkload
	parseErr := json.Unmarshal(createHttpResp.Body.Bytes(), &workload)
	Expect(parseErr).To(BeNil())

	return workload.Meta.Id
}

func createAutoDeleteAsset(testData AssetTest, router chi.Router, randoms *testutils.RandomValues, request *api.AssetCreationRequest) (*api.AssetMeta, []byte) {
	postPayload := testData.factory.marshalCreationReq(*request)

	createHttpReq, err := http.NewRequest("POST", testData.url, strings.NewReader(postPayload))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(randoms)

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	respPayload := createHttpResp.Body.Bytes()

	var createResp GenericGetResponse
	err = json.Unmarshal(respPayload, &createResp)
	Expect(err).To(BeNil())

	return &createResp.Meta, respPayload
}

func setAutoDeleteGrateTime(seconds int) {
	dao.AutoDeleteGraceTimeSec = seconds
}
