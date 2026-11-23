/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const mergeInferenceUrl = urlInferenceV2 + "/merged"

var _ = Describe("Inference workload merge validations", func() {
	router := CreateRouter()
	randoms := testutils.NewRandomValues()

	wsFactory := workloadsServiceClient.SimpleMockClientFactory{
		SimpleMockClient: &workloadsServiceClient.SimpleMockClient{},
	}

	BeforeEach(func() {
		dbTesting.CleanTables()

		authMock := testing.NewRandomsBasedAuthMockClient(randoms)
		authorization.SetMock(authMock)
		org_unit.SetMock(org_unit.NewSimpleMock(randoms))
		workloadsServiceClient.SetMock(&wsFactory)

		wsFactory.Clear()

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms, testing.UnitTestDefaultClusterVersion, false)))
		checker.InitClusterCapabilitiesChecker("", DefaultCheckerClusterConfig, "")
	})

	DescribeTable("validates merged inference queries", func(query string, status int, message string) {
		req := httptest.NewRequest(http.MethodGet, mergeInferenceUrl+query, nil)
		req.Header = *headerForAdmin(randoms)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)
		Expect(resp.Code).To(Equal(status))
		Expect(resp.Body.String()).To(ContainSubstring(message))
	},
		Entry("missing parameters", "", http.StatusBadRequest, "either workloadId or projectId must be provided"),
		Entry("project and workload both given", "?projectId=1&workloadId="+uuid.NewString(), http.StatusBadRequest, "workloadId and projectId are mutually exclusive"),
		Entry("unknown workload", "?workloadId="+uuid.NewString(), http.StatusNotFound, "cannot find"),
	)

	Context("inference policy", func() {
		BeforeEach(func() {
			policyYaml, err := testutils.LoadInputFile("policies/inference_rich", ".")
			Expect(err).NotTo(HaveOccurred())
			policyClient := policies.SimpleMockClient{EffectiveUpdatedAt: time.Now()}
			Expect(yaml.Unmarshal(policyYaml, &policyClient.EffectiveInferencePolicyV2)).To(Succeed())
			policies.SetMock(policies.SimpleMockClientFactory{SimpleMockClient: policyClient})
		})

		getMerged := func(query string) api.InferenceV2 {
			GinkgoHelper()
			req := httptest.NewRequest(http.MethodGet, mergeInferenceUrl+query, nil)
			req.Header = *headerForAdmin(randoms)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())

			var result api.InferenceV2
			Expect(json.Unmarshal(resp.Body.Bytes(), &result)).To(Succeed())
			return result
		}

		It("returns the project policy when creating an inference", func() {
			merged := getMerged("?projectId=" + randoms.ProjectIdStr)
			Expect(merged.PolicyInfo).NotTo(BeNil())
			Expect(merged.SpecificRunParams).NotTo(BeNil())
			Expect(merged.SpecificRunParams.Command).To(HaveValue(Equal("my-command-1")))
			Expect(merged.SpecificRunParams.RunAsUid).To(HaveValue(Equal(int64(500))))
		})

		It("merges policy with a saved inference for Copy", func() {
			td := TestDetails{
				specFname: "inference_rich", expectedFname: "merged_policy_inference",
				environmentWorkloadTypes: &envForInference,
				runtimeData:              RuntimeTestData{router: router, randoms: randoms, basePath: ".", credentials: make(map[string]string)},
			}
			specYaml, err := testutils.LoadInputFile(td.specFname, td.runtimeData.basePath)
			Expect(err).NotTo(HaveOccurred())
			payload := td.inferenceRequest(specYaml, td.specFname)
			req := httptest.NewRequest(http.MethodPost, urlInferenceV2, makePayload(payload))
			req.Header = *headerForAdmin(randoms)
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK), resp.Body.String())
			var created api.InferenceV2
			Expect(json.Unmarshal(resp.Body.Bytes(), &created)).To(Succeed())

			merged := getMerged("?workloadId=" + created.Meta.WorkloadId)
			merged.Meta = api.WorkloadMeta{}
			sanitizeAssetIds(&merged.Assets)
			CompareFiles(td.specFname, map[string]api.InferenceV2{"inference": merged}, td.expectedFname, ".")
		})
	})
})

func sanitizeAssetIds(assets *api.AssetsRef) {
	if assets == nil {
		return
	}
	if assets.Environment.Id != "" {
		assets.Environment.Id = "environment"
	}

	if assets.Compute != nil {
		assets.Compute.Id = "compute"
	}

	if assets.Datasources != nil {
		for idx := 0; idx < len(assets.Datasources); idx++ {
			assets.Datasources[idx].Id = "datasource"
		}
	}
}
