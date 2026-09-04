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
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/go-chi/chi/v5"

	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// the test runs on all assets
var _ = Describe("Test Workspace Service", func() {
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
	})

	Context("count asset invalid scope parameter", func() {
		It("should fail validation", func() {
			countHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?kind=training&scope=xxxyyy", urlCountAssets), nil)
			Expect(err).To(BeNil())
			countHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, countHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("invalid value provided for scope parameter"))
		})
	})

	Context("count asset invalid kind parameter", func() {
		It("should fail validation", func() {
			countHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?kind=xxxyyy", urlCountAssets), nil)
			Expect(err).To(BeNil())
			countHttpReq.Header = *headerForAdmin(randoms)

			countHttpResp := httptest.NewRecorder()
			router.ServeHTTP(countHttpResp, countHttpReq)
			Expect(countHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(countHttpResp.Body.String()).To(ContainSubstring("invalid value provided for kind parameter"))
		})
	})

	Context("count parameters checking", func() {
		It("should return the correct result", func() {
			header := headerForAdmin(randoms)

			createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, header)
			createEnvironmentAsset(buildRandomClusterEnvironmentAsset(randoms), router, header)

			createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), router, header)

			createComputeAsset(buildRandomClusterCompute(randoms), router, header)

			createComputeAsset(buildRandomDepartmentCompute(randoms), router, header)
			createComputeAsset(buildRandomDepartmentCompute(randoms), router, header)
			createComputeAsset(buildRandomDepartmentCompute(randoms), router, header)

			validateCount("", router, header, 7, "total assets count")
			validateCount("kind=environment", router, header, 3, "environment assets")
			validateCount("kind=compute", router, header, 4, "compute assets")
			validateCount(fmt.Sprintf("scope=cluster&clusterId=%s", randoms.ClusterId), router, header, 3, "cluster scoped assets")
			validateCount(fmt.Sprintf("scope=department&departmentId=%s", randoms.DepartmentId), router, header, 4, "department scoped assets")
		})
	})
})

func validateCount(queryString string, router chi.Router, header *http.Header, expecting int64, description string) {
	countHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", urlCountAssets, queryString), nil)
	Expect(err).To(BeNil())
	countHttpReq.Header = *header

	countHttpResp := httptest.NewRecorder()
	router.ServeHTTP(countHttpResp, countHttpReq)
	Expect(countHttpResp.Code).To(Equal(http.StatusOK), description)

	var result api.AssetCountResponse
	err = json.Unmarshal(countHttpResp.Body.Bytes(), &result)
	Expect(err).To(BeNil())

	Expect(result.Result).To(Equal(expecting), description)
}
