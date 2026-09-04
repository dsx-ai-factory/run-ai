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

	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	workloadsServiceClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/workloads-service"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	policies "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
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
			Expect(getHttpResp.Body.String()).To(ContainSubstring("either workloadId, projectId or templateId must be provided"))
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

	Context("template and workload both given", func() {
		DescribeTable("should be rejected", func(url string) {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s&workloadId=%s", url, uuid.NewString(), uuid.NewString()), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("workloadId and templateId are mutually exclusive"))
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

	Context("template not found", func() {
		DescribeTable("should be rejected", func(url string) {
			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s", url, uuid.NewString()), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusNotFound))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("requested resource does not exist"))
		}, MergeUrls)
	})

	Context("project in template is different than project in query param", func() {
		DescribeTable("should be rejected", func(url string) {
			templateFactory := TemplateTestFactory{router: router}
			template, _, _ := templateFactory.createProjectBasedAsset(randoms, headerForAdmin(randoms))

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s&projectId=%s",
				url, template.Id, randoms.CloneWithDifferentProject().ProjectIdStr), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("template is not associated with projectId"))
		}, MergeUrls)
	})

	Context("project is mandatory for non project template", func() {
		DescribeTable("should be rejected", func(url string) {
			templateFactory := TemplateTestFactory{router: router}
			template, _, _ := templateFactory.createDepartmentBasedAsset(randoms, headerForAdmin(randoms))

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s", url, template.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusBadRequest))
			Expect(getHttpResp.Body.String()).To(ContainSubstring("projectId must be provided for non-project scoped template"))
		}, MergeUrls)
	})

	Context("merge defaults and template", func() {
		DescribeTable("should return correct payload", func(url string) {
			adminHeader := headerForAdmin(randoms)

			workspaceDefaults := policyService.WorkspacePolicyDefaultsV2{
				Command: *policyService.NewNullableString(convutils.Ref("command-from-policy-to-skip")),
				Args:    *policyService.NewNullableString(convutils.Ref("args-from-policy-to-use")),
				EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
					Instances: []policyService.EnvironmentVariable{
						{
							Name:  *policyService.NewNullableString(convutils.Ref("env-1")),
							Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-to-skip")),
						},
						{
							Name:  *policyService.NewNullableString(convutils.Ref("env-2")),
							Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-to-skip")),
						},
						{
							Name:  *policyService.NewNullableString(convutils.Ref("env-3")),
							Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-to-use")),
						},
					},
				}),
			}

			var trainingDefaults policyService.TrainingPolicyDefaultsV2
			err := convutils.CopySpecs(&trainingDefaults, workspaceDefaults)
			Expect(err).To(BeNil())

			var inferenceDefaults policyService.InferencePolicyDefaultsV2
			err = convutils.CopySpecs(&inferenceDefaults, workspaceDefaults)
			Expect(err).To(BeNil())

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveWorkspacePolicyV2: &policyService.WorkspacePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableWorkspacePolicyDefaultsV2(&workspaceDefaults),
					},
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableTrainingPolicyDefaultsV2(&trainingDefaults),
					},
					EffectiveInferencePolicyV2: &policyService.InferencePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableInferencePolicyDefaultsV2(&inferenceDefaults),
					},
				},
			})

			envAsset, _, _ := createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), router, adminHeader)

			templatePayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			templatePayload.Spec.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("spec-command"),
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{
						Name:  convutils.Ref("env-1"),
						Value: convutils.Ref("value-1-from-spec"),
					},
					{
						Name:  convutils.Ref("env-2"),
						Value: convutils.Ref("spec-2-from-spec"),
					},
				},
			}

			template, _, _ := createTemplateAsset(templatePayload, router, adminHeader)

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s", url, template.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			switch url {
			case mergeWorkspaceUrl:
				var actualResults api.WorkspaceV2
				err := json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults)
				Expect(err).To(BeNil())

				Expect(actualResults.Assets.Environment.Id).NotTo(Equal(""))
				Expect(actualResults.Assets.Environment.Name).NotTo(Equal(""))
				actualResults.Assets = api.AssetsRef{}

				CompareFiles(fmt.Sprintf("merge with %s template", url), &actualResults, "merged_template_defaults_ws", ".")

			case mergeTrainingUrl:
				var actualResults api.TrainingV2
				err := json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults)
				Expect(err).To(BeNil())

				Expect(actualResults.Assets.Environment.Id).NotTo(Equal(""))
				Expect(actualResults.Assets.Environment.Name).NotTo(Equal(""))
				actualResults.Assets = api.AssetsRef{}

				CompareFiles(fmt.Sprintf("merge with %s template", url), &actualResults, "merged_template_defaults_train", ".")

			case mergeInferenceUrl:
				var actualResults api.InferenceV2
				err := json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults)
				Expect(err).To(BeNil())

				Expect(actualResults.Assets.Environment.Id).NotTo(Equal(""))
				Expect(actualResults.Assets.Environment.Name).NotTo(Equal(""))
				actualResults.Assets = api.AssetsRef{}

				CompareFiles(fmt.Sprintf("merge with %s template", url), &actualResults, "merged_template_defaults_inference", ".")

			default:
				Expect(true).To(Equal(false), "url not supported")
			}
		}, MergeUrls)
	})

	Context("merge defaults and template - distributed", func() {
		It("should return correct payload", func() {
			adminHeader := headerForAdmin(randoms)

			distributedDefaults := policyService.DistributedPolicyDefaultsV2{
				Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
					Command: *policyService.NewNullableString(convutils.Ref("command-from-policy-worker-to-skip")),
					Args:    *policyService.NewNullableString(convutils.Ref("args-from-policy-worker-to-use")),
					EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
						Instances: []policyService.EnvironmentVariable{
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-1")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-worker-to-skip")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-2")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-worker-to-skip")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-3")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-worker-to-use")),
							},
						},
					}),
				}),
				Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
					Command: *policyService.NewNullableString(convutils.Ref("command-from-policy-master-to-skip")),
					Args:    *policyService.NewNullableString(convutils.Ref("args-from-policy-master-to-use")),
					EnvironmentVariables: *policyService.NewNullableEnvironmentVariablesDefaults(&policyService.EnvironmentVariablesDefaults{
						Instances: []policyService.EnvironmentVariable{
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-1")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-master-to-skip")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-2")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-master-to-skip")),
							},
							{
								Name:  *policyService.NewNullableString(convutils.Ref("env-3")),
								Value: *policyService.NewNullableString(convutils.Ref("value-from-policy-master-to-use")),
							},
						},
					}),
				}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
					EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedPolicyDefaultsV2(&distributedDefaults),
					},
				},
			})

			envAsset, _, _ := createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), router, adminHeader)

			templatePayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			templatePayload.Spec.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("spec-command"),
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{
						Name:  convutils.Ref("env-1"),
						Value: convutils.Ref("value-1-from-spec"),
					},
					{
						Name:  convutils.Ref("env-2"),
						Value: convutils.Ref("spec-2-from-spec"),
					},
				},
			}

			template, _, _ := createTemplateAsset(templatePayload, router, adminHeader)

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s&isDistributed=true", mergeTrainingUrl, template.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualResults api.TrainingV2
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults)
			Expect(err).To(BeNil())

			Expect(actualResults.Assets.Environment.Id).NotTo(Equal(""))
			Expect(actualResults.Assets.Environment.Name).NotTo(Equal(""))
			actualResults.Assets = api.AssetsRef{}

			Expect(actualResults.Distributed.Master.Assets.Environment.Id).NotTo(Equal(""))
			Expect(actualResults.Distributed.Master.Assets.Environment.Name).NotTo(Equal(""))
			actualResults.Distributed.Master.Assets = &api.AssetsRefOptional{}

			CompareFiles(fmt.Sprintf("merge with %s template", mergeTrainingUrl), &actualResults, "merged_template_defaults_dist", ".")
		})
	})

	Context("merge defaults and template - distributed - no policy", func() {
		It("should return correct payload", func() {
			adminHeader := headerForAdmin(randoms)

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: time.Now(),
				},
			})

			envAsset, _, _ := createEnvironmentAsset(buildRandomDepartmentEnvironmentAsset(randoms), router, adminHeader)

			templatePayload := buildRandomProjectTemplate(randoms, envAsset.Meta.Id)
			templatePayload.Spec.SpecificEnv = &api.SpecificRunParams{
				Command: convutils.Ref("spec-command"),
				EnvironmentVariables: &[]api.EnvironmentVariableOfAsset{
					{
						Name:  convutils.Ref("env-1"),
						Value: convutils.Ref("value-1-from-spec"),
					},
					{
						Name:  convutils.Ref("env-2"),
						Value: convutils.Ref("spec-2-from-spec"),
					},
				},
			}

			template, _, _ := createTemplateAsset(templatePayload, router, adminHeader)

			getHttpReq, err := http.NewRequest("GET", fmt.Sprintf("%s?templateId=%s&isDistributed=true", mergeTrainingUrl, template.Meta.Id), nil)
			Expect(err).To(BeNil())
			getHttpReq.Header = *headerForAdmin(randoms)

			getHttpResp := httptest.NewRecorder()
			router.ServeHTTP(getHttpResp, getHttpReq)
			Expect(getHttpResp.Code).To(Equal(http.StatusOK))

			var actualResults api.TrainingV2
			err = json.Unmarshal(getHttpResp.Body.Bytes(), &actualResults)
			Expect(err).To(BeNil())

			Expect(actualResults.Assets.Environment.Id).NotTo(Equal(""))
			Expect(actualResults.Assets.Environment.Name).NotTo(Equal(""))
			actualResults.Assets = api.AssetsRef{}

			Expect(actualResults.Distributed.Master.Assets.Environment.Id).NotTo(Equal(""))
			Expect(actualResults.Distributed.Master.Assets.Environment.Name).NotTo(Equal(""))
			actualResults.Distributed.Master.Assets = &api.AssetsRefOptional{}

			CompareFiles(fmt.Sprintf("merge with %s template no policy", mergeTrainingUrl), &actualResults, "merged_template_defaults_no_policy", ".")
		})
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
