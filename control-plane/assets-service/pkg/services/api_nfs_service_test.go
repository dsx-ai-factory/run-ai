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
	"sort"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
)

const (
	urlNFS = "/api/v1/asset/datasource/nfs"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type NFSTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewNFSTestFactory(router chi.Router) *NFSTestFactory {
	return &NFSTestFactory{router: router}
}

func (c *NFSTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.NfsListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *NFSTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.NfsCreationRequest{
		Meta: creationMeta,
		Spec: api.NfsAssetSpec{
			Path:      convutils.Ref(testutils.RandString(5)),
			ReadOnly:  convutils.Ref(false),
			Server:    convutils.Ref(testutils.RandString(5)),
			MountPath: convutils.Ref(testutils.RandString(5)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *NFSTestFactory) marshalUpdateReq([]byte) string {
	request := api.NfsUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: api.NfsAssetSpec{
			Path:      convutils.Ref(testutils.RandString(10)),
			ReadOnly:  convutils.Ref(false),
			Server:    convutils.Ref(testutils.RandString(10)),
			MountPath: convutils.Ref(testutils.RandString(10)),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *NFSTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNFSAsset(buildRandomSystemNFSAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *NFSTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNFSAsset(buildRandomTenantNFSAsset(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *NFSTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNFSAsset(buildRandomClusterNFSAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *NFSTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNFSAsset(buildRandomProjectNFSAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *NFSTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createNFSAsset(buildRandomDepartmentNFSAsset(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *NFSTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.NfsAssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *NFSTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = append(request.Assets.Datasources, api.AssetIdAndKind{
		Kind: api.ASSETKIND_NFS,
		Id:   assetMeta.Id,
	})
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomNFSHelper() api.NfsCreationRequest {
	return api.NfsCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.NfsAssetSpec{
			Path:      convutils.Ref(testutils.RandString(5)),
			ReadOnly:  convutils.Ref(false),
			Server:    convutils.Ref(testutils.RandString(5)),
			MountPath: convutils.Ref(testutils.RandString(5)),
		},
	}
}

func buildRandomSystemNFSAsset() api.NfsCreationRequest {
	ans := buildRandomNFSHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func buildRandomTenantNFSAsset() api.NfsCreationRequest {
	ans := buildRandomNFSHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterNFSAsset(randoms *testutils.RandomValues) api.NfsCreationRequest {
	ans := buildRandomNFSHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomProjectNFSAsset(randoms *testutils.RandomValues) api.NfsCreationRequest {
	ans := buildRandomNFSHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentNFSAsset(randoms *testutils.RandomValues) api.NfsCreationRequest {
	ans := buildRandomNFSHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func createNFSAsset(nfsCreation api.NfsCreationRequest, router chi.Router, header *http.Header) (api.NfsAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlNFS, makePayload(nfsCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.NfsAsset
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

	Context("delete nfs asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			nfs1, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))
			nfs2, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))

			templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfs1.Meta.Id,
				},
				{
					Kind: api.ASSETKIND_NFS,
					Id:   nfs2.Meta.Id,
				},
			}
			sort.Slice(templatePayload.Spec.Assets.Datasources, func(i, j int) bool {
				return templatePayload.Spec.Assets.Datasources[i].Id < templatePayload.Spec.Assets.Datasources[j].Id
			})

			template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlNFS, nfs2.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested nfs asset is being used by template '%s'", template.Meta.Name)))
		})
	})

	Context("update nfs asset's spec", func() {
		It("should allow changing all fields", func() {
			url := urlNFS

			origAsset := buildRandomTenantNFSAsset()

			origAsset.Spec = api.NfsAssetSpec{
				ReadOnly:  convutils.Ref(false),
				Path:      convutils.Ref("/orig/path"),
				MountPath: convutils.Ref("/orig/mount/path"),
				Server:    convutils.Ref("orig-server"),
			}

			createdAsset, _, _ := createNFSAsset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.NfsUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: api.NfsAssetSpec{
					ReadOnly:  convutils.Ref(true),
					Path:      convutils.Ref("/updated/path"),
					MountPath: convutils.Ref("/updated/mount/path"),
					Server:    convutils.Ref("updated-server"),
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
