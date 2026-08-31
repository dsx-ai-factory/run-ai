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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/policies"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"

	"github.com/run-ai/runai/runai-application-packages/testutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/run-ai/runai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/run-ai/runai/control-plane/assets-service/pkg/database/testing"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/testing"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
)

const (
	urlAssets = "/api/v1/asset"
	urlToFlat = "/api/v1/asset/flat"
)

var _ = Describe("assets validation", func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		dbTesting.CleanTables()

		randoms = testutils.NewRandomValues()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		authorization.SetMock(authMock)
		workloadsServiceClient.SetMock(workloadsServiceClient.NewSimpleMockClientFactory())

		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))

		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: time.Now(),
			},
		})
	})

	Context("request flat info", func() {
		It("asset should exist ", func() {
			flatRequestPayload := api.AssetFlatRequest{
				AssetId:      uuid.NewString(),
				WorkloadType: api.WORKLOADTYPE_WORKSPACE,
				ClusterId:    randoms.ClusterId,
				ProjectId:    randoms.ProjectIdStr,
			}

			flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
			Expect(err).To(BeNil())
			flatRequest.Header = *headerForAdmin(randoms)

			flatResponse := httptest.NewRecorder()
			router.ServeHTTP(flatResponse, flatRequest)
			Expect(flatResponse.Code).To(Equal(http.StatusNotFound))
		})

		It("single asset returned correctly - workspace", func() {
			pvcAsset, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			flatRequestPayload := api.AssetFlatRequest{
				AssetId:      pvcAsset.Meta.Id,
				WorkloadType: api.WORKLOADTYPE_WORKSPACE,
				ClusterId:    randoms.ClusterId,
				ProjectId:    randoms.ProjectIdStr,
			}

			flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
			Expect(err).To(BeNil())
			flatRequest.Header = *headerForAdmin(randoms)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, flatRequest)
			Expect(response.Code).To(Equal(http.StatusOK))

			var flatResponse api.AssetFlatResponse
			jsonErr := json.Unmarshal(response.Body.Bytes(), &flatResponse)
			Expect(jsonErr).To(BeNil())

			Expect(flatResponse.Worker).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage.Pvc).NotTo(BeNil())
			Expect((*flatResponse.Worker.Storage.Pvc)[0].Path).To(Equal(pvcAsset.Spec.Path))
			Expect(*(*flatResponse.Worker.Storage.Pvc)[0].ExistingPvc).To(Equal(true))
			Expect((*flatResponse.Worker.Storage.Pvc)[0].ClaimName).To(Equal(pvcAsset.Spec.ClaimName))
		})

		It("multiple assets returned correctly - workspace", func() {
			pvcAsset1, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))
			pvcAsset2, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))
			nfsAsset1, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))
			nfsAsset2, _, _ := createNFSAsset(buildRandomTenantNFSAsset(), router, headerForAdmin(randoms))

			flatRequestPayload := api.AssetFlatRequest{
				AssetId:      pvcAsset1.Meta.Id,
				AssetIds:     []string{pvcAsset2.Meta.Id, nfsAsset1.Meta.Id, nfsAsset2.Meta.Id},
				WorkloadType: api.WORKLOADTYPE_WORKSPACE,
				ClusterId:    randoms.ClusterId,
				ProjectId:    randoms.ProjectIdStr,
			}

			flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
			Expect(err).To(BeNil())
			flatRequest.Header = *headerForAdmin(randoms)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, flatRequest)
			Expect(response.Code).To(Equal(http.StatusOK))

			var flatResponse api.AssetFlatResponse
			jsonErr := json.Unmarshal(response.Body.Bytes(), &flatResponse)
			Expect(jsonErr).To(BeNil())

			Expect(flatResponse.Worker).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage.Pvc).NotTo(BeNil())
			Expect(len(*flatResponse.Worker.Storage.Pvc)).To(Equal(2))
			Expect((*flatResponse.Worker.Storage.Pvc)[0].ClaimName).NotTo(Equal((*flatResponse.Worker.Storage.Pvc)[1].ClaimName))
			Expect((*flatResponse.Worker.Storage.Nfs)[0].Path).NotTo(Equal((*flatResponse.Worker.Storage.Nfs)[1].Path))
		})

		It("single asset returned correctly - distributed", func() {
			pvcAsset, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			flatRequestPayload := api.AssetFlatRequest{
				AssetId:      pvcAsset.Meta.Id,
				WorkloadType: api.WORKLOADTYPE_DISTRIBUTED,
				ClusterId:    randoms.ClusterId,
				ProjectId:    randoms.ProjectIdStr,
			}

			flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
			Expect(err).To(BeNil())
			flatRequest.Header = *headerForAdmin(randoms)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, flatRequest)
			Expect(response.Code).To(Equal(http.StatusOK))

			var flatResponse api.AssetFlatResponse
			jsonErr := json.Unmarshal(response.Body.Bytes(), &flatResponse)
			Expect(jsonErr).To(BeNil())

			Expect(flatResponse.Worker).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage.Pvc).NotTo(BeNil())
			Expect((*flatResponse.Worker.Storage.Pvc)[0].Path).To(Equal(pvcAsset.Spec.Path))
			Expect(*(*flatResponse.Worker.Storage.Pvc)[0].ExistingPvc).To(Equal(true))
			Expect((*flatResponse.Worker.Storage.Pvc)[0].ClaimName).To(Equal(pvcAsset.Spec.ClaimName))

			Expect(flatResponse.Master).NotTo(BeNil())
			Expect(flatResponse.Master.Storage).NotTo(BeNil())
			Expect(flatResponse.Master.Storage.Pvc).NotTo(BeNil())
			Expect((*flatResponse.Master.Storage.Pvc)[0].Path).To(Equal(pvcAsset.Spec.Path))
			Expect(*(*flatResponse.Master.Storage.Pvc)[0].ExistingPvc).To(Equal(true))
			Expect((*flatResponse.Master.Storage.Pvc)[0].ClaimName).To(Equal(pvcAsset.Spec.ClaimName))
		})

		It("template asset returned correctly", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			pvcAsset, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			templateReq := buildRandomTemplate(api.SCOPE_PROJECT, randoms, envAsset.Meta.Id, computeAsset.Meta, pvcAsset.Meta)
			templateReq.Spec.SpecificEnv = &api.SpecificRunParams{
				Annotations: &[]api.Annotation{
					{
						Name:  convutils.Ref("annotation_1"),
						Value: convutils.Ref("value_1"),
					},
					{
						Name:  convutils.Ref("annotation_2"),
						Value: convutils.Ref("value_2"),
					},
				},
				Command: convutils.Ref("command-spec"),
			}
			templateAsset, _, _ := createTemplateAsset(templateReq, router, headerForAdmin(randoms))

			flatRequestPayload := api.AssetFlatRequest{
				AssetId:      templateAsset.Meta.Id,
				WorkloadType: api.WORKLOADTYPE_WORKSPACE,
				ClusterId:    randoms.ClusterId,
				ProjectId:    randoms.ProjectIdStr,
			}

			flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
			Expect(err).To(BeNil())
			flatRequest.Header = *headerForAdmin(randoms)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, flatRequest)
			Expect(response.Code).To(Equal(http.StatusOK))

			var flatResponse api.AssetFlatResponse
			jsonErr := json.Unmarshal(response.Body.Bytes(), &flatResponse)
			Expect(jsonErr).To(BeNil())

			Expect(flatResponse.Worker).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage).NotTo(BeNil())
			Expect(flatResponse.Worker.Storage.Pvc).NotTo(BeNil())

			Expect((*flatResponse.Worker.Storage.Pvc)[0].Path).To(Equal(pvcAsset.Spec.Path))
			Expect(*(*flatResponse.Worker.Storage.Pvc)[0].ExistingPvc).To(Equal(true))
			Expect((*flatResponse.Worker.Storage.Pvc)[0].ClaimName).To(Equal(pvcAsset.Spec.ClaimName))

			Expect(flatResponse.Worker.Image).To(Equal(envAsset.Spec.Image))

			Expect(flatResponse.Worker.Command).NotTo(BeNil())
			Expect(*flatResponse.Worker.Command).To(Equal(*templateReq.Spec.SpecificEnv.Command))

			Expect(flatResponse.Worker.Annotations).NotTo(BeNil())
			Expect(len(*flatResponse.Worker.Annotations)).To(Equal(2))

			Expect(flatResponse.Worker.Compute).NotTo(BeNil())
			Expect(flatResponse.Worker.Compute.CpuMemoryLimit).To(Equal(computeAsset.Spec.CpuMemoryLimit))
			Expect(flatResponse.Worker.Compute.ExtendedResources).NotTo(BeNil())
			Expect(len(*flatResponse.Worker.Compute.ExtendedResources)).To(Equal(len(*computeAsset.Spec.ExtendedResources)))
		})
	})

	Context("validations", func() {
		It("should validate query params", func() {
			listRequest1, err := http.NewRequest("GET", urlAssets, nil)
			Expect(err).To(BeNil())
			listRequest1.Header = *headerForAdmin(randoms)

			listResp1 := httptest.NewRecorder()
			router.ServeHTTP(listResp1, listRequest1)
			Expect(listResp1.Code).To(Equal(http.StatusBadRequest))
			Expect(listResp1.Body.String()).To(ContainSubstring("required field 'id' is zero value."))
		})

		It("should validate the uuid", func() {
			listRequest1, err := http.NewRequest("GET", fmt.Sprintf("%s?id=abc", urlAssets), nil)
			Expect(err).To(BeNil())
			listRequest1.Header = *headerForAdmin(randoms)

			listResp1 := httptest.NewRecorder()
			router.ServeHTTP(listResp1, listRequest1)
			Expect(listResp1.Code).To(Equal(http.StatusBadRequest))
			Expect(listResp1.Body.String()).To(ContainSubstring("malformed uuid abc"))
		})
	})

	Context("get flat info, asset not found", func() {
		It("all relevant info is returned ", func() {
			envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			computeAsset, _, _ := createComputeAsset(buildRandomTenantCompute(), router, headerForAdmin(randoms))
			pvcAsset, _, _ := createPVCAsset(testing.BuildRandomProjectPVCAsset(randoms), router, headerForAdmin(randoms))

			listRequest1, err := http.NewRequest("GET", fmt.Sprintf("%s?id=%s,%s", urlAssets, envAsset.Meta.Id, computeAsset.Meta.Id), nil)
			Expect(err).To(BeNil())
			listRequest1.Header = *headerForAdmin(randoms)

			listResp1 := httptest.NewRecorder()
			router.ServeHTTP(listResp1, listRequest1)
			Expect(listResp1.Code).To(Equal(http.StatusOK))

			var resp api.AssetListByIdsResponse
			respJsonErr := json.Unmarshal(listResp1.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(2))

			listRequest2, err := http.NewRequest("GET", fmt.Sprintf("%s?id=%s,%s,%s", urlAssets, envAsset.Meta.Id, computeAsset.Meta.Id, pvcAsset.Meta.Id), nil)
			Expect(err).To(BeNil())
			listRequest2.Header = *headerForAdmin(randoms)

			listResp2 := httptest.NewRecorder()
			router.ServeHTTP(listResp2, listRequest2)
			Expect(listResp1.Code).To(Equal(http.StatusOK))

			respJsonErr = json.Unmarshal(listResp2.Body.Bytes(), &resp)
			Expect(respJsonErr).To(BeNil())
			Expect(len(resp.Entries)).To(Equal(3))
		})
	})

	Context("get assets flat affects last used", func() {
		It("last used must be affected", func() {
			envAsset1, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAsset2, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))
			envAsset3, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

			for cycle := 1; cycle <= 2; cycle++ {
				// cycle1: touch 1, then 2, then 3, with small gap b/w them
				// cycle2: touch 3, then 2, then 1
				if cycle == 1 {
					touchAsset(router, randoms, envAsset1.Meta.Id)
					time.Sleep(100 * time.Millisecond)
					touchAsset(router, randoms, envAsset2.Meta.Id)
					time.Sleep(100 * time.Millisecond)
					touchAsset(router, randoms, envAsset3.Meta.Id)
				} else {
					touchAsset(router, randoms, envAsset3.Meta.Id)
					time.Sleep(100 * time.Millisecond)
					touchAsset(router, randoms, envAsset2.Meta.Id)
					time.Sleep(100 * time.Millisecond)
					touchAsset(router, randoms, envAsset1.Meta.Id)
				}

				listRequest, err := http.NewRequest("GET", fmt.Sprintf("%s?complyToWorkloadType=%s&complyToProject=%s",
					urlEnvironment, api.WORKLOADTYPE_WORKSPACE, randoms.ProjectIdStr), nil)
				Expect(err).To(BeNil())
				listRequest.Header = *headerForAdmin(randoms)

				listResp := httptest.NewRecorder()
				router.ServeHTTP(listResp, listRequest)
				Expect(listResp.Code).To(Equal(http.StatusOK))

				var resp api.EnvironmentListResponse
				respJsonErr := json.Unmarshal(listResp.Body.Bytes(), &resp)
				Expect(respJsonErr).To(BeNil())
				Expect(len(resp.Entries)).To(Equal(3))

				lastUsedTime := make(map[string]time.Time)
				for _, entry := range resp.Entries {
					lastUsedTime[entry.Meta.Id] = *entry.UsageTimes.LastUsedByWorkload
				}

				lastUsedEnv1 := lastUsedTime[envAsset1.Meta.Id].UnixMilli()
				lastUsedEnv2 := lastUsedTime[envAsset2.Meta.Id].UnixMilli()
				lastUsedEnv3 := lastUsedTime[envAsset3.Meta.Id].UnixMilli()

				var compareOper string
				if cycle == 1 {
					compareOper = "<"
				} else {
					compareOper = ">"
				}
				Expect(lastUsedEnv1).To(BeNumerically(compareOper, lastUsedEnv2))
				Expect(lastUsedEnv2).To(BeNumerically(compareOper, lastUsedEnv3))
			}
		})

		Context("Validations", func() {
			It("template and assets cannot be provided together", func() {
				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				templateReq := buildRandomTemplate(api.SCOPE_TENANT, randoms, envAsset.Meta.Id)
				templateAsset, _, _ := createTemplateAsset(templateReq, router, headerForAdmin(randoms))

				flatRequestPayload := api.AssetFlatRequest{
					AssetIds:     []string{templateAsset.Meta.Id, envAsset.Meta.Id},
					WorkloadType: api.WORKLOADTYPE_WORKSPACE,
					ClusterId:    randoms.ClusterId,
					ProjectId:    randoms.ProjectIdStr,
				}

				flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
				Expect(err).To(BeNil())
				flatRequest.Header = *headerForAdmin(randoms)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, flatRequest)
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body.Bytes()).To(ContainSubstring("multiple assetIds cannot be provided when assetId of a workload template is specified"))
			})

			It("multiple defined asset", func() {
				envAsset, _, _ := createEnvironmentAsset(buildRandomTenantEnvironmentAsset(randoms), router, headerForAdmin(randoms))

				flatRequestPayload := api.AssetFlatRequest{
					AssetId:      envAsset.Meta.Id,
					AssetIds:     []string{envAsset.Meta.Id},
					WorkloadType: api.WORKLOADTYPE_WORKSPACE,
					ClusterId:    randoms.ClusterId,
					ProjectId:    randoms.ProjectIdStr,
				}

				flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
				Expect(err).To(BeNil())
				flatRequest.Header = *headerForAdmin(randoms)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, flatRequest)
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body.Bytes()).To(ContainSubstring("provided more than once"))
			})

		})

	})
})

func touchAsset(router chi.Router, randoms *testutils.RandomValues, assetId string) {
	flatRequestPayload := api.AssetFlatRequest{
		WorkloadType: api.WORKLOADTYPE_WORKSPACE,
		ClusterId:    randoms.ClusterId,
		ProjectId:    randoms.ProjectIdStr,
		AssetId:      assetId,
	}

	flatRequest, err := http.NewRequest("POST", urlToFlat, makePayload(flatRequestPayload))
	Expect(err).To(BeNil())
	flatRequest.Header = *headerForAdmin(randoms)

	flatResponse := httptest.NewRecorder()
	router.ServeHTTP(flatResponse, flatRequest)
	Expect(flatResponse.Code).To(Equal(http.StatusOK))
}
