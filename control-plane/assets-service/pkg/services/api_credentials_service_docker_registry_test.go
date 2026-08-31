/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	k8sObjectsTracker "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/encryption"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
)

const (
	urlDockerRegistry = "/api/v1/asset/credentials/docker-registry"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory for DockerRegistry
/////////////////////////////////////////////////////////////////////////////////////////////

type DockerRegistryTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewDockerRegistryTestFactory(router chi.Router) *DockerRegistryTestFactory {
	return &DockerRegistryTestFactory{router: router}
}

func (c *DockerRegistryTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.DockerRegistryListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *DockerRegistryTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.DockerRegistryCreationRequest{
		Meta: creationMeta,
		Spec: api.DockerRegistryCreationSpec{
			User:     convutils.Ref(testutils.RandString(5)),
			Password: convutils.Ref(testutils.RandString(5)),
			Url:      convutils.Ref("http://runai.com/testing"),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *DockerRegistryTestFactory) marshalUpdateReq([]byte) string {
	request := api.DockerRegistryUpdateRequest{
		Meta: api.CredentialUpdateRequest{
			// Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		   Spec: &api.DockerRegistryUpdateSpec{
		   	User:     convutils.Ref(testutils.RandString(5)),
		   	Password: convutils.Ref(testutils.RandString(5)),
		   	Url:      convutils.Ref("http://runai.com/testing"),
		   },
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *DockerRegistryTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createDockerRegistry(&randoms.ClusterId, buildRandomSystemDockerRegistry(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *DockerRegistryTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createDockerRegistry(&randoms.ClusterId, buildRandomTenantDockerRegistry(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *DockerRegistryTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createDockerRegistry(&randoms.ClusterId, buildRandomClusterDockerRegistry(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *DockerRegistryTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createDockerRegistry(&randoms.ClusterId, buildRandomProjectDockerRegistry(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *DockerRegistryTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createDockerRegistry(&randoms.ClusterId, buildRandomDepartmentDockerRegistry(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *DockerRegistryTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.DockerRegistrySpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *DockerRegistryTestFactory) addToWorkloadCreationReq(*api.WorkloadCreationRequest, *api.AssetMeta) bool {
	return false // docker registry cannot be part of workload creation request
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomDockerRegistryHelper() api.DockerRegistryCreationRequest {
	return api.DockerRegistryCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.DockerRegistryCreationSpec{
			User:     convutils.Ref(testutils.RandString(10)),
			Password: convutils.Ref(testutils.RandString(10)),
			Url:      convutils.Ref("http://runai.com/test"),
		},
	}
}

func buildRandomProjectDockerRegistry(randoms *testutils.RandomValues) api.DockerRegistryCreationRequest {
	ans := buildRandomDockerRegistryHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentDockerRegistry(randoms *testutils.RandomValues) api.DockerRegistryCreationRequest {
	ans := buildRandomDockerRegistryHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantDockerRegistry() api.DockerRegistryCreationRequest {
	ans := buildRandomDockerRegistryHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterDockerRegistry(randoms *testutils.RandomValues) api.DockerRegistryCreationRequest {
	ans := buildRandomDockerRegistryHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemDockerRegistry() api.DockerRegistryCreationRequest {
	ans := buildRandomDockerRegistryHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createDockerRegistry(clusterId *string, dockerRegistryCreation api.DockerRegistryCreationRequest, router chi.Router, header *http.Header) (api.DockerRegistry, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlDockerRegistry, makePayload(dockerRegistryCreation))
	Expect(err).To(BeNil())

	if header != nil {
		createHttpReq.Header = *header
	}

	// for existing assets, make sure their cluster resource exists
	if clusterId != nil && dockerRegistryCreation.Spec.ExistingSecretName != nil {
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithDockerRegistry(*clusterId, &dockerRegistryCreation, nil))
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.DockerRegistry
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestCredentialsDockerRegistry", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test docker registry routes", func() {
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

		Context("Validate empty id is not allowed", func() {
			err := validateAssetVsCredentials(auth.MockAuthInfo(), api.SCOPE_TENANT, "", convutils.Ref(""))
			Expect(err).NotTo(BeNil())
			Expect(err.Message).To(ContainSubstring("must not be empty"))
		})

		Context("List/Get credentials asset", func() {
			It("should mask the secret", func() {
				dockerRegistry, _, _ := createDockerRegistry(&randoms.ClusterId, buildRandomProjectDockerRegistry(randoms), router, headerForAdmin(randoms))

				//
				//   List all
				//
				listHttpReq, err := http.NewRequest("GET", urlDockerRegistry, nil)
				Expect(err).To(BeNil())
				listHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var listResp api.DockerRegistryListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
				Expect(err).To(BeNil())
				Expect(len(listResp.Entries)).To(Equal(1))
				Expect(*listResp.Entries[0].Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(*listResp.Entries[0].Spec.Password).To(Equal(encryption.SecretMask))
				Expect(*listResp.Entries[0].Spec.Url).NotTo(Equal(encryption.SecretMask))

				//
				//   get by id
				//
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlDockerRegistry, dockerRegistry.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp api.DockerRegistry
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(*getResp.Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(*getResp.Spec.Password).To(Equal(encryption.SecretMask))
				Expect(*getResp.Spec.Url).NotTo(Equal(encryption.SecretMask))
			})
		})

		Context("create global (tenant) scope docker registry", func() {
			It("should succeed, tenant scope is allowed", func() {
				createPayload := buildRandomTenantDockerRegistry()
				createHttpReq, err := http.NewRequest("POST", urlDockerRegistry, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
		})
		Context("Test private docker registry", func() {

			It("base test - create a private docker registry", func() {
				creationRequest := api.DockerRegistryCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:    testutils.RandString(5),
						Scope:   api.SCOPE_TENANT,
						Private: convutils.Ref(true),
					},
					Spec: api.DockerRegistryCreationSpec{
						User:     convutils.Ref(testutils.RandString(5)),
						Password: convutils.Ref(testutils.RandString(5)),
						Url:      convutils.Ref("http://runai.com/testing"),
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlDockerRegistry, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})

			It("create a private docker registry in project scope - should fail", func() {
				creationRequest := api.DockerRegistryCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
						Private:   convutils.Ref(true),
					},
					Spec: api.DockerRegistryCreationSpec{
						User:     convutils.Ref(testutils.RandString(5)),
						Password: convutils.Ref(testutils.RandString(5)),
						Url:      convutils.Ref("http://runai.com/testing"),
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlDockerRegistry, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})

			It("test multiply flow for private docker registers(create/list/sync/update/delete)", func() {
				// template for the cred creation request
				description := testutils.RandString(10)
				creationRequest := api.DockerRegistryCreationRequest{
					Meta: api.AssetCreationRequest{
						Description: convutils.Ref(description),
						Name:        testutils.RandString(5),
						Scope:       api.SCOPE_TENANT,
						Private:     convutils.Ref(true),
					},
					Spec: api.DockerRegistryCreationSpec{
						User:     convutils.Ref(testutils.RandString(5)),
						Password: convutils.Ref(testutils.RandString(5)),
						Url:      convutils.Ref("http://runai.com/testing"),
					},
				}

				// createAndExpectDockerRegistry helper func to send cred creation request and the expected result
				createAndExpectDockerRegistry := func(creationRequest api.DockerRegistryCreationRequest, headers *http.Header, expectStatus int, msg string) string {
					createHttpReq, err := http.NewRequest("POST", urlDockerRegistry, makePayload(creationRequest))
					Expect(err).To(BeNil())
					createHttpReq.Header = *headers

					createHttpResp := httptest.NewRecorder()
					router.ServeHTTP(createHttpResp, createHttpReq)
					Expect(createHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusAccepted {
						return ""
					}
					var resp api.DockerRegistry
					respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return resp.Meta.Id
				}

				// Create a private docker register for admin (should succeed)
				private_Id1 := createAndExpectDockerRegistry(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private docker registry")
				// Create another private docker register for admin with the same name (should fail)
				createAndExpectDockerRegistry(creationRequest, headerForAdmin(randoms), http.StatusConflict, "expected to fail creating a private docker registry with the same name for the same user")
				// Create a private docker register for admin with different name (should succeed)
				creationRequest.Meta.Name = testutils.RandString(5)
				private_Id2 := createAndExpectDockerRegistry(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private docker registry with different name")
				// Create a private docker register for a different admin with the same name (should succeed)
				otherPrivate_Id := createAndExpectDockerRegistry(creationRequest, headerForOtherAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private docker register with the same name for a different user")
				// Create a "public" docker register, check there is no conflict with the private one
				creationRequest.Meta.Private = nil
				creationRequest.Spec = api.DockerRegistryCreationSpec{ // reassign the spec value, to id the asset when syncing
					User:     convutils.Ref(testutils.RandString(5)),
					Password: convutils.Ref(testutils.RandString(5)),
					Url:      convutils.Ref("http://runai.com/testing/public"),
				}
				publicId := createAndExpectDockerRegistry(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a public docker registry")

				// listAndExpectDockerRegistry is a helper function to send a list cred request and the expect ids of that request
				listAndExpectDockerRegistry := func(url string, headers *http.Header, expectedIds []string, msg string) {
					listHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					listHttpReq.Header = *headers

					listHttpResp := httptest.NewRecorder()
					router.ServeHTTP(listHttpResp, listHttpReq)
					Expect(listHttpResp.Code).To(Equal(http.StatusOK), msg)

					var listResp api.DockerRegistryListResponse
					err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
					Expect(err).To(BeNil())
					actualIds := make([]string, 0, len(listResp.Entries))
					for _, entry := range listResp.Entries {
						actualIds = append(actualIds, entry.Meta.Id)
					}
					Expect(actualIds).To(ContainElements(expectedIds), msg)
				}

				listAndExpectDockerRegistry(fmt.Sprintf("%s", urlDockerRegistry), headerForAdmin(randoms), []string{publicId}, "expected to list the only public docker registry")
				listAndExpectDockerRegistry(fmt.Sprintf("%s?privateOnly=true", urlDockerRegistry), headerForAdmin(randoms), []string{private_Id1, private_Id2}, "expected to list both private docker registries of the admin")
				listAndExpectDockerRegistry(fmt.Sprintf("%s?includePrivate=true", urlDockerRegistry), headerForAdmin(randoms), []string{private_Id1, private_Id2, publicId}, "expected to list both private docker registries and the public one")

				listAndExpectDockerRegistry(fmt.Sprintf("%s", urlDockerRegistry), headerForOtherAdmin(randoms), []string{publicId}, "expected to list the only public docker registry")
				listAndExpectDockerRegistry(fmt.Sprintf("%s?privateOnly=true", urlDockerRegistry), headerForOtherAdmin(randoms), []string{otherPrivate_Id}, "expected to list the only private docker registry of the other admin")
				listAndExpectDockerRegistry(fmt.Sprintf("%s?includePrivate=true", urlDockerRegistry), headerForOtherAdmin(randoms), []string{publicId, otherPrivate_Id}, "expected to list both private and public docker registries")

				// getDockerRegistryByIdAndExpect is a helper function to send a get a cred by id request the expect the result of that request
				getDockerRegistryByIdAndExpect := func(url string, headers *http.Header, expectStatus int, msg string) *api.DockerRegistry {
					getHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					getHttpReq.Header = *headers

					getHttpResp := httptest.NewRecorder()
					router.ServeHTTP(getHttpResp, getHttpReq)
					Expect(getHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusOK {
						return nil
					}

					var resp api.DockerRegistry
					respJsonErr := json.Unmarshal(getHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return &resp
				}

				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, publicId), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the non-private docker registry by id")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, publicId), headerForOtherAdmin(randoms), http.StatusOK, "expect to successfully get the non-private docker registry by id by the other admin")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the private docker registry by id by the admin")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect to fail getting the private docker registry by id by the other admin")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "expect to success getting the private docker registry by id by other admin")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForAdmin(randoms), http.StatusNotFound, "expect to fail getting the private docker registry by id by the other admin")

				// updateSecretAndExpect is a helper function to send a update cred by id request the expect the result of that request
				updateSecretAndExpect := func(url string, headers *http.Header, request api.GenericSecretUpdateRequest, expectStatus int, msg string) {
					updateHttpReq, err := http.NewRequest("PUT", url, makePayload(request))
					Expect(err).To(BeNil())
					updateHttpReq.Header = *headers

					updateHttpResp := httptest.NewRecorder()
					router.ServeHTTP(updateHttpResp, updateHttpReq)
					Expect(updateHttpResp.Code).To(Equal(expectStatus), msg)
				}

				newDescription := testutils.RandString(10)
				updateRequest := api.GenericSecretUpdateRequest{
					Meta: api.CredentialUpdateRequest{
						Description: convutils.Ref(newDescription),
					},
				}

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForOtherAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the secret by another user")
				privateSecret := getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(description), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), updateRequest, http.StatusAccepted, "expect to successfully update the secret")
				privateSecret = getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(newDescription), "expect the description to change")

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the secret by another user")
				privateSecret = getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, ""), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForOtherAdmin(randoms), updateRequest, http.StatusAccepted, "expect to expect to successfully update the secret")
				privateSecret = getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
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

				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect failing deleting the secret")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to find the secret")
				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusAccepted, "expect to successfully delete the secret")
				getDockerRegistryByIdAndExpect(fmt.Sprintf("%s/%s", urlDockerRegistry, private_Id1), headerForAdmin(randoms), http.StatusNotFound, "expect to not find the deleted secret")
			})
		})
	})
})
