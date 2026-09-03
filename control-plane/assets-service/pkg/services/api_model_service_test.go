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
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	urlModel     = "/api/v1/asset/model"
	urlModelById = "/api/v1/asset/model/%s"
)

/////////////////////////////////////////////////////////////////////////////////////////////
//    Implementation of the AssetTestFactory
/////////////////////////////////////////////////////////////////////////////////////////////

/////////////////////////////////////////////////////////////////////////////////////////////
//    Asset Builders
/////////////////////////////////////////////////////////////////////////////////////////////

func buildRandomModelHelper(environment api.EnvironmentAsset, compute api.ComputeAsset) api.ModelCreationRequest {
	return api.ModelCreationRequest{
		Meta: api.AssetCreationRequest{
			Name: testutils.RandString(5),
		},
		Spec: api.ModelAssetSpec{
			Title:         testutils.RandString(5),
			Description:   convutils.Ref(testutils.RandString(5)),
			Thumbnail:     convutils.Ref(testutils.RandString(5)),
			HuggingfaceId: convutils.Ref(testutils.RandString(5)),
			License:       convutils.Ref(testutils.RandString(5)),
			Assets: api.AssetsIds{
				Environment: environment.Meta.Id,
				Compute:     convutils.Ref(compute.Meta.Id),
			},
		},
	}
}

func buildRandomTenantModelAsset(environment api.EnvironmentAsset, compute api.ComputeAsset) api.ModelCreationRequest {
	ans := buildRandomModelHelper(environment, compute)
	ans.Meta.Scope = api.SCOPE_TENANT
	return ans
}

func buildRandomProjectModelAsset(randoms *testutils.RandomValues, environment api.EnvironmentAsset, compute api.ComputeAsset) api.ModelCreationRequest {
	ans := buildRandomModelHelper(environment, compute)
	ans.Meta.Scope = api.SCOPE_PROJECT
	ans.Meta.ProjectId = convutils.Ref(int32(randoms.ProjectId))
	ans.Meta.Description = convutils.Ref(testutils.RandString(10))
	return ans
}

func buildRandomDepartmentModelAsset(randoms *testutils.RandomValues, environment api.EnvironmentAsset, compute api.ComputeAsset) api.ModelCreationRequest {
	ans := buildRandomModelHelper(environment, compute)
	ans.Meta.Scope = api.SCOPE_DEPARTMENT
	ans.Meta.DepartmentId = convutils.Ref(randoms.DepartmentId)
	return ans
}

func createModelAsset(modelCreation api.ModelCreationRequest, router chi.Router, header *http.Header) (api.ModelAsset, []byte, []byte) {
	createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(modelCreation))
	createHttpReq.Header = *header
	Expect(err).To(BeNil())

	createHttpResp := httptest.NewRecorder()
	router.ServeHTTP(createHttpResp, createHttpReq)
	Expect(createHttpResp.Code).To(Equal(http.StatusCreated))

	var resp api.ModelAsset
	respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
	Expect(respJsonErr).To(BeNil())
	return resp, createHttpResp.Body.Bytes(), nil
}

var _ = Describe("model tests", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	notExist := uuid.NewString()

	BeforeEach(func() {
		dbTesting.CleanTables()
		randoms = testutils.NewRandomValues()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())
	})

	Context("model test", func() {
		It("create should fail, not admin - not permitted", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(buildRandomTenantModelAsset(env, compute)))
			createHttpReq.Header = *headerForNonAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("not permitted to perform this operation"))
		})

		It("list should fail, different tenant", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			_, _, _ = createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(buildRandomTenantModelAsset(env, compute)))
			differentTenant := randoms.CloneWithDifferentTenant()
			createHttpReq.Header = *headerForNonAdmin(differentTenant)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusForbidden))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("not permitted to perform this operation"))

			// list all
			createHttpReq, err = http.NewRequest("GET", urlModel, nil)
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())
			Expect(createHttpResp.Body.String()).To(ContainSubstring("not permitted to perform this operation"))
		})

		It("create should success, validate create a basic model without optional fields", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			randomTenantModelAsset := buildRandomTenantModelAsset(env, compute)
			randomTenantModelAsset.Spec.Description = nil
			randomTenantModelAsset.Spec.Thumbnail = nil
			randomTenantModelAsset.Spec.HuggingfaceId = nil
			_, _, _ = createModelAsset(randomTenantModelAsset, router, headerForAdmin(randoms))
		})

		It("create should success, and list tenant, project, department", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			_, _, _ = createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))
			_, _, _ = createModelAsset(buildRandomProjectModelAsset(randoms, env, compute), router, headerForAdmin(randoms))
			_, _, _ = createModelAsset(buildRandomDepartmentModelAsset(randoms, env, compute), router, headerForAdmin(randoms))

			// list all
			createHttpReq, err := http.NewRequest("GET", urlModel, nil)
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			var resp api.ModelListResponse
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(3))
		})

		It("create should success, validate list response content", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			randomTenantModelAsset := buildRandomTenantModelAsset(env, compute)
			_, _, _ = createModelAsset(randomTenantModelAsset, router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("GET", urlModel, nil)
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			var resp api.ModelListResponse
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(1))

			// validate the response content
			entry := resp.Entries[0]
			Expect(entry.Meta.Id).ToNot(BeNil())
			Expect(entry.Meta.CreatedAt).ToNot(BeNil())
			Expect(entry.Meta.UpdatedAt).ToNot(BeNil())
			Expect(entry.Spec.Title).To(Equal(randomTenantModelAsset.Spec.Title))
			Expect(entry.Spec.Description).To(Equal(randomTenantModelAsset.Spec.Description))
			Expect(entry.Spec.HuggingfaceId).To(Equal(randomTenantModelAsset.Spec.HuggingfaceId))
			Expect(entry.Spec.License).To(Equal(randomTenantModelAsset.Spec.License))
			Expect(entry.Spec.Assets.Environment).To(Equal(randomTenantModelAsset.Spec.Assets.Environment))
			Expect(entry.Spec.Assets.Compute).To(Equal(randomTenantModelAsset.Spec.Assets.Compute))
			Expect(entry.Meta.Scope).To(Equal(randomTenantModelAsset.Meta.Scope))
			Expect(entry.Meta.DepartmentId).To(Equal(randomTenantModelAsset.Meta.DepartmentId))
			Expect(entry.Meta.ProjectId).To(Equal(randomTenantModelAsset.Meta.ProjectId))
		})

		It("validate list response content by environment id ", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			randomTenantModelAsset := buildRandomTenantModelAsset(env, compute)
			randomTenantModelAsset.Spec.License = convutils.Ref("license")
			_, _, _ = createModelAsset(randomTenantModelAsset, router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?environmentId=%s", urlModel, env.Meta.Id), nil)
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))
			var resp api.ModelListResponse
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(1))

			// validate the response content
			entry := resp.Entries[0]
			Expect(entry.Meta.Id).ToNot(BeNil())
			Expect(entry.Spec.License).To(Equal(convutils.Ref("license")))

			// list fake environment id
			createHttpReq, err = http.NewRequest("GET", fmt.Sprintf("%s?environmentId=123", urlModel), nil)
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			respJsonErr = json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(0))
		})

		It("create should fail, missing environment", func() {
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(buildRandomTenantModelAsset(api.EnvironmentAsset{}, compute)))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("value for Environment must have at least 1 characters"))
		})

		It("create should fail, not valid environment asset", func() {
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			compute2, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			randomTenantModelAsset := buildRandomTenantModelAsset(api.EnvironmentAsset{}, compute)
			randomTenantModelAsset.Spec.Assets.Environment = compute2.Meta.Id // not valid environment asset

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(randomTenantModelAsset))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("compute rather than environment"))
		})

		It("create should fail, missing compute asset", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(buildRandomTenantModelAsset(env, api.ComputeAsset{})))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("value for Compute must have at least 1 characters"))
		})

		It("update should success", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			createRequest, _, _ := createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))

			updateRequest := api.ModelUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: "new name",
				},
				Spec: createRequest.Spec,
			}
			createHttpReq, err := http.NewRequest("PUT", fmt.Sprintf(urlModelById, createRequest.Meta.Id), makePayload(updateRequest))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))

			updateRes := api.ModelAsset{}
			err = json.Unmarshal(createHttpResp.Body.Bytes(), &updateRes)
			Expect(err).To(BeNil())

			// validate the response content
			Expect(updateRes.Meta.Name).To(Equal("new name"))
			Expect(updateRes.Spec.Assets.Environment).To(Equal(createRequest.Spec.Assets.Environment))
			Expect(updateRes.Spec.Assets.Compute).To(Equal(createRequest.Spec.Assets.Compute))
		})

		It("update should fail, not valid asset", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			createdModelAsset, _, _ := createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))

			// update with missing environment id
			createdModelAsset.Spec.Assets.Environment = ""
			request := api.ModelUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: testutils.RandString(5),
				},
				Spec: createdModelAsset.Spec,
			}

			createHttpReq, err := http.NewRequest("PUT", fmt.Sprintf(urlModelById, createdModelAsset.Meta.Id), makePayload(request))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("value for Environment must have at least 1 characters"))
		})

		It("update should succeed, update with missing compute", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			randomTenantModelAsset := buildRandomTenantModelAsset(env, compute)

			createHttpReq, err := http.NewRequest("POST", urlModel, makePayload(randomTenantModelAsset))
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)
			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusCreated))

			createRes := api.ModelAsset{}
			respJsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &createRes)
			Expect(respJsonErr).To(BeNil())

			randomTenantModelAsset.Spec.Assets.Compute = nil
			request := api.ModelUpdateRequest{
				Meta: api.AssetUpdateRequest{
					Name: testutils.RandString(5),
				},
				Spec: randomTenantModelAsset.Spec,
			}

			createHttpReq, err = http.NewRequest("PUT", fmt.Sprintf(urlModelById, createRes.Meta.Id), makePayload(request))
			createHttpReq.Header = *headerForAdmin(randoms)
			Expect(err).To(BeNil())

			createHttpResp = httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("delete should success", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			createRequest, _, _ := createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf(urlModelById, createRequest.Meta.Id), nil)
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusAccepted))
		})

		It("delete should fail, not exist", func() {
			createHttpReq, err := http.NewRequest("DELETE", fmt.Sprintf(urlModelById, notExist), nil)
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("requested asset does not exist"))
		})

		It("get should success", func() {
			env, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			compute, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))

			createRequest, _, _ := createModelAsset(buildRandomTenantModelAsset(env, compute), router, headerForAdmin(randoms))

			createHttpReq, err := http.NewRequest("GET", fmt.Sprintf(urlModelById, createRequest.Meta.Id), nil)
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusOK))

			getByIdRes := api.ModelAsset{}
			err = json.Unmarshal(createHttpResp.Body.Bytes(), &getByIdRes)
			Expect(err).To(BeNil())

			// validate the response content
			Expect(getByIdRes.Meta.Name).To(Equal(createRequest.Meta.Name))
			Expect(getByIdRes.Meta.Id).To(Equal(createRequest.Meta.Id))
			Expect(getByIdRes.Spec.Assets.Environment).To(Equal(createRequest.Spec.Assets.Environment))
			Expect(getByIdRes.Spec.Assets.Compute).To(Equal(createRequest.Spec.Assets.Compute))
		})

		It("get should fail", func() {
			createHttpReq, err := http.NewRequest("GET", fmt.Sprintf(urlModelById, notExist), nil)
			Expect(err).To(BeNil())
			createHttpReq.Header = *headerForAdmin(randoms)

			createHttpResp := httptest.NewRecorder()
			router.ServeHTTP(createHttpResp, createHttpReq)
			Expect(createHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(createHttpResp.Body.String()).To(ContainSubstring("requested asset does not exist"))
		})
	})
})
