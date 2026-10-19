/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	clusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const DefaultClusterVersionForTesting = "9.9.9"

var _ = Describe("policy compliance validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues
	var router chi.Router

	var clusterClientMock clusterClient.SimpleMockClient
	defaultCheckerClusterConfig := &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}

	mockTenantSettingsClient := &bc_mocks.TenantSettingsClient{}

	BeforeEach(func() {
		// The shared process-wide tenant id (see testutils.NewRandomValues) means any
		// leftover template rows from other suites can be picked up by our listings,
		// so start every spec from a known-empty state.
		testing.CleanTables()

		randoms = testutils.NewRandomValues()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		defaultNodeAffinity := &policyService.NodeAffinityRequired{NodeSelectorTerms: []policyService.NodeSelectorTerm{
			{
				MatchExpressions: []policyService.MatchExpression{
					{
						Key:      "key",
						Operator: policyService.MATCHEXPRESSIONOPERATOR_EXISTS,
					},
				},
			},
		}}

		policyMockUpdateTime := time.Now()

		policyMock := policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
				EffectiveWorkspacePolicyV2: &policyService.WorkspacePolicyDefaultsAndRulesV2{
					Rules: policyService.NullableWorkspacePolicyRulesV2{},
				},
				EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
					Rules: policyService.NullableTrainingPolicyRulesV2{},
				},
				EffectiveDistributedPolicyV2: &policyService.DistributedPolicyDefaultsAndRulesV2{
					Rules:    policyService.NullableDistributedPolicyRulesV2{},
					Defaults: policyService.NullableDistributedPolicyDefaultsV2{},
				},
			},
		}

		policyMock.EffectiveWorkspacePolicyV2.Rules.Set(convutils.Ref(policyService.WorkspacePolicyRulesV2{
			Compute: *policyService.NewNullableSupersetRulesAllOfCompute(convutils.Ref(policyService.SupersetRulesAllOfCompute{
				GpuDevicesRequest: *policyService.NewNullableIntegerRules(convutils.Ref(policyService.IntegerRules{
					Min: *policyService.NewNullableInt32(convutils.Ref(int32(3))),
					Max: *policyService.NewNullableInt32(convutils.Ref(int32(8))),
				})),
				CpuCoreRequest: *policyService.NewNullableNumberRules(convutils.Ref(policyService.NumberRules{
					Max: *policyService.NewNullableFloat32(convutils.Ref(float32(20))),
					DefaultFrom: *policyService.NewNullableDefaultFromRule(convutils.Ref(policyService.DefaultFromRule{
						Field:  *policyService.NewNullableString(convutils.Ref("compute.cpuCoreLimit")),
						Factor: *policyService.NewNullableFloat64(convutils.Ref(0.2)),
					})),
				})),
			})),
			NodeAffinityRequired: *policyService.NewNullableArbitraryRules(convutils.Ref(policyService.ArbitraryRules{
				CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
			})),
		}))

		policyMock.EffectiveWorkspacePolicyV2.Defaults.Set(convutils.Ref(policyService.WorkspacePolicyDefaultsV2{
			NodeAffinityRequired: *policyService.NewNullableNodeAffinityRequired(defaultNodeAffinity),
		}))

		rules, _ := gojsonq.LoadJson(policyMock.EffectiveWorkspacePolicyV2.Rules.Get())
		policyMock.EffectiveWorkspacePolicyV2.DefaultFromOrder.Set(calculateOrder(rules, rules))

		policyMock.EffectiveTrainingPolicyV2.Rules.Set(convutils.Ref(policyService.TrainingPolicyRulesV2{
			Compute: *policyService.NewNullableSupersetRulesAllOfCompute(convutils.Ref(policyService.SupersetRulesAllOfCompute{
				GpuDevicesRequest: *policyService.NewNullableIntegerRules(convutils.Ref(policyService.IntegerRules{
					Min: *policyService.NewNullableInt32(convutils.Ref(int32(5))),
					Max: *policyService.NewNullableInt32(convutils.Ref(int32(10))),
				})),
			})),
		}))

		rules, _ = gojsonq.LoadJson(policyMock.EffectiveTrainingPolicyV2.Rules.Get())
		policyMock.EffectiveTrainingPolicyV2.DefaultFromOrder.Set(calculateOrder(rules, rules))

		policyMock.EffectiveDistributedPolicyV2.Rules.Set(convutils.Ref(policyService.DistributedPolicyRulesV2{
			Master: *policyService.NewNullableReplicaRulesV2(&policyService.ReplicaRulesV2{
				Compute: *policyService.NewNullableSupersetRulesAllOfCompute(convutils.Ref(policyService.SupersetRulesAllOfCompute{
					GpuDevicesRequest: *policyService.NewNullableIntegerRules(convutils.Ref(policyService.IntegerRules{
						CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
					})),
					CpuCoreRequest: *policyService.NewNullableNumberRules(convutils.Ref(policyService.NumberRules{
						Max: *policyService.NewNullableFloat32(convutils.Ref(float32(5))),
						DefaultFrom: *policyService.NewNullableDefaultFromRule(convutils.Ref(policyService.DefaultFromRule{
							Field:  *policyService.NewNullableString(convutils.Ref("compute.cpuCoreLimit")),
							Factor: *policyService.NewNullableFloat64(convutils.Ref(0.3)),
						})),
					})),
				})),
				NodeAffinityRequired: *policyService.NewNullableArbitraryRules(convutils.Ref(policyService.ArbitraryRules{
					CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
				})),
			}),
			Worker: *policyService.NewNullableDistributedPolicyRulesV2Worker(&policyService.DistributedPolicyRulesV2Worker{
				Compute: *policyService.NewNullableSupersetRulesAllOfCompute(convutils.Ref(policyService.SupersetRulesAllOfCompute{
					GpuDevicesRequest: *policyService.NewNullableIntegerRules(convutils.Ref(policyService.IntegerRules{
						CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
					})),
					CpuCoreRequest: *policyService.NewNullableNumberRules(convutils.Ref(policyService.NumberRules{
						Max: *policyService.NewNullableFloat32(convutils.Ref(float32(10))),
						DefaultFrom: *policyService.NewNullableDefaultFromRule(convutils.Ref(policyService.DefaultFromRule{
							Field:  *policyService.NewNullableString(convutils.Ref("compute.cpuCoreLimit")),
							Factor: *policyService.NewNullableFloat64(convutils.Ref(0.25)),
						})),
					})),
				})),
				NodeAffinityRequired: *policyService.NewNullableArbitraryRules(convutils.Ref(policyService.ArbitraryRules{
					CanEdit: *policyService.NewNullableBool(convutils.Ref(true)),
				})),
			}),
		}))

		policyMock.EffectiveDistributedPolicyV2.Defaults.Set(convutils.Ref(policyService.DistributedPolicyDefaultsV2{
			Master: *policyService.NewNullableReplicaDefaultsV2(&policyService.ReplicaDefaultsV2{
				Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(convutils.Ref(policyService.SupersetDefaultsAllOfCompute{
					GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(2))),
				})),
				NodeAffinityRequired: *policyService.NewNullableNodeAffinityRequired(defaultNodeAffinity),
			}),
			Worker: *policyService.NewNullableDistributedPolicyDefaultsV2Worker(&policyService.DistributedPolicyDefaultsV2Worker{
				Compute: *policyService.NewNullableSupersetDefaultsAllOfCompute(convutils.Ref(policyService.SupersetDefaultsAllOfCompute{
					GpuDevicesRequest: *policyService.NewNullableInt32(convutils.Ref(int32(4))),
				})),
				NodeAffinityRequired: *policyService.NewNullableNodeAffinityRequired(defaultNodeAffinity),
			}),
		}))

		workerRules, _ := gojsonq.LoadJson(policyMock.EffectiveDistributedPolicyV2.Rules.Get().Worker)
		masterRules, _ := gojsonq.LoadJson(policyMock.EffectiveDistributedPolicyV2.Rules.Get().Master)
		policyMock.EffectiveDistributedPolicyV2.DefaultFromOrder.Set(calculateOrder(workerRules, masterRules))

		policies.SetMock(policyMock)

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedDepartmentsToContext(&ctx)

		authorization.SetMock(authMock)
		orgunit.SetMock(authMock.OrgUnitMock())
		mockTenantSettingsClient.On("GetHybridSubmissionEnabled", mock.Anything).Return(false, nil)

		clusterClientMock = clusterClient.NewSimpleMockClient().WithCluster(randoms.ClusterId, clusterApi.DisplayedCluster{
			Uuid:    randoms.ClusterId,
			Version: *clusterApi.NewNullableString(convutils.Ref(DefaultClusterVersionForTesting)),
		})
		clusterClient.SetMock(clusterClientMock)

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, nil)))
		checker.InitClustersCache(defaultCheckerClusterConfig)

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))

		router = createRouterInt(mockTenantSettingsClient)
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("apply training policy in create", func() {
		It("should be sensitive to workload type", func() {
			requestTrainingOkWorkspaceBad := api.TrainingCreationRequest{
				Name:      "tr-good",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(10)),
					},
				},
			}

			requestWorkspaceOkTrainingBad := api.TrainingCreationRequest{
				Name:      "tr-bad",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
					},
				},
			}

			createResponse, _ := CreateTraining(ctx, mockTenantSettingsClient, requestTrainingOkWorkspaceBad)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createResponse, _ = CreateTraining(ctx, mockTenantSettingsClient, requestWorkspaceOkTrainingBad)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))

			policyErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(policyErr.ComplianceIssues).NotTo(BeNil())
			Expect(len(policyErr.ComplianceIssues)).To(Equal(1))
			Expect(policyErr.ComplianceIssues[0].Field).To(ContainSubstring("compute.gpuDevicesRequest"))
			Expect(policyErr.ComplianceIssues[0].Details).To(ContainSubstring("must be no less than 5"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[0].Rule.Get(), "")).To(Equal(policy_compliance.PolicyMin))
		})
	})

	Context("apply workspace policy in create", func() {
		It("should reject workload for error", func() {
			requestTrainingOkWorkspaceBad := api.WorkspaceCreationRequest{
				Name:      "ws-bad",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(10)),
						CpuCoreRequest:    convutils.Ref(float64(40)),
					},
				},
			}

			requestWorkspaceOkTrainingBad := api.WorkspaceCreationRequest{
				Name:      "ws-policy",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(5)),
					},
				},
			}

			createResponse, _ := CreateWorkspace(ctx, mockTenantSettingsClient, requestTrainingOkWorkspaceBad)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			policyErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(policyErr.ComplianceIssues)).To(Equal(2))

			// prevent rest flakiness due to undetermined sort order of the fields
			cpuIndex := 0
			gpuIndex := 1
			if policyErr.ComplianceIssues[0].Field == "compute.gpuDevicesRequest" {
				cpuIndex = 1
				gpuIndex = 0
			}

			Expect(policyErr.ComplianceIssues[cpuIndex].Field, "").To(Equal("compute.cpuCoreRequest"))
			Expect(policyErr.ComplianceIssues[cpuIndex].Details).To(ContainSubstring("must be no greater than 20"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[cpuIndex].Rule.Get(), "")).To(Equal(policy_compliance.PolicyMax))

			Expect(policyErr.ComplianceIssues[gpuIndex].Field, "").To(Equal("compute.gpuDevicesRequest"))
			Expect(policyErr.ComplianceIssues[gpuIndex].Details).To(ContainSubstring("must be no greater than 8"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[gpuIndex].Rule.Get(), "")).To(Equal(policy_compliance.PolicyMax))

			createResponse, _ = CreateWorkspace(ctx, mockTenantSettingsClient, requestWorkspaceOkTrainingBad)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))
		})

		It("should contain compliance info", func() {
			// Use unique cluster to avoid interference from other tests
			testRandoms := randoms.CloneWithDifferentCluster()

			templateGood := api.WorkspaceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "good",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(testRandoms.ClusterId),
				},
				Spec: api.WorkspaceTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(3)),
					},
				},
			}

			templateBad := api.WorkspaceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "bad",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(testRandoms.ClusterId),
				},
				Spec: api.WorkspaceTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(10)),
					},
				},
			}

			createTemplateForTest(router, fmt.Sprintf("%s/workspaces", urlTemplates), &templateGood, testRandoms)
			createTemplateForTest(router, fmt.Sprintf("%s/workspaces", urlTemplates), &templateBad, testRandoms)

			result := listTemplatesForTestType[api.WorkspaceTemplatesListResponse](router, testRandoms,
				fmt.Sprintf("%s/workspaces", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", testRandoms.ProjectId))
			Expect(result.Entries).To(HaveLen(2))

			for _, entry := range result.Entries {
				if entry.Meta.Name == "good" {
					Expect(entry.ComplianceIssues).To(BeNil())
				} else {
					Expect(entry.ComplianceIssues).To(HaveLen(1))
					Expect(entry.ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
						Field:   convutils.Ref("compute.gpuDevicesRequest"),
						Rule:    convutils.Ref(api.POLICYRULEENUM_MAX),
						Details: "must be no greater than 8",
					}))

				}
			}
		})

		It("should reject workload for error - arbitrary rule", func() {
			requestWorkspaceBad := api.WorkspaceCreationRequest{
				Name:      "ws-policy",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					NodeAffinityRequired: &api.NodeAffinityRequired{
						NodeSelectorTerms: []api.NodeSelectorTerm{
							{
								MatchExpressions: []api.MatchExpression{
									{
										Key:      "key",
										Operator: api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST,
									},
								},
							},
						},
					},
				},
			}

			createResponse, _ := CreateWorkspace(ctx, mockTenantSettingsClient, requestWorkspaceBad)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			policyErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(policyErr.ComplianceIssues)).To(Equal(1))

			Expect(policyErr.ComplianceIssues[0].Field, "").To(Equal("nodeAffinityRequired"))
			Expect(policyErr.ComplianceIssues[0].Details).To(ContainSubstring("the administrator prohibited changing the value of this field"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[0].Rule.Get(), "")).To(Equal(policy_compliance.PolicyCanEdit))
		})
	})

	Context("training with compliance issues", func() {
		It("list should contain the errors", func() {
			templateGood := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "good",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(8)),
					},
				},
			}

			templateBad := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "bad",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(2)),
					},
				},
			}

			createTemplateForTest(router, fmt.Sprintf("%s/trainings", urlTemplates), &templateGood, randoms)
			createTemplateForTest(router, fmt.Sprintf("%s/trainings", urlTemplates), &templateBad, randoms)

			result := listTemplatesForTestType[api.TrainingTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/trainings", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(2))

			for _, entry := range result.Entries {
				if entry.Meta.Name == "good" {
					Expect(entry.ComplianceIssues).To(BeNil())
				} else {
					Expect(entry.ComplianceIssues).To(HaveLen(1))
					Expect(entry.ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
						Field:   convutils.Ref("compute.gpuDevicesRequest"),
						Rule:    convutils.Ref(api.POLICYRULEENUM_MIN),
						Details: "must be no less than 5",
					}))

				}
			}
		})
	})

	Context("training with compliance caching", func() {
		It("compliance info should be cached", func() {
			policyMockUpdateTime := time.Now()

			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: policyMockUpdateTime,
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
							Completions: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(10))),
							}),
						}),
					},
				}}

			policies.SetMock(policyMock)

			templateBad := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "bad",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.TrainingTemplateSpec{
					Image:       convutils.Ref("ubuntu"),
					Completions: convutils.Ref(int32(20)),
				},
			}

			createTemplateForTest(router, fmt.Sprintf("%s/trainings", urlTemplates), &templateBad, randoms)

			result := listTemplatesForTestType[api.TrainingTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/trainings", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(1))
			Expect(result.Entries[0].ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
				Field:   convutils.Ref("completions"),
				Rule:    convutils.Ref(api.POLICYRULEENUM_MAX),
				Details: "must be no greater than 10",
			}))

			// now we change the rules of the policy, but we do not change the date
			// so the mechanism should load the compliance from the DB - same error
			policyMock = policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: policyMockUpdateTime,
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
							Completions: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(11))),
							}),
						}),
					},
				}}

			policies.SetMock(policyMock)

			result = listTemplatesForTestType[api.TrainingTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/trainings", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(1))
			Expect(result.Entries[0].ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
				Field:   convutils.Ref("completions"),
				Rule:    convutils.Ref(api.POLICYRULEENUM_MAX),
				Details: "must be no greater than 10", // no change!!! same rule
			}))

			// now we change the date of the policy, which impact removal of the cached
			// policy and calculate again based om the new rule
			policyMockUpdateTime = time.Now()

			policyMock = policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: policyMockUpdateTime,
					EffectiveTrainingPolicyV2: &policyService.TrainingPolicyDefaultsAndRulesV2{
						Rules: *policyService.NewNullableTrainingPolicyRulesV2(&policyService.TrainingPolicyRulesV2{
							Completions: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
								Max: *policyService.NewNullableInt32(convutils.Ref(int32(12))),
							}),
						}),
					},
				}}

			policies.SetMock(policyMock)

			result = listTemplatesForTestType[api.TrainingTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/trainings", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(1))
			Expect(result.Entries[0].ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
				Field:   convutils.Ref("completions"),
				Rule:    convutils.Ref(api.POLICYRULEENUM_MAX),
				Details: "must be no greater than 12", // the new rule applies
			}))
		})
	})

	Context("distributed with compliance issues", func() {
		It("list should contain the errors", func() {
			templateGood := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "good",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.DistributedTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(4)),
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(2)),
						},
					},
				},
			}

			templateBad := api.DistributedTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "bad",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.DistributedTemplateSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(8)),
					},
				},
				Master: &api.TemplateMasterSection{
					Spec: &api.MasterTemplateSpec{
						Image: convutils.Ref("ubuntu"),
						Compute: &api.WorkspaceSpecSpecAllOfCompute{
							GpuDevicesRequest: convutils.Ref(int32(8)),
						},
					},
				},
			}

			createTemplateForTest(router, fmt.Sprintf("%s/distributed", urlTemplates), &templateGood, randoms)
			createTemplateForTest(router, fmt.Sprintf("%s/distributed", urlTemplates), &templateBad, randoms)

			result := listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/distributed", urlTemplates),
				fmt.Sprintf("&complyToProject=%d&complyToReplica=worker", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(2))

			for _, entry := range result.Entries {
				if entry.Meta.Name == "good" {
					Expect(entry.ComplianceIssues).To(BeNil())
				} else {
					Expect(entry.ComplianceIssues).To(HaveLen(1))
					Expect(entry.ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
						Field:   convutils.Ref("compute.gpuDevicesRequest"),
						Rule:    convutils.Ref(api.POLICYRULEENUM_CAN_EDIT),
						Details: "the administrator prohibited changing the value of this field",
					}))
				}
			}

			result = listTemplatesForTestType[api.DistributedTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/distributed", urlTemplates),
				fmt.Sprintf("&complyToProject=%d&complyToReplica=master", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(2))

			for _, entry := range result.Entries {
				if entry.Meta.Name == "good" {
					Expect(entry.ComplianceIssues).To(BeNil())
				} else {
					Expect(entry.ComplianceIssues).To(HaveLen(1))
					Expect(entry.ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
						Field:   convutils.Ref("compute.gpuDevicesRequest"),
						Rule:    convutils.Ref(api.POLICYRULEENUM_CAN_EDIT),
						Details: "the administrator prohibited changing the value of this field",
					}))
				}
			}
		})
	})

	Context("apply distributed policy in create", func() {
		It("should reject the workload for any error", func() {
			///////////////////////////////////////////////////////
			//   default for gpuDevicesRequest worker 4, master 2
			///////////////////////////////////////////////////////

			requestDistributedBadMaster := api.DistributedCreationRequest{
				Name:      "ws-change-default-master",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image:                convutils.Ref("ubuntu"),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(4)),
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("redhat"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(4)),
					},
				},
			}

			createResponse, _ := CreateDistributed(ctx, mockTenantSettingsClient, requestDistributedBadMaster)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			policyErr := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(policyErr.ComplianceIssues)).To(Equal(1))
			Expect(policyErr.ComplianceIssues[0].Field, "").To(Equal("masterSpec.compute.gpuDevicesRequest"))
			Expect(policyErr.ComplianceIssues[0].Details).To(ContainSubstring("the administrator prohibited changing the value of this field"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[0].Rule.Get(), "")).To(Equal(policy_compliance.PolicyCanEdit))

			requestDistributedBadWorker := api.DistributedCreationRequest{
				Name:      "ws-change-default-worker",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image:                convutils.Ref("ubuntu"),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(2)),
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("redhat"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						GpuDevicesRequest: convutils.Ref(int32(2)),
					},
				},
			}

			createResponse, _ = CreateDistributed(ctx, mockTenantSettingsClient, requestDistributedBadWorker)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			policyErr = createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(policyErr.ComplianceIssues)).To(Equal(1))
			Expect(policyErr.ComplianceIssues[0].Field, "").To(Equal("compute.gpuDevicesRequest"))
			Expect(policyErr.ComplianceIssues[0].Details).To(ContainSubstring("the administrator prohibited changing the value of this field"))
			Expect(convutils.Deref(policyErr.ComplianceIssues[0].Rule.Get(), "")).To(Equal(policy_compliance.PolicyCanEdit))
		})
	})

	Context("apply distributed policy in create", func() {
		It("should mutate the workload properly", func() {
			requestDistributed := api.DistributedCreationRequest{
				Name:      "ws-mutate-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image:                convutils.Ref("ubuntu"),
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_PY_TORCH),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreLimit: convutils.Ref(1.1),
					},
					NodeAffinityRequired: &api.NodeAffinityRequired{
						NodeSelectorTerms: []api.NodeSelectorTerm{
							{
								MatchExpressions: []api.MatchExpression{
									{
										Key:      "workerKey",
										Operator: api.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST,
									},
								},
							},
						},
					},
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("redhat"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreLimit: convutils.Ref(1.0),
					},
				},
			}

			createResponse, _ := CreateDistributed(ctx, mockTenantSettingsClient, requestDistributed)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			response := createResponse.Body.(workloadsApi.SupersetWorkload)

			Expect(*response.Spec.Compute.Get().CpuCoreLimit.Get()).To(Equal(1.1))
			Expect(*response.Spec.Compute.Get().CpuCoreRequest.Get()).To(Equal(0.275))

			Expect(*response.MasterSpec.Get().Compute.Get().CpuCoreLimit.Get()).To(Equal(1.0))
			Expect(*response.MasterSpec.Get().Compute.Get().CpuCoreRequest.Get()).To(Equal(0.3))

			// from spec
			workerAffinity := response.Spec.NodeAffinityRequired.Get()
			Expect(workerAffinity).NotTo(BeNil())
			Expect(workerAffinity.NodeSelectorTerms).To(HaveLen(1))
			Expect(workerAffinity.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
			Expect(workerAffinity.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("workerKey"))
			Expect(workerAffinity.NodeSelectorTerms[0].MatchExpressions[0].Operator).To(Equal(workloadsApi.MATCHEXPRESSIONOPERATOR_DOES_NOT_EXIST))

			// from policy
			masterAffinity := response.MasterSpec.Get().NodeAffinityRequired.Get()
			Expect(masterAffinity).NotTo(BeNil())
			Expect(masterAffinity.NodeSelectorTerms).To(HaveLen(1))
			Expect(masterAffinity.NodeSelectorTerms[0].MatchExpressions).To(HaveLen(1))
			Expect(masterAffinity.NodeSelectorTerms[0].MatchExpressions[0].Key).To(Equal("key"))
			Expect(masterAffinity.NodeSelectorTerms[0].MatchExpressions[0].Operator).To(Equal(workloadsApi.MATCHEXPRESSIONOPERATOR_EXISTS))
		})
	})

	Context("apply workspace policy in create", func() {
		It("should mutate the workload properly and calculated field should return policy conflict", func() {
			requestWorkspace := api.WorkspaceCreationRequest{
				Name:      "ws-mutate-2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						CpuCoreLimit: convutils.Ref(8.0),
					},
				},
			}

			createResponse, _ := CreateWorkspace(ctx, mockTenantSettingsClient, requestWorkspace)

			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			workload := createResponse.Body.(workloadsApi.SupersetWorkload)
			Expect(convutils.Deref(workload.Spec.Compute.Get().CpuCoreLimit.Get(), 0.0)).To(Equal(8.0))
			Expect(convutils.Deref(workload.Spec.Compute.Get().CpuCoreRequest.Get(), 0.0)).To(Equal(1.6))
		})
	})

	Context("distributed inference worker section handling", func() {
		var (
			distributedInferencesService *DistributedInferenceService

			policyDefaults policyService.DistributedInferenceDefaultsV2
			policyRules    policyService.DistributedInferenceRulesV2
		)

		BeforeEach(func() {
			serviceInterface := NewDistributedInferenceService(mockTenantSettingsClient)
			var ok bool
			distributedInferencesService, ok = serviceInterface.(*DistributedInferenceService)
			Expect(ok).To(BeTrue(), "Failed to cast service to correct type")

			// Set up mocks
			policyMockUpdateTime := time.Now()

			policyDefaults = policyService.DistributedInferenceDefaultsV2{
				ServingPort: *policyService.NewNullableDistributedInferenceServingPort(
					&policyService.DistributedInferenceServingPort{
						Port: *policyService.NewNullableInt32(convutils.Ref(int32(8080))),
					},
				),
				Worker: *policyService.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(
					&policyService.DistributedInferenceLeaderWorkerDefaultsV2{
						Image: *policyService.NewNullableString(convutils.Ref("ubuntu")),
					},
				),
			}

			policyRules = policyService.DistributedInferenceRulesV2{
				ServingPort: *policyService.NewNullableDistributedInferenceServingPortRules(
					&policyService.DistributedInferenceServingPortRules{
						Port: *policyService.NewNullableIntegerRules(&policyService.IntegerRules{
							Step: *policyService.NewNullableInt32(convutils.Ref(int32(3))),
						}),
					}),
				Worker: *policyService.NewNullableDistributedInferenceLeaderWorkerRulesV2(
					&policyService.DistributedInferenceLeaderWorkerRulesV2{
						Image: *policyService.NewNullableStringRules(&policyService.StringRules{
							CanEdit: *policyService.NewNullableBool(convutils.Ref(false)),
						}),
					}),
			}

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: policyMockUpdateTime,
					EffectiveDistributedInferencePolicyV2: &policyService.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedInferenceDefaultsV2(&policyDefaults),
					},
				},
			})
		})

		DescribeTable("resolves replica defaultFrom rules during distributed inference creation",
			func(requestField string) {
				policyRules = policyService.DistributedInferenceRulesV2{}
				Expect(json.Unmarshal([]byte(`{
					"leader":{"compute":{"cpuCoreRequest":{"defaultFrom":{"field":"compute.gpuPortionRequest","factor":8}}}},
					"worker":{"compute":{"cpuCoreRequest":{"defaultFrom":{"field":"compute.gpuPortionRequest","factor":8}}}}
				}`), &policyRules)).To(Succeed())
				policies.SetMock(policies.SimpleMockClientFactory{
					SimpleMockClient: policies.SimpleMockClient{
						EffectiveUpdatedAt: time.Now(),
						EffectiveDistributedInferencePolicyV2: &policyService.DistributedInferencePolicyDefaultsAndRulesV2{
							Defaults: *policyService.NewNullableDistributedInferenceDefaultsV2(&policyDefaults),
							Rules:    *policyService.NewNullableDistributedInferenceRulesV2(&policyRules),
						},
					},
				})
				var spec api.DistributedInferenceSpecSpec
				Expect(json.Unmarshal([]byte(fmt.Sprintf(`{
					"workers":1,
					"leader":{"image":"ubuntu","compute":{"gpuRequestType":"portion",%q:0.25}},
					"worker":{"image":"ubuntu","compute":{"gpuRequestType":"portion",%q:0.5}}
				}`, requestField, requestField)), &spec)).To(Succeed())
				response, err := distributedInferencesService.CreateDistributedInference(ctx, api.DistributedInferenceCreationRequest{
					Name: "test-dinf-default-from", ProjectId: randoms.ProjectIdStr, ClusterId: randoms.ClusterId, Spec: spec,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Code).To(Equal(http.StatusAccepted), "%+v", response.Body)
				workload := response.Body.(workloadsApi.SupersetWorkload)
				Expect(workload.Spec.Leader.Get().Compute.Get().GetCpuCoreRequest()).To(Equal(2.0))
				Expect(workload.Spec.Worker.Get().Compute.Get().GetCpuCoreRequest()).To(Equal(4.0))
			},
			Entry("legacy GPU portion", "gpuPortionRequest"),
			Entry("new GPU memory portion", "gpuMemoryPortionRequest"),
		)

		Context("distributed inference worker section handling", func() {
			It("should not set worker spec when workers field is not sent", func() {
				request := api.DistributedInferenceCreationRequest{
					Name:      "test-dinf-1",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.DistributedInferenceSpecSpec{
						Leader: &api.DistributedInferenceLeaderWorkerSpec{
							Image: convutils.Ref("leader-image:v1"),
						},
					},
				}

				resp, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
				Expect(resp.Code).To(Equal(http.StatusAccepted))
				workload := resp.Body.(workloadsApi.SupersetWorkload)

				getResp, _ := distributedInferencesService.GetDistributedInference(ctx, workload.WorkloadId)
				getPayload := getResp.Body.(workloadsApi.SupersetWorkload)
				Expect(getPayload.Spec.Worker.IsSet()).To(BeFalse())
			})

			It("should not set worker spec when workers is set to 0", func() {
				request := api.DistributedInferenceCreationRequest{
					Name:      "test-dinf-2",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.DistributedInferenceSpecSpec{
						Leader: &api.DistributedInferenceLeaderWorkerSpec{
							Image: convutils.Ref("leader-image:v1"),
						},
						Workers: convutils.Ref(int32(0)),
					},
				}

				resp, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
				Expect(resp.Code).To(Equal(http.StatusAccepted))
				workload := resp.Body.(workloadsApi.SupersetWorkload)

				getResp, _ := distributedInferencesService.GetDistributedInference(ctx, workload.WorkloadId)
				getPayload := getResp.Body.(workloadsApi.SupersetWorkload)
				Expect(getPayload.Spec.Worker.IsSet()).To(BeFalse())
			})

			It("should set worker spec with policy defaults when workers is positive", func() {
				request := api.DistributedInferenceCreationRequest{
					Name:      "test-dinf-3",
					ProjectId: randoms.ProjectIdStr,
					ClusterId: randoms.ClusterId,
					Spec: api.DistributedInferenceSpecSpec{
						Leader: &api.DistributedInferenceLeaderWorkerSpec{
							Image: convutils.Ref("leader-image:v1"),
						},
						Workers: convutils.Ref(int32(2)),
					},
				}

				resp, _ := distributedInferencesService.CreateDistributedInference(ctx, request)
				Expect(resp.Code).To(Equal(http.StatusAccepted))
				workload := resp.Body.(workloadsApi.SupersetWorkload)

				getResp, _ := distributedInferencesService.GetDistributedInference(ctx, workload.WorkloadId)
				getPayload := getResp.Body.(workloadsApi.SupersetWorkload)
				Expect(getPayload.Spec.Worker).NotTo(BeNil())
				Expect(*getPayload.Spec.Worker.Get().Image.Get()).To(Equal("ubuntu"))
			})
		})

		It("list should contain the errors", func() {
			policyMockUpdateTime := time.Now()

			policies.SetMock(policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveUpdatedAt: policyMockUpdateTime,
					EffectiveDistributedInferencePolicyV2: &policyService.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyService.NewNullableDistributedInferenceDefaultsV2(&policyDefaults),
						Rules:    *policyService.NewNullableDistributedInferenceRulesV2(&policyRules),
					},
				},
			})

			templateGood := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "good",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("ubuntu"),
					},
					Workers: convutils.Ref(int32(6)),
				},
			}

			templateBad := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "bad",
					ScopeType: api.SCOPETYPE_CLUSTER,
					ScopeId:   convutils.Ref(randoms.ClusterId),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Worker: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("worker-image:v1"),
					},
					Workers: convutils.Ref(int32(7)),
				},
			}

			createTemplateForTest(router, fmt.Sprintf("%s/distributed-inferences", urlTemplates), &templateGood, randoms)
			createTemplateForTest(router, fmt.Sprintf("%s/distributed-inferences", urlTemplates), &templateBad, randoms)

			result := listTemplatesForTestType[api.DistributedInferenceTemplatesListResponse](router, randoms,
				fmt.Sprintf("%s/distributed-inferences", urlTemplates),
				fmt.Sprintf("&complyToProject=%d", randoms.ProjectId))
			Expect(result.Entries).To(HaveLen(2))

			for _, entry := range result.Entries {
				if entry.Meta.Name == "good" {
					Expect(entry.ComplianceIssues).To(BeNil())
				} else {
					Expect(entry.ComplianceIssues).To(HaveLen(1))
					Expect(entry.ComplianceIssues[0]).To(Equal(api.ComplianceInfoReason{
						Field:   convutils.Ref("worker.image"),
						Rule:    convutils.Ref(api.POLICYRULEENUM_CAN_EDIT),
						Details: "the administrator prohibited changing the value of this field",
					}))
				}
			}
		})
	})
})

func calculateOrder(workerRules map[string]interface{}, masterRules map[string]interface{}) *policyService.DefaultFromOrder {
	orderWorker, circleErr := policy_compliance.NewDefaultsFromOrder(workerRules)
	Expect(circleErr).To(BeNil())

	orderMaster, circleErr := policy_compliance.NewDefaultsFromOrder(masterRules)
	Expect(circleErr).To(BeNil())

	return &policyService.DefaultFromOrder{
		Worker: convutils.Ref(map[string][]string(orderWorker)),
		Master: convutils.Ref(map[string][]string(orderMaster)),
	}
}
