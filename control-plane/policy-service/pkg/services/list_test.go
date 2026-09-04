/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListTest", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var authMock *PoliciesAuthClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		tree.Restart()

		backendMock = NewBackendMockClient(2, 1)
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))

		authMock = NewPoliciesAuthClient(backendMock)
		authorization.SetMock(authMock)

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Behavior
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("non numeric project id", func() {
		It("should return with error", func() {
			requestGet, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=abcde", urlPolicy), nil)
			Expect(err).To(BeNil())
			requestGet.Header = *headerForAdmin(2)

			responseGet := httptest.NewRecorder()
			router.ServeHTTP(responseGet, requestGet)
			Expect(responseGet.Code).To(Equal(http.StatusBadRequest))
			Expect(responseGet.Body.String()).To(ContainSubstring("project id 'abcde' is not numeric"))
		})
	})

	Context("list policies", func() {
		It("includes policies of all types and does not show the rules/defaults", func() {
			workspaceBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-w",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2)),
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
					},
				},
			}

			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
				},
			}

			inferenceBody := api.InferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-i",
				},
				Policy: &api.InferencePolicyDefaultsAndRulesV2{
					Rules: &api.InferencePolicyRulesV2{
						Security: &api.InferencePolicyRulesV2AllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
				},
			}

			distributedBody := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-d",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Worker: &api.DistributedPolicyRulesV2Worker{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
						},
						Master: &api.ReplicaRulesV2{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(2)),
								},
							},
						},
					},
				},
			}

			fallbackBody := api.FallbackPolicyChangeRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-f",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Rules: &api.SupersetRules{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2)),
							},
						},
					},
					Defaults: &api.SupersetDefaults{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
					},
				},
			}

			//
			//   Phase 1 -> add them
			//
			updateWorkspacePolicyForTest(router, &workspaceBody, 2)
			updateTrainingPolicyForTest(router, &trainingBody, 2)
			updateInferencePolicyForTest(router, &inferenceBody, 2)
			updateDistributedPolicyForTest(router, &distributedBody, 2)
			updateFallbackPolicyForTest(router, &fallbackBody, 2)

			//
			//   Phase 2 get them back (should not be visible in list)
			//
			actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s", urlPolicy))
			Expect(len(actualResponse.Policies)).To(Equal(4))

			workspaceEntry := -1
			trainingEntry := -1
			inferenceEntry := -1
			distributedEntry := -1

			for index, entry := range actualResponse.Policies {
				if entry.Type == api.POLICYTYPE_WORKSPACE {
					workspaceEntry = index
					Expect(entry.Meta.Name).To(Equal(workspaceBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_TRAINING {
					trainingEntry = index
					Expect(entry.Meta.Name).To(Equal(trainingBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_DISTRIBUTED {
					distributedEntry = index
					Expect(entry.Meta.Name).To(Equal(distributedBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_INFERENCE {
					inferenceEntry = index
					Expect(entry.Meta.Name).To(Equal(inferenceBody.Meta.Name))
				}

				Expect(entry.Meta.Scope).To(Equal(workspaceBody.Meta.Scope))
			}

			Expect(workspaceEntry).NotTo(Equal(-1), "expecting one workspace policy")
			Expect(trainingEntry).NotTo(Equal(-1), "expecting one training policy")
			Expect(distributedEntry).NotTo(Equal(-1), "expecting one distributed policy")
			Expect(inferenceEntry).NotTo(Equal(-1), "expecting one inference policy")

			//
			//   Phase 3 get them back with fallback included
			//
			actualResponse = listPoliciesForTest(router, 2, fmt.Sprintf("%s?includeFallbackPolicies=true", urlPolicy))
			Expect(len(actualResponse.Policies)).To(Equal(5))

			workspaceEntry = -1
			trainingEntry = -1
			inferenceEntry = -1
			distributedEntry = -1
			fallbackEntry := -1

			for index, entry := range actualResponse.Policies {
				if entry.Type == api.POLICYTYPE_WORKSPACE {
					workspaceEntry = index
					Expect(entry.Meta.Name).To(Equal(workspaceBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_TRAINING {
					trainingEntry = index
					Expect(entry.Meta.Name).To(Equal(trainingBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_DISTRIBUTED {
					distributedEntry = index
					Expect(entry.Meta.Name).To(Equal(distributedBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_INFERENCE {
					inferenceEntry = index
					Expect(entry.Meta.Name).To(Equal(inferenceBody.Meta.Name))
				}
				if entry.Type == api.POLICYTYPE_FALLBACK {
					fallbackEntry = index
					Expect(entry.Meta.Name).To(Equal(fallbackBody.Meta.Name))
				}

				Expect(entry.Meta.Scope).To(Equal(workspaceBody.Meta.Scope))
			}

			Expect(workspaceEntry).NotTo(Equal(-1), "expecting one workspace policy")
			Expect(trainingEntry).NotTo(Equal(-1), "expecting one training policy")
			Expect(distributedEntry).NotTo(Equal(-1), "expecting one distributed policy")
			Expect(inferenceEntry).NotTo(Equal(-1), "expecting one inference policy")
			Expect(fallbackEntry).NotTo(Equal(-1), "expecting one fallback policy")
		})

		Context("gvk_workload filter on list", func() {
			It("excludes policies with gvk_workload=true from the normal list", func() {
				rules := api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(1)),
						},
					},
				}
				rulesAsJson, err := gojsonq.LoadJson(rules)
				Expect(err).To(BeNil())

				nativeID := uuid.New()
				native := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        nativeID,
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:         "native-list-gvk-filter",
					WorkloadType: string(tree.WorkloadTypeWorkspace),
					Scope:        api.SCOPETYPE_TENANT,
					TenantId:     2,
					RulesV2:      rulesAsJson,
					GvkWorkload:  false,
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &native)).To(BeNil())

				trainingRules := api.TrainingPolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(2)),
						},
					},
				}
				trainingRulesJson, errTr := gojsonq.LoadJson(trainingRules)
				Expect(errTr).To(BeNil())

				gvkOnly := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        uuid.New(),
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:         "gvk-only-hidden-from-list",
					WorkloadType: "apps/Deployment/v1",
					Scope:        api.SCOPETYPE_TENANT,
					TenantId:     2,
					RulesV2:      trainingRulesJson,
					GvkWorkload:  true,
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &gvkOnly)).To(BeNil())

				query := "?workloadType=Workspace&scope=tenant"
				actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s%s", urlPolicy, query))
				Expect(actualResponse.Policies).To(HaveLen(1))
				Expect(actualResponse.Policies[0].Meta.Name).To(Equal(native.Name))
			})
		})

		Context("availability of gvk workloads in list", func() {
			It("populates availability with phase and issues for gvk policies and leaves it nil for native ones", func() {
				rules := api.WorkspacePolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(1)),
						},
					},
				}
				rulesAsJson, err := gojsonq.LoadJson(rules)
				Expect(err).To(BeNil())

				native := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        uuid.New(),
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:         "native-availability",
					WorkloadType: string(tree.WorkloadTypeWorkspace),
					Scope:        api.SCOPETYPE_TENANT,
					TenantId:     2,
					RulesV2:      rulesAsJson,
					GvkWorkload:  false,
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &native)).To(BeNil())

				degraded := api.POLICYPHASE_DEGRADED
				gvkDegraded := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        uuid.New(),
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:             "gvk-degraded",
					WorkloadType:     "apps/Deployment/v1",
					Scope:            api.SCOPETYPE_TENANT,
					TenantId:         2,
					RulesV2:          rulesAsJson,
					GvkWorkload:      true,
					ValidationStatus: &degraded,
					ValidationIssues: []string{"specSelector portByName is not defined"},
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &gvkDegraded)).To(BeNil())

				ready := api.POLICYPHASE_READY
				gvkReady := dbmodel.Policy{
					BaseModel: dbmodels.BaseModel{
						ID:        uuid.New(),
						CreatedBy: "test",
						UpdatedBy: "test",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Name:             "gvk-ready",
					WorkloadType:     "apps/StatefulSet/v1",
					Scope:            api.SCOPETYPE_TENANT,
					TenantId:         2,
					RulesV2:          rulesAsJson,
					GvkWorkload:      true,
					ValidationStatus: &ready,
					ValidationIssues: []string{},
				}
				Expect(dao.AddPolicy(context.Background(), runaidb.DB(), &gvkReady)).To(BeNil())

				actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s?scope=tenant&includeGvkPolicies=true", urlPolicy))
				Expect(actualResponse.Policies).To(HaveLen(3))

				byName := map[string]api.PolicyListEntry{}
				for _, p := range actualResponse.Policies {
					byName[p.Meta.Name] = p
				}

				Expect(byName).To(HaveKey(native.Name))
				Expect(byName[native.Name].Availability).To(BeNil())

				Expect(byName).To(HaveKey(gvkDegraded.Name))
				Expect(byName[gvkDegraded.Name].Availability).NotTo(BeNil())
				Expect(byName[gvkDegraded.Name].Availability.Phase).To(Equal(api.POLICYPHASE_DEGRADED))
				Expect(byName[gvkDegraded.Name].Availability.Issues).To(Equal([]string{"specSelector portByName is not defined"}))

				Expect(byName).To(HaveKey(gvkReady.Name))
				Expect(byName[gvkReady.Name].Availability).NotTo(BeNil())
				Expect(byName[gvkReady.Name].Availability.Phase).To(Equal(api.POLICYPHASE_READY))
				Expect(byName[gvkReady.Name].Availability.Issues).To(BeEmpty())
			})
		})

		Context("list policies following deletion", func() {
			It("should get pending deletion status", func() {
				tenantId := int32(2)

				workspaceBody := api.WorkspacePolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope: api.SCOPETYPE_TENANT,
						Name:  "pol-w",
					},
					Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
						Rules: &api.WorkspacePolicyRulesV2{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(2)),
								},
							},
						},
						Defaults: &api.WorkspacePolicyDefaultsV2{
							Security: &api.SupersetDefaultsAllOfSecurity{
								RunAsUid: convutils.Ref(int64(5)),
							},
						},
					},
				}

				//
				//   Phase 1 -> add the policy
				//
				updateWorkspacePolicyForTest(router, &workspaceBody, tenantId)

				//
				//   delete the workspace policy
				//
				deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

				//
				//   list is empty (because cluster never heard of this policy so can be deleted permanently)
				//
				actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s", urlPolicy))
				Expect(len(actualResponse.Policies)).To(Equal(0))
			})
		})

		Context("list policies scopes", func() {
			It("should support all scopes", func() {
				workspaceBody := api.WorkspacePolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:        api.SCOPETYPE_DEPARTMENT,
						Name:         "pol-d1",
						DepartmentId: convutils.Ref("111"),
					},
					Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
						Defaults: &api.WorkspacePolicyDefaultsV2{
							Image: convutils.Ref("ubuntu"),
						},
					},
				}

				trainingBody := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_PROJECT,
						Name:      "pol-p1",
						ProjectId: convutils.Ref(int32(1111)),
					},
					Policy: &api.TrainingPolicyDefaultsAndRulesV2{
						Defaults: &api.TrainingPolicyDefaultsV2{
							Image: convutils.Ref("ubuntu"),
						},
					},
				}

				//
				//   Phase 1 -> add them both
				//
				updateWorkspacePolicyForTest(router, &workspaceBody, 2)
				updateTrainingPolicyForTest(router, &trainingBody, 2)

				//
				//   Phase 2 get them back - we expect correct scope in the list
				//
				actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s", urlPolicy))
				Expect(len(actualResponse.Policies)).To(Equal(2))

				for _, entry := range actualResponse.Policies {
					if entry.Type == api.POLICYTYPE_WORKSPACE {
						Expect(entry.Meta.Name).To(Equal(workspaceBody.Meta.Name))
						Expect(entry.Meta.Scope).To(Equal(workspaceBody.Meta.Scope))
						Expect(entry.Meta.DepartmentId).To(Equal(workspaceBody.Meta.DepartmentId))
					}
					if entry.Type == api.POLICYTYPE_TRAINING {
						Expect(entry.Meta.Name).To(Equal(trainingBody.Meta.Name))
						Expect(entry.Meta.Scope).To(Equal(trainingBody.Meta.Scope))
						Expect(entry.Meta.ProjectId).To(Equal(trainingBody.Meta.ProjectId))
					}

				}
			})
		})

		Context("list policies query", func() {
			It("should return the correct policies in accordance with the query", func() {
				// creating 3 policies (1 for the department and 2 under the same project)

				policyProjectId := int32(1111)

				departmentTrainingPolicy := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:        api.SCOPETYPE_DEPARTMENT,
						Name:         "policy-department-training",
						DepartmentId: convutils.Ref("111"),
					},
					Policy: &api.TrainingPolicyDefaultsAndRulesV2{
						Defaults: &api.TrainingPolicyDefaultsV2{
							Image: convutils.Ref("ubuntu"),
						},
					},
				}

				trainingBody := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_PROJECT,
						Name:      "policy-project-training",
						ProjectId: convutils.Ref(policyProjectId),
					},
					Policy: &api.TrainingPolicyDefaultsAndRulesV2{
						Defaults: &api.TrainingPolicyDefaultsV2{
							Image: convutils.Ref("ubuntu"),
						},
					},
				}

				workspaceBody := api.WorkspacePolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_PROJECT,
						Name:      "policy-project-workspace",
						ProjectId: convutils.Ref(policyProjectId),
					},
					Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
						Defaults: &api.WorkspacePolicyDefaultsV2{
							Image: convutils.Ref("ubuntu"),
						},
					},
				}

				//
				//   Phase 1 -> add all 3
				//
				updateTrainingPolicyForTest(router, &departmentTrainingPolicy, 2)
				updateWorkspacePolicyForTest(router, &workspaceBody, 2)
				updateTrainingPolicyForTest(router, &trainingBody, 2)

				//
				//   Phase 2 send a list request with query, and expect only one to match
				//
				query := fmt.Sprintf("?workloadType=Training&scope=project&projectId=%d", policyProjectId)

				actualResponse := listPoliciesForTest(router, 2, fmt.Sprintf("%s%s", urlPolicy, query))
				Expect(len(actualResponse.Policies)).To(Equal(1))
				Expect(actualResponse.Policies[0].Meta.Name).To(Equal(trainingBody.Meta.Name))
				Expect(actualResponse.Policies[0].Meta.Scope).To(Equal(trainingBody.Meta.Scope))
				Expect(actualResponse.Policies[0].Meta.ProjectId).To(Equal(trainingBody.Meta.ProjectId))
			})
		})

		Context("list policies permitted scopes", func() {
			It("should consider the READ permissions of the user", func() {
				policyBody := api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				}

				policyProjectId := int32(1111)

				tenantPolicy := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope: api.SCOPETYPE_TENANT,
						Name:  "tenant",
					},
					Policy: &policyBody,
				}
				clusterPolicy := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_CLUSTER,
						Name:      "cluster",
						ClusterId: convutils.Ref(backendMock.ClusterIdPerTenant["1"][0]),
					},
					Policy: &policyBody,
				}
				deptPolicy := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:        api.SCOPETYPE_DEPARTMENT,
						Name:         "dept",
						DepartmentId: convutils.Ref("111"),
					},
					Policy: &policyBody,
				}

				projPolicy := api.TrainingPolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_PROJECT,
						Name:      "policy-project-training",
						ProjectId: convutils.Ref(policyProjectId),
					},
					Policy: &policyBody,
				}

				updateTrainingPolicyForTest(router, &tenantPolicy, 1)
				updateTrainingPolicyForTest(router, &clusterPolicy, 1)
				updateTrainingPolicyForTest(router, &deptPolicy, 1)
				updateTrainingPolicyForTest(router, &projPolicy, 1)

				//
				// permitted scopes all tenant (default for our mock) all policies returned
				//
				actualResponse := listPoliciesForTest(router, 1, urlPolicy)
				Expect(len(actualResponse.Policies)).To(Equal(4))

				//
				// only one project
				//
				authMock.SetPermittedScopes(openapi.ACTION_READ, &openapi.PermittedScopes{
					Projects: []string{"1111"},
				})

				actualResponse = listPoliciesForTest(router, 1, urlPolicy)
				Expect(len(actualResponse.Policies)).To(Equal(1))
				Expect(actualResponse.Policies[0].Meta.Scope).To(Equal(api.SCOPETYPE_PROJECT))

				//
				// the entire department
				//
				authMock.SetPermittedScopes(openapi.ACTION_READ, &openapi.PermittedScopes{
					Departments: []string{"111"},
				})

				actualResponse = listPoliciesForTest(router, 1, urlPolicy)
				Expect(len(actualResponse.Policies)).To(Equal(2))

				types := datastructs.NewSet[api.ScopeType]()
				types.Add(actualResponse.Policies[0].Meta.Scope)
				types.Add(actualResponse.Policies[1].Meta.Scope)
				Expect(types.Has(api.SCOPETYPE_PROJECT)).To(Equal(true))
				Expect(types.Has(api.SCOPETYPE_DEPARTMENT)).To(Equal(true))
			})
		})

	})
})
