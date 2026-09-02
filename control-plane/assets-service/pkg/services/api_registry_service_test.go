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

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const (
	urlRegistry  = "/api/v1/asset/registries"
	fakeUser     = "admin"
	fakePassword = "password"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type RegistryTestFactory struct {
	_          AssetTestFactory
	router     chi.Router
	mockServer *httptest.Server
}

type Catalog struct {
	Repositories []string
}

type TagList struct {
	Name string
	Tags []string
}

func newRegistryMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Docker Registry HTTP API v2\n"))
		case "/v2/_catalog":
			catalog := Catalog{
				Repositories: []string{"nginx", "mysql", "redis"},
			}
			json, err := json.Marshal(catalog)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(json)
		case "/v2/nginx/tags/list":
			tags := TagList{
				Name: "nginx",
				Tags: []string{"latest", "1.21.0", "1.20.1"},
			}
			json, err := json.Marshal(tags)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(json)
		default:
			http.NotFound(w, r)
			return
		}
	}))
}

func NewRegistryTestFactory(router chi.Router) *RegistryTestFactory {
	return &RegistryTestFactory{router: router, mockServer: newRegistryMockServer()}
}

func (c *RegistryTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.RegistryListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *RegistryTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.RegistryCreationRequest{
		Meta: creationMeta,
		Spec: api.RegistrySpec{
			Url:            c.mockServer.URL,
			CredentialKind: api.CREDENTIALKIND_PASSWORD,
			User:           fakeUser,
			Password:       fakePassword,
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *RegistryTestFactory) marshalUpdateReq([]byte) string {
	request := api.RegistryUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		   Spec: api.RegistrySpec{
		   	Url:            c.mockServer.URL,
		   	CredentialKind: api.CREDENTIALKIND_PASSWORD,
		   	User:           fakeUser,
		   	Password:       fakePassword,
		   },
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

// registry scope is always tenant, we implement the rest just for completeness, as the interface requires
func (c *RegistryTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, _ *http.Header) (api.AssetMeta, []byte, []byte) {
	registry, payload, checkSum := createRegistry(buildRegistryCreationRequest(c.mockServer.URL), randoms, c.router, nil, APITestOptions{})
	return registry.Meta, payload, checkSum
}

func (c *RegistryTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	return c.createTenantBasedAsset(randoms, header)
}

func (c *RegistryTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	return c.createTenantBasedAsset(randoms, header)
}

func (c *RegistryTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	return c.createTenantBasedAsset(randoms, header)
}

func (c *RegistryTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	return c.createTenantBasedAsset(randoms, header)
}

func (c *RegistryTestFactory) validateListContent(_ *http.Request, _ datastructs.Set[string], createdSpecs ...[]byte) {
}

func (c *RegistryTestFactory) addToWorkloadCreationReq(*api.WorkloadCreationRequest, *api.AssetMeta) bool {
	return false // registry asset cannot be part of workspace creation request
}

// ///////////////////////////////////////////////////////////////////////////////////////////
//
//	Asset Builders
//
// ///////////////////////////////////////////////////////////////////////////////////////////
func buildRegistrySpec(registryUrl string) api.RegistrySpec {
	return api.RegistrySpec{
		Url:            registryUrl,
		CredentialKind: api.CREDENTIALKIND_PASSWORD,
		User:           fakeUser,
		Password:       fakePassword,
	}
}

func buildRegistryCreationRequest(registryUrl string) api.RegistryCreationRequest {
	return api.RegistryCreationRequest{
		Meta: api.AssetCreationRequest{
			Name:  testutils.RandString(5),
			Scope: api.SCOPE_TENANT,
		},
		Spec: buildRegistrySpec(registryUrl),
	}
}

func buildRegistryUpdateRequest(registryUrl string) api.RegistryUpdateRequest {
	return api.RegistryUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name: testutils.RandString(5),
		},
		/* At this point we are not supporting spec update
		   Spec: buildRegistrySpec(registryUrl),
		*/
	}
}

type APITestOptions struct {
	shouldFail         bool
	expectedStatusCode int
}

func createRegistry(creationRequest api.RegistryCreationRequest, randoms *testutils.RandomValues, router chi.Router, header *http.Header, options APITestOptions) (api.Registry, []byte, []byte) {
	createHttpReq, err := http.NewRequest(http.MethodPost, urlRegistry, makePayload(creationRequest))
	Expect(err).To(BeNil())

	if header != nil {
		createHttpReq.Header = *header
	} else {
		createHttpReq.Header = *headerForAdmin(randoms)
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	if options.shouldFail {
		Expect(createHttpResp.Code).To(Equal(options.expectedStatusCode))
		return api.Registry{}, nil, nil
	}

	Expect(createHttpResp.Code).To(Equal(http.StatusCreated))

	var resp api.Registry
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

func updateRegistry(updateRequest api.RegistryUpdateRequest, id string, randoms *testutils.RandomValues, router chi.Router, header *http.Header, options APITestOptions) api.Registry {
	updateHttpReq, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s", urlRegistry, id), makePayload(updateRequest))
	Expect(err).To(BeNil())

	if header != nil {
		updateHttpReq.Header = *header
	} else {
		updateHttpReq.Header = *headerForAdmin(randoms)
	}

	updateHttpResp := httptest.NewRecorder()
	router.ServeHTTP(updateHttpResp, updateHttpReq)
	if options.shouldFail {
		Expect(updateHttpResp.Code).To(Equal(options.expectedStatusCode))
		return api.Registry{}
	}

	Expect(updateHttpResp.Code).To(Equal(http.StatusOK))

	var resp api.Registry
	respJsonErr := json.Unmarshal(updateHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp
}

func deleteRegistry(id string, randoms *testutils.RandomValues, router chi.Router, header *http.Header, options APITestOptions) {
	deleteHttpReq, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s", urlRegistry, id), nil)
	Expect(err).To(BeNil())

	if header != nil {
		deleteHttpReq.Header = *header
	} else {
		deleteHttpReq.Header = *headerForAdmin(randoms)
	}

	deleteHttpResp := httptest.NewRecorder()
	router.ServeHTTP(deleteHttpResp, deleteHttpReq)
	if options.shouldFail {
		Expect(deleteHttpResp.Code).To(Equal(options.expectedStatusCode))
	} else {
		Expect(deleteHttpResp.Code).To(Equal(http.StatusNoContent))
	}
}

var _ = Describe("TestRegistry", Ordered, func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()
	mockServer := newRegistryMockServer()

	AfterAll(func() {
		mockServer.Close()
	})
	Describe("Test registry routes", func() {
		BeforeEach(func() {
			dbTesting.CleanTables()
			randoms = testutils.NewRandomValues()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms))
			authorization.SetMock(authMock)
		})

		Context("Test Crete registry with wrong parameters", func() {
			It("Project scope is not supported for Registries", func() {
				request := buildRegistryCreationRequest(mockServer.URL)
				request.Meta.Scope = api.SCOPE_PROJECT
				request.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
				request.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
				createRegistry(request, randoms, router, nil, APITestOptions{
					shouldFail:         true,
					expectedStatusCode: http.StatusBadRequest,
				})
			})
			It("Wrong url should fail", func() {
				request := buildRegistryCreationRequest(mockServer.URL)
				request.Spec.Url = testutils.RandString(16)
				createRegistry(request, randoms, router, nil, APITestOptions{
					shouldFail:         true,
					expectedStatusCode: http.StatusBadRequest,
				})
			})
		})

		Context("Tests with existing registry", func() {
			var createdRegistry *api.Registry
			BeforeEach(func() {
				registry, _, _ := createRegistry(buildRegistryCreationRequest(mockServer.URL), randoms, router, nil, APITestOptions{})
				createdRegistry = &registry
			})

			AfterEach(func() {
				if createdRegistry != nil {
					deleteRegistry(createdRegistry.Meta.Id, randoms, router, nil, APITestOptions{})
				}
			})

			/* At this point we are not supporting spec update
			   It("Update with wrong url should fail", func() {
			   	request := buildRegistryUpdateRequest(mockServer.URL)
			   	request.Spec.Url = testutils.RandString(16)
			   	updateRegistry(request, createdRegistry.Meta.Id, randoms, router, nil, APITestOptions{
			   		shouldFail:         true,
			   		expectedStatusCode: http.StatusBadRequest,
			   	})
			   })
			*/

			It("Should not list the registry password in passwords", func() {
				listPasswordsHttpReq, err := http.NewRequest(http.MethodGet, urlPassword, nil)
				Expect(err).To(BeNil())
				listPasswordsHttpReq.Header = *headerForAdmin(randoms)

				listPasswordsHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listPasswordsHttpResp, listPasswordsHttpReq)
				Expect(listPasswordsHttpResp.Code).To(Equal(http.StatusOK))

				var listPasswordsResp api.PasswordListResponse
				err = json.Unmarshal(listPasswordsHttpResp.Body.Bytes(), &listPasswordsResp)
				Expect(err).To(BeNil())
				Expect(len(listPasswordsResp.Entries)).To(Equal(0))

				listRegistriesHttpReq, err := http.NewRequest(http.MethodGet, urlRegistry, nil)
				Expect(err).To(BeNil())
				listRegistriesHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listRegistriesHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var listRegistriesResp api.RegistryListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &listRegistriesResp)
				Expect(err).To(BeNil())
				Expect(len(listRegistriesResp.Entries)).To(Equal(1))
				Expect(listRegistriesResp.Entries[0].Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(listRegistriesResp.Entries[0].Spec.Password).To(Equal(encryption.SecretMask))
			})

			It("Should mask password on Get", func() {
				getHttpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", urlRegistry, createdRegistry.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp api.Registry
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(getResp.Spec.User).NotTo(Equal(encryption.SecretMask))
				Expect(getResp.Spec.Password).To(Equal(encryption.SecretMask))
			})

			It("Update with wrong permissions", func() {
				request := buildRegistryUpdateRequest(mockServer.URL)
				updateRegistry(request, createdRegistry.Meta.Id, randoms, router, headerWithMissingInfo(randoms), APITestOptions{
					shouldFail:         true,
					expectedStatusCode: http.StatusForbidden,
				})
			})

			It("Delete with wrong permissions", func() {
				deleteRegistry(createdRegistry.Meta.Id, randoms, router, headerWithMissingInfo(randoms), APITestOptions{
					shouldFail:         true,
					expectedStatusCode: http.StatusForbidden,
				})
			})

			It("Create with wrong permissions", func() {
				request := buildRegistryCreationRequest(mockServer.URL)
				createRegistry(request, randoms, router, headerWithMissingInfo(randoms), APITestOptions{
					shouldFail:         true,
					expectedStatusCode: http.StatusForbidden,
				})
			})
		})

		Context("Repositories and Tags", func() {
			var createdRegistry *api.Registry
			BeforeEach(func() {
				registry, _, _ := createRegistry(buildRegistryCreationRequest(mockServer.URL), randoms, router, nil, APITestOptions{})
				createdRegistry = &registry
			})

			AfterEach(func() {
				if createdRegistry != nil {
					deleteRegistry(createdRegistry.Meta.Id, randoms, router, nil, APITestOptions{})
				}
			})

			It("Get Repositories and Tags", func() {
				repositoriesHttpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/repositories", urlRegistry, createdRegistry.Meta.Id), nil)
				Expect(err).To(BeNil())
				repositoriesHttpReq.Header = *headerForAdmin(randoms)

				repositoriesHttpResp := httptest.NewRecorder()
				router.ServeHTTP(repositoriesHttpResp, repositoriesHttpReq)
				Expect(repositoriesHttpResp.Code).To(Equal(http.StatusOK))

				var repositories api.Repositories
				err = json.Unmarshal(repositoriesHttpResp.Body.Bytes(), &repositories)
				Expect(err).To(BeNil())
				Expect(repositories.Repositories).NotTo(BeNil())

				if len(repositories.Repositories) > 0 {
					repoName := repositories.Repositories[0]
					tagsHttpReq, err := http.NewRequest(
						http.MethodGet,
						fmt.Sprintf("%s/%s/repositories/tags?repository=%s",
							urlRegistry,
							createdRegistry.Meta.Id,
							repoName),
						nil)
					Expect(err).To(BeNil())
					tagsHttpReq.Header = *headerForAdmin(randoms)

					tagsHttpResp := httptest.NewRecorder()
					router.ServeHTTP(tagsHttpResp, tagsHttpReq)
					Expect(tagsHttpResp.Code).To(Equal(http.StatusOK))

					var tags api.RepositoryTags
					err = json.Unmarshal(tagsHttpResp.Body.Bytes(), &tags)
					Expect(err).To(BeNil())
					Expect(tags.Name).To(Equal(repoName))
					Expect(tags.Tags).ToNot(BeNil())
				}
			})
			It("Get Repositories without registry", func() {
				uuid, _ := uuid.NewRandom()
				repositoriesHttpReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s/repositories", urlRegistry, uuid), nil)
				Expect(err).To(BeNil())
				repositoriesHttpReq.Header = *headerForAdmin(randoms)

				repositoriesHttpResp := httptest.NewRecorder()
				router.ServeHTTP(repositoriesHttpResp, repositoriesHttpReq)
				Expect(repositoriesHttpResp.Code).To(Equal(http.StatusNotFound))
			})
			It("Get Tags without registry", func() {
				uuid, _ := uuid.NewRandom()
				repoName := testutils.RandString(16)
				tagsHttpReq, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("%s/%s/repositories/tags?repository=%s",
						urlRegistry,
						uuid,
						repoName),
					nil)
				Expect(err).To(BeNil())
				tagsHttpReq.Header = *headerForAdmin(randoms)

				tagsHttpResp := httptest.NewRecorder()
				router.ServeHTTP(tagsHttpResp, tagsHttpReq)
				Expect(tagsHttpResp.Code).To(Equal(http.StatusNotFound))
			})
			It("Get Tags with wrong repository", func() {
				repoName := testutils.RandString(16)
				tagsHttpReq, err := http.NewRequest(
					http.MethodGet,
					fmt.Sprintf("%s/%s/repositories/tags?repository=%s",
						urlRegistry,
						createdRegistry.Meta.Id,
						repoName),
					nil)
				Expect(err).To(BeNil())
				tagsHttpReq.Header = *headerForAdmin(randoms)

				tagsHttpResp := httptest.NewRecorder()
				router.ServeHTTP(tagsHttpResp, tagsHttpReq)
				Expect(tagsHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
