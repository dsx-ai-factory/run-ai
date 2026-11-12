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

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	mergeTrainingUrl  = urlTrainingV2 + "/merged"
	mergeWorkspaceUrl = urlWorkspaceV2 + "/merged"
	mergeInferenceUrl = urlInferenceV2 + "/merged"
)

type MergeTestDetails struct {
	TestDetails

	isDistributed bool
	policyFname   string

	expectedMergeError string
}

var MergeUrls = []TableEntry{
	Entry("workspace url", mergeWorkspaceUrl),
	Entry("training url", mergeTrainingUrl),
	Entry("inference url", mergeInferenceUrl),
}

var MergeTests = []TableEntry{
	Entry("policy only - workspace", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:  string(workloadsApi.WORKLOADTYPE_WORKSPACE),
			expectedFname: "merged_policy_only_ws",
		},
		policyFname: "policies/workspace_rich",
	}),

	Entry("policy only - distributed", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
			expectedFname: "merged_policy_only_dist",
		},
		isDistributed: true,
		policyFname:   "policies/distributed_rich",
	}),

	Entry("policy only - worker dist", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:  string(workloadsApi.WORKLOADTYPE_TRAINING),
			expectedFname: "merged_policy_only_dist_worker",
		},
		isDistributed: true,
		policyFname:   "policies/distributed_worker_only",
	}),

	Entry("distributed - no specific params", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
			expectedFname:            "merged_dist_no_specific_params",
			specFname:                "distributed_no_specific_params",
			environmentWorkloadTypes: &envForTraining,
		},
		isDistributed: true,
		policyFname:   "policies/distributed_commands",
	}),

	Entry("distributed - different commands", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
			expectedFname:            "merged_dist_different_commands",
			specFname:                "distributed_different_commands",
			environmentWorkloadTypes: &envForTraining,
		},
		isDistributed: true,
		policyFname:   "policies/distributed_commands",
	}),

	Entry("distributed - masterSameAsWorker - defaults of master taken from worker", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:             string(workloadsApi.WORKLOADTYPE_TRAINING),
			specFname:                "distributed_master_same",
			expectedFname:            "merged_dist_same_as_worker",
			environmentWorkloadTypes: &envForTraining,
		},
		isDistributed: true,
		policyFname:   "policies/distributed_worker_only",
	}),

	Entry("jupyter and tensor", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:             string(workloadsApi.WORKLOADTYPE_WORKSPACE),
			specFname:                "jupyter_tensor",
			expectedFname:            "merged_policy_ws_and_jupyter_tensor",
			environmentWorkloadTypes: &envForWorkspace,
		},
		policyFname: "policies/workspace_rich",
	}),

	Entry("inference", MergeTestDetails{
		TestDetails: TestDetails{
			workloadType:             string(workloadsApi.WORKLOADTYPE_INFERENCE),
			specFname:                "inference_rich",
			expectedFname:            "merged_policy_inference",
			environmentWorkloadTypes: &envForInference,
		},
		policyFname: "policies/inference_rich",
	}),
}

type MergeResults struct {
	Training  *api.TrainingV2  `json:"training,omitempty"`
	Workspace *api.WorkspaceV2 `json:"workspace,omitempty"`
	Inference *api.WorkspaceV2 `json:"inference,omitempty"`
}

var _ = Describe("Workload merge validations", func() {
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

	Context("missing parameters", func() {
		It("should be rejected", func() {
			getHttpReq, err := http.NewRequest("GET", mergeTrainingUrl, nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("either workloadId or projectId must be provided"))
		})
	})

	Context("project and workload both given", func() {
		DescribeTable("should be rejected", func(url string) {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?projectId=%s&workloadId=%s", url, randoms.ProjectIdStr, uuid.NewString()), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("workloadId and projectId are mutually exclusive"))
		}, MergeUrls)
	})

	Context("distributed and workload both given", func() {
		It("should be rejected", func() {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?isDistributed=true&workloadId=%s", mergeTrainingUrl, uuid.NewString()), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("isDistributed and workloadId are mutually exclusive"))
		})
	})

	Context("workload not found", func() {
		DescribeTable("should be rejected", func(url string) {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?workloadId=%s", url, uuid.NewString()), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("cannot find"))
		}, MergeUrls)
	})

	Context("input VS expected tests", func() {
		DescribeTable("should merge workload and policy properly", func(testDetails MergeTestDetails) {
			testDetails.runtimeData.router = router
			testDetails.runtimeData.randoms = randoms
			testDetails.runtimeData.basePath = "."
			testDetails.runtimeData.wsFactory = wsFactory
			testDetails.runtimeData.credentials = make(map[string]string)
			testDetails.RunMergeTest()
		}, MergeTests)
	})
})

func (td *MergeTestDetails) RunMergeTest() {
	testBrief := fmt.Sprintf("Spec: %s, Policy: %s, Type: %s", td.specFname, td.policyFname, td.workloadType)

	if td.policyFname != "" {
		policyYaml, err := testutils.LoadInputFile(td.policyFname, td.runtimeData.basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", td.specFname))

		td.updatePolicyMock(policyYaml, testBrief)
	}

	//
	//   if we want defaults + workload (copy-and-edit mode) we need to create the workload first
	//
	var url string
	var payload map[string]interface{}
	var workloadId string

	if td.TestDetails.specFname != "" {
		specYaml, err := testutils.LoadInputFile(td.specFname, td.runtimeData.basePath)
		Expect(err).To(BeNil(), fmt.Sprintf("failed to load of %s.yaml", td.specFname))

		url, payload = td.TestDetails.determineUrlAndPayload(specYaml, testBrief)

		createHttpReq, err := http.NewRequest("POST", url, strings.NewReader(string(gojsonq.ToBytes(payload))))
		Expect(err).To(BeNil())
		createHttpReq.Header = *headerForAdmin(td.runtimeData.randoms)

		createHttpResp := httptest.NewRecorder()
		td.runtimeData.router.ServeHTTP(createHttpResp, createHttpReq)
		if td.expectedError != "" {
			Expect(createHttpResp.Code).NotTo(Equal(http.StatusOK))
			Expect(createHttpResp.Body.String()).To(ContainSubstring(td.expectedError))
			return
		}
		Expect(createHttpResp.Code).To(Equal(http.StatusOK), createHttpResp.Body.String())

		// we need the workload id from the metadata, so the type is not important
		var resp api.WorkspaceV2
		jsonErr := json.Unmarshal(createHttpResp.Body.Bytes(), &resp)
		Expect(jsonErr).To(BeNil())

		workloadId = resp.Meta.WorkloadId
	} else {
		url, payload = td.TestDetails.determineUrlAndPayload(nil, testBrief)
	}

	//
	//   call the merge workload API either wit project (policy defaults only) or workload id
	//
	if workloadId == "" {
		url = fmt.Sprintf("%s/merged?projectId=%d", url, td.TestDetails.runtimeData.randoms.ProjectId)
		if td.isDistributed {
			url += "&isDistributed=true"
		}
	} else {
		url = fmt.Sprintf("%s/merged?workloadId=%s", url, workloadId)
	}

	//
	//   now run the merge workload with the combination of parameters that we composed above
	//
	getHttpReq, err := http.NewRequest("GET", url, nil)
	Expect(err).To(BeNil())
	getHttpReq.Header = *headerForAdmin(td.runtimeData.randoms)

	getHttpResp := httptest.NewRecorder()
	td.runtimeData.router.ServeHTTP(getHttpResp, getHttpReq)
	if td.expectedMergeError != "" {
		Expect(getHttpResp.Code).NotTo(Equal(http.StatusOK))
		Expect(getHttpResp.Body.String()).To(ContainSubstring(td.expectedMergeError))
		return
	}
	Expect(getHttpResp.Code).To(Equal(http.StatusOK))

	var jsonErr error
	actualResults := MergeResults{}

	switch td.TestDetails.workloadType {
	case string(workloadsApi.WORKLOADTYPE_WORKSPACE):
		jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults.Workspace)
		actualResults.Workspace.Meta = api.WorkloadMeta{} // sanitize metadata, we do not need it in this test
		sanitizeAssetIds(&actualResults.Workspace.Assets)
	case string(workloadsApi.WORKLOADTYPE_TRAINING):
		jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults.Training)
		actualResults.Training.Meta = api.WorkloadMeta{} // sanitize metadata, we do not need it in this test
		sanitizeAssetIds(&actualResults.Training.Assets)
		if actualResults.Training.Distributed != nil && actualResults.Training.Distributed.Master != nil {
			sanitizeAssetIdsMaster(actualResults.Training.Distributed.Master.Assets)
		}
	case string(workloadsApi.WORKLOADTYPE_INFERENCE):
		jsonErr = json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults.Inference)
		actualResults.Inference.Meta = api.WorkloadMeta{} // sanitize metadata, we do not need it in this test
		sanitizeAssetIds(&actualResults.Inference.Assets)
	default:
		jsonErr = fmt.Errorf("workload type %s not supported", td.TestDetails.workloadType)
	}
	Expect(jsonErr).To(BeNil(), testBrief)

	CompareFiles(testBrief, &actualResults, td.expectedFname, td.runtimeData.basePath)
}

func (td *MergeTestDetails) updatePolicyMock(policyYaml []byte, testBrief string) {
	mock := policies.SimpleMockClient{
		EffectiveUpdatedAt: time.Now(),
	}

	var err error

	switch td.workloadType {
	case string(workloadsApi.WORKLOADTYPE_WORKSPACE):
		err = yaml.Unmarshal(policyYaml, &mock.EffectiveWorkspacePolicyV2)

	case string(workloadsApi.WORKLOADTYPE_TRAINING):
		if td.isDistributed {
			err = yaml.Unmarshal(policyYaml, &mock.EffectiveDistributedPolicyV2)
		} else {
			err = yaml.Unmarshal(policyYaml, &mock.EffectiveTrainingPolicyV2)
		}

	case string(workloadsApi.WORKLOADTYPE_INFERENCE):
		err = yaml.Unmarshal(policyYaml, &mock.EffectiveInferencePolicyV2)
	}

	Expect(err).To(BeNil(), testBrief)

	policies.SetMock(policies.SimpleMockClientFactory{
		SimpleMockClient: mock,
	})
}

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

func sanitizeAssetIdsMaster(assets *api.AssetsRefOptional) {
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
