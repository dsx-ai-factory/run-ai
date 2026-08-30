package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	policies "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

const (
	urlS3 = "/api/v1/asset/datasource/s3"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type S3TestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewS3TestFactory(router chi.Router) *S3TestFactory {
	return &S3TestFactory{router: router}
}

func (c *S3TestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.S3ListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *S3TestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.S3CreationRequest{
		Meta: creationMeta,
		Spec: api.S3AssetSpec{
			Bucket: convutils.Ref(testutils.RandString(5)),
			Path:   convutils.Ref(testutils.RandString(10)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *S3TestFactory) marshalUpdateReq(origPayload []byte) string {
	var origS3 api.S3Asset
	err := json.Unmarshal(origPayload, &origS3)
	Expect(err).To(BeNil())

	request := api.S3UpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.S3AssetSpec{
			Bucket:           convutils.Ref(testutils.RandString(5)),
			Path:             convutils.Ref(testutils.RandString(10)),
			AccessKeyAssetId: origS3.Spec.AccessKeyAssetId, // preserve access key id because it is validated for existence
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *S3TestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	s3resp, payload, checkSum := createS3Asset(buildRandomSystemS3(nil), c.router, header)
	return s3resp.Meta, payload, checkSum
}

func (c *S3TestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	s3resp, payload, checkSum := createS3Asset(buildRandomTenantS3(nil), c.router, header)
	return s3resp.Meta, payload, checkSum
}

func (c *S3TestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	s3resp, payload, checkSum := createS3Asset(buildRandomClusterS3(randoms, nil), c.router, header)
	return s3resp.Meta, payload, checkSum
}

func (c *S3TestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	s3resp, payload, checkSum := createS3Asset(buildRandomProjectS3(randoms, nil), c.router, header)
	return s3resp.Meta, payload, checkSum
}

func (c *S3TestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	s3resp, payload, checkSum := createS3Asset(buildRandomDepartmentS3(randoms, nil), c.router, header)
	return s3resp.Meta, payload, checkSum
}

func (c *S3TestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.S3AssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *S3TestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = []api.AssetIdAndKind{
		{
			Id:   assetMeta.Id,
			Kind: api.ASSETKIND_S3,
		},
	}
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomS3Helper(accessKeyAssetId *string) api.S3CreationRequest {
	return api.S3CreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.S3AssetSpec{
			Bucket:           convutils.Ref(testutils.RandString(5)),
			Path:             convutils.Ref(testutils.RandString(5)),
			AccessKeyAssetId: accessKeyAssetId,
			Url:              convutils.Ref(fmt.Sprintf("https://%s.com", testutils.RandString(5))),
		},
	}
}

func buildRandomProjectS3(randoms *testutils.RandomValues, accessKeyAssetId *string) api.S3CreationRequest {
	ans := buildRandomS3Helper(accessKeyAssetId)
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentS3(randoms *testutils.RandomValues, accessKeyAssetId *string) api.S3CreationRequest {
	ans := buildRandomS3Helper(accessKeyAssetId)
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantS3(accessKeyAssetId *string) api.S3CreationRequest {
	ans := buildRandomS3Helper(accessKeyAssetId)
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterS3(randoms *testutils.RandomValues, accessKeyAssetId *string) api.S3CreationRequest {
	ans := buildRandomS3Helper(accessKeyAssetId)
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemS3(accessKeyAssetId *string) api.S3CreationRequest {
	ans := buildRandomS3Helper(accessKeyAssetId)
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createS3Asset(s3Creation api.S3CreationRequest, router chi.Router, header *http.Header) (api.S3Asset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlS3, makePayload(s3Creation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.S3Asset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("s3 specific tests", func() {
	var randoms *testutils.RandomValues
	var differentProject *testutils.RandomValues

	router := CreateRouter()

	notExist := uuid.NewString()

	BeforeEach(func() {
		dbTesting.CleanTables()
		randoms = testutils.NewRandomValues()
		differentProject = randoms.CloneWithDifferentProject()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms, differentProject)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms, differentProject))
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

	Context("create s3 with access key from different project", func() {
		It("should fail, both should be from the same project", func() {
			accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(differentProject), router, headerForAdmin(randoms))

			createPayload := buildRandomProjectS3(randoms, &accessKey.Meta.Id)
			createHttpReq, err := http.NewRequest("POST", urlS3, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("create s3 with access key which does not exist", func() {
		It("should fail, both should be from the same project", func() {
			createPayload := buildRandomProjectS3(randoms, convutils.Ref(notExist))
			createHttpReq, err := http.NewRequest("POST", urlS3, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Context("create global scope unauthenticated s3", func() {
		It("should succeed, public s3 are allowed to be global", func() {
			createPayload := buildRandomTenantS3(nil)
			createHttpReq, err := http.NewRequest("POST", urlS3, makePayload(createPayload))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("delete S3 asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
			s3, _, _ := createS3Asset(buildRandomProjectS3(randoms, &accessKey.Meta.Id), router, headerForAdmin(randoms))

			templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_S3,
					Id:   s3.Meta.Id,
				},
			}
			template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlS3, s3.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested s3 asset is being used by template '%s'", template.Meta.Name)))
		})
	})

	Context("update S3 asset's spec", func() {
		It("should allow changing all fields", func() {
			url := urlS3

			origAsset := buildRandomTenantS3(nil)

			origAsset.Spec = api.S3AssetSpec{
				Path:   convutils.Ref("/orig/path"),
				Bucket: convutils.Ref("/orig/bucket"),
				Url:    convutils.Ref("http://orig.url.com"),
			}

			createdAsset, _, _ := createS3Asset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.S3UpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: api.S3AssetSpec{
					Path:   convutils.Ref("/updated/path"),
					Bucket: convutils.Ref("/updated/bucket"),
					Url:    convutils.Ref("http://updated.url.com"),
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
