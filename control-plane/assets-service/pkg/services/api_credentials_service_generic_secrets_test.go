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
	"strings"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
)

const urlGenericSecret = "/api/v1/asset/credentials/generic-secret"

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type GenericSecretsTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewGenericSecretsTestFactory(router chi.Router) *GenericSecretsTestFactory {
	return &GenericSecretsTestFactory{router: router}
}

func (c *GenericSecretsTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.GenericSecretListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *GenericSecretsTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.GenericSecretCreationRequest{
		Meta: creationMeta,
		Spec: api.GenericSecretCreationSpec{
			ExistingSecretName: convutils.Ref(testutils.RandString(5)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *GenericSecretsTestFactory) marshalUpdateReq([]byte) string {
	request := api.GenericSecretUpdateRequest{
		Meta: api.CredentialUpdateRequest{
			// Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		           Spec: &api.GenericSecretUpdateSpec{
		   			ExistingSecretName: convutils.Ref(testutils.RandString(5)),
		   		},
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *GenericSecretsTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGenericSecret(&randoms.ClusterId, buildRandomSystemGenericSecret(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GenericSecretsTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGenericSecret(&randoms.ClusterId, buildRandomTenantGenericSecret(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GenericSecretsTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGenericSecret(&randoms.ClusterId, buildRandomClusterGenericSecret(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GenericSecretsTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGenericSecret(&randoms.ClusterId, buildRandomDepartmentGenericSecret(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GenericSecretsTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createGenericSecret(&randoms.ClusterId, buildRandomProjectGenericSecret(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *GenericSecretsTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.GenericSecretSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *GenericSecretsTestFactory) addToWorkloadCreationReq(*api.WorkloadCreationRequest, *api.AssetMeta) bool {
	return false // generic secret asset cannot be part of workload creation request
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomGenericSecretHelper() api.GenericSecretCreationRequest {
	return api.GenericSecretCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.GenericSecretCreationSpec{
			ExistingSecretName: convutils.Ref(strings.ToLower(testutils.RandString(10))),
		},
	}
}

func buildRandomProjectGenericSecret(randoms *testutils.RandomValues) api.GenericSecretCreationRequest {
	ans := buildRandomGenericSecretHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentGenericSecret(randoms *testutils.RandomValues) api.GenericSecretCreationRequest {
	ans := buildRandomGenericSecretHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantGenericSecret() api.GenericSecretCreationRequest {
	ans := buildRandomGenericSecretHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterGenericSecret(randoms *testutils.RandomValues) api.GenericSecretCreationRequest {
	ans := buildRandomGenericSecretHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemGenericSecret() api.GenericSecretCreationRequest {
	ans := buildRandomGenericSecretHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createGenericSecret(clusterId *string, genericSecretCreation api.GenericSecretCreationRequest, router chi.Router, header *http.Header) (api.GenericSecret, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(genericSecretCreation))
	Expect(err).To(BeNil())

	if header != nil {
		createHttpReq.Header = *header
	}

	// for existing assets, make sure their cluster resource exists
	if clusterId != nil && genericSecretCreation.Spec.ExistingSecretName != nil {
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithGenericSecret(*clusterId, &genericSecretCreation, nil))
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.GenericSecret
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestCredentialsGenericSecrets", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test generic secrets routes", func() {
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

		// Context("Test rename of generic secret", func() {
		// 	It("asset should contain the new name", func() {
		// 		genericSecret, _, _ := createGenericSecret(&randoms.ClusterId, buildRandomProjectGenericSecret(randoms), router, headerForAdmin(randoms))

		// 		//
		// 		//   simulate sync of the asset)
		// 		//
		// 		var metaForUpdate dbmodel.AssetMetaRec
		// 		copier.Copy(&metaForUpdate.AssetMeta, &genericSecret.Meta)
		// 		metaForUpdate.AssetMeta.DeletedAt.Valid = false

		// 		syncErr := dao.UpdateSyncStatusSendToAgent(context.Background(), *genericSecret.Meta.TenantId, *genericSecret.Meta.ClusterId, []dbmodel.AssetsToSync{
		// 			{
		// 				AssetMetaRec:        metaForUpdate,
		// 				AssetsGenericSecret: &dbmodel.GenericSecretRecord{},
		// 			},
		// 		})
		// 		Expect(syncErr).To(BeNil())

		// 		dao.HandleAgentSyncUpdates(context.Background(), *genericSecret.Meta.TenantId, *genericSecret.Meta.ClusterId, []api.ClusterSyncReportUpdate{
		// 			{
		// 				Id:          genericSecret.Meta.Id,
		// 				UpdateCount: genericSecret.Meta.UpdateCount,
		// 			},
		// 		})

		// 		renamePayload := api.GenericSecretUpdateRequest{
		// 			Meta: api.AssetUpdateRequest{
		// 				Name: "new-name",
		// 			},
		// 		}
		// 		updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlGenericSecret, genericSecret.Meta.Id), makePayload(renamePayload))
		// 		Expect(err).To(BeNil())
		// 		updateHttpReq.Header = *headerForAdmin(randoms)

		// 		updateHttpResp := httptest.NewRecorder()
		// 		router.ServeHTTP(updateHttpResp, updateHttpReq)
		// 		Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))
		// 		Expect(updateHttpResp.Body.String()).To(ContainSubstring("new-name"))
		// 	})
		// })

		Context("Submission of generic secret", func() {
			It("should fail if providing both properties", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.GenericSecretCreationSpec{
						ExistingSecretName: convutils.Ref(strings.ToLower(testutils.RandString(10))),
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("keyValuePairs should not be provided when creating credentials from existing secret"))
			})
		})

		Context("existing secret not in the cluster", func() {
			It("should fail on precondition", func() {
				request := buildRandomProjectGenericSecret(randoms)

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("of project scope could not be found in the cluster"))
			})
		})

		Context("existing secret scope different than cluster resource", func() {
			It("should fail on precondition", func() {
				request := buildRandomProjectGenericSecret(randoms)

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithGenericSecret(randoms.ClusterId, &request, nil))

				// we need to test with old cluster, one that supports non-project scope
				checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
					WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18.0", false)))
				checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

				request.Meta.Scope = api.SCOPE_CLUSTER
				request.Meta.ProjectId = nil
				request.Meta.ClusterId = convutils.Ref(randoms.ClusterId)

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("of cluster scope could not be found in the cluster"))
			})
		})

		Context("ListGet credentials asset", func() {
			It("should mask the secret", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

				var resp api.GenericSecret
				respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
				Expect(respJsonErr).To(BeNil())

				//
				//   List all
				//
				listHttpReq, err := http.NewRequest("GET", urlGenericSecret, nil)
				Expect(err).To(BeNil())
				listHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var listResp api.GenericSecretListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
				Expect(err).To(BeNil())
				Expect(len(listResp.Entries)).To(Equal(1))

				Expect((*listResp.Entries[0].Spec.KeyValuePairs)[0].Key).To(Equal(convutils.Ref(encryption.SecretMask)))
				Expect((*listResp.Entries[0].Spec.KeyValuePairs)[0].Value).To(Equal(convutils.Ref(encryption.SecretMask)))

				//
				//   get by id
				//
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlGenericSecret, resp.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp api.GenericSecret
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect((*getResp.Spec.KeyValuePairs)[0].Key).To(Equal(convutils.Ref(encryption.SecretMask)))
				Expect((*getResp.Spec.KeyValuePairs)[0].Value).To(Equal(convutils.Ref(encryption.SecretMask)))
			})
		})

		Context("create different scopes generic secret", func() {
			It("should fail, tenant scope is not allowed", func() {
				createPayload := buildRandomTenantGenericSecret()

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
			It("should succeed, department scope is allowed", func() {
				createPayload := buildRandomDepartmentGenericSecret(randoms)

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithGenericSecret(randoms.ClusterId, &createPayload, nil))

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
			It("should succeed, cluster scope is allowed", func() {
				createPayload := buildRandomClusterGenericSecret(randoms)

				k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithGenericSecret(randoms.ClusterId, &createPayload, nil))

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
		})
		Context("Test Delete generic secret with existing data source which used it", func() {
			It("deletion should return precondition failed used by Secret Volume", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

				var resp api.GenericSecret
				respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
				Expect(respJsonErr).To(BeNil())

				createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, &resp.Meta.Id), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGenericSecret, resp.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
			It("deletion should return precondition failed used by Git", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

				var gs api.GenericSecret
				respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &gs)
				Expect(respJsonErr).To(BeNil())

				createGitAsset(buildRandomProjectGit(randoms, convutils.Ref(gs.Meta.Id)), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGenericSecret, gs.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})
		Context("Test Delete generic secret used by an environment asset", func() {
			It("deletion should return precondition failed", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

				var resp api.GenericSecret
				respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
				Expect(respJsonErr).To(BeNil())

				environmentCreationReq := buildRandomTenantEnvironmentAsset(randoms)
				addEnvVarToEnvironmentCreationRequest(&environmentCreationReq, api.EnvironmentVariableOfAsset{
					Name: convutils.Ref(testutils.RandString(5)),
					Credential: &api.EnvironmentVariableCredential{
						AssetId: &resp.Meta.Id,
						Key:     convutils.Ref(testutils.RandString(5)),
					},
				})
				createEnvironmentAsset(environmentCreationReq, router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGenericSecret, resp.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})

		Context("Test private generic secret", func() {
			It("base test - create a private generic secret", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:    testutils.RandString(5),
						Scope:   api.SCOPE_TENANT,
						Private: convutils.Ref(true),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})

			It("create a private generic secret in project scope - should fail", func() {
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:      testutils.RandString(5),
						Scope:     api.SCOPE_PROJECT,
						ProjectId: convutils.Ref(int32(randoms.ProjectId)),
						Private:   convutils.Ref(true),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})

			It("test multiply flow for private generic secrets(create/list/sync/update/delete)", func() {
				// template for the cred creation request
				description := testutils.RandString(10)
				creationRequest := api.GenericSecretCreationRequest{
					Meta: api.AssetCreationRequest{
						Name:        testutils.RandString(5),
						Scope:       api.SCOPE_TENANT,
						Private:     convutils.Ref(true),
						Description: convutils.Ref(description),
					},
					Spec: api.GenericSecretCreationSpec{
						KeyValuePairs: &[]api.KeyValuePair{
							{
								Key:   convutils.Ref(testutils.RandString(10)),
								Value: convutils.Ref(testutils.RandString(10)),
							},
						},
					},
				}

				// createAndExpectGenericSecret helper func to send cred creation request and the expected result
				createAndExpectGenericSecret := func(creationRequest api.GenericSecretCreationRequest, headers *http.Header, expectStatus int, msg string) string {
					createHttpReq, err := http.NewRequest("POST", urlGenericSecret, makePayload(creationRequest))
					Expect(err).To(BeNil())
					createHttpReq.Header = *headers

					createHttpResp := httptest.NewRecorder()
					router.ServeHTTP(createHttpResp, createHttpReq)
					Expect(createHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusAccepted {
						return ""
					}
					var resp api.GenericSecret
					respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return resp.Meta.Id
				}

				// Create a private generic secret for admin (should succeed)
				private_Id1 := createAndExpectGenericSecret(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private generic secret")
				// Create another private generic secret for admin with the same name (should fail)
				createAndExpectGenericSecret(creationRequest, headerForAdmin(randoms), http.StatusConflict, "expected to fail creating a private generic secret with the same name for the same user")
				// Create a private generic secret for admin with different (should succeed)
				creationRequest.Meta.Name = testutils.RandString(5)
				private_Id2 := createAndExpectGenericSecret(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private generic secret with different name")
				// Create a private generic secret for a different admin with the same name (should succeed)
				otherPrivate_Id := createAndExpectGenericSecret(creationRequest, headerForOtherAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a private generic secret with the same name for a different user")

				// Create "public" private generic secret use for list testing
				creationRequest.Meta.Private = nil
				creationRequest.Meta.Scope = api.SCOPE_PROJECT
				creationRequest.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
				publicId := createAndExpectGenericSecret(creationRequest, headerForAdmin(randoms), http.StatusAccepted, "expected to successfully creation of a public generic secret")

				// listAndExpectGenericSecret is a helper function to send a list cred request and the expect ids of that request
				listAndExpectGenericSecret := func(url string, headers *http.Header, expectedIds []string, msg string) {
					listHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					listHttpReq.Header = *headers

					listHttpResp := httptest.NewRecorder()
					router.ServeHTTP(listHttpResp, listHttpReq)
					Expect(listHttpResp.Code).To(Equal(http.StatusOK), msg)

					var listResp api.GenericSecretListResponse
					err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
					Expect(err).To(BeNil())
					actualIds := make([]string, 0, len(listResp.Entries))
					for _, entry := range listResp.Entries {
						actualIds = append(actualIds, entry.Meta.Id)
					}
					Expect(actualIds).To(ContainElements(expectedIds), msg)
				}

				listAndExpectGenericSecret(fmt.Sprintf("%s", urlGenericSecret), headerForAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectGenericSecret(fmt.Sprintf("%s?privateOnly=true", urlGenericSecret), headerForAdmin(randoms), []string{private_Id1, private_Id2}, "expected to list both private secrets of the admin")
				listAndExpectGenericSecret(fmt.Sprintf("%s?includePrivate=true", urlGenericSecret), headerForAdmin(randoms), []string{private_Id1, private_Id2, publicId}, "expected to list both private secrets and the public one")

				listAndExpectGenericSecret(fmt.Sprintf("%s", urlGenericSecret), headerForOtherAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectGenericSecret(fmt.Sprintf("%s?privateOnly=true", urlGenericSecret), headerForOtherAdmin(randoms), []string{otherPrivate_Id}, "expected to list the only private secret of the other admin")
				listAndExpectGenericSecret(fmt.Sprintf("%s?includePrivate=true", urlGenericSecret), headerForOtherAdmin(randoms), []string{publicId, otherPrivate_Id}, "expected to list both private and public secret")

				// listAndExpectGenericSecret is a helper function to send a list cred request and the expect ids of that request
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
				listAndExpectCredentials(fmt.Sprintf("%s", urlCredentials), headerForAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectCredentials(fmt.Sprintf("%s?privateOnly=true", urlCredentials), headerForAdmin(randoms), []string{private_Id1, private_Id2}, "expected to list both private secrets of the admin")
				listAndExpectCredentials(fmt.Sprintf("%s?includePrivate=true", urlCredentials), headerForAdmin(randoms), []string{private_Id1, private_Id2, publicId}, "expected to list both private secrets and the public one")

				listAndExpectCredentials(fmt.Sprintf("%s", urlCredentials), headerForOtherAdmin(randoms), []string{publicId}, "expected to list the only public secret")
				listAndExpectCredentials(fmt.Sprintf("%s?privateOnly=true", urlCredentials), headerForOtherAdmin(randoms), []string{otherPrivate_Id}, "expected to list the only private secret of the other admin")
				listAndExpectCredentials(fmt.Sprintf("%s?includePrivate=true", urlGenericSecret), headerForOtherAdmin(randoms), []string{publicId, otherPrivate_Id}, "expected to list both private and public secret")

				// getGenericSecretByIdAndExpect is a helper function to send a get a cred by id request the expect the result of that request
				getGenericSecretByIdAndExpect := func(url string, headers *http.Header, expectStatus int, msg string) *api.GenericSecret {
					getHttpReq, err := http.NewRequest("GET", url, nil)
					Expect(err).To(BeNil())
					getHttpReq.Header = *headers

					getHttpResp := httptest.NewRecorder()
					router.ServeHTTP(getHttpResp, getHttpReq)
					Expect(getHttpResp.Code).To(Equal(expectStatus), msg)

					if expectStatus != http.StatusOK {
						return nil
					}
					var resp api.GenericSecret
					respJsonErr := json.Unmarshal(getHttpResp.Body.Bytes(), &resp)
					Expect(respJsonErr).To(BeNil())
					return &resp
				}

				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, publicId), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the non-private generic secret by id")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, publicId), headerForOtherAdmin(randoms), http.StatusOK, "expect to successfully get the non-private generic secret by id by the other admin")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to successfully get the private generic secret by id by the admin")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect to fail getting the private generic secret by id by the other admin")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "expect to success getting the private generic secret by id by other admin")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForAdmin(randoms), http.StatusNotFound, "expect to fail getting the private generic secret by id by the other admin")

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

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForOtherAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the private docker registry by id")
				privateSecret := getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(description), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), updateRequest, http.StatusAccepted, "expect to update the secret")
				privateSecret = getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, "")).To(Equal(newDescription), "expect the description to change")

				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForAdmin(randoms), updateRequest, http.StatusNotFound, "expect failing updating the private docker registry by id")
				privateSecret = getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
				Expect(convutils.Deref(privateSecret.Meta.Description, ""), "expect the description to be the same")
				updateSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForOtherAdmin(randoms), updateRequest, http.StatusAccepted, "expect to fail getting the private docker registry by id")
				privateSecret = getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, otherPrivate_Id), headerForOtherAdmin(randoms), http.StatusOK, "")
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
				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForOtherAdmin(randoms), http.StatusNotFound, "expect failing deleting the private docker registry by id")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusOK, "expect to find the secret")
				// success to delete the private asset by the admin
				deleteSecretAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusAccepted, "expect to successfully delete the secret")
				getGenericSecretByIdAndExpect(fmt.Sprintf("%s/%s", urlGenericSecret, private_Id1), headerForAdmin(randoms), http.StatusNotFound, "expect to not find the deleted secret")
			})
		})
	})
})
