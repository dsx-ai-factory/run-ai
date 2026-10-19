/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type UrlAndPayload struct {
	Url       string
	GetParams string
	Payload   interface{}
	Response  *httptest.ResponseRecorder
}

var _ = Describe("GetTest", func() {
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
	// Validation
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("missing project id", func() {
		It("should return with error", func() {
			requestGet, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			requestGet.Header = *headerForAdmin(2)

			responseGet := httptest.NewRecorder()
			router.ServeHTTP(responseGet, requestGet)
			Expect(responseGet.Code).To(Equal(http.StatusBadRequest))
			Expect(responseGet.Body.String()).To(ContainSubstring("mandatory projectId is not provided"))
		})
	})

	Context("non numeric project", func() {
		DescribeTable("get with non numeric project", func(url string) {
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s&projectId=abcde", url, api.SCOPETYPE_PROJECT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("project id 'abcde' is not numeric"))
		}, forWorkspaceUrl, forTrainingUrl, forDistributedUrl, forInferenceUrl)
	})

	Context("non numeric project id", func() {
		It("should return with error", func() {
			requestGet, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=abcde", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			requestGet.Header = *headerForAdmin(2)

			responseGet := httptest.NewRecorder()
			router.ServeHTTP(responseGet, requestGet)
			Expect(responseGet.Code).To(Equal(http.StatusBadRequest))
			Expect(responseGet.Body.String()).To(ContainSubstring("project id 'abcde' is not numeric"))
		})
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Behavior
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("system scope", func() {
		It("should return with error", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=%s", urlPolicyWorkspace, api.SCOPETYPE_SYSTEM), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("system scope is not supported for policies"))
		})
	})

	Context("get from recovered tree", func() {
		It("should return the correct policy", func() {
			//
			//   Part 1: Add to DB, tree is empty at this point
			//
			rules := api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						Min: convutils.Ref(int32(1)),
					},
				},
			}
			rulesAsJson, err := gojsonq.LoadJson(rules)
			Expect(err).To(BeNil())

			policy := dbmodel.Policy{
				BaseModel: dbmodels.BaseModel{
					ID:        uuid.New(),
					CreatedBy: "test",
					UpdatedBy: "test",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				Name:         "pol-t",
				WorkloadType: string(tree.WorkloadTypeWorkspace),
				Scope:        api.SCOPETYPE_TENANT,
				TenantId:     1,
				RulesV2:      rulesAsJson,
			}
			dbErr := dao.AddPolicy(context.Background(), runaidb.DB(), &policy)
			Expect(dbErr).To(BeNil())

			//
			//  Part 2: Get from the tree - expected effective to be calculated properly
			//
			responseGet := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))
			Expect(responseGet.Meta.Name).To(Equal("empty-policy-department-111"))
			Expect(string(responseGet.Meta.Scope)).To(Equal("department"))
			Expect(responseGet.Meta.ClusterId).NotTo(BeNil()) // because it's a department, need to have cluster id inside
			Expect(convutils.Deref(responseGet.Meta.DepartmentId, "")).To(Equal("111"))

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			expectedEffective.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						SourceOfRule: &api.SourceOfRule{Scope: "tenant"},
						Min:          convutils.Ref(int32(1)),
					},
				},
			}

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*responseGet.Effective.Rules).ToEqual(expectedEffective.Rules)
		})
	})

	Context("noCreationMode", func() {
		It("should return true only if tenant created policy", func() {
			urlPolicy := fmt.Sprintf("%s?scope=project&projectId=1111", urlPolicyWorkspace)
			noCreateStr := "&noCreationMode=true"

			expectingNotToFindPolicy(router, 1, urlPolicy+noCreateStr)

			// we want to make sure that the first get did not create any policy
			expectingNotToFindPolicy(router, 1, urlPolicy+noCreateStr)

			// get without noCreationMode will create the policy
			getWorkspacePolicyForTest(router, 1, urlPolicy)
		})
	})

	Context("get empty and then apply policy", func() {
		It("given + effective returned correctly", func() {
			//
			//  Part 1: Tenant with two env vars and rules on gpu portion
			//
			bodySetTenant := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("TENANT-ENV-A"),
									Value: convutils.Ref("tenant-value-a"),
								},
								{
									Name:  convutils.Ref("TENANT-ENV-B"),
									Value: convutils.Ref("tenant-value-b"),
								},
							},
						},
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							GpuPortionRequest: &api.NumberRules{
								Max: convutils.Ref(float32(10)),
							},
						},
					},
				},
			}

			responseSetTenant := updateWorkspacePolicyForTest(router, &bodySetTenant, 1)

			var actualRespSetTenant api.WorkspacePolicyV2
			err := json.Unmarshal(responseSetTenant.Body.Bytes(), &actualRespSetTenant)
			Expect(err).To(BeNil())

			expectedRespSetTenant := api.WorkspacePolicyDefaultsAndRulesV2{
				Rules:    bodySetTenant.Policy.Rules,
				Defaults: bodySetTenant.Policy.Defaults,
			}
			expectedRespSetTenant.Rules.Compute.GpuPortionRequest.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedRespSetTenant.Rules.Compute.GpuMemoryPortionRequest = expectedRespSetTenant.Rules.Compute.GpuPortionRequest
			dao.InjectSystemPolicyForTesting(&expectedRespSetTenant)
			Expect(*actualRespSetTenant.Effective).To(Equal(expectedRespSetTenant))

			//
			//   Part 2 - GET the department policy. Given policy => nil, Effective policy => same as tenant
			//
			actualRespDept := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))
			Expect(*actualRespDept.Policy).To(Equal(api.WorkspacePolicyDefaultsAndRulesV2{}))
			Expect(*actualRespDept.Effective).To(Equal(*actualRespSetTenant.Effective))

			//
			//   Part 3 - GET the tenant policy to make sure it has not changed
			//
			actualRespGetTenant := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))
			Expect(actualRespGetTenant.Policy).To(Equal(actualRespSetTenant.Policy))

			expectedRespGetTenant := *actualRespSetTenant.Effective
			expectedRespGetTenant.Rules.Compute.GpuPortionRequest.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedRespGetTenant.Rules.Compute.GpuMemoryPortionRequest = expectedRespGetTenant.Rules.Compute.GpuPortionRequest
			dao.InjectSystemPolicyForTesting(&expectedRespGetTenant)
			Expect(*actualRespGetTenant.Effective).To(Equal(expectedRespGetTenant))
		})
	})

	Context("get empty policy", func() {
		It("created with invented name", func() {
			actualRespDeptA := getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyWorkspace))

			Expect(actualRespDeptA.Meta.Name).To(Equal("empty-policy-department-111"))
			Expect(*actualRespDeptA.Policy).To(Equal(api.WorkspacePolicyDefaultsAndRulesV2{}))
			testutils.Expect(*actualRespDeptA.Effective).ToEqual(dao.SystemPolicy.Policy)
		})
	})

	Context("rename policy", func() {
		It("name of the policy changes", func() {
			//
			//   Phase 1 -> create a policy
			//
			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					Name:         "initial-name",
					DepartmentId: convutils.Ref("111"),
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						NodeType: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "training"},
							},
						},
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						// the code generator minimized the files, so it created only one set for workspace,
						// and it uses them for both... oh well
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
						NodeType: convutils.Ref("training"),
					},
				},
			}

			updateTrainingPolicyForTest(router, &trainingBody, 1)

			//
			//  Part 2: rename
			//
			bodySetDept := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "new-name",
				},
			}

			responseSetDept := updateTrainingPolicyForTest(router, &bodySetDept, 1)

			var actualRespSetDept api.TrainingPolicyV2
			err := json.Unmarshal(responseSetDept.Body.Bytes(), &actualRespSetDept)
			Expect(err).To(BeNil())
			Expect(actualRespSetDept.Meta.Name).To(Equal("new-name"))

			//
			//   Part 3 - GET now return the new name
			//
			actualRespDeptB := getTrainingPolicyForTest(router, 1, fmt.Sprintf("%s?scope=department&departmentId=111", urlPolicyTraining))
			Expect(actualRespDeptB.Meta.Name).To(Equal("new-name"))
			Expect(*actualRespDeptB.Policy).To(Equal(*trainingBody.Policy))
		})
	})

	Context("policies for different workload types are distinct", func() {
		It("get each one returns the correct values", func() {
			workspaceBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-w1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(2)),
							},
						},
						NodeType: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "workspace"},
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
						NodeType: convutils.Ref("workspace"),
					},
				},
			}

			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
						NodeType: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "training"},
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						// the code generator minimized the files, so it created only one set for workspace,
						// and it uses them for both... oh well
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
						NodeType: convutils.Ref("training"),
					},
				},
			}

			distributedBody := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-d1",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Worker: &api.DistributedPolicyRulesV2Worker{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
							NodeType: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "dist-worker"},
								},
							},
						},
						Master: &api.ReplicaRulesV2{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
							NodeType: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "dist-master"},
								},
							},
						},
					},
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							Security: &api.SupersetDefaultsAllOfSecurity{
								RunAsUid: convutils.Ref(int64(6)),
							},
						},
						Master: &api.ReplicaDefaultsV2{
							Security: &api.SupersetDefaultsAllOfSecurity{
								RunAsUid: convutils.Ref(int64(7)),
							},
						},
					},
				},
			}

			inferenceBody := api.InferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-i1",
				},
				Policy: &api.InferencePolicyDefaultsAndRulesV2{
					Rules: &api.InferencePolicyRulesV2{
						Security: &api.InferencePolicyRulesV2AllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
						NodeType: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "inference"},
							},
						},
					},
					Defaults: &api.InferencePolicyDefaultsV2{
						// the code generator minimized the files, so it created only one set for workspace,
						// and it uses them for both... oh well
						Security: &api.InferencePolicyDefaultsV2AllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
					},
				},
			}

			distributedInferenceBody := api.DistributedInferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-i1",
				},
				Policy: &api.DistributedInferencePolicyDefaultsAndRulesV2{
					Rules: &api.DistributedInferenceRulesV2{
						Worker: &api.DistributedInferenceLeaderWorkerRulesV2{
							Security: &api.DistributedInferenceLeaderWorkerRulesV2Security{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
							NodeType: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "dist-inf-worker"},
								},
							},
						},
						Leader: &api.DistributedInferenceLeaderWorkerRulesV2{
							Security: &api.DistributedInferenceLeaderWorkerRulesV2Security{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
							NodeType: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "inference"},
								},
							},
						},
					},
					Defaults: &api.DistributedInferenceDefaultsV2{
						Worker: &api.DistributedInferenceLeaderWorkerDefaultsV2{
							Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{
								RunAsUid: convutils.Ref(int64(6)),
							},
						},
						Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{
							Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{
								RunAsUid: convutils.Ref(int64(5)),
							},
						},
					},
				},
			}

			//
			//   Phase 1 -> add them all
			//
			updateWorkspacePolicyForTest(router, &workspaceBody, 2)
			updateTrainingPolicyForTest(router, &trainingBody, 2)
			updateDistributedPolicyForTest(router, &distributedBody, 2)
			updateInferencePolicyForTest(router, &inferenceBody, 2)
			updateDistributedInferencePolicyForTest(router, &distributedInferenceBody, 2)

			//
			//   Phase 2 get them back
			//
			actualRespGetWs := getWorkspacePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))
			actualRespGetTrain := getTrainingPolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyTraining))
			actualRespGetDist := getDistributedPolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyDistributed))
			actualRespGetInference := getInferencePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyInference))
			actualRespGetDistInference := getDistributedInferencePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyDistributedInference))

			//
			//  check the values
			//
			expectedGetWs := api.WorkspacePolicyDefaultsAndRulesV2{
				Rules:    workspaceBody.Policy.Rules,
				Defaults: workspaceBody.Policy.Defaults,
			}
			expectedGetWs.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedGetWs.Rules.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForTesting(&expectedGetWs)

			Expect(*actualRespGetWs.Effective).To(Equal(expectedGetWs))

			expectedTrainEffective := api.TrainingPolicyDefaultsAndRulesV2{
				Rules:    trainingBody.Policy.Rules,
				Defaults: trainingBody.Policy.Defaults,
			}
			expectedTrainEffective.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedTrainEffective.Rules.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForTesting(&expectedTrainEffective)

			Expect(*actualRespGetTrain.Effective).To(Equal(expectedTrainEffective))

			expectedDistEffective := api.DistributedPolicyDefaultsAndRulesV2{
				Rules:    distributedBody.Policy.Rules,
				Defaults: distributedBody.Policy.Defaults,
			}
			expectedDistEffective.Rules.Worker.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistEffective.Rules.Worker.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistEffective.Rules.Master.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistEffective.Rules.Master.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForDistributedTesting(&expectedDistEffective)

			testutils.Expect(*actualRespGetDist.Effective).ToEqual(expectedDistEffective)

			expectedInferenceEffective := api.InferencePolicyDefaultsAndRulesV2{
				Rules:    inferenceBody.Policy.Rules,
				Defaults: inferenceBody.Policy.Defaults,
			}
			expectedInferenceEffective.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedInferenceEffective.Rules.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForTesting(&expectedInferenceEffective)

			Expect(*actualRespGetInference.Effective).To(Equal(expectedInferenceEffective))

			expectedDistributedInferenceEffective := api.DistributedInferencePolicyDefaultsAndRulesV2{
				Rules:    distributedInferenceBody.Policy.Rules,
				Defaults: distributedInferenceBody.Policy.Defaults,
			}
			expectedDistributedInferenceEffective.Rules.Worker.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistributedInferenceEffective.Rules.Worker.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistributedInferenceEffective.Rules.Leader.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDistributedInferenceEffective.Rules.Leader.NodeType.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForDistributedInferenceTesting(&expectedDistributedInferenceEffective)

			testutils.Expect(*actualRespGetDistInference.Effective).ToEqual(expectedDistributedInferenceEffective)
		})
	})

	Context("project-scoped user reading fallback policy", func() {
		BeforeEach(func() {
			authMock.SetPermittedScopes(authClient.ACTION_READ, &authClient.PermittedScopes{
				Projects: []string{"1111"},
			})
		})

		It("should allow reading tenant-scoped fallback policy", func() {
			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s", urlPolicyFallback, api.SCOPETYPE_TENANT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should allow reading department-scoped fallback policy in same tenant", func() {
			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s&departmentId=111", urlPolicyFallback, api.SCOPETYPE_DEPARTMENT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should allow reading project-scoped fallback policy in same tenant", func() {
			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s&projectId=1111", urlPolicyFallback, api.SCOPETYPE_PROJECT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("should allow reading cluster-scoped fallback policy in same tenant", func() {
			clusterOfTenant1 := backendMock.ClusterIdPerTenant[tenantIdStr(1)][0]

			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s&clusterId=%s", urlPolicyFallback, api.SCOPETYPE_CLUSTER, clusterOfTenant1), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})

	Context("cross-tenant cluster access for fallback policy", func() {
		It("should return 403 when tenant B requests fallback for tenant A's cluster", func() {
			tenantA := int32(1)
			tenantB := int32(2)
			clusterOfTenantA := backendMock.ClusterIdPerTenant[tenantIdStr(tenantA)][0]

			authMock.SetPermittedScopes(authClient.ACTION_READ, &authClient.PermittedScopes{
				Tenants: []string{tenantIdStr(tenantB)},
			})

			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s&clusterId=%s", urlPolicyFallback, api.SCOPETYPE_CLUSTER, clusterOfTenantA), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantB)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	Context("cross-tenant tenant-scope access for fallback policy", func() {
		It("should not leak tenant A's fallback policy to tenant B", func() {
			tenantA := int32(1)
			tenantB := int32(2)

			tenantAPolicy := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "fallback-tenant-a",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Args: convutils.Ref("tenant-a-secret"),
					},
				},
			}
			applyPolicyForTest(router, urlPolicyFallback, makePayload(tenantAPolicy), tenantA, http.MethodPut)

			authMock.SetPermittedScopes(authClient.ACTION_READ, &authClient.PermittedScopes{
				Projects: []string{"2111"},
			})

			request, err := http.NewRequest("GET",
				fmt.Sprintf("%s?scope=%s", urlPolicyFallback, api.SCOPETYPE_TENANT), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantB)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			var fallback api.FallbackPolicy
			err = json.Unmarshal(response.Body.Bytes(), &fallback)
			Expect(err).To(BeNil())

			if fallback.Effective != nil && fallback.Effective.Defaults != nil && fallback.Effective.Defaults.Args != nil {
				Expect(*fallback.Effective.Defaults.Args).NotTo(Equal("tenant-a-secret"))
			}
		})
	})

	Context("multiple level workspace + training", func() {
		It("get each one returns the correct values", func() {
			workspaceDept := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-int-dept",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							HostPath: &api.HostPathsDefaults{
								Instances: &[]api.HostPathInstance{
									{
										Path:      convutils.Ref("workspace"),
										MountPath: convutils.Ref("department"),
									},
								},
							},
						},
					},
				},
			}

			workspaceProject := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1112)),
					Name:      "pol-int-proj",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							HostPath: &api.HostPathsDefaults{
								Instances: &[]api.HostPathInstance{
									{
										Path:      convutils.Ref("workspace"),
										MountPath: convutils.Ref("project"),
									},
								},
							},
						},
					},
				},
			}

			trainingDept := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-train-dept",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							HostPath: &api.HostPathsDefaults{
								Instances: &[]api.HostPathInstance{
									{
										Path:      convutils.Ref("training"),
										MountPath: convutils.Ref("department"),
									},
								},
							},
						},
					},
				},
			}

			trainingProject := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1112)),
					Name:      "pol-train-proj",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							HostPath: &api.HostPathsDefaults{
								Instances: &[]api.HostPathInstance{
									{
										Path:      convutils.Ref("training"),
										MountPath: convutils.Ref("project"),
									},
								},
							},
						},
					},
				},
			}

			reqID := UrlAndPayload{Url: urlPolicyWorkspace, Payload: workspaceDept, GetParams: "scope=department&departmentId=111"}
			reqIP := UrlAndPayload{Url: urlPolicyWorkspace, Payload: workspaceProject, GetParams: "scope=project&projectId=1112"}
			reqTD := UrlAndPayload{Url: urlPolicyTraining, Payload: trainingDept, GetParams: "scope=department&departmentId=111"}
			reqTP := UrlAndPayload{Url: urlPolicyTraining, Payload: trainingProject, GetParams: "scope=project&projectId=1112"}

			for _, req := range []UrlAndPayload{reqID, reqIP, reqTD, reqTP} {
				applyPolicyForTest(router, req.Url, makePayload(req.Payload), 2, http.MethodPatch)
			}

			//
			//  now get each one and see that each of them is ok
			//
			for _, req := range []*UrlAndPayload{&reqID, &reqIP, &reqTD, &reqTP} {
				request, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", req.Url, req.GetParams), makePayload(req.Payload))
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(2)

				req.Response = httptest.NewRecorder()
				router.ServeHTTP(req.Response, request)
				Expect(req.Response.Code).To(Equal(http.StatusOK))
			}

			var actualRespID api.WorkspacePolicyV2
			var actualRespIP api.WorkspacePolicyV2
			var actualRespTD api.TrainingPolicyV2
			var actualRespTP api.TrainingPolicyV2

			err := json.Unmarshal(reqID.Response.Body.Bytes(), &actualRespID)
			Expect(err).To(BeNil())
			err = json.Unmarshal(reqIP.Response.Body.Bytes(), &actualRespIP)
			Expect(err).To(BeNil())
			err = json.Unmarshal(reqTD.Response.Body.Bytes(), &actualRespTD)
			Expect(err).To(BeNil())
			err = json.Unmarshal(reqTP.Response.Body.Bytes(), &actualRespTP)
			Expect(err).To(BeNil())

			Expect(actualRespID.Policy.Defaults).To(Equal(workspaceDept.Policy.Defaults))
			Expect(actualRespIP.Policy.Defaults).To(Equal(workspaceProject.Policy.Defaults))
			Expect(actualRespTD.Policy.Defaults).To(Equal(trainingDept.Policy.Defaults))
			Expect(actualRespTP.Policy.Defaults).To(Equal(trainingProject.Policy.Defaults))

			//
			//   now delete all of them
			//
			for _, req := range []UrlAndPayload{reqID, reqIP, reqTD, reqTP} {
				request, err := http.NewRequest("DELETE", fmt.Sprintf("%s?%s", req.Url, req.GetParams), makePayload(req.Payload))
				Expect(err).To(BeNil())
				request.Header = *headerForAdmin(2)

				req.Response = httptest.NewRecorder()
				router.ServeHTTP(req.Response, request)
				Expect(req.Response.Code).To(Equal(http.StatusNoContent))
			}
		})
	})
})
