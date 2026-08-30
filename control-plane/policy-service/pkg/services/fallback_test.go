package services

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/policy-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/assets"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/run-ai/runai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/policy-service/pkg/tree"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/authorization"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
)

var _ = Describe("ApplyTest", func() {
	router := CreateRouter()

	var backendMock *BackendMockClient
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		database.DeleteAllTablesForTesting()

		tree.Restart()

		backendMock = NewBackendMockClient(2, 1)
		authorization.SetMock(NewPoliciesAuthClient(backendMock))
		orgUnitService.SetMock(NewPoliciesOrgUnitClient(backendMock))
		orgunit.SetMock(NewPoliciesOrgUnitClient(backendMock))

		clusterMock = cluster.NewSimpleMockClient().WithClusters(backendMock.ClusterIdPerTenant)
		cluster.SetMock(clusterMock)

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})
	})

	Context("fallback policy created in tenant level", func() {
		It("if it affecting the entire matrix", func() {
			tenantId := int32(2)

			//
			//   apply a policy to fallback/tenant
			//
			givenPolicy := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "fallback-tenant",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Args: convutils.Ref("a=1"),
					},
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(givenPolicy), tenantId, http.MethodPut)

			fallbackPolicy := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=tenant", urlPolicyFallback))
			Expect(*fallbackPolicy.Effective.Defaults.Args).To(Equal(*givenPolicy.Policy.Defaults.Args))
			Expect(*fallbackPolicy.Effective.Rules.Args.SourceOfRule).To(Equal(api.SourceOfRule{
				Scope:      api.SCOPETYPE_TENANT,
				IsFallback: true,
			}))

			for _, scope := range []string{"tenant", "department&departmentId=211", "project&projectId=2111"} {

				trainingPolicy := getTrainingPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyTraining, scope))
				Expect(trainingPolicy.Policy.Defaults).To(BeNil())
				Expect(trainingPolicy.Policy.Rules).To(BeNil())
				testutils.Expect(trainingPolicy.Effective.Defaults).ToEqual(fallbackPolicy.Effective.Defaults)
				testutils.Expect(trainingPolicy.Effective.Rules).ToEqual(fallbackPolicy.Effective.Rules)

				workspacePolicy := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyWorkspace, scope))
				Expect(workspacePolicy.Policy.Defaults).To(BeNil())
				Expect(workspacePolicy.Policy.Rules).To(BeNil())
				testutils.Expect(workspacePolicy.Effective.Defaults).ToEqual(fallbackPolicy.Effective.Defaults)
				testutils.Expect(workspacePolicy.Effective.Rules).ToEqual(fallbackPolicy.Effective.Rules)

				distributedPolicy := getDistributedPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyDistributed, scope))
				Expect(distributedPolicy.Policy.Defaults).To(BeNil())
				Expect(distributedPolicy.Policy.Rules).To(BeNil())
				testutils.Expect(*distributedPolicy.Effective.Defaults.Master).ToEqual(fallbackPolicy.Effective.Defaults)
				testutils.Expect(*distributedPolicy.Effective.Rules.Master).ToEqual(fallbackPolicy.Effective.Rules)
				testutils.Expect(*distributedPolicy.Effective.Defaults.Worker).ToEqual(fallbackPolicy.Effective.Defaults)
				testutils.Expect(*distributedPolicy.Effective.Rules.Worker).ToEqual(fallbackPolicy.Effective.Rules)

				var inferenceExpected api.InferencePolicyDefaultsAndRulesV2
				copyErr := convutils.CopySpecs(&inferenceExpected, fallbackPolicy.Effective)
				Expect(copyErr).To(BeNil())

				inferencePolicy := getInferencePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyInference, scope))
				Expect(inferencePolicy.Policy.Defaults).To(BeNil())
				Expect(inferencePolicy.Policy.Rules).To(BeNil())
				testutils.Expect(inferencePolicy.Effective).ToEqual(inferenceExpected)
			}
		})
	})

	Context("fallback policy created in project level", func() {
		It("affecting only project policies", func() {
			tenantId := int32(2)
			projectId := int32(2111)

			//
			//   apply a policy to fallback/project
			//
			givenPolicy := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: &projectId,
					Name:      "fallback-proj",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Args: convutils.Ref("a=1"),
					},
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(givenPolicy), tenantId, http.MethodPut)

			fallbackPolicyTenant := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=tenant", urlPolicyFallback))
			testutils.Expect(fallbackPolicyTenant.Effective).ToEqual(dao.SystemPolicy.Policy)

			fallbackPolicyDept := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))
			testutils.Expect(fallbackPolicyDept.Effective).ToEqual(dao.SystemPolicy.Policy)

			fallbackPolicyProject := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyFallback))
			Expect(*fallbackPolicyProject.Effective.Defaults.Args).To(Equal(*givenPolicy.Policy.Defaults.Args))
			Expect(*fallbackPolicyProject.Effective.Rules.Args.SourceOfRule).To(Equal(api.SourceOfRule{
				Scope:      api.SCOPETYPE_PROJECT,
				ProjectId:  &projectId,
				IsFallback: true,
			}))

			fallbackPolicy := []api.FallbackPolicy{
				fallbackPolicyTenant,
				fallbackPolicyDept,
				fallbackPolicyProject,
			}

			for idx, scope := range []string{"tenant", "department&departmentId=211", "project&projectId=2111"} {
				trainPolicy := getTrainingPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyTraining, scope))
				Expect(trainPolicy.Policy.Defaults).To(BeNil())
				Expect(trainPolicy.Policy.Rules).To(BeNil())
				testutils.Expect(trainPolicy.Effective.Defaults).ToEqual(fallbackPolicy[idx].Effective.Defaults)
				testutils.Expect(trainPolicy.Effective.Rules).ToEqual(fallbackPolicy[idx].Effective.Rules)

			}
		})
	})

	Context("fallback policy removed", func() {
		It("it is clearing also the affected policies", func() {
			tenantId := int32(2)
			deptId := "211"

			//
			//   apply a policy to fallback/department
			//
			givenPolicy := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "fallback-dept",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Args: convutils.Ref("a=1"),
					},
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(givenPolicy), tenantId, http.MethodPut)

			fallbackPolicyDept := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))
			Expect(*fallbackPolicyDept.Effective.Defaults.Args).To(Equal(*givenPolicy.Policy.Defaults.Args))
			Expect(*fallbackPolicyDept.Effective.Rules.Args.SourceOfRule).To(Equal(api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
				IsFallback:   true,
			}))

			// project and department scope has the fallback stuff
			for _, scope := range []string{"department&departmentId=211", "project&projectId=2111"} {
				trainingPolicy := getTrainingPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyTraining, scope))
				Expect(trainingPolicy.Policy.Defaults).To(BeNil())
				Expect(trainingPolicy.Policy.Rules).To(BeNil())
				testutils.Expect(trainingPolicy.Effective.Defaults).ToEqual(fallbackPolicyDept.Effective.Defaults)
				testutils.Expect(trainingPolicy.Effective.Rules).ToEqual(fallbackPolicyDept.Effective.Rules)
			}

			// delete the fallback policy
			deletePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))

			// project and department scope has the fallback stuff
			for _, scope := range []string{"department&departmentId=211", "project&projectId=2111"} {
				trainingPolicy := getTrainingPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=%s", urlPolicyTraining, scope))
				Expect(trainingPolicy.Policy.Defaults).To(BeNil())
				Expect(trainingPolicy.Policy.Rules).To(BeNil())
				testutils.Expect(trainingPolicy.Effective).ToEqual(dao.SystemPolicy.Policy)
			}
		})
	})

	Context("security combo", func() {
		It("should be calculated correctly", func() {
			tenantId := int32(2)
			deptId := "211"

			fallbackDeptGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "fallback-dept",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Image: convutils.Ref("fallback-dept-image"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							AllowPrivilegeEscalation: convutils.Ref(false),
						},
					},
					Rules: &api.SupersetRules{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
						},
						Security: &api.SupersetRulesAllOfSecurity{
							AllowPrivilegeEscalation: &api.BooleanRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},
				},
			}

			workspaceDeptGiven := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "workspace-dept",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("workspace-dept-image"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							AllowPrivilegeEscalation: convutils.Ref(true),
						},
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
						},
						Security: &api.SupersetRulesAllOfSecurity{
							AllowPrivilegeEscalation: &api.BooleanRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackDeptGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(workspaceDeptGiven), tenantId, http.MethodPut)

			sourceFallbackDept := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
				IsFallback:   true,
			}
			sourceWorkspaceDept := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
			}

			fallbackPolicyDept := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))

			Expect(*fallbackPolicyDept.Effective.Defaults.Image).To(Equal("fallback-dept-image"))
			Expect(*fallbackPolicyDept.Effective.Defaults.Security.AllowPrivilegeEscalation).To(Equal(false))
			Expect(*fallbackPolicyDept.Effective.Rules.Image.SourceOfRule).To(Equal(sourceFallbackDept))
			Expect(*fallbackPolicyDept.Effective.Rules.Security.AllowPrivilegeEscalation.SourceOfRule).To(Equal(sourceFallbackDept))

			//
			//   since we do not have fallback policy for project, the fallback policy of the project will be identical
			//   to the fallback policy of the department
			//
			fallbackPolicyProject := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyFallback))

			Expect(*fallbackPolicyProject.Effective.Defaults.Image).To(Equal("fallback-dept-image"))
			Expect(*fallbackPolicyProject.Effective.Defaults.Security.AllowPrivilegeEscalation).To(Equal(false))
			Expect(*fallbackPolicyProject.Effective.Rules.Image.SourceOfRule).To(Equal(sourceFallbackDept))
			Expect(*fallbackPolicyProject.Effective.Rules.Security.AllowPrivilegeEscalation.SourceOfRule).To(Equal(sourceFallbackDept))

			//
			//    since we have workspace policy for the department, it overrides everything from the fallback policy
			//
			workspacePolicyDept := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyWorkspace))

			Expect(*workspacePolicyDept.Effective.Defaults.Image).To(Equal("workspace-dept-image"))
			Expect(*workspacePolicyDept.Effective.Defaults.Security.AllowPrivilegeEscalation).To(Equal(true))
			Expect(*workspacePolicyDept.Effective.Rules.Image.SourceOfRule).To(Equal(sourceWorkspaceDept))
			Expect(*workspacePolicyDept.Effective.Rules.Security.AllowPrivilegeEscalation.SourceOfRule).To(Equal(sourceWorkspaceDept))

			workspacePolicyProj := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyWorkspace))

			Expect(*workspacePolicyProj.Effective.Defaults.Image).To(Equal("workspace-dept-image"))
			Expect(*workspacePolicyProj.Effective.Defaults.Security.AllowPrivilegeEscalation).To(Equal(true))
			Expect(*workspacePolicyProj.Effective.Rules.Image.SourceOfRule).To(Equal(sourceWorkspaceDept))
			Expect(*workspacePolicyProj.Effective.Rules.Security.AllowPrivilegeEscalation.SourceOfRule).To(Equal(sourceWorkspaceDept))
		})
	})

	Context("security combo two levels", func() {
		It("should be calculated correctly", func() {
			tenantId := int32(2)
			deptId := "211"

			fallbackTenantGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "fallback-tenant",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Image: convutils.Ref("fallback-tenant"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							AllowPrivilegeEscalation: convutils.Ref(false),
						},
					},
					Rules: &api.SupersetRules{
						Security: &api.SupersetRulesAllOfSecurity{
							AllowPrivilegeEscalation: &api.BooleanRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},
				},
			}

			workspaceDeptGiven := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "workspace-dept",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("workspace-dept-image"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							AllowPrivilegeEscalation: convutils.Ref(true),
						},
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
						},
					},
				},
			}

			sourceFallbackTenant := api.SourceOfRule{
				Scope:      api.SCOPETYPE_TENANT,
				IsFallback: true,
			}
			sourceWorkspaceDept := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackTenantGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(workspaceDeptGiven), tenantId, http.MethodPut)

			workspacePolicyProj := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyWorkspace))

			Expect(*workspacePolicyProj.Effective.Defaults.Image).To(Equal("workspace-dept-image"))
			Expect(*workspacePolicyProj.Effective.Defaults.Security.AllowPrivilegeEscalation).To(Equal(true))
			Expect(*workspacePolicyProj.Effective.Rules.Image.SourceOfRule).To(Equal(sourceWorkspaceDept))
			Expect(*workspacePolicyProj.Effective.Rules.Security.AllowPrivilegeEscalation.SourceOfRule).To(Equal(sourceFallbackTenant))
		})
	})

	Context("values combo", func() {
		It("should be calculated correctly", func() {
			tenantId := int32(2)
			deptId := "211"
			projectId := int32(2111)

			fallbackDeptGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "fallback-dept",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Image: convutils.Ref("fallback/dept/can-be-overridden"),
					},
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "fallback/dept/can-be-overridden"},
							},
						},
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "fallback/dept/can-be-overridden"},
										},
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false), // cannot-be-overridden (instance rules)
								},
							},
						},
					},
				},
			}

			fallbackProjectGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: &projectId,
					Name:      "fallback-proj",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Image: convutils.Ref("fallback/proj/can-be-overridden"),
					},
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "fallback/proj/can-be-overridden"},
							},
						},
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "fallback/proj/can-be-overridden"},
										},
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false), // cannot-be-overridden (instance rules)
								},
							},
						},
					},
				},
			}

			workspaceDeptGiven := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: &deptId,
					Name:         "workspace-dept",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("workspace/dept/can-be-overridden"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Args: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "workspace/dept/can-be-overridden"},
							},
						},
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "workspace/proj/can-be-overridden"},
										},
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false), // cannot-be-overridden (instance rules)
								},
							},
						},
					},
				},
			}

			workspaceProjGiven := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: &projectId,
					Name:      "workspace-proj",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("workspace/proj/can-be-overridden"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Args: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "workspace/proj/can-be-overridden"},
							},
						},
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "workspace/proj/can-be-overridden"},
										},
									},
								},
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false), // cannot-be-overridden (instance rules)
								},
							},
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackDeptGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackProjectGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(workspaceDeptGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(workspaceProjGiven), tenantId, http.MethodPut)

			sourceFallbackDept := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
				IsFallback:   true,
			}
			sourceFallbackProj := api.SourceOfRule{
				Scope:      api.SCOPETYPE_PROJECT,
				ProjectId:  &projectId,
				IsFallback: true,
			}
			sourceWorkspaceDept := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: &deptId,
			}
			sourceWorkspaceProj := api.SourceOfRule{
				Scope:     api.SCOPETYPE_PROJECT,
				ProjectId: &projectId,
			}

			fallbackPolicyDept := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))

			Expect(*fallbackPolicyDept.Effective.Defaults.Image).To(Equal("fallback/dept/can-be-overridden"))
			Expect(*fallbackPolicyDept.Effective.Rules.Args.SourceOfRule).To(Equal(sourceFallbackDept))
			Expect(*fallbackPolicyDept.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceFallbackDept))
			Expect(*fallbackPolicyDept.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceFallbackDept)) // instances - parent takes over

			fallbackPolicyProject := getFallbackPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyFallback))

			Expect(*fallbackPolicyProject.Effective.Defaults.Image).To(Equal("fallback/proj/can-be-overridden"))
			Expect(*fallbackPolicyProject.Effective.Rules.Args.SourceOfRule).To(Equal(sourceFallbackProj))
			Expect(*fallbackPolicyProject.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceFallbackProj)) // attributes - overridable
			Expect(*fallbackPolicyProject.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceFallbackDept))           // instances - parent takes over

			workspacePolicyDept := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyWorkspace))

			Expect(*workspacePolicyDept.Effective.Defaults.Image).To(Equal("workspace/dept/can-be-overridden"))
			Expect(*workspacePolicyDept.Effective.Rules.Args.SourceOfRule).To(Equal(sourceWorkspaceDept))
			Expect(*workspacePolicyDept.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceWorkspaceDept))
			Expect(*workspacePolicyDept.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceWorkspaceDept)) // instances - parent takes over

			workspacePolicyProj := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyWorkspace))

			Expect(*workspacePolicyProj.Effective.Defaults.Image).To(Equal("workspace/proj/can-be-overridden"))
			Expect(*workspacePolicyProj.Effective.Rules.Args.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceWorkspaceDept)) // instances - parent takes over

			//
			//  now delete the fallback policy / department
			//
			deletePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=department&departmentId=211", urlPolicyFallback))

			workspacePolicyProj = getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyWorkspace))

			Expect(*workspacePolicyProj.Effective.Defaults.Image).To(Equal("workspace/proj/can-be-overridden"))
			Expect(*workspacePolicyProj.Effective.Rules.Args.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceWorkspaceDept)) // instances - parent takes over

			//
			//  now delete the fallback policy / project
			//
			deletePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyFallback))

			workspacePolicyProj = getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=2111", urlPolicyWorkspace))

			Expect(*workspacePolicyProj.Effective.Defaults.Image).To(Equal("workspace/proj/can-be-overridden"))
			Expect(*workspacePolicyProj.Effective.Rules.Args.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Attributes.Quantity.SourceOfRule).To(Equal(sourceWorkspaceProj))
			Expect(*workspacePolicyProj.Effective.Rules.Compute.ExtendedResources.Instances.SourceOfRule).To(Equal(sourceWorkspaceDept)) // instances - parent takes over
		})
	})

	Context("distributed combo", func() {
		It("fallback only defaults, distributed: master rules/worker defaults", func() {
			tenantId := int32(2)
			clusterId := backendMock.ClusterIdPerTenant["2"][0]

			fallbackGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "fallback-cluster",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Defaults: &api.SupersetDefaults{
						Image: convutils.Ref("fallback"),
					},
				},
			}

			distributedGiven := api.DistributedPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "distributed-cluster",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Master: &api.ReplicaRulesV2{
							Args: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "distributed/master"},
								},
							},
						},
					},
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							Image: convutils.Ref("distributed/worker"),
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyDistributed, makePayload(distributedGiven), tenantId, http.MethodPut)

			sourceDistributed := api.SourceOfRule{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: &clusterId,
			}

			// defaults + master = fallback
			// defaults + worker = distributed
			// rules + master = distributed
			// rules + worker = none
			workspacePolicy := getDistributedPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=cluster&clusterId=%s", urlPolicyDistributed, clusterId))

			Expect(*workspacePolicy.Effective.Defaults.Master.Image).To(Equal("fallback"))
			Expect(*workspacePolicy.Effective.Defaults.Worker.Image).To(Equal("distributed/worker"))
			Expect(*workspacePolicy.Effective.Rules.Master.Args.SourceOfRule).To(Equal(sourceDistributed))
			testutils.Expect(workspacePolicy.Effective.Rules.Worker).ToEqual(dao.SystemPolicy.Policy.Rules)
		})

		It("fallback only rules, distributed: master defaults/worker rules", func() {
			tenantId := int32(2)
			clusterId := backendMock.ClusterIdPerTenant["2"][0]

			fallbackGiven := api.FallbackPolicyOverwriteRequest{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "fallback-cluster",
				},
				Policy: &api.FallbackPolicyDefaultsAndRules{
					Rules: &api.SupersetRules{
						Args: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "fallback"},
							},
						},
					},
				},
			}

			distributedGiven := api.DistributedPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					ClusterId: &clusterId,
					Name:      "distributed-cluster",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Master: &api.ReplicaRulesV2{
							Image: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "distributed/master"},
								},
							},
						},
					},
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							WorkingDir: convutils.Ref("distributed/worker"),
						},
					},
				},
			}

			applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackGiven), tenantId, http.MethodPut)
			applyPolicyForTest(router, urlPolicyDistributed, makePayload(distributedGiven), tenantId, http.MethodPut)

			sourceFallback := api.SourceOfRule{
				Scope:      api.SCOPETYPE_CLUSTER,
				ClusterId:  &clusterId,
				IsFallback: true,
			}

			sourceDistributed := api.SourceOfRule{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: &clusterId,
			}

			// rules + master = distributed
			// rules + worker = fallback
			// defaults + worker = distributed
			// defaults + master = none
			workspacePolicy := getDistributedPolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=cluster&clusterId=%s", urlPolicyDistributed, clusterId))

			Expect(*workspacePolicy.Effective.Defaults.Worker.WorkingDir).To(Equal("distributed/worker"))
			testutils.Expect(workspacePolicy.Effective.Defaults.Master).ToEqual(dao.SystemPolicy.Policy.Defaults)
			Expect(*workspacePolicy.Effective.Rules.Master.Image.SourceOfRule).To(Equal(sourceDistributed))
			Expect(*workspacePolicy.Effective.Rules.Worker.Args.SourceOfRule).To(Equal(sourceFallback))
		})

		Context("storage blocking", func() {
			It("fallback only nfs, workspace also pvc", func() {
				tenantId := int32(2)
				projectId := int32(2111)

				fallbackGiven := api.FallbackPolicyOverwriteRequest{
					Meta: api.PolicyCreationFields{
						Scope: api.SCOPETYPE_TENANT,
						Name:  "fallback",
					},
					Policy: &api.FallbackPolicyDefaultsAndRules{
						Rules: &api.SupersetRules{
							Storage: &api.SupersetRulesAllOfStorage{
								Nfs: &api.NfssRules{
									Instances: &api.ItemRules{
										CanAdd: convutils.Ref(false),
									},
								},
							},
						},
					},
				}

				projectGiven := api.WorkspacePolicyOverwriteRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:     api.SCOPETYPE_PROJECT,
						ProjectId: &projectId,
						Name:      "project",
					},
					Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
						Rules: &api.WorkspacePolicyRulesV2{
							Storage: &api.SupersetRulesAllOfStorage{
								Pvc: &api.PvcsRules{
									Instances: &api.ItemRules{
										CanAdd: convutils.Ref(false),
									},
								},
							},
						},
					},
				}

				applyPolicyForTest(router, urlPolicyFallback, makePayload(fallbackGiven), tenantId, http.MethodPut)

				// only NFS at this point is blocked

				workspacePolicy := getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyWorkspace, projectId))

				Expect(workspacePolicy.Effective.Rules.Storage.Nfs).NotTo(BeNil())
				Expect(workspacePolicy.Effective.Rules.Storage.Pvc).To(BeNil())

				// now add PVC block to the project level

				applyPolicyForTest(router, urlPolicyWorkspace, makePayload(projectGiven), tenantId, http.MethodPut)

				workspacePolicy = getWorkspacePolicyForTest(router, tenantId, fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyWorkspace, projectId))

				Expect(workspacePolicy.Effective.Rules.Storage.Nfs).NotTo(BeNil())
				Expect(workspacePolicy.Effective.Rules.Storage.Pvc).NotTo(BeNil())
			})
		})
	})
})
