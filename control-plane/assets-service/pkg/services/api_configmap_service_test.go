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
	"time"

	k8sObjectsTracker "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlConfigMap = "/api/v1/asset/datasource/config-map"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type ConfigMapTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewConfigMapTestFactory(router chi.Router) *ConfigMapTestFactory {
	return &ConfigMapTestFactory{router: router}
}

func (c *ConfigMapTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.ConfigMapListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *ConfigMapTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.ConfigMapCreationRequest{
		Meta: creationMeta,
		Spec: api.ConfigMapAssetSpec{
			ConfigMap:   convutils.Ref(strings.ToLower(testutils.RandString(10))),
			MountPath:   convutils.Ref(testutils.RandString(5)),
			SubPath:     convutils.Ref(testutils.RandString(5)),
			DefaultMode: convutils.Ref("0644"),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *ConfigMapTestFactory) marshalUpdateReq([]byte) string {
	request := api.ConfigMapUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		/* At this point we are not supporting spec update
		   Spec: api.ConfigMapAssetSpec{
		   	ConfigMap: convutils.Ref(testutils.RandString(5)),
		   	MountPath: convutils.Ref(testutils.RandString(10)),
		   },
		*/
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *ConfigMapTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createConfigMapAsset(testing.BuildRandomSystemConfigMapAsset(), c.router, header, http.StatusBadRequest)
	return resp.Meta, payload, checkSum
}

func (c *ConfigMapTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createConfigMapAsset(testing.BuildRandomTenantConfigMapAsset(), c.router, header, http.StatusBadRequest)
	return resp.Meta, payload, checkSum
}

func (c *ConfigMapTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createConfigMapAsset(testing.BuildRandomClusterConfigMapAsset(randoms), c.router, header, http.StatusAccepted)
	return resp.Meta, payload, checkSum
}

func (c *ConfigMapTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createConfigMapAsset(testing.BuildRandomProjectConfigMapAsset(randoms), c.router, header, http.StatusAccepted)
	return resp.Meta, payload, checkSum
}

func (c *ConfigMapTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createConfigMapAsset(testing.BuildRandomDepartmentConfigMapAsset(randoms), c.router, header, http.StatusAccepted)
	return resp.Meta, payload, checkSum
}

func (c *ConfigMapTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.ConfigMapAssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *ConfigMapTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	if request.Assets.Datasources == nil {
		request.Assets.Datasources = make([]api.AssetIdAndKind, 0)
	}
	request.Assets.Datasources = append(request.Assets.Datasources, api.AssetIdAndKind{
		Kind: api.ASSETKIND_CONFIG_MAP,
		Id:   assetMeta.Id,
	})
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func createConfigMapAsset(request api.ConfigMapCreationRequest, router chi.Router, header *http.Header, expectedResCode int) (api.ConfigMapAsset, []byte, []byte) {
	var clusterId string
	if request.Meta.ClusterId == nil {
		clusterId = ""
	} else {
		clusterId = *request.Meta.ClusterId
	}
	// When creating a ConfigMap asset it must exist in the k8s tracker
	k8sObjectsTracker.SetMock(k8sObjectsTracker.NewSimpleMockClient().WithConfigMap(clusterId, &request, nil))

	createHttpReq, err := http.NewRequest("POST", urlConfigMap, makePayload(request))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(expectedResCode))
	if createHttpResp.Code != http.StatusAccepted {
		return api.ConfigMapAsset{}, nil, nil
	}

	var resp api.ConfigMapAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())

	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("TestConfigMapAsset", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	Describe("Test ConfigMap Routes", func() {
		BeforeEach(func() {
			dbTesting.CleanTables()
			randoms = testutils.NewRandomValues()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms))
			authorization.SetMock(authMock)
			workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18", false)))

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
				},
			})
		})

		Context("existing config map not in the cluster", func() {
			It("should fail on precondition", func() {
				request := testing.BuildRandomProjectConfigMapAsset(randoms)

				createHttpReq, err := http.NewRequest("POST", urlConfigMap, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
				Expect(createHttpResp.Body.String()).To(ContainSubstring("of project scope could not be found in the cluster"))
			})
		})

		Context("delete ConfigMap asset used by template(s)", func() {
			It("should fail on precondition ", func() {
				env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
				configMap, _, _ := createConfigMapAsset(testing.BuildRandomProjectConfigMapAsset(randoms), router, headerForAdmin(randoms), http.StatusAccepted)

				templatePayload := buildRandomProjectTemplate(randoms, env.Meta.Id)
				templatePayload.Spec.Assets.Datasources = []api.AssetIdAndKind{
					{
						Kind: api.ASSETKIND_CONFIG_MAP,
						Id:   configMap.Meta.Id,
					},
				}
				template, _, _ := createTemplateAsset(templatePayload, router, headerForAdmin(randoms))

				deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlConfigMap, configMap.Meta.Id), nil)
				Expect(err).To(BeNil())
				deleteHttpReq.Header = *headerForAdmin(randoms)

				deleteHttpResp := httptest.NewRecorder()
				router.ServeHTTP(deleteHttpResp, deleteHttpReq)
				Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
				Expect(deleteHttpResp.Body.String()).To(ContainSubstring(fmt.Sprintf("the requested config-map asset is being used by template '%s'", template.Meta.Name)))
			})
		})
		// TODO Remove this once ConfigMap as data source is supported in all scopes
		Context("Fail to create configmap in non-project scope for clusters OLDER than 2.19", func() {
			It("department scope", func() {
				request := testing.BuildRandomDepartmentConfigMapAsset(randoms)
				createHttpReq, err := http.NewRequest("POST", urlConfigMap, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
			It("cluster scope", func() {
				request := testing.BuildRandomClusterConfigMapAsset(randoms)
				createHttpReq, err := http.NewRequest("POST", urlConfigMap, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
			It("tenant scope", func() {
				request := testing.BuildRandomTenantConfigMapAsset()
				createHttpReq, err := http.NewRequest("POST", urlConfigMap, makePayload(request))
				createHttpReq.Header = *headerForAdmin(randoms)
				Expect(err).To(BeNil())

				createHttpResp := httptest.NewRecorder()
				router.ServeHTTP(createHttpResp, createHttpReq)
				Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
