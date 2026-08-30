package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

const (
	urlHostPath = "/api/v1/asset/datasource/host-path"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type HostPathTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewHostPathTestFactory(router chi.Router) *HostPathTestFactory {
	return &HostPathTestFactory{router: router}
}

func (c *HostPathTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.HostPathListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *HostPathTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.HostPathCreationRequest{
		Meta: creationMeta,
		Spec: api.HostPathSpec{
			Path:      convutils.Ref(testutils.RandString(5)),
			ReadOnly:  convutils.Ref(false),
			MountPath: convutils.Ref(testutils.RandString(5)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *HostPathTestFactory) marshalUpdateReq([]byte) string {
	request := api.HostPathUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: api.HostPathSpec{
			Path:      convutils.Ref(testutils.RandString(10)),
			ReadOnly:  convutils.Ref(false),
			MountPath: convutils.Ref(testutils.RandString(10)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *HostPathTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createHostPathAsset(buildRandomSystemHostPathAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *HostPathTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createHostPathAsset(buildRandomTenantHostPathAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *HostPathTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createHostPathAsset(buildRandomClusterHostPathAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *HostPathTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *HostPathTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createHostPathAsset(buildRandomDepartmentHostPathAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *HostPathTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.HostPathSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *HostPathTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = append(request.Assets.Datasources, api.AssetIdAndKind{
		Kind: api.ASSETKIND_HOST_PATH,
		Id:   assetMeta.Id,
	})
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomHostPatHelper() api.HostPathCreationRequest {
	return api.HostPathCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.HostPathSpec{
			Path:      convutils.Ref(testutils.RandString(5)),
			ReadOnly:  convutils.Ref(false),
			MountPath: convutils.Ref(testutils.RandString(5)),
		},
	}
}

func buildRandomProjectHostPathAsset(randoms *testutils.RandomValues) api.HostPathCreationRequest {
	ans := buildRandomHostPatHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentHostPathAsset(randoms *testutils.RandomValues) api.HostPathCreationRequest {
	ans := buildRandomHostPatHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantHostPathAsset() api.HostPathCreationRequest {
	ans := buildRandomHostPatHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterHostPathAsset(randoms *testutils.RandomValues) api.HostPathCreationRequest {
	ans := buildRandomHostPatHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemHostPathAsset() api.HostPathCreationRequest {
	ans := buildRandomHostPatHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createHostPathAsset(hostPathCreation api.HostPathCreationRequest, router chi.Router, header *http.Header) (api.HostPathAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlHostPath, makePayload(hostPathCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.HostPathAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("host path specific tests", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()
		randoms = testutils.NewRandomValues()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})
	})

	Context("delete host path asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			hostPath, _, _ := createHostPathAsset(buildRandomProjectHostPathAsset(randoms), router, headerForAdmin(randoms))

			templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_HOST_PATH,
					Id:   hostPath.Meta.Id,
				},
			}
			template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlHostPath, hostPath.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested hostPath asset is being used by template '%s'", template.Meta.Name)))
		})
	})

	Context("update host path's spec", func() {
		It("should allow changing all fields", func() {
			url := urlHostPath

			origAsset := buildRandomTenantHostPathAsset()

			origAsset.Spec = api.HostPathSpec{
				ReadOnly:         convutils.Ref(false),
				Path:             convutils.Ref("/orig/path"),
				MountPath:        convutils.Ref("/orig/mount/path"),
				MountPropagation: convutils.Ref(api.HostPathMountPropagation(api.HOSTPATHMOUNTPROPAGATION_NONE)),
			}

			createdAsset, _, _ := createHostPathAsset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.HostPathUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: api.HostPathSpec{
					ReadOnly:         convutils.Ref(true),
					Path:             convutils.Ref("/updated/path"),
					MountPath:        convutils.Ref("/updated/mount/path"),
					MountPropagation: convutils.Ref(api.HostPathMountPropagation(api.HOSTPATHMOUNTPROPAGATION_HOST_TO_CONTAINER)),
				},
			}

			updateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &updateRequest)

			//
			//  validate that the update changed the fields correctly
			//
			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &updateRequest.Spec)
		})
	})
})
