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

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/encryption"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlAccessKey = "/api/v1/asset/credentials/access-key"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type AccessKeysTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewAccessKeysTestFactory(router chi.Router) *AccessKeysTestFactory {
	return &AccessKeysTestFactory{router: router}
}

func (c *AccessKeysTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.AccessKeyListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *AccessKeysTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.AccessKeyCreationRequest{
		Meta: creationMeta,
		Spec: api.AccessKeyCreationSpec{
			AccessKeyId:     convutils.Ref(testutils.RandString(5)),
			SecretAccessKey: convutils.Ref(testutils.RandString(5)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *AccessKeysTestFactory) marshalUpdateReq([]byte) string {
	request := api.AccessKeyUpdateRequest{
		Meta: api.CredentialUpdateRequest{
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		   Spec: &api.AccessKeyUpdateSpec{
		   	AccessKeyId:     convutils.Ref(testutils.RandString(5)),
		   	SecretAccessKey: convutils.Ref(testutils.RandString(5)),
		   },
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *AccessKeysTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createAccessKey(&randoms.ClusterId, buildRandomSystemAccessKey(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *AccessKeysTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createAccessKey(&randoms.ClusterId, buildRandomTenantAccessKey(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *AccessKeysTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createAccessKey(&randoms.ClusterId, buildRandomClusterAccessKey(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *AccessKeysTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *AccessKeysTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createAccessKey(&randoms.ClusterId, buildRandomDepartmentAccessKey(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *AccessKeysTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.AccessKeySpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *AccessKeysTestFactory) addToWorkloadCreationReq(*api.WorkloadCreationRequest, *api.AssetMeta) bool {
	return false // access key asset cannot be part of workload creation request
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomAccessKeyHelper() api.AccessKeyCreationRequest {
	return api.AccessKeyCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.AccessKeyCreationSpec{
			AccessKeyId:     convutils.Ref(testutils.RandString(10)),
			SecretAccessKey: convutils.Ref(testutils.RandString(10)),
		},
	}
}

func buildRandomProjectAccessKey(randoms *testutils.RandomValues) api.AccessKeyCreationRequest {
	ans := buildRandomAccessKeyHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentAccessKey(randoms *testutils.RandomValues) api.AccessKeyCreationRequest {
	ans := buildRandomAccessKeyHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func buildRandomTenantAccessKey() api.AccessKeyCreationRequest {
	ans := buildRandomAccessKeyHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterAccessKey(randoms *testutils.RandomValues) api.AccessKeyCreationRequest {
	ans := buildRandomAccessKeyHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomSystemAccessKey() api.AccessKeyCreationRequest {
	ans := buildRandomAccessKeyHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func createAccessKey(clusterId *string, accessKeyCreation api.AccessKeyCreationRequest, router chi.Router, header *http.Header) (api.AccessKey, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlAccessKey, makePayload(accessKeyCreation))
	Expect(err).To(BeNil())

	if header != nil {
		createHttpReq.Header = *header
	}

	// for existing assets, make sure their cluster resource exists
	if clusterId != nil && accessKeyCreation.Spec.ExistingSecretName != nil {
		k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithAccessKey(*clusterId, &accessKeyCreation, nil))
	}

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.AccessKey
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestCredentialsAccessKeys", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test access keys routes", func() {
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

		// Context("Test rename of access key", func() {
		// 	It("deletion should return precondition failed", func() {
		// 		accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
		// 		createS3Asset(buildRandomProjectS3(randoms, &accessKey.Meta.Id), router, headerForAdmin(randoms))

		// 		//
		// 		//   simulate sync of the asset)
		// 		//
		// 		var metaForUpdate dbmodel.AssetMetaRec
		// 		copier.Copy(&metaForUpdate.AssetMeta, &accessKey.Meta)
		// 		metaForUpdate.AssetMeta.DeletedAt.Valid = false

		// 		syncErr := dao.UpdateSyncStatusSendToAgent(context.Background(), *accessKey.Meta.TenantId, *accessKey.Meta.ClusterId, []dbmodel.AssetsToSync{
		// 			{
		// 				AssetMetaRec:        metaForUpdate,
		// 				AssetsGenericSecret: &dbmodel.GenericSecretRecord{},
		// 			},
		// 		})
		// 		Expect(syncErr).To(BeNil())

		// 		dao.HandleAgentSyncUpdates(context.Background(), *accessKey.Meta.TenantId, *accessKey.Meta.ClusterId, []api.ClusterSyncReportUpdate{
		// 			{
		// 				Id:          accessKey.Meta.Id,
		// 				UpdateCount: accessKey.Meta.UpdateCount,
		// 			},
		// 		})

		// 		renamePayload := api.AccessKeyUpdateRequest{
		// 			Meta: api.AssetUpdateRequest{
		// 				Name: "new-name",
		// 			},
		// 		}
		// 		updateHttpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAccessKey, accessKey.Meta.Id), makePayload(renamePayload))
		// 		Expect(err).To(BeNil())
		// 		updateHttpReq.Header = *headerForAdmin(randoms)

		// 		updateHttpResp := httptest.NewRecorder()
		// 		router.ServeHTTP(updateHttpResp, updateHttpReq)
		// 		Expect(updateHttpResp.Code).To(Equal(http.StatusAccepted))
		// 		Expect(updateHttpResp.Body.String()).To(ContainSubstring("new-name"))
		// 	})
		// })

		Context("Test Delete access key with existing data source which used it", func() {
			It("deletion should return precondition failed used by S3", func() {
				accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
				createS3Asset(buildRandomProjectS3(randoms, &accessKey.Meta.Id), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlAccessKey, accessKey.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})

			It("deletion should return precondition failed used by Secret Volume", func() {
				accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))
				createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, &accessKey.Meta.Id), router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlAccessKey, accessKey.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})

		Context("ListGet credentials asset", func() {
			It("should mask the secret", func() {
				accessKey, _, _ := createAccessKey(&randoms.ClusterId, buildRandomProjectAccessKey(randoms), router, headerForAdmin(randoms))

				//
				//   List all
				//
				listHttpReq, err := http.NewRequest("GET", urlAccessKey, nil)
				Expect(err).To(BeNil())
				listHttpReq.Header = *headerForAdmin(randoms)

				listHttpResp := httptest.NewRecorder()
				router.ServeHTTP(listHttpResp, listHttpReq)
				Expect(listHttpResp.Code).To(Equal(http.StatusOK))

				var listResp api.AccessKeyListResponse
				err = json.Unmarshal(listHttpResp.Body.Bytes(), &listResp)
				Expect(err).To(BeNil())
				Expect(len(listResp.Entries)).To(Equal(1))
				Expect(*listResp.Entries[0].Spec.AccessKeyId).NotTo(Equal(encryption.SecretMask))
				Expect(*listResp.Entries[0].Spec.SecretAccessKey).To(Equal(encryption.SecretMask))

				//
				//   get by id
				//
				getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", urlAccessKey, accessKey.Meta.Id), nil)
				Expect(err).To(BeNil())
				getHttpReq.Header = *headerForAdmin(randoms)

				getHttpResp := httptest.NewRecorder()
				router.ServeHTTP(getHttpResp, getHttpReq)
				Expect(getHttpResp.Code).To(Equal(http.StatusOK))

				var getResp api.AccessKey
				err = json.Unmarshal(getHttpResp.Body.Bytes(), &getResp)
				Expect(err).To(BeNil())
				Expect(*getResp.Spec.AccessKeyId).NotTo(Equal(encryption.SecretMask))
				Expect(*getResp.Spec.SecretAccessKey).To(Equal(encryption.SecretMask))
			})
		})

		Context("create global (tenant) scope access key", func() {
			It("should succeed, tenant scope is allowed", func() {
				createPayload := buildRandomTenantAccessKey()
				createHttpReq, err := http.NewRequest("POST", urlAccessKey, makePayload(createPayload))
				Expect(err).To(BeNil())
				createHttpReq.Header = *headerForAdmin(randoms)

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
		})
		Context("Test Delete generic secret used by an environment asset", func() {
			It("deletion should return precondition failed", func() {
				genericSecret, _, _ := createGenericSecret(&randoms.ClusterId, buildRandomProjectGenericSecret(randoms), router, headerForAdmin(randoms))

				environmentCreationReq := buildRandomTenantEnvironmentAsset(randoms)
				addEnvVarToEnvironmentCreationRequest(&environmentCreationReq, api.EnvironmentVariableOfAsset{
					Name: convutils.Ref(testutils.RandString(5)),
					Credential: &api.EnvironmentVariableCredential{
						AssetId: &genericSecret.Meta.Id,
						Key:     convutils.Ref(testutils.RandString(5)),
					},
				})
				createEnvironmentAsset(environmentCreationReq, router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlGenericSecret, genericSecret.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			})
		})

	})
})
