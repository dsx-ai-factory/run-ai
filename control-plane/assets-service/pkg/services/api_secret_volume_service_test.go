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
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/google/uuid"

	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlSecretVolume = "/api/v1/asset/datasource/secrets"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type SecretVolumeTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

// NewSecretVolumeTestFactory returns a factory for secret storage assets
func NewSecretVolumeTestFactory(router chi.Router) *SecretVolumeTestFactory {
	return &SecretVolumeTestFactory{router: router}
}

func (f *SecretVolumeTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.SecretAssetListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (f *SecretVolumeTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.SecretAssetCreationRequest{
		Meta: creationMeta,
		Spec: api.SecretAssetSpec{
			CredentialAssetId: convutils.Ref(testutils.RandString(10)),
			MountPath:         convutils.Ref(testutils.RandString(5)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (f *SecretVolumeTestFactory) marshalUpdateReq([]byte) string {
	request := api.SecretAssetUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		   Spec: api.SecretAssetSpec{
		   	CredentialAssetId: convutils.Ref(testutils.RandString(10)),
		   	MountPath:         convutils.Ref(testutils.RandString(10)),
		   },
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (f *SecretVolumeTestFactory) createSystemBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomSystemSecretVolumeAsset(), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *SecretVolumeTestFactory) createTenantBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomTenantSecretVolumeAsset(), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *SecretVolumeTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomClusterSecretVolumeAsset(randoms, nil), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *SecretVolumeTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, nil), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *SecretVolumeTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomDepartmentSecretVolumeAsset(randoms, nil), f.router, header)
	return resp.Meta, payload, checkSum
}

func (f *SecretVolumeTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.SecretAssetSpec]{}
	tg.validateListContent(f.router, listRequest, expectedUsedBy, createdSpecs)
}

func (f *SecretVolumeTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = append(request.Assets.Datasources, api.AssetIdAndKind{
		Kind: api.ASSETKIND_SECRET_VOLUME,
		Id:   assetMeta.Id,
	})
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildGenericSecret(clusterId *string, projectId *int32, router chi.Router, header *http.Header) string {
	secretRequest := buildRandomGenericSecretHelper()
	secretRequest.Meta.Scope = api.SCOPE_PROJECT
	secretRequest.Meta.ProjectId = projectId
	secretResp, _, _ := createGenericSecret(clusterId, secretRequest, router, header)
	return secretResp.Meta.Id
}

func createSecretVolumeAsset(clusterId *string, request api.SecretAssetCreationRequest, router chi.Router, header *http.Header) (api.SecretAsset, []byte, []byte) {
	// SecretVolumeAsset creation requires a secret asset to be created first, if not provided
	// create a generic secret asset
	if request.Spec.CredentialAssetId == nil {
		request.Spec.CredentialAssetId = convutils.Ref(buildGenericSecret(clusterId, request.Meta.ProjectId, router, header))
	}

	createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.SecretAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestSecretVolumeAsset", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	notExist := uuid.NewString()
	var genericSecret api.GenericSecret

	Describe("Test Secret Storage Routes", func() {
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

			genericSecret, _, _ = createGenericSecret(convutils.Ref(randoms.ClusterId), buildRandomClusterGenericSecret(randoms), router, headerForAdmin(randoms))

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
				},
			})
		})

		Context("Successful creation by scopes", func() {
			It("cluster scope", func() {

				request := testing.BuildRandomClusterSecretVolumeAsset(randoms, convutils.Ref(genericSecret.Meta.Id))
				createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})

			It("department scope", func() {
				request := testing.BuildRandomDepartmentSecretVolumeAsset(randoms, convutils.Ref(genericSecret.Meta.Id))
				createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
			It("project scope", func() {
				request := testing.BuildRandomProjectSecretVolumeAsset(randoms, convutils.Ref(genericSecret.Meta.Id))
				createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
			})
		})

		Context("delete Secret Storage asset used by template(s)", func() {
			It("should fail on precondition ", func() {
				env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				secret, _, _ := createSecretVolumeAsset(&randoms.ClusterId, testing.BuildRandomProjectSecretVolumeAsset(randoms, nil), router, headerForAdmin(randoms))

				templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
				templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
					{
						Kind: api.ASSETKIND_SECRET_VOLUME,
						Id:   secret.Meta.Id,
					},
				}
				template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlSecretVolume, secret.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
				Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested %s asset is being used by template '%s'", api.ASSETKIND_SECRET_VOLUME, template.Meta.Name)))
			})
		})

		Context("Fail to create secret in non-implemented scopes", func() {
			It("tenant scope", func() {
				request := testing.BuildRandomTenantSecretVolumeAsset()
				createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("Fail to create secret in invalid cred id", func() {
			It("department scope", func() {
				request := testing.BuildRandomProjectSecretVolumeAsset(randoms, convutils.Ref(notExist))
				createHttpReq, err := http.NewRequest("POST", urlSecretVolume, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("Credential id wasn't found. credential id %s", *request.Spec.CredentialAssetId)))
			})
		})
	})
})
