/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/gomega"
)

func expectResetErrorWorkspace(router chi.Router, patchBody *api.WorkspacePolicyChangeRequestV2, resetField string, expectedMessage string) {
	patchBody.Reset = []string{
		resetField,
	}

	request, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(patchBody))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(2)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if expectedMessage == "" {
		Expect(response.Code).To(Equal(http.StatusOK))
	} else {
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring(expectedMessage))
	}
}

func expectResetErrorDistributed(router chi.Router, patchBody *api.DistributedPolicyChangeRequestV2, resetField string, expectedMessage string) {
	patchBody.Reset = []string{
		resetField,
	}

	request, err := http.NewRequest(http.MethodPatch, urlPolicyDistributed, makePayload(patchBody))
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(2)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if expectedMessage == "" {
		Expect(response.Code).To(Equal(http.StatusOK))
	} else {
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring(expectedMessage))
	}
}

func listPoliciesForTest(router chi.Router, tenantId int32, url string) api.PolicyListResponse {
	requestList, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	requestList.Header = *headerForAdmin(tenantId)

	responseList := httptest.NewRecorder()
	router.ServeHTTP(responseList, requestList.WithContext(tenantContextForTest(tenantId)))
	Expect(responseList.Code).To(Equal(http.StatusOK))

	var actualListResponse api.PolicyListResponse
	err = json.Unmarshal(responseList.Body.Bytes(), &actualListResponse)
	Expect(err).To(BeNil())

	return actualListResponse
}

type SupportedPolicyForTest interface {
	api.WorkspacePolicyV2 | api.TrainingPolicyV2 | api.DistributedPolicyV2 | api.InferencePolicyV2 | api.DistributedInferencePolicyV2 | api.FallbackPolicy | api.NimServicePolicy
}

func getPolicyForTest[T SupportedPolicyForTest](router chi.Router, tenantId int32, url string) T {
	requestGet, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	requestGet.Header = *headerForAdmin(tenantId)

	responseGet := httptest.NewRecorder()
	router.ServeHTTP(responseGet, requestGet.WithContext(tenantContextForTest(tenantId)))
	Expect(responseGet.Code).To(Equal(http.StatusOK))

	var actualGetResponse T
	err = json.Unmarshal(responseGet.Body.Bytes(), &actualGetResponse)
	Expect(err).To(BeNil())

	return actualGetResponse
}

func getWorkspacePolicyForTest(router chi.Router, tenantId int32, url string) api.WorkspacePolicyV2 {
	return getPolicyForTest[api.WorkspacePolicyV2](router, tenantId, url)
}

func getTrainingPolicyForTest(router chi.Router, tenantId int32, url string) api.TrainingPolicyV2 {
	return getPolicyForTest[api.TrainingPolicyV2](router, tenantId, url)
}

func getDistributedInferencePolicyForTest(router chi.Router, tenantId int32, url string) api.DistributedInferencePolicyV2 {
	return getPolicyForTest[api.DistributedInferencePolicyV2](router, tenantId, url)
}

func getDistributedPolicyForTest(router chi.Router, tenantId int32, url string) api.DistributedPolicyV2 {
	return getPolicyForTest[api.DistributedPolicyV2](router, tenantId, url)
}

func getInferencePolicyForTest(router chi.Router, tenantId int32, url string) api.InferencePolicyV2 {
	return getPolicyForTest[api.InferencePolicyV2](router, tenantId, url)
}

func getFallbackPolicyForTest(router chi.Router, tenantId int32, url string) api.FallbackPolicy {
	return getPolicyForTest[api.FallbackPolicy](router, tenantId, url)
}

func expectingNotToFindPolicy(router chi.Router, tenantId int32, url string) {
	requestGet, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	requestGet.Header = *headerForAdmin(tenantId)

	responseGet := httptest.NewRecorder()
	router.ServeHTTP(responseGet, requestGet.WithContext(tenantContextForTest(tenantId)))
	Expect(responseGet.Code).To(Equal(http.StatusNotFound))
}

func deletePolicyForTest(router chi.Router, tenantId int32, url string) {
	requestDelete, err := http.NewRequest("DELETE", url, nil)
	Expect(err).To(BeNil())
	requestDelete.Header = *headerForAdmin(tenantId)

	responseDelete := httptest.NewRecorder()
	router.ServeHTTP(responseDelete, requestDelete.WithContext(tenantContextForTest(tenantId)))
	Expect(responseDelete.Code).To(Equal(http.StatusNoContent))
}

func applyPolicyForTest(router chi.Router, url string, bodyStr *strings.Reader, tenantId int32, verb string) *httptest.ResponseRecorder {
	request, err := http.NewRequest(verb, url, bodyStr)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	return response
}

func applySyncConfig(router chi.Router, url string, bodyStr *strings.Reader, tenantId int32) *httptest.ResponseRecorder {
	request, err := http.NewRequest(http.MethodPatch, url, bodyStr)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())

	return response
}

func expectError(router chi.Router, url string, bodyStr *strings.Reader, tenantId int32, verb string, expectedErrCode int) string {
	request, err := http.NewRequest(verb, url, bodyStr)
	Expect(err).To(BeNil())
	request.Header = *headerForAdmin(tenantId)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request.WithContext(tenantContextForTest(tenantId)))
	Expect(response.Code).To(Equal(expectedErrCode), response.Body.String())

	return response.Body.String()
}

func updateWorkspacePolicyForTest(router chi.Router, body *api.WorkspacePolicyChangeRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyWorkspace, makePayload(*body), tenantId, http.MethodPatch)
}

func overwriteWorkspacePolicyForTest(router chi.Router, body *api.WorkspacePolicyOverwriteRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyWorkspace, makePayload(*body), tenantId, http.MethodPut)
}

func updateTrainingPolicyForTest(router chi.Router, body *api.TrainingPolicyChangeRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyTraining, makePayload(*body), tenantId, http.MethodPatch)
}

func overwriteTrainingPolicyForTest(router chi.Router, body *api.TrainingPolicyOverwriteRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyTraining, makePayload(*body), tenantId, http.MethodPut)
}

func updateDistributedPolicyForTest(router chi.Router, body *api.DistributedPolicyChangeRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyDistributed, makePayload(*body), tenantId, http.MethodPatch)
}

func updateFallbackPolicyForTest(router chi.Router, body *api.FallbackPolicyChangeRequest, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyFallback, makePayload(*body), tenantId, http.MethodPatch)
}

func overwriteDistributedPolicyForTest(router chi.Router, body *api.DistributedPolicyOverwriteRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyDistributed, makePayload(*body), tenantId, http.MethodPut)
}

func updateInferencePolicyForTest(router chi.Router, body *api.InferencePolicyChangeRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyInference, makePayload(*body), tenantId, http.MethodPatch)
}

func overwriteInferencePolicyForTest(router chi.Router, body *api.InferencePolicyOverwriteRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyInference, makePayload(*body), tenantId, http.MethodPut)
}

func updateDistributedInferencePolicyForTest(router chi.Router, body *api.DistributedInferencePolicyChangeRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyDistributedInference, makePayload(*body), tenantId, http.MethodPatch)
}

func overwriteDistributedInferencePolicyForTest(router chi.Router, body *api.DistributedInferencePolicyOverwriteRequestV2, tenantId int32) *httptest.ResponseRecorder {
	return applyPolicyForTest(router, urlPolicyDistributedInference, makePayload(*body), tenantId, http.MethodPut)
}
