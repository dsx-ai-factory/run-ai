package services

import (
	"encoding/json"
	"fmt"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"net/http"
	"net/http/httptest"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
)

const (
	urlNgcApiKey      = "/api/v1/asset/credentials/ngc-api-key"
	internalNgcApiKey = "/internal/api/v1/asset/ngc-api-key"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type NgcApiKeyTestFactory struct {
	router chi.Router
	gt     GenericTest[api.NgcApiKeySpec]
}

func NewNgcApiKeyTestFactory(router chi.Router) *NgcApiKeyTestFactory {
	return &NgcApiKeyTestFactory{router: router}
}

func (f *NgcApiKeyTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listResponse api.NgcApiKeyListResponse
	err := json.Unmarshal(bytes, &listResponse)
	if err != nil {
		return nil
	}

	var result []api.AssetMeta
	for _, entry := range listResponse.Entries {
		result = append(result, entry.Meta)
	}
	return result
}

func (f *NgcApiKeyTestFactory) marshalCreationReq(request api.AssetCreationRequest) string {
	creationReq := api.NgcApiKeyCreationRequest{
		Meta: request,
		Spec: api.NgcApiKeyCreationSpec{
			NgcApiKey: testutils.RandString(5),
		},
	}

	bytes, err := json.Marshal(creationReq)
	Expect(err).To(BeNil())
	return string(bytes)
}

func (f *NgcApiKeyTestFactory) marshalUpdateReq(createdSpec []byte) string {
	var createdAsset api.NgcApiKey
	err := json.Unmarshal(createdSpec, &createdAsset)
	if err != nil {
		return ""
	}

	updateReq := api.NgcApiKeyUpdateRequest{
		Meta: api.CredentialUpdateRequest{
			/* At this point we are not supporting spec update */
			Description: convutils.Ref(testutils.RandString(5)),
		},
	}

	bytes, err := json.Marshal(updateReq)
	Expect(err).To(BeNil())
	return string(bytes)
}

func (f *NgcApiKeyTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNgcApiKey(buildRandomSystemNgcApiKey(), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *NgcApiKeyTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNgcApiKey(buildRandomTenantNgcApiKey(), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *NgcApiKeyTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNgcApiKey(buildRandomClusterNgcApiKey(randoms), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *NgcApiKeyTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNgcApiKey(buildRandomDepartmentNgcApiKey(randoms), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *NgcApiKeyTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNgcApiKey(buildRandomProjectNgcApiKey(randoms), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *NgcApiKeyTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.NgcApiKeySpec]{}
	tg.validateListContent(f.router, listRequest, expectedUsedBy, createdSpecs)
}

func (f *NgcApiKeyTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	return false // NGC API Key cannot be directly used in workloads
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomNgcApiKeyHelper() api.NgcApiKeyCreationRequest {
	return api.NgcApiKeyCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.NgcApiKeyCreationSpec{
			NgcApiKey: testutils.RandString(10),
		},
	}
}

func buildRandomProjectNgcApiKey(randoms *testutils.RandomValues) api.NgcApiKeyCreationRequest {
	ans := buildRandomNgcApiKeyHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentNgcApiKey(randoms *testutils.RandomValues) api.NgcApiKeyCreationRequest {
	ans := buildRandomNgcApiKeyHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantNgcApiKey() api.NgcApiKeyCreationRequest {
	ans := buildRandomNgcApiKeyHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterNgcApiKey(randoms *testutils.RandomValues) api.NgcApiKeyCreationRequest {
	ans := buildRandomNgcApiKeyHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemNgcApiKey() api.NgcApiKeyCreationRequest {
	ans := buildRandomNgcApiKeyHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createNgcApiKey(requset api.NgcApiKeyCreationRequest, router chi.Router, header *http.Header) (api.NgcApiKey, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlNgcApiKey, makePayload(requset))
	Expect(err).To(BeNil())

	if header != nil {
		createHttpReq.Header = *header
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.NgcApiKey
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestCredentialsNgcApiKey", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test ngc api key", func() {
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
		})

		Context("Test private ngc api key", func() {
			It("base test - create a ngc api key", func() {
				creationRequest := api.NgcApiKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:    testutils.RandString(5),
						Scope:   api.SCOPE_TENANT,
						Private: convutils.Ref(true),
					},
					Spec: api.NgcApiKeyCreationSpec{
						NgcApiKey: testutils.RandString(10),
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlNgcApiKey, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})

			It("create a private ngc api key in project scope - should fail", func() {
				creationRequest := api.NgcApiKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:  testutils.RandString(5),
						Scope: api.SCOPE_PROJECT,
					},
					Spec: api.NgcApiKeyCreationSpec{
						NgcApiKey: testutils.RandString(10),
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlNgcApiKey, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})

			It("test multiply flow for private ngc api key(create/list/sync/update/delete)", func() {
				// template for the cred creation request
				description := testutils.RandString(10)
				creationRequest := api.NgcApiKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:        testutils.RandString(5),
						Scope:       api.SCOPE_TENANT,
						Private:     convutils.Ref(true),
						Description: convutils.Ref(description),
					},
					Spec: api.NgcApiKeyCreationSpec{
						NgcApiKey: testutils.RandString(10),
					},
				}

				// createAndExpectNgcApiKey helper func to send cred creation request and the expected result
				createAndExpectNgcApiKey := func(creationRequest api.NgcApiKeyCreationRequest, headers *http.Header, expectStatus int, msg string) string {
					createHttpReq, err := http.NewRequest("POST", urlNgcApiKey, makePayload(creationRequest))
					Expect(err).To(BeNil())
					createHttpReq.Header = *headers

					createHttpResp := httptest.NewRecorder()
					router.ServeHTTP(createHttpResp, createHttpReq)
					Expect(createHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusAccepted {
						return ""
					}
					var resp api.NgcApiKey
					respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return resp.Meta.Id
				}

				// Create a private ngc api key for admin (should succeed)
				private_Id1 := createAndExpectNgcApiKey(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private ngc api key")
				// Create another private ngc api key for admin with the same name (should fail)
				createAndExpectNgcApiKey(creationRequest, headerForAdmin(randoms), http.StatusConflict, "expected to fail creating a private ngc api key with the same name for the same user")
				// Create a private ngc api keyfor admin with different (should succeed)
				creationRequest.Meta.Name = testutils.RandString(5)
				private_Id2 := createAndExpectNgcApiKey(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private ngc api key with different name")
				// Create a private ngc api key for a different admin with the same name (should succeed)
				otherPrivate_Id := createAndExpectNgcApiKey(creationRequest, headerForOtherAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private ngc api key with the same name for a different user")

				// Create "public" private ngc api key use for list testing
				creationRequest.Meta.Private = nil
				creationRequest.Meta.Scope = api.SCOPE_PROJECT
				creationRequest.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
				publicId := createAndExpectNgcApiKey(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a public ngc api key")

				// listAndExpectNgcApiKey is a helper function to send a list cred request and the expect ids of that request
				listAndExpectNgcApiKey := func(url string, headers *http.Header, expectedIds []string, msg string) {
					listHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					listHttpReq.Header = *headers

					listHttpResp := httptest.NewRecorder()
					router.ServeHTTP(listHttpResp, listHttpReq)
					Expect(listHttpResp.Code).To(Equal(http.StatusOK), msg)

					var listResp api.NgcApiKeyListResponse
					err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
					Expect(err).To(BeNil())
					actualIds := make([]string, 0, len(listResp.Entries))
					for _, entry := range listResp.Entries {
						actualIds = append(actualIds, entry.Meta.Id)
					}
					Expect(actualIds).To(ContainElements(expectedIds), msg)
				}

				listAndExpectNgcApiKey(fmt.Sprintf("%s", urlNgcApiKey), headerForAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectNgcApiKey(fmt.Sprintf("%s?privateOnly=true", urlNgcApiKey), headerForAdmin(randoms), []string{private_Id1, private_Id2}, "expected to list both private secrets of the admin")
				listAndExpectNgcApiKey(fmt.Sprintf("%s?includePrivate=true", urlNgcApiKey), headerForAdmin(randoms), []string{private_Id1, private_Id2, publicId}, "expected to list both private secrets and the public one")

				listAndExpectNgcApiKey(fmt.Sprintf("%s", urlNgcApiKey), headerForOtherAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectNgcApiKey(fmt.Sprintf("%s?privateOnly=true", urlNgcApiKey), headerForOtherAdmin(randoms), []string{otherPrivate_Id}, "expected to list the only private secret of the other admin")
				listAndExpectNgcApiKey(fmt.Sprintf("%s?includePrivate=true", urlNgcApiKey), headerForOtherAdmin(randoms), []string{publicId, otherPrivate_Id}, "expected to list both private and public secret")

				// listAndExpectCredentials is a helper function to send a list cred request and the expect ids of that request
				listAndExpectCredentials := func(url string, headers *http.Header, expectedIds []string, msg string) {
					listHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					listHttpReq.Header = *headers

					listHttpResp := httptest.NewRecorder()
					router.ServeHTTP(listHttpResp, listHttpReq)
					Expect(listHttpResp.Code).To(Equal(http.StatusOK), msg)

					var listResp api.CredentialsListResponse
					err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
					Expect(err).To(BeNil())
					actualIds := make([]string, 0, len(listResp.Entries))
					for _, entry := range listResp.Entries {
						actualIds = append(actualIds, entry.Meta.Id)
					}
					Expect(actualIds).To(ContainElements(expectedIds), msg)
				}
				listAndExpectCredentials(fmt.Sprintf("%s", urlNgcApiKey), headerForAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectCredentials(fmt.Sprintf("%s?privateOnly=true", urlNgcApiKey), headerForAdmin(randoms), []string{private_Id1, private_Id2}, "expected to list both private secrets of the admin")
				listAndExpectCredentials(fmt.Sprintf("%s?includePrivate=true", urlNgcApiKey), headerForAdmin(randoms), []string{private_Id1, private_Id2, publicId}, "expected to list both private secrets and the public one")

				listAndExpectCredentials(fmt.Sprintf("%s", urlNgcApiKey), headerForOtherAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectCredentials(fmt.Sprintf("%s?privateOnly=true", urlNgcApiKey), headerForOtherAdmin(randoms), []string{otherPrivate_Id}, "expected to list the only private secret of the other admin")
				listAndExpectCredentials(fmt.Sprintf("%s?includePrivate=true", urlNgcApiKey), headerForOtherAdmin(randoms), []string{publicId, otherPrivate_Id}, "expected to list both private and public secret")

				// getNgcApiKeyByIdAndExpect is a helper function to send a get a cred by id request the expect the result of that request
				getNgcApiKeyByIdAndExpect := func(url string, headers *http.Header, expectStatus int, msg string) *api.NgcApiKey {
					getHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					getHttpReq.Header = *headers

					getHttpResp := httptest.NewRecorder()
					router.ServeHTTP(getHttpResp, getHttpReq)
					Expect(getHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusOK {
						return nil
					}
					var resp api.NgcApiKey
					respJsonErr := json.Unmarshal(getHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return &resp
				}

				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, publicId), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the non-private generic secret by id")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, publicId), headerForOtherAdmin(randoms), http.StatusOK, "expect to successfully get the non-private generic secret by id by the other admin")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the private generic secret by id by the admin")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect to fail getting the private generic secret by id by the other admin")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "expect to success getting the private generic secret by id by other admin")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForAdmin(randoms), http.StatusNotFound, "expect to fail getting the private generic secret by id by the other admin")

				// updateSecretAndExpect is a helper function to send a update cred by id request the expect the result of that request
				updateSecretAndExpect := func(url string, headers *http.Header, request api.NgcApiKeyUpdateRequest, expectStatus int, msg string) {
					updateHttpReq, err := http.NewRequest("PUT", url, makePayload(request))
					Expect(err).To(BeNil())
					updateHttpReq.Header = *headers

					updateHttpResp := httptest.NewRecorder()
					router.ServeHTTP(updateHttpResp, updateHttpReq)
					Expect(updateHttpResp.Code).To(Equal(expectStatus), msg)
				}

				newDescription := testutils.RandString(10)
				updateRequest := api.NgcApiKeyUpdateRequest{
					Meta: api.CredentialUpdateRequest{
						Description: convutils.Ref(newDescription),
					},
				}

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForOtherAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the private docker registry by id")
				privateSecret := getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(description), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), updateRequest, http.StatusAccepted, "expect to update the secret")
				privateSecret = getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(newDescription), "expect the description to change")

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the private docker registry by id")
				privateSecret = getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, ""), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForOtherAdmin(randoms), updateRequest, http.StatusAccepted, "expect to fail getting the private docker registry by id")
				privateSecret = getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(newDescription), "expect the description to change")

				// deleteSecretAndExpect is a helper function to send a delete cred by id request the expect the result of that request
				deleteSecretAndExpect := func(url string, headers *http.Header, expectStatus int, msg string) {
					deleteHttpReq, err := http.NewRequest("DELETE", url, nil)
					Expect(err).To(BeNil())
					deleteHttpReq.Header = *headers

					deleteHttpResp := httptest.NewRecorder()
					router.ServeHTTP(deleteHttpResp, deleteHttpReq)
					Expect(deleteHttpResp.Code).To(Equal(expectStatus), msg)
				}

				// fail to delete the private asset by another admin
				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect failing deleting the private docker registry by id")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to find the secret")
				// success to delete the private asset by the admin
				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusAccepted, "expect to successfully delete the secret")
				getNgcApiKeyByIdAndExpect(fmt.Sprintf("%s/%s", urlNgcApiKey, private_Id1), headerForAdmin(randoms), http.StatusNotFound, "expect to not find the deleted secret")
			})
		})

		Context("Test internal ngc api key", func() {
			It("test get ngc api key by id internal", func() {
				creationRequest := api.NgcApiKeyCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     clusterScope,
						ClusterId: &randoms.ClusterId,
					},
					Spec: api.NgcApiKeyCreationSpec{
						NgcApiKey: "nvapi-key",
					},
				}
				createHttpReq, err := http.NewRequest("POST", urlNgcApiKey, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

				var createResp api.NgcApiKey
				respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createResp)

				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", internalNgcApiKey, createResp.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var resp api.NgcApiKeyInternal
				respJsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &resp)
				Expect(respJsonErr).To(BeNil())
				Expect(resp.Key).To(Equal(creationRequest.Spec.NgcApiKey))

				getHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s/%s", urlNgcApiKey, createResp.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForOtherAdmin(randoms)

				getHttpResp = httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))
			})
		})
	})
})
