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
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
)

const (
	urlCompute = "/api/v1/asset/compute"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

type ComputeTestFactory struct {
	_      AssetTestFactory
	router chi.Router
}

func NewComputeTestFactory(router chi.Router) *ComputeTestFactory {
	return &ComputeTestFactory{router: router}
}

func (c *ComputeTestFactory) unmarshalList(bytes []byte) []api.AssetMeta {
	var listAllResp api.ComputeListResponse
	ans := make([]api.AssetMeta, 0)
	err := json.Unmarshal(bytes, &listAllResp)
	Expect(err).To(BeNil())
	for _, val := range listAllResp.Entries {
		ans = append(ans, val.Meta)
	}
	return ans
}

func (c *ComputeTestFactory) marshalCreationReq(creationMeta api.AssetCreationRequest) string {
	request := api.ComputeCreationRequest{
		Meta: creationMeta,
		Spec: &api.ComputeAssetSpec{
			GpuRequestType:   nil,
			CpuCoreRequest:   convutils.Ref(2.0),
			CpuCoreLimit:     convutils.Ref(3.5),
			CpuMemoryRequest: convutils.Ref("5M"),
			CpuMemoryLimit:   convutils.Ref("8G"),
			LargeShmRequest:  convutils.Ref(true),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *ComputeTestFactory) marshalUpdateReq([]byte) string {
	request := api.ComputeUpdateRequest{
		Meta: api.AssetUpdateRequest{
			Name:        testutils.RandString(5),
			Description: convutils.Ref(testutils.RandString(5)),
		},
		Spec: &api.ComputeAssetSpec{
			GpuRequestType:   nil,
			CpuCoreRequest:   convutils.Ref(1.0),
			CpuCoreLimit:     convutils.Ref(3.2),
			CpuMemoryRequest: convutils.Ref("30G"),
			CpuMemoryLimit:   convutils.Ref("1T"),
		},
	}
	payload, err := json.Marshal(request)
	Expect(err).To(BeNil())
	return string(payload)
}

func (c *ComputeTestFactory) createSystemBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createComputeAsset(buildRandomSystemCompute(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *ComputeTestFactory) createTenantBasedAsset(_ *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createComputeAsset(buildRandomTenantCompute(), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *ComputeTestFactory) createClusterBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createComputeAsset(buildRandomClusterCompute(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *ComputeTestFactory) createProjectBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createComputeAsset(buildRandomProjectCompute(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *ComputeTestFactory) createDepartmentBasedAsset(randoms *testutils.RandomValues, header *http.Header) (api.AssetMeta, []byte, []byte) {
	resp, payload, checkSum := createComputeAsset(buildRandomDepartmentCompute(randoms), c.router, header)
	return resp.Meta, payload, checkSum
}

func (c *ComputeTestFactory) validateListContent(listRequest *http.Request, expectedUsedBy datastructs.Set[string], createdSpecs ...[]byte) {
	tg := GenericTest[api.ComputeAssetSpec]{}
	tg.validateListContent(c.router, listRequest, expectedUsedBy, createdSpecs)
}

func (c *ComputeTestFactory) addToWorkloadCreationReq(request *api.WorkloadCreationRequest, assetMeta *api.AssetMeta) bool {
	request.Assets.Compute = convutils.Ref(assetMeta.Id)
	return true
}

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomComputeHelper() api.ComputeCreationRequest {
	extendedResources := []api.ExtendedResource{
		{
			Resource: convutils.Ref("gpu/amd"),
			Quantity: convutils.Ref("12M"),
		},
		{
			Resource: convutils.Ref("tesla/car"),
			Quantity: convutils.Ref("12345"),
		},
	}
	return api.ComputeCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: &api.ComputeAssetSpec{
			GpuRequestType:    nil,
			CpuCoreLimit:      convutils.Ref(3.5),
			CpuMemoryLimit:    convutils.Ref("20M"),
			LargeShmRequest:   convutils.Ref(true),
			ExtendedResources: &extendedResources,
		},
	}
}

func buildRandomSystemCompute() api.ComputeCreationRequest {
	ans := buildRandomComputeHelper()
	ans.Meta.Scope = api.SCOPE_SYSTEM
	return ans
}

func buildRandomTenantCompute() api.ComputeCreationRequest {
	ans := buildRandomComputeHelper()
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomClusterCompute(randoms *testutils.RandomValues) api.ComputeCreationRequest {
	ans := buildRandomComputeHelper()
	ans.Meta.Scope = api.SCOPE_CLUSTER
	ans.Meta.ClusterId = convutils.Ref(randoms.ClusterId)
	return ans
}

func buildRandomProjectCompute(randoms *testutils.RandomValues) api.ComputeCreationRequest {
	ans := buildRandomComputeHelper()
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	return ans
}

func buildRandomDepartmentCompute(randoms *testutils.RandomValues) api.ComputeCreationRequest {
	ans := buildRandomComputeHelper()
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func createComputeAsset(computeCreation api.ComputeCreationRequest, router chi.Router, header *http.Header) (api.ComputeAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)

	Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

	var resp api.ComputeAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

func fetchComputeAssets(router chi.Router, randoms *testutils.RandomValues, url string) api.ComputeListResponse {
	listHttpReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	listHttpReq.Header = *headerForAdmin(randoms)

	listHttpResp := httptest.NewRecorder()
	router.ServeHTTP(listHttpResp, listHttpReq)

	Expect(listHttpResp.Code).To(Equal(http.StatusOK))

	var resp api.ComputeListResponse
	respJsonErr := json.Unmarshal(listHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp
}

func expectComputeAssets(entries []api.ComputeAsset, expectedIds []string) {
	Expect(len(entries)).To(Equal(len(expectedIds)), "Expected %d compute assets but got %d", len(expectedIds), len(entries))

	actualIds := make([]string, 0, len(entries))
	for _, entry := range entries {
		actualIds = append(actualIds, entry.Meta.Id)
	}

	for _, expectedId := range expectedIds {
		Expect(actualIds).To(ContainElement(expectedId))
	}
}

var _ = Describe("TestComputeService", func() {
})

/////////////////////////////////////////////////////////////////////////////////////////////
// 	Test Compute formats
/////////////////////////////////////////////////////////////////////////////////////////////

var _ = Describe("compute asset specific tests", func() {
	var randoms *testutils.RandomValues
	var policyMockUpdateTime time.Time

	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		policyMockUpdateTime = time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		randoms = testutils.NewRandomValues()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18", false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")

	})

	Context("compute gpu requests", func() {
		It("no gpu request type", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("when gpuDevices is 1, gpuRequestType must be specified"))
		})

		It("create compute with portion AND memory requests", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(0.5)
			computeCreation.Spec.GpuMemoryRequest = convutils.Ref("10G")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryRequest/gpuMemoryLimit should not be provided when GPU request type is 'portion'"))
		})

		It("create compute with portion AND memory limit", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(0.5)
			computeCreation.Spec.GpuMemoryLimit = convutils.Ref("10G")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryRequest/gpuMemoryLimit should not be provided when GPU request type is 'portion'"))
		})

		It("create compute with memory AND portion requests", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(0.5)
			computeCreation.Spec.GpuMemoryRequest = convutils.Ref("10G")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryPortionRequest/gpuMemoryPortionLimit (deprecated aliases: gpuPortionRequest/gpuPortionLimit) should not be provided when GPU request type is 'memory'"))
		})

		It("create compute with gpu portion range issue", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(1.2)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("out of range (0 to 1)"))
		})

		It("create compute with gpu portion w/o the portion", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryPortionRequest (deprecated alias: gpuPortionRequest) is mandatory when GPU request of type 'portion'"))
		})

		It("create compute with gpu memory w/o the memory", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryRequest is mandatory when GPU request of type 'memory'"))
		})

		It("devices > 1 and request type given", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(4))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuRequestType can be provided only when gpuDevices is 1"))
		})

		It("devices = 0 and request type given", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(0))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuRequestType can be provided only when gpuDevices is 1"))
		})

		It("create empty compute request", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("tests gpu memory format", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
			computeCreation.Spec.GpuMemoryRequest = convutils.Ref("10GBccc")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("value for GpuMemoryRequest must comply to the pattern"))
		})

		It("gpuPortion limit range issue", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(0.5)
			computeCreation.Spec.GpuPortionLimit = convutils.Ref(0.4)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("out of range (0.500000 to 1)"))
		})

		It("gpuPortion limit range OK", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypePortion)
			computeCreation.Spec.GpuPortionRequest = convutils.Ref(0.5)
			computeCreation.Spec.GpuPortionLimit = convutils.Ref(0.5)

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("gpuMemory limit range issue", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
			computeCreation.Spec.GpuMemoryRequest = convutils.Ref("1G")
			computeCreation.Spec.GpuMemoryLimit = convutils.Ref("999M")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("gpuMemoryLimit must be no less than gpuMemoryRequest"))
		})

		It("gpuMemory limit range ok", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.GpuDevicesRequest = convutils.Ref(int32(1))
			computeCreation.Spec.GpuRequestType = GpuRequestTypePtr(GpuRequestTypeMemory)
			computeCreation.Spec.GpuMemoryRequest = convutils.Ref("1G")
			computeCreation.Spec.GpuMemoryLimit = convutils.Ref("1G")

			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})
	})

	Context("compute memory formatted fields", func() {
		It("create compute bad gpu request", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.CpuMemoryLimit = convutils.Ref("1A2")
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)

			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
		It("create compute request with bad CpuMemoryRequest", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.CpuMemoryLimit = convutils.Ref("500X")
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
		})
	})

	Context("gpu fields validation", func() {
		It("should reject incorrect combination of fields", func() {
			err := validateComputeSpec(api.ComputeAssetSpec{
				GpuDevicesRequest: convutils.Ref(int32(4)),
				GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("gpuRequestType can be provided only when gpuDevices is 1"))

			err = validateComputeSpec(api.ComputeAssetSpec{
				GpuDevicesRequest: convutils.Ref(int32(4)),
				GpuMemoryRequest:  convutils.Ref("10M"),
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("gpuMemoryRequest/gpuMemoryLimit can be provided only when gpuDevices is 1"))

			err = validateComputeSpec(api.ComputeAssetSpec{
				GpuDevicesRequest: convutils.Ref(int32(4)),
				GpuPortionRequest: convutils.Ref(12.3),
			}, false)
			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(err.Message).To(ContainSubstring("gpuMemoryPortionRequest/gpuMemoryPortionLimit (deprecated aliases: gpuPortionRequest/gpuPortionLimit) can be provided only when gpuDevices is 1"))
		})
	})

	Context("extended resource validation", func() {
		It("multiple resource name should give error", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.ExtendedResources = &[]api.ExtendedResource{
				{
					Resource: convutils.Ref("abc/def"),
					Quantity: convutils.Ref("10m"),
				},
				{
					Resource: convutils.Ref("xxx/yyy"),
					Quantity: convutils.Ref("15"),
				},
				{
					Resource: convutils.Ref("abc/def"),
					Quantity: convutils.Ref("20m"),
				},
			}
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)
			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body).To(ContainSubstring("resource abc/def specified more than once"))
		})

		It("invalid resource name should give error", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.ExtendedResources = &[]api.ExtendedResource{
				{
					Resource: convutils.Ref("xxx/yyy"),
					Quantity: convutils.Ref("10m"),
				},
				{
					Resource: convutils.Ref("XXX$YYY"),
					Quantity: convutils.Ref("15"),
				},
			}
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)
			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body).To(ContainSubstring("resource XXX$YYY does not comply k8s resource naming rules"))
		})

		It("invalid quantity should give error", func() {
			computeCreation := buildRandomProjectCompute(randoms)
			computeCreation.Spec.ExtendedResources = &[]api.ExtendedResource{
				{
					Resource: convutils.Ref("xxx/yyy"),
					Quantity: convutils.Ref("10m"),
				},
				{
					Resource: convutils.Ref("aaa/bbb"),
					Quantity: convutils.Ref("15X"),
				},
			}
			createHttpReq, err := http.NewRequest("POST", urlCompute, makePayload(computeCreation))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createHttpResp)
			Expect(respJsonErr).To(BeNil())
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("Quantity: value for Quantity must comply to the pattern"))
		})
	})

	Context("delete compute asset used by template(s)", func() {
		It("should fail on precondition ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			templatePayload1 := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload2 := buildRandomProjectTemplate(randoms, env.Meta.Id)
			templatePayload1.Spec.Assets.Compute = convutils.Ref(compute.Meta.Id)
			templatePayload2.Spec.Assets.Compute = convutils.Ref(compute.Meta.Id)
			template1, _, _ := createTemplateAsset(templatePayload1, router, headerForAdmin(randoms))
			template2, _, _ := createTemplateAsset(templatePayload2, router, headerForAdmin(randoms))

			deleteHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/%s", urlCompute, compute.Meta.Id), nil)
			Expect(err).To(BeNil())
			deleteHttpReq.Header = *headerForAdmin(randoms)

			deleteHttpResp := httptest.NewRecorder()
			router.ServeHTTP(deleteHttpResp, deleteHttpReq)
			Expect(deleteHttpResp.Code).To(Equal(http.StatusPreconditionFailed))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring("the requested compute asset is being used by 2 templates"))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(template1.Meta.Name))
			Expect(deleteHttpResp.Body.String()).To(ContainSubstring(template2.Meta.Name))
		})
	})

	Context("update compute asset's spec", func() {
		It("should allow changing all fields", func() {
			url := urlCompute

			origAsset := buildRandomTenantCompute()

			origAsset.Spec = &api.ComputeAssetSpec{
				GpuDevicesRequest: convutils.Ref(int32(1)),
				GpuRequestType:    convutils.Ref(api.GPUREQUESTTYPE_MEMORY),
				GpuComputeMode:    convutils.Ref(api.GPUCOMPUTEMODE_TIME_SLICING),
				GpuMemoryRequest:  convutils.Ref("2G"),
				GpuMemoryLimit:    convutils.Ref("4G"),
				CpuCoreRequest:    convutils.Ref(2.5),
				CpuCoreLimit:      convutils.Ref(5.0),
				CpuMemoryRequest:  convutils.Ref("16G"),
				CpuMemoryLimit:    convutils.Ref("32G"),
				ExtendedResources: &[]api.ExtendedResource{
					{
						Resource: convutils.Ref("orig/res"),
						Quantity: convutils.Ref("10G"),
					},
				},
			}

			createdAsset, _, _ := createComputeAsset(origAsset, router, headerForAdmin(randoms))

			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &origAsset.Spec)

			//
			//  now modify the asset
			//
			updateRequest := api.ComputeUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: origAsset.Meta.Name,
				},
				Spec: &api.ComputeAssetSpec{
					GpuDevicesRequest:       convutils.Ref(int32(1)),
					GpuRequestType:          convutils.Ref(api.GPUREQUESTTYPE_PORTION),
					GpuComputeMode:          convutils.Ref(api.GPUCOMPUTEMODE_SM_SHARING),
					GpuMemoryPortionRequest: convutils.Ref(0.2),
					GpuMemoryPortionLimit:   convutils.Ref(0.4),
					CpuCoreRequest:          convutils.Ref(5.0),
					CpuCoreLimit:            convutils.Ref(7.5),
					CpuMemoryRequest:        convutils.Ref("20G"),
					CpuMemoryLimit:          convutils.Ref("40G"),
					ExtendedResources: &[]api.ExtendedResource{
						{
							Resource: convutils.Ref("updated/res"),
							Quantity: convutils.Ref("20G"),
						},
					},
				},
			}

			updateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &updateRequest)

			//
			//  validate that the update changed the fields correctly
			//
			expectedSpec := *updateRequest.Spec
			expectedSpec.GpuPortionRequest = convutils.Ref(0.2)
			expectedSpec.GpuPortionLimit = convutils.Ref(0.4)
			validateAssetSpec(randoms, router, url, createdAsset.Meta.Id, &expectedSpec)
		})
	})

	Context("test includeDescendants param", func() {
		var anotherCluster *testutils.RandomValues
		var anotherProj *testutils.RandomValues

		BeforeEach(func() {
			dbTesting.CleanTables()

			anotherProj = randoms.CloneWithDifferentProject()
			anotherCluster = randoms.CloneWithDifferentCluster()

			authMock := testing.NewRandomsBasedAuthMockClient(randoms, anotherProj, anotherCluster)
			org_unit.SetMock(org_unit.NewSimpleMock(randoms, anotherProj, anotherCluster))
			authorization.SetMock(authMock)
			checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
				WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, "2.18", false)).
				WithCluster(anotherCluster.ClusterId, testing.BuildDisplayedCluster(anotherCluster, "2.18", false)))
			checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		})

		It("should return correct assets with includeDescendants flag", func() {
			// Create compute assets at different scope levels
			tenantCompute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			clusterCompute, _, _ := createComputeAsset(buildRandomClusterCompute(randoms), router, headerForAdmin(randoms))
			anotherClusterCompute, _, _ := createComputeAsset(buildRandomClusterCompute(anotherCluster), router, headerForAdmin(anotherCluster))
			departmentCompute, _, _ := createComputeAsset(buildRandomDepartmentCompute(randoms), router, headerForAdmin(randoms))
			projectCompute, _, _ := createComputeAsset(buildRandomProjectCompute(randoms), router, headerForAdmin(randoms))
			anotherProjectCompute, _, _ := createComputeAsset(buildRandomProjectCompute(anotherProj), router, headerForAdmin(anotherProj))

			// Test with no params - expect to get all assets
			resp := fetchComputeAssets(router, randoms, urlCompute)
			expectComputeAssets(resp.Entries, []string{tenantCompute.Meta.Id, clusterCompute.Meta.Id, anotherClusterCompute.Meta.Id, departmentCompute.Meta.Id, projectCompute.Meta.Id, anotherProjectCompute.Meta.Id})

			// Test with cluster id - expect to get the cluster asset and all assets that are above (1 tenant)
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s", urlCompute, randoms.ClusterId))
			expectComputeAssets(resp.Entries, []string{clusterCompute.Meta.Id, tenantCompute.Meta.Id})

			// Test with cluster id and includeDescendants - expect to get the cluster asset, all assets that are above (1 tenant) and all that are below (1 department, 2 projects)
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&includeDescendants=true", urlCompute, randoms.ClusterId))
			expectComputeAssets(resp.Entries, []string{clusterCompute.Meta.Id, tenantCompute.Meta.Id, departmentCompute.Meta.Id, projectCompute.Meta.Id, anotherProjectCompute.Meta.Id})

			// Test with cluster id and scope - expect to get just the cluster asset
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?clusterId=%s&scope=cluster", urlCompute, randoms.ClusterId))
			expectComputeAssets(resp.Entries, []string{clusterCompute.Meta.Id})

			// Test with department id - expect to get the department asset and all assets that are above (1 cluster, 1 tenant)
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s", urlCompute, randoms.DepartmentId))
			expectComputeAssets(resp.Entries, []string{departmentCompute.Meta.Id, clusterCompute.Meta.Id, tenantCompute.Meta.Id})

			// Test with department id and scope - expect to get just the department asset
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&scope=department", urlCompute, randoms.DepartmentId))
			expectComputeAssets(resp.Entries, []string{departmentCompute.Meta.Id})

			// Test with department id and includeDescendants - expect to get the department asset, all assets that are above (1 cluster, 1 tenant) and no descendants (department doesn't support includeDescendants yet)
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?departmentId=%s&includeDescendants=true", urlCompute, randoms.DepartmentId))
			expectComputeAssets(resp.Entries, []string{clusterCompute.Meta.Id, tenantCompute.Meta.Id, departmentCompute.Meta.Id})

			// Test with project id - expect to get the project asset and all assets that are above (1 department, 1 cluster, 1 tenant)
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?projectId=%v", urlCompute, randoms.ProjectId))
			expectComputeAssets(resp.Entries, []string{projectCompute.Meta.Id, departmentCompute.Meta.Id, clusterCompute.Meta.Id, tenantCompute.Meta.Id})

			// Test with project id and scope - expect to get just the project asset
			resp = fetchComputeAssets(router, randoms, fmt.Sprintf("%s?projectId=%v&scope=project", urlCompute, randoms.ProjectId))
			expectComputeAssets(resp.Entries, []string{projectCompute.Meta.Id})
		})

		It("should reject includeDescendants with scope parameter", func() {
			// Test that includeDescendants and scope are mutually exclusive
			listHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?clusterId=%s&scope=cluster&includeDescendants=true", urlCompute, randoms.ClusterId), nil)
			Expect(err).To(BeNil())
			listHttpReq.Header = *headerForAdmin(randoms)

			listHttpResp := httptest.NewRecorder()
			router.ServeHTTP(listHttpResp, listHttpReq)

			Expect(listHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(listHttpResp.Body.String()).To(ContainSubstring("includeDescendants and scope are mutual exclusive"))
		})
	})
})
