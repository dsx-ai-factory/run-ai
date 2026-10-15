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
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	assetsService "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	runaictx_http "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx/http"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clusterServiceApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/cluster"
	orgUnitService "github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/tree"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
)

var (
	urlPolicy                     = "/api/v2/policy"
	urlPolicyWorkspace            = "/api/v2/policy/workspaces"
	urlPolicyTraining             = "/api/v2/policy/trainings"
	urlPolicyDistributed          = "/api/v2/policy/distributed"
	urlPolicyInference            = "/api/v2/policy/inferences"
	urlPolicyDistributedInference = "/api/v2/policy/distributed-inferences"
	urlPolicyFallback             = "/api/v2/policy/fallback"
	urlPolicyNimService           = "/api/v2/policy/nim-services"

	urlAssetHook = "/internal/policy/asset"
)

type ExpectedResult struct {
	Code    int
	Message string
}

var _ = Describe("ApplyTest", Ordered, func() {
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

		assets.SetMock(assets.SimpleMockClientFactory{
			SimpleMockClient: assets.NewSimpleMockClient(),
		})
	})

	Context("GPU memory defaultFrom compatibility", func() {
		applyPolicyRules := func(url string, rules map[string]interface{}) *httptest.ResponseRecorder {
			body := map[string]interface{}{
				"meta":   map[string]interface{}{"scope": "tenant", "name": "gpu-memory-dependencies"},
				"policy": map[string]interface{}{"rules": rules},
			}
			request, err := http.NewRequest(http.MethodPatch, url, makePayload(body))
			Expect(err).NotTo(HaveOccurred())
			request.Header = *headerForAdmin(1)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			return response
		}
		applyRules := func(rules map[string]interface{}) *httptest.ResponseRecorder {
			return applyPolicyRules(urlPolicyWorkspace, map[string]interface{}{"compute": rules})
		}
		derivedFrom := func(field string) map[string]interface{} {
			return map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute." + field}}
		}
		effective := func(response *httptest.ResponseRecorder) map[string]interface{} {
			var body map[string]interface{}
			Expect(json.Unmarshal(response.Body.Bytes(), &body)).To(Succeed())
			return convutils.MapMe(body["effective"])
		}
		getPolicy := func() *httptest.ResponseRecorder {
			request, err := http.NewRequest(http.MethodGet, urlPolicyWorkspace+"?scope=tenant", nil)
			Expect(err).NotTo(HaveOccurred())
			request.Header = *headerForAdmin(1)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
			return response
		}

		DescribeTable("rejects a default derived from the same field",
			func(target, source string) {
				response := applyRules(map[string]interface{}{target: derivedFrom(source)})
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body.String()).To(ContainSubstring("defaultFrom.field must be a different field"))
			},
			Entry("ordinary CPU self reference", "cpuCoreRequest", "cpuCoreRequest"),
			Entry("legacy request from new request", "gpuPortionRequest", "gpuMemoryPortionRequest"),
			Entry("new request from legacy request", "gpuMemoryPortionRequest", "gpuPortionRequest"),
			Entry("legacy limit from new limit", "gpuPortionLimit", "gpuMemoryPortionLimit"),
			Entry("new limit from legacy limit", "gpuMemoryPortionLimit", "gpuPortionLimit"),
		)

		DescribeTable("rejects alias self references in each policy section",
			func(url, section string) {
				rules := map[string]interface{}{"compute": map[string]interface{}{"gpuPortionRequest": derivedFrom("gpuMemoryPortionRequest")}}
				if section != "" {
					rules = map[string]interface{}{section: rules}
				}
				response := applyPolicyRules(url, rules)
				Expect(response.Code).To(Equal(http.StatusBadRequest))
				Expect(response.Body.String()).To(ContainSubstring("defaultFrom.field must be a different field"))
			},
			Entry("training", urlPolicyTraining, ""),
			Entry("inference", urlPolicyInference, ""),
			Entry("NIM", urlPolicyNimService, ""),
			Entry("fallback", urlPolicyFallback, ""),
			Entry("distributed worker", urlPolicyDistributed, "worker"),
			Entry("distributed master", urlPolicyDistributed, "master"),
			Entry("distributed inference worker", urlPolicyDistributedInference, "worker"),
			Entry("distributed inference leader", urlPolicyDistributedInference, "leader"),
		)

		DescribeTable("keeps response dependency paths consistent for distinct fields",
			func(target, source, legacyTarget, legacySource string) {
				response := applyRules(map[string]interface{}{target: derivedFrom(source)})
				Expect(response.Code).To(Equal(http.StatusOK))
				applied := effective(response)
				compute := convutils.MapMe(convutils.MapMe(applied["rules"])["compute"])
				reference := convutils.MapMe(convutils.MapMe(compute[legacyTarget])["defaultFrom"])
				Expect(reference["field"]).To(Equal("compute." + legacySource))
				order := convutils.MapMe(convutils.MapMe(applied["defaultFromOrder"])["worker"])
				Expect(order[""]).To(Equal([]interface{}{"compute." + legacySource, "compute." + legacyTarget}))
				Expect(effective(getPolicy())).To(Equal(applied))
			},
			Entry("CPU from legacy request", "cpuCoreRequest", "gpuPortionRequest", "cpuCoreRequest", "gpuPortionRequest"),
			Entry("CPU from new request", "cpuCoreRequest", "gpuMemoryPortionRequest", "cpuCoreRequest", "gpuPortionRequest"),
			Entry("new limit from legacy request", "gpuMemoryPortionLimit", "gpuPortionRequest", "gpuPortionLimit", "gpuPortionRequest"),
			Entry("legacy limit from new request", "gpuPortionLimit", "gpuMemoryPortionRequest", "gpuPortionLimit", "gpuPortionRequest"),
		)

		It("includes distributed inference leader dependencies in the primary replica order", func() {
			response := applyPolicyRules(urlPolicyDistributedInference, map[string]interface{}{
				"leader": map[string]interface{}{"compute": map[string]interface{}{"gpuMemoryPortionLimit": derivedFrom("gpuPortionRequest")}},
			})
			Expect(response.Code).To(Equal(http.StatusOK))
			order := convutils.MapMe(convutils.MapMe(effective(response)["defaultFromOrder"])["master"])
			Expect(order[""]).To(Equal([]interface{}{"compute.gpuPortionRequest", "compute.gpuPortionLimit"}))
		})

		It("rejects a mixed-alias cycle introduced by PATCH and preserves the saved policy", func() {
			initial := applyRules(map[string]interface{}{"cpuCoreRequest": derivedFrom("gpuPortionRequest")})
			Expect(initial.Code).To(Equal(http.StatusOK))
			before := effective(getPolicy())

			response := applyRules(map[string]interface{}{"gpuMemoryPortionRequest": derivedFrom("cpuCoreRequest")})
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("cycle"))
			Expect(effective(getPolicy())).To(Equal(before))
		})
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Authentication
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("no authentication header", func() {
		It("should fail forbidden error", func() {
			body := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("11"),
					Name:         "pol1",
				},
			}

			request, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(body))
			Expect(err).To(BeNil())

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusForbidden))
		})
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Validation
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("system scope", func() {
		It("should return with error", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_SYSTEM,
					Name:  "pol-s",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{},
			}

			requestAdd, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(2)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("system scope is not supported for policies"))
		})
	})

	Context("system admin with tenant permission accesses cluster policy", func() {
		It("should succeed when GetAncestors fails but user has tenant permission", func() {
			clusterIdNotInBackend := "f57cb0f6-df05-4c2f-8d60-998125440ce4"

			clusterMock.Clusters[clusterIdNotInBackend] = &clusterServiceApi.DisplayedCluster{
				Uuid:     clusterIdNotInBackend,
				TenantId: 1,
				Version:  *clusterServiceApi.NewNullableString(convutils.Ref(testutils.DefaultClusterVersionForTesting)),
			}
			cluster.SetMock(clusterMock)

			authMock.SetPermittedScopes(authClient.ACTION_READ, &authClient.PermittedScopes{
				Tenants: []string{"1"},
			})

			url := fmt.Sprintf("%s?scope=cluster&clusterId=%s", urlPolicyWorkspace, clusterIdNotInBackend)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request.WithContext(tenantContextForTest(1)))
			Expect(response.Code).To(Equal(http.StatusOK), response.Body.String())
		})
	})

	Context("empty policy", func() {
		It("should return with error", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{},
			}

			requestAdd, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(2)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("canEdit=false without default value", func() {
		It("should return with error", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "policy1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						ImagePullPolicy: &api.ImagePullPolicyRules{
							CanEdit: convutils.Ref(false),
							Options: &[]api.ImagePullPolicyOptionsOptionsInner{
								{Value: convutils.Ref(api.IMAGEPULLPOLICY_ALWAYS)},
								{Value: convutils.Ref(api.IMAGEPULLPOLICY_NEVER)},
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						// imagePullPolicy default is intentionally omitted
					},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(2)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)

			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("imagePullPolicy"))
			Expect(responseAdd.Body.String()).To(ContainSubstring("default value must be provided for policy fields having canEdit=false"))
		})
	})

	Context("canEdit=false at project level, and default is defined at tenant level", func() {
		It("should allow creation", func() {
			// Level 1: Tenant policy with default
			bodyTenant := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "tenant-with-default",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						ImagePullPolicy: convutils.Ref(api.IMAGEPULLPOLICY_NEVER),
					},
				},
			}

			reqTenant, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenant))
			Expect(err).To(BeNil())
			reqTenant.Header = *headerForAdmin(1)

			respTenant := httptest.NewRecorder()
			router.ServeHTTP(respTenant, reqTenant)
			Expect(respTenant.Code).To(Equal(http.StatusOK))

			// Level 2: Project policy with canEdit=false (no default)
			bodyProject := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "proj-canEdit-false",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						ImagePullPolicy: &api.ImagePullPolicyRules{
							CanEdit: convutils.Ref(false),
							Options: &[]api.ImagePullPolicyOptionsOptionsInner{
								{Value: convutils.Ref(api.IMAGEPULLPOLICY_ALWAYS)},
								{Value: convutils.Ref(api.IMAGEPULLPOLICY_NEVER)},
							},
						},
					},
				},
			}

			reqProj, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProject))
			Expect(err).To(BeNil())
			reqProj.Header = *headerForAdmin(1)

			respProj := httptest.NewRecorder()
			router.ServeHTTP(respProj, reqProj)
			Expect(respProj.Code).To(Equal(http.StatusOK))

			var actualResp api.WorkspacePolicyV2
			err = json.Unmarshal(respProj.Body.Bytes(), &actualResp)
			Expect(err).To(BeNil())

			Expect(*actualResp.Effective.Defaults.ImagePullPolicy).To(Equal(api.IMAGEPULLPOLICY_NEVER))
			Expect(*actualResp.Effective.Rules.ImagePullPolicy.CanEdit).To(BeFalse())
		})
	})

	Context("rules/defaults + reset for the same field", func() {
		It("should return with error", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
							RunAsGid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Compute: &api.SupersetDefaultsAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesDefaults{
								Instances: &[]api.ExtendedResource{
									{
										Resource: convutils.Ref("aaa/bbb"),
										Quantity: convutils.Ref("10G"),
									},
								},
							},
						},
						Storage: &api.SupersetDefaultsAllOfStorage{
							Git: &api.GitsDefaults{
								Attributes: &api.GitInstance{
									Branch: convutils.Ref("hello"),
								},
							},
						},
					},
				},
			}

			expectResetErrorWorkspace(router, &addBody, "security.runAsGid",
				"security.runAsGid: reset and rules are mutually excluded")
			expectResetErrorWorkspace(router, &addBody, "security.runAsNonRoot", "")
			expectResetErrorWorkspace(router, &addBody, "compute.extendedResources",
				"compute.extendedResources: reset and defaults are mutually excluded")
			expectResetErrorWorkspace(router, &addBody, "compute.gpuDevicesRequest", "")
			expectResetErrorWorkspace(router, &addBody, "storage.git",
				"storage.git: reset and defaults are mutually excluded")
			expectResetErrorWorkspace(router, &addBody, "storage.pvc", "")
		})
	})

	Context("rules/defaults + reset for the same field in distributes", func() {
		It("should return with error", func() {
			addBody := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
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
								RunAsGid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
						},
					},
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							Compute: &api.SupersetDefaultsAllOfCompute{
								ExtendedResources: &api.ExtendedResourcesDefaults{
									Instances: &[]api.ExtendedResource{
										{
											Resource: convutils.Ref("aaa/bbb"),
											Quantity: convutils.Ref("10G"),
										},
									},
								},
							},
						},
						Master: &api.ReplicaDefaultsV2{
							Storage: &api.SupersetDefaultsAllOfStorage{
								Git: &api.GitsDefaults{
									Attributes: &api.GitInstance{
										Branch: convutils.Ref("hello"),
									},
								},
							},
						},
					},
				},
			}

			expectResetErrorDistributed(router, &addBody, "worker.security.runAsUid",
				"worker.security.runAsUid: reset and rules are mutually excluded")
			expectResetErrorDistributed(router, &addBody, "master.security.runAsUid", "")
			expectResetErrorDistributed(router, &addBody, "master.security.runAsGid",
				"master.security.runAsGid: reset and rules are mutually excluded")
			expectResetErrorDistributed(router, &addBody, "worker.security.runAsGid", "")

			expectResetErrorDistributed(router, &addBody, "worker.compute.extendedResources",
				"worker.compute.extendedResources: reset and defaults are mutually excluded")
			expectResetErrorDistributed(router, &addBody, "master.compute.extendedResources", "")
			expectResetErrorDistributed(router, &addBody, "master.storage.git",
				"master.storage.git: reset and defaults are mutually excluded")
			expectResetErrorDistributed(router, &addBody, "worker.storage.git", "")
		})
	})

	Context("get project policy -> apply tenant policy -> get project", func() {
		It("first project is created empty, then the 2nd get should return the effective", func() {
			projectId := 2111

			//
			//   get policy of the project should return empty effective
			//
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyWorkspace, projectId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			var emptyPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(response.Body.Bytes(), &emptyPolicy)
			Expect(err).To(BeNil())

			Expect(emptyPolicy.Policy.Defaults).To(BeNil())
			Expect(emptyPolicy.Policy.Rules).To(BeNil())
			testutils.Expect(emptyPolicy.Effective).ToEqual(dao.SystemPolicy.Policy)

			//
			//   now apply a policy to the tenant
			//
			tenantPolicy := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Args: convutils.Ref("rotem=1"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Args: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}
			applyPolicyForTest(router, urlPolicyWorkspace, makePayload(tenantPolicy), 2, http.MethodPatch)

			//
			//   get policy of the project should return empty given, cluster effective
			//
			request, err = http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyWorkspace, projectId), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			startTime := time.Now()
			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			var hasEffective api.WorkspacePolicyV2
			err = json.Unmarshal(response.Body.Bytes(), &hasEffective)
			Expect(err).To(BeNil())

			Expect(hasEffective.Policy.Defaults).To(BeNil())
			Expect(hasEffective.Policy.Rules).To(BeNil())
			Expect(hasEffective.Effective.Defaults).NotTo(BeNil())
			Expect(hasEffective.Effective.Rules).NotTo(BeNil())

			// while testing timestamps, we allow some time deviation between DB and e2e test host. I took a 1 sec tolerance.
			Expect(hasEffective.Meta.CreatedAt).To(BeTemporally(">=", hasEffective.Meta.UpdatedAt.Add(-time.Second)))
			Expect(hasEffective.Meta.CreatedAt).To(BeTemporally(">=", startTime.Add(-time.Second)))
			Expect(hasEffective.Meta.CreatedAt).To(BeTemporally("<=", time.Now().Add(time.Second)))

			Expect(convutils.Deref(hasEffective.Effective.Defaults.Args, "")).To(Equal(*tenantPolicy.Policy.Defaults.Args))
			Expect(hasEffective.Effective.Rules.Args.SourceOfRule.Scope).To(Equal(api.SCOPETYPE_TENANT))
		})
	})

	Context("training policy - Privileged field", func() {
		It("apply policy with default privileged field", func() {
			projectID := int32(1111)
			policyName := "pol-t1"
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							Privileged: convutils.Ref(true),
						},
					},
				},
			}

			// Apply the policy
			response := overwriteTrainingPolicyForTest(router, &addBody, 1)
			Expect(response.Code).To(Equal(http.StatusOK))

			// Verify the response
			var actualResp api.TrainingPolicyV2
			err := json.Unmarshal(response.Body.Bytes(), &actualResp)
			Expect(err).To(BeNil())
			Expect(*actualResp.Policy.Defaults.Security.Privileged).To(Equal(*addBody.Policy.Defaults.Security.Privileged))

			// Verify the GET training policies response
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyTraining, projectID), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			err = json.Unmarshal(response.Body.Bytes(), &actualResp)
			Expect(err).To(BeNil())
			Expect(*actualResp.Policy.Defaults.Security.Privileged).To(Equal(*addBody.Policy.Defaults.Security.Privileged))
		})
	})

	Context("distributed policy - Privileged field", func() {
		It("apply policy with required privileged field rule", func() {
			projectID := int32(1111)
			policyName := "pol-t1"
			addBody := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Rules: &api.DistributedPolicyRulesV2{
						Master: &api.ReplicaRulesV2{
							Security: &api.SupersetRulesAllOfSecurity{
								Privileged: &api.BooleanRules{
									Required: convutils.Ref(true),
								},
							},
						},
					},
				},
			}

			overwriteBody := &api.DistributedPolicyOverwriteRequestV2{
				Meta:   addBody.Meta,
				Policy: addBody.Policy,
			}

			// Apply the policy
			response := overwriteDistributedPolicyForTest(router, overwriteBody, 1)
			Expect(response.Code).To(Equal(http.StatusOK))

			// Verify the response
			var actualResp api.DistributedPolicyV2
			err := json.Unmarshal(response.Body.Bytes(), &actualResp)
			Expect(err).To(BeNil())
			Expect(*actualResp.Policy.Rules.Master.Security.Privileged.Required).To(Equal(*overwriteBody.Policy.Rules.Master.Security.Privileged.Required))

			// Verify the GET distributed policies response
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyDistributed, projectID), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(1)

			response = httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			err = json.Unmarshal(response.Body.Bytes(), &actualResp)
			Expect(err).To(BeNil())
			Expect(*actualResp.Policy.Rules.Master.Security.Privileged.Required).To(Equal(*overwriteBody.Policy.Rules.Master.Security.Privileged.Required))
		})
	})

	Context("empty policy", func() {
		It("should be rejected", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: nil,
				},
			}

			request, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("empty policy", func() {
		It("should be rejected", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{},
			}

			request, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("attempt to delete all fields", func() {
		It("should be rejected", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image"),
					},
				},
			}

			updateWorkspacePolicyForTest(router, &addBody, 1)

			//
			//  try to delete all the fields
			//
			deleteFieldsBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Reset: []string{
					"image",
				},
			}

			request, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(deleteFieldsBody))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("policy recreation updates createdAt", func() {
		It("should update createdAt when training policy is recreated", func() {
			projectID := int32(1111)
			policyName := "pol-t1"

			// Phase 1: Create the policy
			initialBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			overwritePolicyResponse1 := overwriteTrainingPolicyForTest(router, &initialBody, 1)

			var actualResp1 api.TrainingPolicyV2
			err := json.Unmarshal(overwritePolicyResponse1.Body.Bytes(), &actualResp1)
			Expect(err).To(BeNil())
			createdAt1 := actualResp1.Meta.CreatedAt
			Expect(createdAt1).NotTo(BeNil())

			updatedAt1 := actualResp1.Meta.UpdatedAt
			Expect(updatedAt1).NotTo(BeNil())

			// Phase 2: Delete the policy
			deletePolicyForTest(router, 1, fmt.Sprintf("%s?scope=project&projectId=%d", urlPolicyTraining, projectID))

			// Wait a bit and recreate policy with same name
			time.Sleep(1 * time.Second)

			recreateBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			// Phase 3: Re-create the policy
			overwritePolicyResponse2 := overwriteTrainingPolicyForTest(router, &recreateBody, 1)

			var actualResp2 api.TrainingPolicyV2
			err = json.Unmarshal(overwritePolicyResponse2.Body.Bytes(), &actualResp2)
			Expect(err).To(BeNil())

			createdAt2 := actualResp2.Meta.CreatedAt
			Expect(createdAt2).NotTo(BeNil())
			updatedAt2 := actualResp2.Meta.UpdatedAt
			Expect(updatedAt2).NotTo(BeNil())

			// Phase 4: Check that the creation time and the update time has been updated
			Expect(createdAt2.After(createdAt1)).To(BeTrue(), "createdAt should be updated after recreation")
			Expect(updatedAt2.After(updatedAt1)).To(BeTrue(), "updatedAt should be updated after recreation")
		})
	})

	Context("overwrite workspace policy", func() {
		It("should replace the existing one", func() {
			addBody := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image"),
					},
				},
			}

			response1 := overwriteWorkspacePolicyForTest(router, &addBody, 1)

			var actualResp1 api.WorkspacePolicyV2
			err := json.Unmarshal(response1.Body.Bytes(), &actualResp1)
			Expect(err).To(BeNil())
			Expect(*actualResp1.Policy.Defaults).To(Equal(*addBody.Policy.Defaults))
			Expect(*actualResp1.Policy.Rules).To(Equal(*addBody.Policy.Rules))

			//
			//   now overwrite it with another policy
			//
			overwriteBody := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(500)),
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(500)),
						},
					},
				},
			}

			response2 := overwriteWorkspacePolicyForTest(router, &overwriteBody, 1)

			var actualResp2 api.WorkspacePolicyV2
			err = json.Unmarshal(response2.Body.Bytes(), &actualResp2)
			Expect(err).To(BeNil())
			Expect(*actualResp2.Policy.Defaults).To(Equal(*overwriteBody.Policy.Defaults))
			Expect(*actualResp2.Policy.Rules).To(Equal(*overwriteBody.Policy.Rules))
		})
	})

	Context("overwrite training policy", func() {
		It("should replace the existing one", func() {
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("image"),
					},
				},
			}

			response1 := overwriteTrainingPolicyForTest(router, &addBody, 1)

			var actualResp1 api.TrainingPolicyV2
			err := json.Unmarshal(response1.Body.Bytes(), &actualResp1)
			Expect(err).To(BeNil())
			Expect(*actualResp1.Policy.Defaults).To(Equal(*addBody.Policy.Defaults))
			Expect(*actualResp1.Policy.Rules).To(Equal(*addBody.Policy.Rules))

			//
			//   now overwrite it with another policy
			//
			overwriteBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(500)),
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(500)),
						},
					},
				},
			}

			response2 := overwriteTrainingPolicyForTest(router, &overwriteBody, 1)

			var actualResp2 api.TrainingPolicyV2
			err = json.Unmarshal(response2.Body.Bytes(), &actualResp2)
			Expect(err).To(BeNil())
			Expect(*actualResp2.Policy.Defaults).To(Equal(*overwriteBody.Policy.Defaults))
			Expect(*actualResp2.Policy.Rules).To(Equal(*overwriteBody.Policy.Rules))
		})
	})

	Context("overwrite of empty rules", func() {
		It("should reject the overwrite if there are no defaults", func() {
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			response1 := overwriteTrainingPolicyForTest(router, &addBody, 1)

			var actualResp1 api.TrainingPolicyV2
			err := json.Unmarshal(response1.Body.Bytes(), &actualResp1)
			Expect(err).To(BeNil())
			Expect(*actualResp1.Policy.Rules).To(Equal(*addBody.Policy.Rules))

			//
			//   now overwrite it with policy having no rules
			//
			overwriteBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{},
				},
			}

			request2, err := http.NewRequest(http.MethodPut, urlPolicyTraining, makePayload(overwriteBody))
			Expect(err).To(BeNil())
			request2.Header = *headerForAdmin(1)

			response2 := httptest.NewRecorder()
			router.ServeHTTP(response2, request2)
			Expect(response2.Code).To(Equal(http.StatusBadRequest))
			Expect(response2.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("attempt to delete all fields", func() {
		It("should be rejected", func() {
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			updateWorkspacePolicyForTest(router, &addBody, 1)

			//
			//  try to delete all the fields
			//
			deleteFieldsBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Reset: []string{
					"image",
				},
			}

			request, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(deleteFieldsBody))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusBadRequest))
			Expect(response.Body.String()).To(ContainSubstring("To apply a policy, there must be at least one default or rule"))
		})
	})

	Context("options with no default", func() {
		It("should be ok, we do not need default in this case", func() {
			body := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Options: &[]api.StringOption{
								{
									Value: "ubuntu",
								},
								{
									Value: "redhat",
								},
							},
						},
					},
				},
			}

			request, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(body))
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(2)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))
		})
	})

	Context("asset deleted, but it is not part of any policy", func() {
		It("no impact on policy", func() {
			tenantId := int32(1)

			//
			//   create pvc asset
			//
			pvcW := assetsService.DatasourceListResponseEntry{
				Meta: assetsService.AssetMeta{
					Id:       uuid.NewString(),
					Kind:     assetsService.ASSETKIND_PVC,
					TenantId: *assetsService.NewNullableInt32(&tenantId),
					Scope:    assetsService.SCOPE_TENANT,
				},
				Spec: assetsService.DatasourceListResponseAssetSpec{
					Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
						ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
						ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
						Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
					}),
				},
			}

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   simulate deletion of the pvc asset
			//
			deletedEvent := api.AssetChangeNotification{
				ChangeType: api.ASSETCHANGETYPE_DELETED,
				ScopeType:  api.SCOPETYPE_TENANT,
				ScopeId:    convutils.Ref(fmt.Sprintf("%d", tenantId)),
			}

			requestDelete, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAssetHook, pvcW.Meta.Id), makePayload(deletedEvent))
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(tenantId)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete)
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))

			//
			//   list is empty
			//
			actualResponse := listPoliciesForTest(router, tenantId, fmt.Sprintf("%s", urlPolicy))
			Expect(len(actualResponse.Policies)).To(Equal(0))
		})
	})

	Context("imposed assets validation", func() {
		It("malformed ids should be rejected", func() {
			tenantId := int32(1)

			//
			//   add a workspace policy with asset having malformed id
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{uuid.NewString(), "xyz"},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("malformed asset id 'xyz'"))
		})

		It("empty id should be rejected", func() {
			tenantId := int32(1)

			//
			//   add a workspace policy with asset having malformed id
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{uuid.NewString(), ""},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("should not be empty"))
		})

		It("duplicate ids rejected", func() {
			tenantId := int32(1)

			id1 := uuid.NewString()
			id2 := uuid.NewString()

			//
			//   add a workspace policy with asset having malformed id
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{id1, id2, id1},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("appears more than once"))
		})

		It("asset not found - rejected", func() {
			tenantId := int32(1)

			id1 := uuid.NewString()
			id2 := uuid.NewString()

			//
			//   create pvc asset
			//
			pvcW := assetsService.DatasourceListResponseEntry{
				Meta: assetsService.AssetMeta{
					Id:       uuid.NewString(),
					Kind:     assetsService.ASSETKIND_PVC,
					TenantId: *assetsService.NewNullableInt32(&tenantId),
					Scope:    assetsService.SCOPE_TENANT,
				},
				Spec: assetsService.DatasourceListResponseAssetSpec{
					Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
						ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
						ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
						Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
					}),
				},
			}

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   add a workspace policy with asset having malformed id
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{id1, id2, pvcW.Meta.Id},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusNotFound))

			msg1 := fmt.Sprintf("asset(s) %s, %s do not exist", id1, id2)
			msg2 := fmt.Sprintf("asset(s) %s, %s do not exist", id2, id1)
			Expect(responseAdd.Body.String()).To(Or(ContainSubstring(msg1), ContainSubstring(msg2)))
		})

		Context("scope validation", func() {
			It("reject the asset if scope is not same as of policy", func() {
				tenantId := int32(1)

				//
				//   create pvc asset
				//
				pvcW := assetsService.DatasourceListResponseEntry{
					Meta: assetsService.AssetMeta{
						Id:       uuid.NewString(),
						Kind:     assetsService.ASSETKIND_PVC,
						TenantId: *assetsService.NewNullableInt32(&tenantId),
						Scope:    assetsService.SCOPE_TENANT,
					},
					Spec: assetsService.DatasourceListResponseAssetSpec{
						Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
							ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
							ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
							Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
						}),
					},
				}

				assets.SetMock(assets.SimpleMockClientFactory{
					SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
				})

				//
				//   add a workspace policy with single asset only
				//
				addBody := api.WorkspacePolicyChangeRequestV2{
					Meta: api.PolicyCreationFields{
						Scope:        api.SCOPETYPE_DEPARTMENT,
						DepartmentId: convutils.Ref("111"),
						Name:         "pol-t1",
					},
					Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
						ImposedAssets: &[]string{pvcW.Meta.Id},
					},
				}

				requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
				Expect(err).To(BeNil())
				requestAdd.Header = *headerForAdmin(tenantId)

				responseAdd := httptest.NewRecorder()
				router.ServeHTTP(responseAdd, requestAdd)
				Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
				Expect(responseAdd.Body.String()).To(ContainSubstring("(tenant:1) does not match to to the scope of the policy (department:111)"))
			})
		})
	})

	Context("kind validation", func() {
		It("reject the asset if it is not datasource", func() {
			tenantId := int32(1)

			//
			//   create pvc asset
			//
			envAsset := assetsService.DatasourceListResponseEntry{
				Meta: assetsService.AssetMeta{
					Id:       uuid.NewString(),
					Kind:     assetsService.ASSETKIND_ENVIRONMENT,
					TenantId: *assetsService.NewNullableInt32(&tenantId),
					Scope:    assetsService.SCOPE_TENANT,
				},
				Spec: assetsService.DatasourceListResponseAssetSpec{
					Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
						ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
						ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
						Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
					}),
				},
			}

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(envAsset),
			})

			//
			//   add a workspace policy with single asset only
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{envAsset.Meta.Id},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusBadRequest))
			Expect(responseAdd.Body.String()).To(ContainSubstring("cannot use asset of kind environment as imposedAsset"))
		})
	})

	Context("imposed asset deleted", func() {
		It("expecting policy to be deleted as well", func() {
			tenantId := int32(1)

			//
			//   create pvc asset
			//
			pvcW := assetsService.DatasourceListResponseEntry{
				Meta: assetsService.AssetMeta{
					Id:       uuid.NewString(),
					Kind:     assetsService.ASSETKIND_PVC,
					TenantId: *assetsService.NewNullableInt32(&tenantId),
					Scope:    assetsService.SCOPE_TENANT,
				},
				Spec: assetsService.DatasourceListResponseAssetSpec{
					Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
						ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
						ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
						Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
					}),
				},
			}

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   add a workspace policy with single asset only
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					ImposedAssets: &[]string{pvcW.Meta.Id},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusOK))

			var createdPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(responseAdd.Body.Bytes(), &createdPolicy)
			Expect(err).To(BeNil())

			Expect(createdPolicy.Effective).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage.Pvc).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage.Pvc.Instances).NotTo(BeNil())
			Expect(len(*createdPolicy.Effective.Defaults.Storage.Pvc.Instances)).To(Equal(1))

			pvcInstance := (*createdPolicy.Effective.Defaults.Storage.Pvc.Instances)[0]
			Expect(convutils.Deref(pvcInstance.ClaimName, "")).To(Equal("claim-w"))
			Expect(convutils.Deref(pvcInstance.Name, "")).To(Equal(pvcW.Meta.Id))

			Expect(createdPolicy.Effective.ImposedAssets).NotTo(BeNil())
			Expect(len(*createdPolicy.Effective.ImposedAssets)).To(Equal(1))
			Expect((*createdPolicy.Effective.ImposedAssets)[0]).To(Equal(pvcW.Meta.Id))

			//
			//   change the claim name of the pvc in the asset
			//
			pvcW.Spec.Pvc.Set(&assetsService.PVCAssetSpec{
				ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
				ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w-modified")),
				Path:        *assetsService.NewNullableString(convutils.Ref("/usr/w")),
			})

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   simulate update event of the pvc asset
			//
			updatedEvent := api.AssetChangeNotification{
				ChangeType: api.ASSETCHANGETYPE_UPDATED,
				ScopeType:  api.SCOPETYPE_TENANT,
				ScopeId:    convutils.Ref(fmt.Sprintf("%d", tenantId)),
			}

			requestUpdate, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAssetHook, pvcW.Meta.Id), makePayload(updatedEvent))
			Expect(err).To(BeNil())
			requestUpdate.Header = *headerForAdmin(tenantId)

			responseUpdate := httptest.NewRecorder()
			router.ServeHTTP(responseUpdate, requestUpdate)
			Expect(responseUpdate.Code).To(Equal(http.StatusNoContent))

			//
			//   get policy of the project should be updated with the new claim
			//
			request, err := http.NewRequest("GET", fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			request.Header = *headerForAdmin(tenantId)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			Expect(response.Code).To(Equal(http.StatusOK))

			var updatedPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(response.Body.Bytes(), &updatedPolicy)
			Expect(err).To(BeNil())

			Expect(updatedPolicy.Effective).NotTo(BeNil())
			Expect(updatedPolicy.Effective.Defaults).NotTo(BeNil())
			Expect(updatedPolicy.Effective.Defaults.Storage).NotTo(BeNil())
			Expect(updatedPolicy.Effective.Defaults.Storage.Pvc).NotTo(BeNil())
			Expect(updatedPolicy.Effective.Defaults.Storage.Pvc.Instances).NotTo(BeNil())
			Expect(len(*updatedPolicy.Effective.Defaults.Storage.Pvc.Instances)).To(Equal(1))

			pvcInstance = (*updatedPolicy.Effective.Defaults.Storage.Pvc.Instances)[0]
			Expect(convutils.Deref(pvcInstance.ClaimName, "")).To(Equal("claim-w-modified"))
			Expect(convutils.Deref(pvcInstance.Name, "")).To(Equal(pvcW.Meta.Id))

			Expect(updatedPolicy.Effective.ImposedAssets).NotTo(BeNil())
			Expect(len(*updatedPolicy.Effective.ImposedAssets)).To(Equal(1))
			Expect((*updatedPolicy.Effective.ImposedAssets)[0]).To(Equal(pvcW.Meta.Id))

			//
			//   simulate deletion of the pvc asset
			//
			deletedEvent := api.AssetChangeNotification{
				ChangeType: api.ASSETCHANGETYPE_DELETED,
				ScopeType:  api.SCOPETYPE_TENANT,
				ScopeId:    convutils.Ref(fmt.Sprintf("%d", tenantId)),
			}

			requestDelete, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAssetHook, pvcW.Meta.Id), makePayload(deletedEvent))
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(tenantId)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete)
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))

			//
			//   list is empty
			//
			actualResponse := listPoliciesForTest(router, tenantId, fmt.Sprintf("%s", urlPolicy))
			Expect(len(actualResponse.Policies)).To(Equal(0))
		})
	})

	Context("imposed asset cannot be updated due to policy conflict", func() {
		It("request should be rejected", func() {
			tenantId := int32(1)

			//
			//   create pvc asset
			//
			pvcW := assetsService.DatasourceListResponseEntry{
				Meta: assetsService.AssetMeta{
					Id:       uuid.NewString(),
					Kind:     assetsService.ASSETKIND_PVC,
					TenantId: *assetsService.NewNullableInt32(&tenantId),
					Scope:    assetsService.SCOPE_TENANT,
				},
				Spec: assetsService.DatasourceListResponseAssetSpec{
					Pvc: *assetsService.NewNullablePVCAssetSpec(&assetsService.PVCAssetSpec{
						ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
						ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w")),
						Path:        *assetsService.NewNullableString(convutils.Ref("/correct/path")),
					}),
				},
			}

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   add a workspace policy with single asset only
			//
			addBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Storage: &api.SupersetRulesAllOfStorage{
							Pvc: &api.PvcsRules{
								Attributes: &api.PvcRules{
									Path: &api.StringRules{
										Options: &[]api.StringOption{
											{
												Value: "/correct/path",
											},
										},
									},
								},
							},
						},
					},
					ImposedAssets: &[]string{pvcW.Meta.Id},
				},
			}

			requestAdd, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(addBody))
			Expect(err).To(BeNil())
			requestAdd.Header = *headerForAdmin(tenantId)

			responseAdd := httptest.NewRecorder()
			router.ServeHTTP(responseAdd, requestAdd)
			Expect(responseAdd.Code).To(Equal(http.StatusOK))

			var createdPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(responseAdd.Body.Bytes(), &createdPolicy)
			Expect(err).To(BeNil())

			Expect(createdPolicy.Effective).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage.Pvc).NotTo(BeNil())
			Expect(createdPolicy.Effective.Defaults.Storage.Pvc.Instances).NotTo(BeNil())
			Expect(len(*createdPolicy.Effective.Defaults.Storage.Pvc.Instances)).To(Equal(1))

			pvcInstance := (*createdPolicy.Effective.Defaults.Storage.Pvc.Instances)[0]
			Expect(convutils.Deref(pvcInstance.ClaimName, "")).To(Equal("claim-w"))
			Expect(convutils.Deref(pvcInstance.Name, "")).To(Equal(pvcW.Meta.Id))

			Expect(createdPolicy.Effective.ImposedAssets).NotTo(BeNil())
			Expect(len(*createdPolicy.Effective.ImposedAssets)).To(Equal(1))
			Expect((*createdPolicy.Effective.ImposedAssets)[0]).To(Equal(pvcW.Meta.Id))

			//
			//   change the path of the pvc in the asset
			//
			pvcW.Spec.Pvc.Set(&assetsService.PVCAssetSpec{
				ExistingPvc: *assetsService.NewNullableBool(convutils.Ref(true)),
				ClaimName:   *assetsService.NewNullableString(convutils.Ref("claim-w-modified")),
				Path:        *assetsService.NewNullableString(convutils.Ref("/incorrect/path")),
			})

			assets.SetMock(assets.SimpleMockClientFactory{
				SimpleMockClient: assets.NewSimpleMockClient().WithAsset(pvcW),
			})

			//
			//   simulate update event of the pvc asset
			//
			updatedEvent := api.AssetChangeNotification{
				ChangeType: api.ASSETCHANGETYPE_UPDATED,
				ScopeType:  api.SCOPETYPE_TENANT,
				ScopeId:    convutils.Ref(fmt.Sprintf("%d", tenantId)),
			}

			requestUpdate, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", urlAssetHook, pvcW.Meta.Id), makePayload(updatedEvent))
			Expect(err).To(BeNil())
			requestUpdate.Header = *headerForAdmin(tenantId)

			responseUpdate := httptest.NewRecorder()
			router.ServeHTTP(responseUpdate, requestUpdate)
			Expect(responseUpdate.Code).To(Equal(http.StatusBadRequest))
			Expect(responseUpdate.Body.String()).To(ContainSubstring("update conflicts Workspace policy"))
			Expect(responseUpdate.Body.String()).To(ContainSubstring("storage.pvc.attributes.path"))
			Expect(responseUpdate.Body.String()).To(ContainSubstring("the value of this field must be one of: /correct/path"))
		})
	})

	/////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Behavior
	/////////////////////////////////////////////////////////////////////////////////////////////////////////

	Context("effective derivation", func() {
		It("should calculate correctly", func() {
			//
			//  Part 1: Tenant with two env vars + two capabilities and rules on gpu portion
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
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

						Image: convutils.Ref("image"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							Capabilities: &[]api.Capability{
								api.CAPABILITY_CHOWN,
								api.CAPABILITY_KILL,
							},
						},
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							GpuPortionRequest: &api.NumberRules{
								Max: convutils.Ref(float32(10.0)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			expectedTenantA := api.WorkspacePolicyDefaultsAndRulesV2{
				Rules:    bodyTenantA.Policy.Rules,
				Defaults: bodyTenantA.Policy.Defaults,
			}
			expectedTenantA.Rules.Compute.GpuPortionRequest.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedTenantA.Rules.Compute.GpuMemoryPortionRequest = expectedTenantA.Rules.Compute.GpuPortionRequest
			dao.InjectSystemPolicyForTesting[api.WorkspacePolicyDefaultsAndRulesV2](&expectedTenantA)

			testutils.Expect(*actualRespTenantA.Effective).ToEqual(expectedTenantA)

			//
			//   Part 2 - Department replaces one variable and adds another, plus - add rule for CpuMemory,
			//            plus - provides new set of capabilities. so for rules we have gpu portion -> tenant, cpu memory -> dept
			//
			bodyDeptA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("image"),
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("TENANT-ENV-A"),
									Value: convutils.Ref("tenant-value-a-replaced-by-dept"),
								},
								{
									Name:  convutils.Ref("DEPT-ENV-C"),
									Value: convutils.Ref("dept-c"),
								},
							},
						},
						Security: &api.SupersetDefaultsAllOfSecurity{
							Capabilities: &[]api.Capability{
								api.CAPABILITY_BLOCK_SUSPEND,
							},
						},
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							CpuMemoryLimit: &api.QuantityRules{
								Required: convutils.Ref(true),
							},
						},
					},
				},
			}

			requestDeptA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDeptA))
			Expect(err).To(BeNil())
			requestDeptA.Header = *headerForAdmin(1)

			responseDeptA := httptest.NewRecorder()
			router.ServeHTTP(responseDeptA, requestDeptA)
			Expect(responseDeptA.Code).To(Equal(http.StatusOK))

			var actualRespDeptA api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptA.Body.Bytes(), &actualRespDeptA)
			Expect(err).To(BeNil())

			expectedDeptA := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					EnvironmentVariables: &api.EnvironmentVariablesDefaults{
						Instances: &[]api.EnvironmentVariable{
							{
								Name:  convutils.Ref("TENANT-ENV-A"),
								Value: convutils.Ref("tenant-value-a-replaced-by-dept"),
							},
							{
								Name:  convutils.Ref("TENANT-ENV-B"),
								Value: convutils.Ref("tenant-value-b"),
							},
							{
								Name:  convutils.Ref("DEPT-ENV-C"),
								Value: convutils.Ref("dept-c"),
							},
						},
					},
					Image: convutils.Ref("image"),
					Security: &api.SupersetDefaultsAllOfSecurity{
						Capabilities: &[]api.Capability{
							// capabilities are array, therefore new array replaces the inherited array
							api.CAPABILITY_BLOCK_SUSPEND,
						},
					},
				},
				Rules: &api.WorkspacePolicyRulesV2{
					Compute: &api.SupersetRulesAllOfCompute{
						CpuMemoryLimit: &api.QuantityRules{
							Required: convutils.Ref(true),
						},
						GpuPortionRequest: &api.NumberRules{
							Max: convutils.Ref(float32(10)),
						},
					},
				},
			}
			expectedDeptA.Rules.Compute.GpuPortionRequest.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedDeptA.Rules.Compute.GpuMemoryPortionRequest = expectedDeptA.Rules.Compute.GpuPortionRequest
			expectedDeptA.Rules.Compute.CpuMemoryLimit.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("111")}
			dao.InjectSystemPolicyForTesting[api.WorkspacePolicyDefaultsAndRulesV2](&expectedDeptA)

			Expect(*actualRespDeptA.Effective).To(Equal(expectedDeptA))

			//
			//  Part 3: Remove gpu request rule, environment variables and capability defaults form the tenant policy
			//       so for tenant we do not have any rules now, department remains with cpu memory
			//
			bodyTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Reset: []string{
					"compute.gpuPortionRequest",
					"environmentVariables",
					"security.capabilities",
				},
			}

			requestTenantB, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantB))
			Expect(err).To(BeNil())
			requestTenantB.Header = *headerForAdmin(1)

			responseTenantB := httptest.NewRecorder()
			router.ServeHTTP(responseTenantB, requestTenantB)
			Expect(responseTenantB.Code).To(Equal(http.StatusOK))

			var actualRespTenantB api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantB.Body.Bytes(), &actualRespTenantB)
			Expect(err).To(BeNil())

			expectedEffective := api.WorkspacePolicyDefaultsAndRulesV2{
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("image"),
				},
				Rules: nil,
			}
			dao.InjectSystemPolicyForTesting(&expectedEffective)
			testutils.Expect(*actualRespTenantB.Effective).ToEqual(expectedEffective)

			//
			//  Part 4: Apply to department in order to re-calculate the effective from the tenant
			//      as for rules -> we only left with department rule for cpu limit
			//
			bodyDeptB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
			}

			requestDeptB, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDeptB))
			Expect(err).To(BeNil())
			requestDeptB.Header = *headerForAdmin(1)

			responseDeptB := httptest.NewRecorder()
			router.ServeHTTP(responseDeptB, requestDeptB)
			Expect(responseDeptB.Code).To(Equal(http.StatusOK))

			var actualRespDeptB api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptB.Body.Bytes(), &actualRespDeptB)
			Expect(err).To(BeNil())

			expectedRespDeptB := *actualRespDeptB.Policy
			expectedRespDeptB.Rules.Compute.CpuMemoryLimit.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_DEPARTMENT, DepartmentId: convutils.Ref("111")}
			dao.InjectSystemPolicyForTesting(&expectedRespDeptB)

			Expect(*actualRespDeptB.Effective).To(Equal(expectedRespDeptB))
		})
	})

	Context("two levels derivation", func() {
		It("should calculate correctly", func() {
			//
			//  Level 1: Tenant
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffective, &actualRespTenantA.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespTenantA.Effective).To(Equal(expectedEffective))

			//
			//   Level 2 - departments
			//
			bodyDeptA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Compute: &api.SupersetDefaultsAllOfCompute{
							CpuMemoryLimit: convutils.Ref("1G"),
						},
					},
				},
			}

			requestDeptA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDeptA))
			Expect(err).To(BeNil())
			requestDeptA.Header = *headerForAdmin(1)

			responseDeptA := httptest.NewRecorder()
			router.ServeHTTP(responseDeptA, requestDeptA)
			Expect(responseDeptA.Code).To(Equal(http.StatusOK))

			var actualRespDeptA api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptA.Body.Bytes(), &actualRespDeptA)
			Expect(err).To(BeNil())

			expectedEffective = api.WorkspacePolicyDefaultsAndRulesV2{
				Rules: nil,
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("ubuntu"),
					Compute: &api.SupersetDefaultsAllOfCompute{
						CpuMemoryLimit: convutils.Ref("1G"),
					},
				},
			}
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespDeptA.Effective).To(Equal(expectedEffective))

			//
			//  Level 3 - project
			//
			bodyProjA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							HostIpc: convutils.Ref(true),
						},
					},
				},
			}

			requestProjA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProjA))
			Expect(err).To(BeNil())
			requestProjA.Header = *headerForAdmin(1)

			responseProjA := httptest.NewRecorder()
			router.ServeHTTP(responseProjA, requestProjA)
			Expect(responseProjA.Code).To(Equal(http.StatusOK))

			var actualRespProjA api.WorkspacePolicyV2
			err = json.Unmarshal(responseProjA.Body.Bytes(), &actualRespProjA)
			Expect(err).To(BeNil())

			expectedEffective = api.WorkspacePolicyDefaultsAndRulesV2{
				Rules: nil,
				Defaults: &api.WorkspacePolicyDefaultsV2{
					Image: convutils.Ref("ubuntu"),
					Security: &api.SupersetDefaultsAllOfSecurity{
						HostIpc: convutils.Ref(true),
					},
					Compute: &api.SupersetDefaultsAllOfCompute{
						CpuMemoryLimit: convutils.Ref("1G"),
					},
				},
			}
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			Expect(*actualRespProjA.Effective).To(Equal(expectedEffective))
		})
	})

	Context("rule for same field given in two scopes", func() {
		It("should calculate correctly", func() {
			//
			//  Level 1: Tenant
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image:      convutils.Ref("image-tenant-opt1"),
						WorkingDir: convutils.Ref("workdir-tenant-opt1"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
							Options: &[]api.StringOption{
								{Value: "image-tenant-opt1"},
								{Value: "image-tenant-opt2"},
							},
						},
						WorkingDir: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "workdir-tenant-opt1"},
								{Value: "workdir-tenant-opt2"},
								{Value: "workdir-dept-opt1"}, // we need it because these rules apply to the department as well
							},
						},
						CreateHomeDir: &api.BooleanRules{
							Required: convutils.Ref(true),
						},

						Storage: &api.SupersetRulesAllOfStorage{
							Git: &api.GitsRules{
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(false),
								},
								Attributes: &api.GitRules{
									Repository: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "repo1-from-tenant"},
											{Value: "repo2-from-tenant"},
										},
									},
									Branch: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "branch1-from-tenant"},
											{Value: "branch2-from-tenant"},
										},
									},
								},
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			sourceTenant := api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}
			sourceDept11 := api.SourceOfRule{
				Scope:        api.SCOPETYPE_DEPARTMENT,
				DepartmentId: convutils.Ref("111"),
			}
			sourceProject111 := api.SourceOfRule{
				Scope:     api.SCOPETYPE_PROJECT,
				ProjectId: convutils.Ref(int32(1111)),
			}

			expectedRules := api.WorkspacePolicyRulesV2{
				Image:         bodyTenantA.Policy.Rules.Image,
				WorkingDir:    bodyTenantA.Policy.Rules.WorkingDir,
				CreateHomeDir: bodyTenantA.Policy.Rules.CreateHomeDir,
				Storage: &api.SupersetRulesAllOfStorage{
					Git: &api.GitsRules{
						Instances: bodyTenantA.Policy.Rules.Storage.Git.Instances,
						Attributes: &api.GitRules{
							Repository: bodyTenantA.Policy.Rules.Storage.Git.Attributes.Repository,
							Branch:     bodyTenantA.Policy.Rules.Storage.Git.Attributes.Branch,
						},
					},
				},
			}

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			expectedRules.Image.SourceOfRule = &sourceTenant
			expectedRules.WorkingDir.SourceOfRule = &sourceTenant
			expectedRules.CreateHomeDir.SourceOfRule = &sourceTenant
			expectedRules.Storage.Git.Instances.SourceOfRule = &sourceTenant
			expectedRules.Storage.Git.Attributes.Repository.SourceOfRule = &sourceTenant
			expectedRules.Storage.Git.Attributes.Branch.SourceOfRule = &sourceTenant
			expectedEffective.Rules = &expectedRules

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespTenantA.Effective.Rules).ToEqual(expectedRules)

			//
			//   Level 2 - departments
			//
			bodyDeptA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image:      convutils.Ref("image-dept-opt1"),
						WorkingDir: convutils.Ref("workdir-dept-opt1"),
					},
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "image-dept-opt1"},
								{Value: "image-dept-opt2"},
							},
						},
						BackoffLimit: &api.IntegerRules{
							Max: convutils.Ref(int32(4)),
						},
						Storage: &api.SupersetRulesAllOfStorage{
							Git: &api.GitsRules{
								Instances: &api.ItemRules{
									CanAdd: convutils.Ref(true),
								},
								Attributes: &api.GitRules{
									Repository: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "repo1-from-dept"},
											{Value: "repo2-from-dept"},
										},
									},
									Path: &api.StringRules{
										Options: &[]api.StringOption{
											{Value: "path1-from-dept"},
											{Value: "path2-from-dept"},
										},
									},
								},
							},
						},
					},
				},
			}

			requestDeptA, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(bodyDeptA))
			Expect(err).To(BeNil())
			requestDeptA.Header = *headerForAdmin(1)

			responseDeptA := httptest.NewRecorder()
			router.ServeHTTP(responseDeptA, requestDeptA)
			Expect(responseDeptA.Code).To(Equal(http.StatusOK))

			var actualRespDeptA api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptA.Body.Bytes(), &actualRespDeptA)
			Expect(err).To(BeNil())

			expectedRules = api.WorkspacePolicyRulesV2{
				Image:         bodyDeptA.Policy.Rules.Image,
				WorkingDir:    bodyTenantA.Policy.Rules.WorkingDir,
				CreateHomeDir: bodyTenantA.Policy.Rules.CreateHomeDir,
				BackoffLimit:  bodyDeptA.Policy.Rules.BackoffLimit,

				Storage: &api.SupersetRulesAllOfStorage{
					Git: &api.GitsRules{
						Instances: bodyDeptA.Policy.Rules.Storage.Git.Instances,
						Attributes: &api.GitRules{
							Repository: bodyDeptA.Policy.Rules.Storage.Git.Attributes.Repository,
							Path:       bodyDeptA.Policy.Rules.Storage.Git.Attributes.Path,
							Branch:     bodyTenantA.Policy.Rules.Storage.Git.Attributes.Branch,
						},
					},
				},
			}

			expectedRules.Image.SourceOfRule = &sourceDept11
			expectedRules.WorkingDir.SourceOfRule = &sourceTenant
			expectedRules.CreateHomeDir.SourceOfRule = &sourceTenant
			expectedRules.BackoffLimit.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Instances.SourceOfRule = &sourceTenant
			expectedRules.Storage.Git.Instances.CanAdd = convutils.Ref(false)
			expectedRules.Storage.Git.Attributes.Repository.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Attributes.Path.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Attributes.Branch.SourceOfRule = &sourceTenant

			expectedEffective.Rules = &expectedRules
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespDeptA.Effective.Rules).ToEqual(expectedEffective.Rules)

			//
			//  Level 3 - project
			//
			bodyProjA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "image-proj-opt1"},
								{Value: "image-proj-opt2"},
								{Value: "image-dept-opt1"}, // because we inherit the value from the department
							},
						},
						WorkingDir: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "workdir-dept-opt1"}, // we inherit this value from the department
								{Value: "workdir-proj-opt1"},
								{Value: "workdir-proj-opt2"},
							},
						},
					},
				},
			}

			requestProjA, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(bodyProjA))
			Expect(err).To(BeNil())
			requestProjA.Header = *headerForAdmin(1)

			responseProjA := httptest.NewRecorder()
			router.ServeHTTP(responseProjA, requestProjA)
			Expect(responseProjA.Code).To(Equal(http.StatusOK))

			var actualRespProjA api.WorkspacePolicyV2
			err = json.Unmarshal(responseProjA.Body.Bytes(), &actualRespProjA)
			Expect(err).To(BeNil())

			expectedRules = api.WorkspacePolicyRulesV2{
				Image:         bodyProjA.Policy.Rules.Image,
				WorkingDir:    bodyProjA.Policy.Rules.WorkingDir,
				CreateHomeDir: bodyTenantA.Policy.Rules.CreateHomeDir,
				BackoffLimit:  bodyDeptA.Policy.Rules.BackoffLimit,

				Storage: &api.SupersetRulesAllOfStorage{
					Git: &api.GitsRules{
						Instances: bodyDeptA.Policy.Rules.Storage.Git.Instances,
						Attributes: &api.GitRules{
							Repository: bodyDeptA.Policy.Rules.Storage.Git.Attributes.Repository,
							Path:       bodyDeptA.Policy.Rules.Storage.Git.Attributes.Path,
							Branch:     bodyTenantA.Policy.Rules.Storage.Git.Attributes.Branch,
						},
					},
				},
			}

			expectedRules.Image.SourceOfRule = &sourceProject111
			expectedRules.WorkingDir.SourceOfRule = &sourceProject111
			expectedRules.CreateHomeDir.SourceOfRule = &sourceTenant
			expectedRules.BackoffLimit.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Instances.SourceOfRule = &sourceTenant
			expectedRules.Storage.Git.Attributes.Repository.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Attributes.Path.SourceOfRule = &sourceDept11
			expectedRules.Storage.Git.Attributes.Branch.SourceOfRule = &sourceTenant

			expectedEffective.Rules = &expectedRules
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespProjA.Effective.Rules).ToEqual(expectedEffective.Rules)
		})
	})

	Context("update only part of the fields", func() {
		It("should keep the other fields unchanged", func() {
			//
			//  Level 1: give defaults to three fields
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
						Ports: &api.PortsDefaults{
							Attributes: &api.Port{
								Container: convutils.Ref(int32(8080)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			expectedTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: nil,
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
						Ports: &api.PortsDefaults{
							Attributes: &api.Port{
								Container: convutils.Ref(int32(8080)),
							},
						},
					},
				},
			}

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffective, &expectedTenantA.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespTenantA.Policy).ToEqual(expectedTenantA.Policy)
			testutils.Expect(*actualRespTenantA.Effective).ToEqual(expectedEffective)

			//
			//  Level 2: update only one field
			//
			bodyTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(40)),
						},
					},
				},
			}

			requestTenantB, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantB))
			Expect(err).To(BeNil())
			requestTenantB.Header = *headerForAdmin(1)

			responseTenantB := httptest.NewRecorder()
			router.ServeHTTP(responseTenantB, requestTenantB)
			Expect(responseTenantB.Code).To(Equal(http.StatusOK))

			var actualRespTenantB api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantB.Body.Bytes(), &actualRespTenantB)
			Expect(err).To(BeNil())

			expectedTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: nil,
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(40)),
						},
						Ports: &api.PortsDefaults{
							Attributes: &api.Port{
								Container: convutils.Ref(int32(8080)),
							},
						},
					},
				},
			}

			copyErr = convutils.CopySpecs(&expectedEffective, &expectedTenantB.Policy)
			Expect(copyErr).To(BeNil())
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespTenantB.Policy).ToEqual(*expectedTenantB.Policy)
			testutils.Expect(*actualRespTenantB.Effective).ToEqual(expectedEffective)
		})
	})

	Context("canEdit no default value", func() {
		It("should reject the policy", func() {
			projectID := int32(1111)
			policyName := "pol-t1"
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
						},
					},
				},
			}

			// Apply the policy
			errText := expectError(router, urlPolicyTraining, makePayload(addBody), 1, http.MethodPut, http.StatusBadRequest)
			Expect(errText).To(ContainSubstring("default value must be provided for policy fields having canEdit"))
		})
	})

	Context("canEdit no default value - attributes", func() {
		It("should reject the policy", func() {
			projectID := int32(1111)
			policyName := "pol-t1"
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										CanEdit: convutils.Ref(false),
									},
								},
							},
						},
					},
				},
			}

			// Apply the policy
			errText := expectError(router, urlPolicyTraining, makePayload(addBody), 1, http.MethodPut, http.StatusBadRequest)
			Expect(errText).To(ContainSubstring("default value must be provided for policy fields having canEdit"))
		})
	})

	Context("canEdit with defaults", func() {
		It("should reject the policy", func() {
			projectID := int32(1111)
			policyName := "pol-t1"
			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					Name:      policyName,
					ProjectId: &projectID,
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Compute: &api.SupersetDefaultsAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesDefaults{
								Attributes: &api.ExtendedResource{
									Quantity: convutils.Ref("1G"),
								},
							},
						},
					},
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							CanEdit: convutils.Ref(false),
						},
						Compute: &api.SupersetRulesAllOfCompute{
							ExtendedResources: &api.ExtendedResourcesRules{
								Attributes: &api.ExtendedResourceRules{
									Quantity: &api.StringRules{
										CanEdit: convutils.Ref(false),
									},
								},
							},
						},
					},
				},
			}

			// Apply the policy
			overwriteTrainingPolicyForTest(router, &addBody, 1)
		})
	})

	Context("policy update should update updated_by field", func() {
		It("should update updated_by when policy is modified by different user", func() {
			// Test the database update function directly
			db := runaidb.DB()

			// Create a test policy in the database
			testPolicy := dbmodel.Policy{
				Name:          "test-policy-updated-by",
				WorkloadType:  string(api.POLICYTYPE_WORKSPACE),
				Scope:         api.SCOPETYPE_TENANT,
				TenantId:      1,
				DefaultsV2:    map[string]interface{}{"image": "ubuntu"},
				RulesV2:       map[string]interface{}{},
				ImposedAssets: nil,
			}

			// Set the BaseModel fields manually
			testPolicy.CreatedBy = "user1"
			testPolicy.UpdatedBy = "user1"
			testPolicy.CreatedAt = time.Now()
			testPolicy.UpdatedAt = time.Now()

			// Insert the policy
			_, err := db.Create(&testPolicy)
			Expect(err).To(BeNil())
			Expect(testPolicy.ID).ToNot(BeZero())

			// Now update the policy with a different user
			testPolicy.UpdatedBy = "user2"
			testPolicy.UpdatedAt = time.Now()
			testPolicy.DefaultsV2 = map[string]interface{}{"image": "alpine"}

			// Call the UpdatePolicySpec function directly
			err = dao.UpdatePolicySpec(context.Background(), db, &testPolicy)
			Expect(err).To(BeNil())

			// Verify the database was updated correctly
			var updatedPolicy dbmodel.Policy
			err = db.Where("id = ?", testPolicy.ID).First(&updatedPolicy)
			Expect(err).To(BeNil())

			// Verify that updated_by was changed to user 2, but created_by remains user 1
			Expect(updatedPolicy.UpdatedBy).To(Equal("user2"))
			Expect(updatedPolicy.CreatedBy).To(Equal("user1"))
			Expect(updatedPolicy.DefaultsV2["image"]).To(Equal("alpine"))

			// Clean up
			_, err = db.Delete(&testPolicy)
			Expect(err).To(BeNil())
		})
	})

	Context("placeholder policy creator fix", func() {
		It("should update creator when applying policy to empty placeholder policy", func() {
			// Step 1: Creating an empty placeholder policy (GET request would cause an empty policy to be created if it doesn't exist)
			getReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			getReq.Header = *headerForAdmin(1)
			getResp := httptest.NewRecorder()
			router.ServeHTTP(getResp, getReq)
			Expect(getResp.Code).To(Equal(http.StatusOK))
			var placeholderPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(getResp.Body.Bytes(), &placeholderPolicy)
			Expect(err).To(BeNil())

			// Verify it's a placeholder - Policy is not nil but an empty struct
			Expect(placeholderPolicy.Policy).NotTo(BeNil())
			Expect(placeholderPolicy.Policy.Rules).To(BeNil())
			Expect(placeholderPolicy.Policy.Defaults).To(BeNil())
			Expect(placeholderPolicy.Policy.ImposedAssets).To(BeNil())
			Expect(placeholderPolicy.Meta.CreatedBy).To(Equal(tree.EmptyPolicyCreator))

			placeholderCreatedAt := placeholderPolicy.Meta.CreatedAt

			// Step 2: Now apply an actual policy with real content
			applyBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "real-policy",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			requestApply, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(applyBody))
			Expect(err).To(BeNil())
			requestApply.Header = *headerForAdmin(1)

			responseApply := httptest.NewRecorder()
			router.ServeHTTP(responseApply, requestApply.WithContext(tenantContextForTest(1)))
			Expect(responseApply.Code).To(Equal(http.StatusOK), responseApply.Body.String())

			// Step 3: Verify the policy creator was updated from EmptyPolicyCreator to the actual user

			var updatedPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(responseApply.Body.Bytes(), &updatedPolicy)
			Expect(err).To(BeNil())

			// Verify creator was changed from EmptyPolicyCreator to actual user
			Expect(updatedPolicy.Meta.CreatedBy).To(Equal("admin"))
			Expect(updatedPolicy.Meta.UpdatedBy).To(Equal("admin"))
			// Verify timestamps were updated
			Expect(updatedPolicy.Meta.CreatedAt).To(BeTemporally(">", placeholderCreatedAt))
		})

		It("should update creator when applying policy to a deleted policy", func() {
			// Step 1: Create a policy and then delete it
			applyBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "policy-to-delete",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			requestApply, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(applyBody))
			Expect(err).To(BeNil())
			requestApply.Header = *headerForAdmin(1)

			responseApply := httptest.NewRecorder()
			router.ServeHTTP(responseApply, requestApply.WithContext(tenantContextForTest(1)))
			Expect(responseApply.Code).To(Equal(http.StatusOK))

			// Extract policy metadata from response
			var applyResponse api.WorkspacePolicyV2
			err = json.Unmarshal(responseApply.Body.Bytes(), &applyResponse)
			Expect(err).To(BeNil())
			originalCreatedAt := applyResponse.Meta.CreatedAt

			// Step 2: Delete the policy
			requestDelete, err := http.NewRequest(http.MethodDelete, urlPolicyWorkspace+"?scope=tenant", nil)
			Expect(err).To(BeNil())
			requestDelete.Header = *headerForAdmin(1)

			responseDelete := httptest.NewRecorder()
			router.ServeHTTP(responseDelete, requestDelete.WithContext(tenantContextForTest(1)))
			Expect(responseDelete.Code).To(Equal(http.StatusNoContent))

			// Step 3: Verify the policy is deleted by trying to GET it
			getReqDeleted, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			getReqDeleted.Header = *headerForAdmin(1)
			getRespDeleted := httptest.NewRecorder()
			router.ServeHTTP(getRespDeleted, getReqDeleted)
			Expect(getRespDeleted.Code).To(Equal(http.StatusOK))

			var deletedPolicyResponse api.WorkspacePolicyV2
			err = json.Unmarshal(getRespDeleted.Body.Bytes(), &deletedPolicyResponse)
			Expect(err).To(BeNil())
			// After deletion, it should return an empty placeholder
			Expect(deletedPolicyResponse.Policy).NotTo(BeNil())
			Expect(deletedPolicyResponse.Policy.Rules).To(BeNil())
			Expect(deletedPolicyResponse.Policy.Defaults).To(BeNil())

			// Step 4: Re-create the policy (same tenant, since deleted policy keeps the tenant)
			recreateBody := api.WorkspacePolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "recreated-policy",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			requestRecreate, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(recreateBody))
			Expect(err).To(BeNil())
			// Use same tenant but verify creator is updated
			requestRecreate.Header = *headerForAdmin(1)

			responseRecreate := httptest.NewRecorder()
			router.ServeHTTP(responseRecreate, requestRecreate.WithContext(tenantContextForTest(1)))
			Expect(responseRecreate.Code).To(Equal(http.StatusOK), responseRecreate.Body.String())

			// Step 5: Verify the creator was updated via the response
			var recreatedPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(responseRecreate.Body.Bytes(), &recreatedPolicy)
			Expect(err).To(BeNil())

			// Verify the policy now has content (not a placeholder)
			Expect(recreatedPolicy.Policy).NotTo(BeNil())
			Expect(recreatedPolicy.Policy.Defaults).NotTo(BeNil())
			Expect(recreatedPolicy.Policy.Defaults.Image).To(Equal(convutils.Ref("ubuntu")))
			// Verify creator was updated (from the previous creator to admin)
			Expect(recreatedPolicy.Meta.CreatedBy).To(Equal("admin"))
			Expect(recreatedPolicy.Meta.UpdatedBy).To(Equal("admin"))
			// Verify timestamps were updated (since it was deleted, it's treated as new)
			Expect(recreatedPolicy.Meta.CreatedAt).To(BeTemporally(">", originalCreatedAt))
		})

		It("should NOT update creator when applying to existing non-placeholder policy", func() {
			// Helper function to create headers with a custom subject
			headerForUser := func(tenantId int32, subjectId string) *http.Header {
				ctx := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantId))
				ctx = context.WithValue(ctx, runaictx.CfSubjectID, subjectId)
				return runaictx_http.GetHeaders(ctx)
			}

			// Step 1: Create a policy with user "user1"
			applyBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "existing-policy",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
				},
			}

			requestApply, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(applyBody))
			Expect(err).To(BeNil())
			requestApply.Header = *headerForUser(1, "user1")

			responseApply := httptest.NewRecorder()
			router.ServeHTTP(responseApply, requestApply.WithContext(tenantContextForTest(1)))
			Expect(responseApply.Code).To(Equal(http.StatusOK))

			// Step 2: Get the policy via URL to capture original metadata
			var applyResponse api.WorkspacePolicyV2
			err = json.Unmarshal(responseApply.Body.Bytes(), &applyResponse)
			Expect(err).To(BeNil())
			Expect(applyResponse.Meta.CreatedBy).To(Equal("user1"))
			Expect(applyResponse.Policy.Rules).NotTo(BeNil())
			Expect(applyResponse.Policy.Rules.Image).NotTo(BeNil())
			originalCreatedAt := applyResponse.Meta.CreatedAt

			// Step 3: Update the same policy with user2 - creator and createdAt should remain unchanged
			updateBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "updated-policy",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("alpine"),
					},
				},
			}

			requestUpdate, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(updateBody))
			Expect(err).To(BeNil())
			requestUpdate.Header = *headerForUser(1, "user2")

			responseUpdate := httptest.NewRecorder()
			router.ServeHTTP(responseUpdate, requestUpdate.WithContext(tenantContextForTest(1)))
			Expect(responseUpdate.Code).To(Equal(http.StatusOK))

			// Step 4: Verify creator and createdAt were NOT changed via the response
			var updatedPolicy api.WorkspacePolicyV2
			err = json.Unmarshal(responseUpdate.Body.Bytes(), &updatedPolicy)
			Expect(err).To(BeNil())

			// Creator should remain the original user (user1)
			Expect(updatedPolicy.Meta.CreatedBy).To(Equal("user1"), "CreatedBy should remain as original creator")
			// UpdatedBy should be the new user (user2)
			Expect(updatedPolicy.Meta.UpdatedBy).To(Equal("user2"), "UpdatedBy should be the updating user")
			// CreatedAt should remain the same (this is the key test - non-placeholder policies preserve createdAt)
			// Use BeTemporally with ~0 to allow for timestamp precision differences between Go and DB
			Expect(updatedPolicy.Meta.CreatedAt).To(BeTemporally("~", originalCreatedAt, time.Millisecond), "CreatedAt should not change for non-placeholder policy")
		})
	})

	Context("reset fields that have only values or only rules", func() {
		It("should clean those fields", func() {
			//
			//  Level 1: give defaults to two fields and rules to other fields
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								Min: convutils.Ref(int32(1)),
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsGid: convutils.Ref(int64(30)),
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			Expect(*actualRespTenantA.Policy).To(Equal(*bodyTenantA.Policy))

			expectedTenantA := *bodyTenantA.Policy
			expectedTenantA.Rules.Image.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			expectedTenantA.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}

			dao.InjectSystemPolicyForTesting(&expectedTenantA)

			Expect(*actualRespTenantA.Effective).To(Equal(expectedTenantA))

			//
			//  Level 2: reset part of the fields
			//
			bodyTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Reset: []string{
					"security.runAsUid",
					"security.runAsGid",
				},
			}

			requestTenantB, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantB))
			Expect(err).To(BeNil())
			requestTenantB.Header = *headerForAdmin(1)

			responseTenantB := httptest.NewRecorder()
			router.ServeHTTP(responseTenantB, requestTenantB)
			Expect(responseTenantB.Code).To(Equal(http.StatusOK))

			var actualRespTenantB api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantB.Body.Bytes(), &actualRespTenantB)
			Expect(err).To(BeNil())

			expectedTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Image: &api.StringRules{
							Required: convutils.Ref(true),
						},
					},
					Defaults: nil,
				},
			}

			Expect(*actualRespTenantB.Policy).To(Equal(*expectedTenantB.Policy))

			expectedTenantB.Policy.Rules.Image.SourceOfRule = &api.SourceOfRule{Scope: api.SCOPETYPE_TENANT}
			dao.InjectSystemPolicyForTesting(expectedTenantB.Policy)
			testutils.Expect(*actualRespTenantB.Effective).ToEqual(expectedTenantB.Policy)
		})
	})

	Context("change privileged for given tenant", func() {
		It("should change to that tenant", func() {
			//
			//  apply a policy that changes the system policy for tenant 1
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							Privileged: &api.BooleanRules{
								CanEdit: convutils.Ref(true),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			Expect(*actualRespTenantA.Policy).To(Equal(*bodyTenantA.Policy))

			Expect(*actualRespTenantA.Effective.Rules.Security.Privileged).To(Equal(api.BooleanRules{
				CanEdit: convutils.Ref(true),
				SourceOfRule: &api.SourceOfRule{
					Scope: api.SCOPETYPE_TENANT,
				},
			}))

			//
			//  Level 2: try another tenant, should get the privilege from the system policy
			//
			requestTenantB, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace), nil)
			Expect(err).To(BeNil())
			requestTenantB.Header = *headerForAdmin(2)

			responseTenantB := httptest.NewRecorder()
			router.ServeHTTP(responseTenantB, requestTenantB)
			Expect(responseTenantB.Code).To(Equal(http.StatusOK))

			var actualRespTenantB api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantB.Body.Bytes(), &actualRespTenantB)
			Expect(err).To(BeNil())

			Expect(*actualRespTenantB.Effective.Rules.Security.Privileged).To(Equal(api.BooleanRules{
				CanEdit: convutils.Ref(false),
				SourceOfRule: &api.SourceOfRule{
					Scope: api.SCOPETYPE_SYSTEM,
				},
			}))
		})
	})

	Context("multi level defaults", func() {
		It("effective calculated properly", func() {
			//
			//  Level 3 - project, defaults for runAsUid, RunAsGid, image, gpuPortionRequest and gpuMemoryRequest
			//
			bodyProjA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_PROJECT,
					ProjectId: convutils.Ref(int32(1111)),
					Name:      "pol-p-1111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							CpuCoreRequest: &api.NumberRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("332"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							// RunAsUid: convutils.Ref(330),
							RunAsGid: convutils.Ref(int64(331)),
						},
						Compute: &api.SupersetDefaultsAllOfCompute{
							CpuCoreRequest:   convutils.Ref(3.3),
							CpuMemoryRequest: convutils.Ref("3M"),
						},
					},
				},
			}

			requestProjA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyProjA))
			Expect(err).To(BeNil())
			requestProjA.Header = *headerForAdmin(1)

			responseProjA := httptest.NewRecorder()
			router.ServeHTTP(responseProjA, requestProjA)
			Expect(responseProjA.Code).To(Equal(http.StatusOK))

			var actualRespProjA api.WorkspacePolicyV2
			err = json.Unmarshal(responseProjA.Body.Bytes(), &actualRespProjA)
			Expect(err).To(BeNil())

			var expectedEffective api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffective, &bodyProjA.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespProjA.Effective.Defaults).ToEqual(expectedEffective.Defaults)

			//
			//   Level 2 - departments
			//
			bodyDeptA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:        api.SCOPETYPE_DEPARTMENT,
					DepartmentId: convutils.Ref("111"),
					Name:         "pol-d-111",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: nil,
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("222"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsGid: convutils.Ref(int64(221)),
						},
						Compute: &api.SupersetDefaultsAllOfCompute{
							CpuCoreRequest:   convutils.Ref(2.2), // taken from here
							CpuMemoryRequest: convutils.Ref("2M"),
						},
					},
				},
			}

			requestDeptA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyDeptA))
			Expect(err).To(BeNil())
			requestDeptA.Header = *headerForAdmin(1)

			responseDeptA := httptest.NewRecorder()
			router.ServeHTTP(responseDeptA, requestDeptA)
			Expect(responseDeptA.Code).To(Equal(http.StatusOK))

			var actualRespDeptA api.WorkspacePolicyV2
			err = json.Unmarshal(responseDeptA.Body.Bytes(), &actualRespDeptA)
			Expect(err).To(BeNil())

			copyErr = convutils.CopySpecs(&expectedEffective, &bodyDeptA.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespDeptA.Effective.Defaults).ToEqual(expectedEffective.Defaults)

			//
			//  Level 1: Tenant
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},

				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: &api.WorkspacePolicyRulesV2{
						Security: &api.SupersetRulesAllOfSecurity{
							RunAsUid: &api.IntegerRulesOptional{
								CanEdit: convutils.Ref(false),
							},
						},
						Compute: &api.SupersetRulesAllOfCompute{
							CpuCoreRequest: &api.NumberRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},

					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("112"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(110)),
							RunAsGid: convutils.Ref(int64(111)),
						},
						Compute: &api.SupersetDefaultsAllOfCompute{
							CpuCoreRequest:   convutils.Ref(1.1),
							CpuMemoryRequest: convutils.Ref("1M"),
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			copyErr = convutils.CopySpecs(&expectedEffective, &bodyTenantA.Policy)
			Expect(copyErr).To(BeNil())
			dao.InjectSystemPolicyForTesting(&expectedEffective)

			testutils.Expect(*actualRespTenantA.Effective.Defaults).ToEqual(expectedEffective.Defaults)

			//
			//  get the current policy of all three levels
			//  here is what we expect
			//               tenant    department    project
			//   memory      1M        2M            3M
			//   cores       1.1 ---->    ---------> 1.1 (canEdit)
			//   image       112       222           332
			//   runAsUid    110 ------>  ---------->
			//   runAsGid    111       221           331
			//
			finalPolicy := make([]api.WorkspacePolicyV2, 0)
			for _, getParam := range []string{"scope=tenant", "scope=department&departmentId=111", "scope=project&projectId=1111"} {
				requestGet, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", urlPolicyWorkspace, getParam), nil)
				Expect(err).To(BeNil())
				requestGet.Header = *headerForAdmin(1)

				responseGet := httptest.NewRecorder()
				router.ServeHTTP(responseGet, requestGet)
				Expect(responseGet.Code).To(Equal(http.StatusOK))

				var actualRespGet api.WorkspacePolicyV2
				err = json.Unmarshal(responseGet.Body.Bytes(), &actualRespGet)
				Expect(err).To(BeNil())

				finalPolicy = append(finalPolicy, actualRespGet)
			}

			var expectedEffectiveTenant api.WorkspacePolicyDefaultsAndRulesV2
			copyErr = convutils.CopySpecs(&expectedEffectiveTenant, actualRespTenantA.Policy)
			Expect(copyErr).To(BeNil())
			expectedEffectiveTenant.Rules.Compute.CpuCoreRequest.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}
			expectedEffectiveTenant.Rules.Security.RunAsUid.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}
			dao.InjectSystemPolicyForTesting(&expectedEffectiveTenant)

			// no change in given/effective policies of the tenant
			testutils.Expect(*finalPolicy[0].Policy).ToEqual(*actualRespTenantA.Policy)
			testutils.Expect(*finalPolicy[0].Effective).ToEqual(expectedEffectiveTenant)

			// department
			Expect(*finalPolicy[1].Policy).To(Equal(*actualRespDeptA.Policy))

			actualRespDeptA.Effective.Defaults.Security.RunAsUid = actualRespTenantA.Effective.Defaults.Security.RunAsUid
			actualRespDeptA.Effective.Defaults.Compute.CpuCoreRequest = actualRespTenantA.Effective.Defaults.Compute.CpuCoreRequest

			dao.InjectSystemPolicyForTesting(actualRespTenantA.Effective)

			testutils.Expect(*finalPolicy[1].Effective.Defaults).ToEqual(actualRespDeptA.Effective.Defaults)

			// project
			Expect(*finalPolicy[2].Policy.Defaults).To(Equal(*actualRespProjA.Policy.Defaults))

			actualRespProjA.Effective.Defaults.Compute.CpuCoreRequest = actualRespTenantA.Effective.Defaults.Compute.CpuCoreRequest
			actualRespProjA.Effective.Defaults.Security.RunAsUid = actualRespTenantA.Effective.Defaults.Security.RunAsUid

			dao.InjectSystemPolicyForTesting(actualRespTenantA.Effective)

			testutils.Expect(*finalPolicy[2].Effective.Defaults).ToEqual(actualRespProjA.Effective.Defaults)
		})
	})

	Context("policy types are distinct", func() {
		It("allows rules for the same field, and it treats each policy separately", func() {
			numPolicies := 0

			workspaceBody := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
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
			numPolicies += 1

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
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						// the code generator minimized the files, so it created only one set for workspace,
						// and it uses them for both... oh well
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(5)),
						},
					},
				},
			}
			numPolicies += 1

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
						},
						Master: &api.ReplicaRulesV2{
							Security: &api.SupersetRulesAllOfSecurity{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
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
			numPolicies += 1

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
			numPolicies += 1

			distributedInferenceBody := api.DistributedInferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-i1",
				},
				Policy: &api.DistributedInferencePolicyDefaultsAndRulesV2{
					Rules: &api.DistributedInferenceRulesV2{
						Leader: &api.DistributedInferenceLeaderWorkerRulesV2{
							Security: &api.DistributedInferenceLeaderWorkerRulesV2Security{
								RunAsUid: &api.IntegerRulesOptional{
									Min: convutils.Ref(int32(1)),
								},
							},
						},
					},
					Defaults: &api.DistributedInferenceDefaultsV2{
						Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{
							// the code generator minimized the files, so it created only one set for workspace,
							// and it uses them for both... oh well
							Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{
								RunAsUid: convutils.Ref(int64(5)),
							},
						},
					},
				},
			}
			numPolicies += 1

			nimServiceBody := api.NimServicePolicyChangeRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-ns1",
				},
				Policy: &api.NimServicePolicyDefaultsAndRules{
					Rules: &api.NimServiceRules{
						Autoscaling: &api.NimServiceRulesAllOfAutoscaling{
							MaxReplicas: &api.IntegerRules{
								Max: convutils.Ref(int32(1)),
							},
						},
					},
					Defaults: &api.NimServiceDefaults{
						Autoscaling: &api.NimServiceDefaultsAllOfAutoscaling{
							MaxReplicas: convutils.Ref(int32(1)),
						},
					},
				},
			}
			numPolicies += 1

			requestWs, err := http.NewRequest(http.MethodPut, urlPolicyWorkspace, makePayload(workspaceBody))
			Expect(err).To(BeNil())
			requestWs.Header = *headerForAdmin(2)

			responseWs := httptest.NewRecorder()
			router.ServeHTTP(responseWs, requestWs)
			Expect(responseWs.Code).To(Equal(http.StatusOK))

			requestTrain, err := http.NewRequest(http.MethodPut, urlPolicyTraining, makePayload(trainingBody))
			Expect(err).To(BeNil())
			requestTrain.Header = *headerForAdmin(2)

			responseTrain := httptest.NewRecorder()
			router.ServeHTTP(responseTrain, requestTrain)
			Expect(responseTrain.Code).To(Equal(http.StatusOK))

			requestDist, err := http.NewRequest(http.MethodPut, urlPolicyDistributed, makePayload(distributedBody))
			Expect(err).To(BeNil())
			requestDist.Header = *headerForAdmin(2)

			responseDist := httptest.NewRecorder()
			router.ServeHTTP(responseDist, requestDist)
			Expect(responseDist.Code).To(Equal(http.StatusOK))

			requestInference, err := http.NewRequest(http.MethodPut, urlPolicyInference, makePayload(inferenceBody))
			Expect(err).To(BeNil())
			requestInference.Header = *headerForAdmin(2)

			responseInference := httptest.NewRecorder()
			router.ServeHTTP(responseInference, requestInference)
			Expect(responseInference.Code).To(Equal(http.StatusOK))

			requestDistributedInference, err := http.NewRequest(http.MethodPut, urlPolicyDistributedInference, makePayload(distributedInferenceBody))
			Expect(err).To(BeNil())
			requestDistributedInference.Header = *headerForAdmin(2)

			responseDistributedInference := httptest.NewRecorder()
			router.ServeHTTP(responseDistributedInference, requestDistributedInference)
			Expect(responseDistributedInference.Code).To(Equal(http.StatusOK))

			requestNimService, err := http.NewRequest(http.MethodPut, urlPolicyNimService, makePayload(nimServiceBody))
			Expect(err).To(BeNil())
			requestNimService.Header = *headerForAdmin(2)

			responseNimService := httptest.NewRecorder()
			router.ServeHTTP(responseNimService, requestNimService)
			Expect(responseNimService.Code).To(Equal(http.StatusOK))

			response := listPoliciesForTest(router, 2, urlPolicy)
			Expect(len(response.Policies)).To(Equal(numPolicies))

			policiesPerType := make(map[api.PolicyType]int)
			for _, policy := range response.Policies {
				policiesPerType[policy.Type] += 1
			}

			// need to have 1 per type
			Expect(len(policiesPerType)).To(Equal(numPolicies))
			for _, count := range policiesPerType {
				Expect(count).To(Equal(1))
			}
		})
	})

	Context("nim policy CRUD", func() {
		It("should maintain the policy properly", func() {
			nimServiceBody := api.NimServicePolicyChangeRequest{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-ns1",
				},
				Policy: &api.NimServicePolicyDefaultsAndRules{
					Rules: &api.NimServiceRules{
						Autoscaling: &api.NimServiceRulesAllOfAutoscaling{
							MaxReplicas: &api.IntegerRules{
								Max: convutils.Ref(int32(1)),
							},
							Metric: &api.StringRules{
								CanEdit: convutils.Ref(false),
							},
						},
					},
					Defaults: &api.NimServiceDefaults{
						Autoscaling: &api.NimServiceDefaultsAllOfAutoscaling{
							MaxReplicas: convutils.Ref(int32(1)),
							Metric:      convutils.Ref("my_metric"),
						},
					},
				},
			}

			requestNimService, err := http.NewRequest(http.MethodPut, urlPolicyNimService, makePayload(nimServiceBody))
			Expect(err).To(BeNil())
			requestNimService.Header = *headerForAdmin(2)

			responseNimService := httptest.NewRecorder()
			router.ServeHTTP(responseNimService, requestNimService)
			Expect(responseNimService.Code).To(Equal(http.StatusOK))

			// add the source of rules to the expected
			var expectedPolicy api.NimServicePolicyDefaultsAndRules
			convutils.CopySpecs(&expectedPolicy, nimServiceBody.Policy)

			expectedPolicy.Rules.Autoscaling.MaxReplicas.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}
			expectedPolicy.Rules.Autoscaling.Metric.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}

			policy := getPolicyForTest[api.NimServicePolicy](router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyNimService))
			Expect(policy.Effective).NotTo(BeNil())
			Expect(*policy.Effective.Defaults.Autoscaling).To(Equal(*expectedPolicy.Defaults.Autoscaling))
			Expect(*policy.Effective.Rules.Autoscaling).To(Equal(*expectedPolicy.Rules.Autoscaling))

			// see it in the list
			policies := listPoliciesForTest(router, 2, urlPolicy)

			Expect(policies.Policies).To(HaveLen(1))
			Expect(policies.Policies[0].Type).To(Equal(api.POLICYTYPE_NIM_SERVICE))
			Expect(policies.Policies[0].Meta.Scope).To(Equal(api.SCOPETYPE_TENANT))
			Expect(policies.Policies[0].Status.Clusters).To(HaveLen(1))
			Expect(policies.Policies[0].Status.Clusters[0].Status).To(Equal(PolicyStatusReady))

			// Patch a policy
			patchRequest := api.NimServicePolicyChangeRequest{
				Meta: api.PolicyCreationFields{
					Name:  "pol-ns1",
					Scope: api.SCOPETYPE_TENANT,
				},
				Policy: &api.NimServicePolicyDefaultsAndRules{
					Defaults: &api.NimServiceDefaults{
						Autoscaling: &api.NimServiceDefaultsAllOfAutoscaling{
							MinReplicas: convutils.Ref(int32(2)),
						},
					},
					Rules: &api.NimServiceRules{
						Autoscaling: &api.NimServiceRulesAllOfAutoscaling{
							MinReplicas: &api.IntegerRules{
								Max: convutils.Ref(int32(5)),
							},
						},
					},
				},
				Reset: []string{
					"autoscaling.maxReplicas",
				},
			}

			requestNimService, err = http.NewRequest(http.MethodPatch, urlPolicyNimService, makePayload(patchRequest))
			Expect(err).To(BeNil())
			requestNimService.Header = *headerForAdmin(2)

			responseNimService = httptest.NewRecorder()
			router.ServeHTTP(responseNimService, requestNimService)
			Expect(responseNimService.Code).To(Equal(http.StatusOK))

			// add the source of rules to the expected also add the maxReplicas remove the minReplicas
			convutils.CopySpecs(&expectedPolicy, nimServiceBody.Policy)

			expectedPolicy.Defaults.Autoscaling.MaxReplicas = nil
			expectedPolicy.Rules.Autoscaling.MaxReplicas = nil
			expectedPolicy.Defaults.Autoscaling.MinReplicas = patchRequest.Policy.Defaults.Autoscaling.MinReplicas
			expectedPolicy.Rules.Autoscaling.MinReplicas = patchRequest.Policy.Rules.Autoscaling.MinReplicas

			expectedPolicy.Rules.Autoscaling.MinReplicas.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}
			expectedPolicy.Rules.Autoscaling.Metric.SourceOfRule = &api.SourceOfRule{
				Scope: api.SCOPETYPE_TENANT,
			}

			policy = getPolicyForTest[api.NimServicePolicy](router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyNimService))
			Expect(policy.Effective).NotTo(BeNil())
			Expect(*policy.Effective.Defaults.Autoscaling).To(Equal(*expectedPolicy.Defaults.Autoscaling))
			Expect(*policy.Effective.Rules.Autoscaling).To(Equal(*expectedPolicy.Rules.Autoscaling))

			// delete policy
			deletePolicyForTest(router, 2, fmt.Sprintf("%s?scope=tenant", urlPolicyNimService))

			// make sure it is removed
			policies = listPoliciesForTest(router, 2, urlPolicy)
			Expect(policies.Policies).To(HaveLen(0))
		})
	})

	Context("policy instances is not aligned with canEdit rule", func() {
		It("should return with error", func() {
			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Storage: &api.SupersetRulesAllOfStorage{
							Pvc: &api.PvcsRules{
								Attributes: &api.PvcRules{
									Path: &api.StringRules{
										CanEdit: convutils.Ref(false),
									},
									ClaimName: &api.StringRules{
										Options: &[]api.StringOption{
											{
												Value: "claim-1",
											},
											{
												Value: "claim-2",
											},
										},
									},
								},
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							Pvc: &api.PvcsDefaults{
								Attributes: &api.PvcInstance{
									Path:      convutils.Ref("/default/path"),
									ClaimName: convutils.Ref("claim-incorrect-1"),
								},
								Instances: &[]api.PvcInstance{
									{
										ClaimName: convutils.Ref("claim-1"),
										Path:      convutils.Ref("/incorrect/path"),
									},
									{
										ClaimName: convutils.Ref("claim-incorrect-2"),
										Path:      convutils.Ref("/default/path"),
									},
								},
							},
						},
					},
				},
			}

			requestTrain, err := http.NewRequest(http.MethodPut, urlPolicyTraining, makePayload(trainingBody))
			Expect(err).To(BeNil())
			requestTrain.Header = *headerForAdmin(2)

			responseTrain := httptest.NewRecorder()
			router.ServeHTTP(responseTrain, requestTrain)
			Expect(responseTrain.Code).To(Equal(http.StatusBadRequest))

			var applyErr api.Error
			jsonErr := json.Unmarshal(responseTrain.Body.Bytes(), &applyErr)
			Expect(jsonErr).To(BeNil())

			messages := strings.Split(applyErr.Message, "\n")
			expectedMessages := []string{
				"storage.pvc.instances[1].path -> storage.pvc.attributes.path: the administrator prohibited changing the value of this field",
				"storage.pvc.instances[2].claimName -> storage.pvc.attributes.claimName: the value of this field must be one of: claim-1, claim-2",
				"storage.pvc.attributes.claimName -> the value of this field must be one of: claim-1, claim-2",
			}

			Expect(messages).To(Equal(expectedMessages))
		})
	})

	Context("imposed asset id always displayed on errors of policy", func() {
		It("should return with error", func() {
			imposedAssetId := uuid.NewString()

			trainingBody := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Storage: &api.SupersetRulesAllOfStorage{
							Pvc: &api.PvcsRules{
								Attributes: &api.PvcRules{
									Path: &api.StringRules{
										CanEdit: convutils.Ref(false),
									},
								},
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Storage: &api.SupersetDefaultsAllOfStorage{
							Pvc: &api.PvcsDefaults{
								Attributes: &api.PvcInstance{
									Path: convutils.Ref("/default/path"),
								},
								Instances: &[]api.PvcInstance{
									{
										Name:      convutils.Ref(imposedAssetId),
										ClaimName: convutils.Ref("claim-1"),
										Path:      convutils.Ref("/incorrect/path"),
									},
								},
							},
						},
					},
				},
			}

			requestTrain, err := http.NewRequest(http.MethodPut, urlPolicyTraining, makePayload(trainingBody))
			Expect(err).To(BeNil())
			requestTrain.Header = *headerForAdmin(2)

			responseTrain := httptest.NewRecorder()
			router.ServeHTTP(responseTrain, requestTrain)
			Expect(responseTrain.Code).To(Equal(http.StatusBadRequest))

			var applyErr api.Error
			jsonErr := json.Unmarshal(responseTrain.Body.Bytes(), &applyErr)
			Expect(jsonErr).To(BeNil())

			messages := strings.Split(applyErr.Message, "\n")
			expectedMessages := []string{
				fmt.Sprintf("storage.pvc.instances[%s].path -> storage.pvc.attributes.path: the administrator prohibited changing the value of this field",
					imposedAssetId),
			}

			Expect(messages).To(Equal(expectedMessages))
		})
	})

	Context("apply + validate only", func() {
		It("policy is unchanged", func() {
			//d
			//  Level 1: give defaults to two fields
			//
			bodyTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, urlPolicyWorkspace, makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusOK))

			var actualRespTenantA api.WorkspacePolicyV2
			err = json.Unmarshal(responseTenantA.Body.Bytes(), &actualRespTenantA)
			Expect(err).To(BeNil())

			expectedTenantA := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Rules: nil,
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
					},
				},
			}

			Expect(*actualRespTenantA.Policy).To(Equal(*expectedTenantA.Policy))

			var expectedEffectiveTenantA api.WorkspacePolicyDefaultsAndRulesV2
			copyErr := convutils.CopySpecs(&expectedEffectiveTenantA, expectedTenantA.Policy)
			Expect(copyErr).To(BeNil())

			dao.InjectSystemPolicyForTesting(&expectedEffectiveTenantA)

			testutils.Expect(*actualRespTenantA.Effective).ToEqual(expectedEffectiveTenantA)

			//
			//  Level 2: update one field + validate
			//
			bodyTenantB := api.WorkspacePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.WorkspacePolicyDefaultsAndRulesV2{
					Defaults: &api.WorkspacePolicyDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(40)),
						},
					},
				},
			}

			requestTenantB, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyWorkspace), makePayload(bodyTenantB))
			Expect(err).To(BeNil())
			requestTenantB.Header = *headerForAdmin(1)

			responseTenantB := httptest.NewRecorder()
			router.ServeHTTP(responseTenantB, requestTenantB)
			Expect(responseTenantB.Code).To(Equal(http.StatusNoContent))

			//
			//   Part 3 - GET the tenant policy to make sure it has not changed
			//
			getWorkspacePolicyForTest(router, 1, fmt.Sprintf("%s?scope=tenant", urlPolicyWorkspace))

			testutils.Expect(*actualRespTenantA.Policy).ToEqual(*expectedTenantA.Policy)
			testutils.Expect(*actualRespTenantA.Effective).ToEqual(expectedEffectiveTenantA)
		})
	})

	Context("apply + validate only (training)", func() {
		It("flag is supported (we already checked functionality for workspaces)", func() {
			//
			//  Level 1: give defaults to two fields
			//
			bodyTenantA := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyTraining), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("apply + validate only (inference)", func() {
		It("flag is supported (we already checked functionality for workspaces)", func() {
			//
			//  Level 1: give defaults to two fields
			//
			bodyTenantA := api.InferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.InferencePolicyDefaultsAndRulesV2{
					Defaults: &api.InferencePolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
						Security: &api.InferencePolicyDefaultsV2AllOfSecurity{
							RunAsUid: convutils.Ref(int64(30)),
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyInference), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("apply + validate only (distributed-inference)", func() {
		It("flag is supported (we already checked functionality for workspaces)", func() {
			bodyTenantA := api.DistributedInferencePolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.DistributedInferencePolicyDefaultsAndRulesV2{
					Defaults: &api.DistributedInferenceDefaultsV2{
						Leader: &api.DistributedInferenceLeaderWorkerDefaultsV2{
							Image: convutils.Ref("ubuntu"),
							Security: &api.DistributedInferenceLeaderWorkerDefaultsV2Security{
								RunAsUid: convutils.Ref(int64(30)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyDistributedInference), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("apply + validate only (distributed)", func() {
		It("flag is supported (we already checked functionality for workspaces)", func() {
			//
			//  Level 1: give defaults to two fields
			//
			bodyTenantA := api.DistributedPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.DistributedPolicyDefaultsAndRulesV2{
					Defaults: &api.DistributedPolicyDefaultsV2{
						Worker: &api.DistributedPolicyDefaultsV2Worker{
							Image: convutils.Ref("ubuntu"),
							Security: &api.SupersetDefaultsAllOfSecurity{
								RunAsUid: convutils.Ref(int64(30)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyDistributed), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusNoContent))
		})
	})

	Context("general validations", func() {
		It("IntegerRules - min/max values", func() {
			bodyTenantA := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							GpuDevicesRequest: &api.IntegerRules{
								Max: convutils.Ref(int32(2)),
								Min: convutils.Ref(int32(5)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyTraining), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusBadRequest))
			Expect(responseTenantA.Body.String()).To(ContainSubstring("max value must be no less than min value"))
		})

		It("QuantityRules - min/max values", func() {
			bodyTenantA := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							GpuMemoryRequest: &api.QuantityRules{
								Max: convutils.Ref("5M"),
								Min: convutils.Ref("1G"),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyTraining), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusBadRequest))
			Expect(responseTenantA.Body.String()).To(ContainSubstring("max value must be no less than min value"))
		})

		It("IntegerRules - step value", func() {
			bodyTenantA := api.TrainingPolicyChangeRequestV2{
				Meta: api.PolicyCreationFields{
					Scope: api.SCOPETYPE_TENANT,
					Name:  "pol-t-1",
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Compute: &api.SupersetRulesAllOfCompute{
							GpuDevicesRequest: &api.IntegerRules{
								Max:  convutils.Ref(int32(2)),
								Step: convutils.Ref(int32(5)),
							},
						},
					},
				},
			}

			requestTenantA, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s?validateOnly=true", urlPolicyTraining), makePayload(bodyTenantA))
			Expect(err).To(BeNil())
			requestTenantA.Header = *headerForAdmin(1)

			responseTenantA := httptest.NewRecorder()
			router.ServeHTTP(responseTenantA, requestTenantA)
			Expect(responseTenantA.Code).To(Equal(http.StatusBadRequest))
			Expect(responseTenantA.Body.String()).To(ContainSubstring("step value must not exceed max value"))
		})
	})
	Context("cluster scope", func() {
		It("create a cluster scoped policy and then overwrite it - simple policy", func() {
			clusterId := backendMock.ClusterIdPerTenant["1"][0]

			addBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					Name:      "pol-cluster-1",
					ClusterId: convutils.Ref(clusterId),
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "ubuntu"},
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("ubuntu"),
					},
				},
			}

			response1 := overwriteTrainingPolicyForTest(router, &addBody, 1)

			var actualResp1 api.TrainingPolicyV2
			err := json.Unmarshal(response1.Body.Bytes(), &actualResp1)
			Expect(err).To(BeNil())
			Expect(*actualResp1.Policy.Defaults).To(Equal(*addBody.Policy.Defaults))
			Expect(*actualResp1.Policy.Rules).To(Equal(*addBody.Policy.Rules))

			//
			//   now overwrite it with another policy
			//
			overwriteBody := api.TrainingPolicyOverwriteRequestV2{
				Meta: api.PolicyCreationFields{
					Scope:     api.SCOPETYPE_CLUSTER,
					Name:      "pol-cluster-1",
					ClusterId: convutils.Ref(clusterId),
				},
				Policy: &api.TrainingPolicyDefaultsAndRulesV2{
					Rules: &api.TrainingPolicyRulesV2{
						Image: &api.StringRules{
							Options: &[]api.StringOption{
								{Value: "ubuntu"},
								{Value: "alpine"},
							},
						},
					},
					Defaults: &api.TrainingPolicyDefaultsV2{
						Image: convutils.Ref("alpine"),
					},
				},
			}

			response2 := overwriteTrainingPolicyForTest(router, &overwriteBody, 1)

			var actualResp2 api.TrainingPolicyV2
			err = json.Unmarshal(response2.Body.Bytes(), &actualResp2)
			Expect(err).To(BeNil())
			Expect(*actualResp2.Policy.Defaults).To(Equal(*overwriteBody.Policy.Defaults))
			Expect(*actualResp2.Policy.Rules).To(Equal(*overwriteBody.Policy.Rules))
		})
	})

	It("create a cluster scoped policy and then overwrite it - complex policy", func() {
		clusterId := backendMock.ClusterIdPerTenant["1"][0]

		addBody := api.TrainingPolicyOverwriteRequestV2{
			Meta: api.PolicyCreationFields{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: convutils.Ref(clusterId),
				Name:      "pol-cluster-2",
			},
			Policy: &api.TrainingPolicyDefaultsAndRulesV2{
				Rules: &api.TrainingPolicyRulesV2{
					Image: &api.StringRules{
						Required: convutils.Ref(true),
					},
				},
				Defaults: &api.TrainingPolicyDefaultsV2{
					Image: convutils.Ref("image"),
				},
			},
		}

		response1 := overwriteTrainingPolicyForTest(router, &addBody, 1)

		var actualResp1 api.TrainingPolicyV2
		err := json.Unmarshal(response1.Body.Bytes(), &actualResp1)
		Expect(err).To(BeNil())
		Expect(*actualResp1.Policy.Defaults).To(Equal(*addBody.Policy.Defaults))
		Expect(*actualResp1.Policy.Rules).To(Equal(*addBody.Policy.Rules))

		//
		//   now overwrite it with another policy
		//
		overwriteBody := api.TrainingPolicyOverwriteRequestV2{
			Meta: api.PolicyCreationFields{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: convutils.Ref(clusterId),
				Name:      "pol-cluster-2",
			},
			Policy: &api.TrainingPolicyDefaultsAndRulesV2{
				Rules: &api.TrainingPolicyRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(500)),
						},
					},
				},
				Defaults: &api.TrainingPolicyDefaultsV2{
					Security: &api.SupersetDefaultsAllOfSecurity{
						RunAsUid: convutils.Ref(int64(500)),
					},
				},
			},
		}

		response2 := overwriteTrainingPolicyForTest(router, &overwriteBody, 1)

		var actualResp2 api.TrainingPolicyV2
		err = json.Unmarshal(response2.Body.Bytes(), &actualResp2)
		Expect(err).To(BeNil())
		Expect(*actualResp2.Policy.Defaults).To(Equal(*overwriteBody.Policy.Defaults))
		Expect(*actualResp2.Policy.Rules).To(Equal(*overwriteBody.Policy.Rules))
	})

	It("lock environment variables", func() {
		clusterId := backendMock.ClusterIdPerTenant["1"][0]

		addBody := api.TrainingPolicyOverwriteRequestV2{
			Meta: api.PolicyCreationFields{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: convutils.Ref(clusterId),
				Name:      "pol-cluster-2",
			},
			Policy: &api.TrainingPolicyDefaultsAndRulesV2{
				Rules: &api.TrainingPolicyRulesV2{
					EnvironmentVariables: &api.InstancesRules{
						Instances: &api.ItemRules{
							Locked: []string{
								"first-var",
							},
						},
					},
				},
				Defaults: &api.TrainingPolicyDefaultsV2{
					EnvironmentVariables: &api.EnvironmentVariablesDefaults{
						Instances: &[]api.EnvironmentVariable{
							{
								Name:  convutils.Ref("first-var"),
								Value: convutils.Ref("first"),
							},
							{
								Name:  convutils.Ref("second-var"),
								Value: convutils.Ref("second"),
							},
						},
					},
				},
			},
		}

		response1 := overwriteTrainingPolicyForTest(router, &addBody, 1)
		Expect(response1.Code).To(Equal(http.StatusOK))
	})

	It("lock environment variables - distributed", func() {
		clusterId := backendMock.ClusterIdPerTenant["1"][0]

		addBody := api.DistributedPolicyOverwriteRequestV2{
			Meta: api.PolicyCreationFields{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: convutils.Ref(clusterId),
				Name:      "pol-cluster-2",
			},
			Policy: &api.DistributedPolicyDefaultsAndRulesV2{
				Rules: &api.DistributedPolicyRulesV2{
					Master: &api.ReplicaRulesV2{
						EnvironmentVariables: &api.InstancesRules{
							Instances: &api.ItemRules{
								Locked: []string{
									"first-master-var",
								},
							},
						},
					},
					Worker: &api.DistributedPolicyRulesV2Worker{
						EnvironmentVariables: &api.InstancesRules{
							Instances: &api.ItemRules{
								Locked: []string{
									"first-worker-var",
								},
							},
						},
					},
				},
				Defaults: &api.DistributedPolicyDefaultsV2{
					Master: &api.ReplicaDefaultsV2{
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("first-master-var"),
									Value: convutils.Ref("first"),
								},
								{
									Name:  convutils.Ref("second-master-var"),
									Value: convutils.Ref("second"),
								},
							},
						},
					},
					Worker: &api.DistributedPolicyDefaultsV2Worker{
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("first-worker-var"),
									Value: convutils.Ref("first"),
								},
								{
									Name:  convutils.Ref("second-worker-var"),
									Value: convutils.Ref("second"),
								},
							},
						},
					},
				},
			},
		}

		response1 := overwriteDistributedPolicyForTest(router, &addBody, 1)
		Expect(response1.Code).To(Equal(http.StatusOK))
	})

	It("lock environment variables - distributed - negative", func() {
		clusterId := backendMock.ClusterIdPerTenant["1"][0]

		addBody := api.DistributedPolicyOverwriteRequestV2{
			Meta: api.PolicyCreationFields{
				Scope:     api.SCOPETYPE_CLUSTER,
				ClusterId: convutils.Ref(clusterId),
				Name:      "pol-cluster-2",
			},
			Policy: &api.DistributedPolicyDefaultsAndRulesV2{
				Rules: &api.DistributedPolicyRulesV2{
					Master: &api.ReplicaRulesV2{
						EnvironmentVariables: &api.InstancesRules{
							Instances: &api.ItemRules{
								Locked: []string{
									"first-master-var",
									"third-master-var",
								},
							},
						},
					},
					Worker: &api.DistributedPolicyRulesV2Worker{
						EnvironmentVariables: &api.InstancesRules{
							Instances: &api.ItemRules{
								Locked: []string{
									"first-worker-var",
									"third-worker-var",
								},
							},
						},
					},
				},
				Defaults: &api.DistributedPolicyDefaultsV2{
					Master: &api.ReplicaDefaultsV2{
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("first-master-var"),
									Value: convutils.Ref("first"),
								},
								{
									Name:  convutils.Ref("second-master-var"),
									Value: convutils.Ref("second"),
								},
							},
						},
					},
					Worker: &api.DistributedPolicyDefaultsV2Worker{
						EnvironmentVariables: &api.EnvironmentVariablesDefaults{
							Instances: &[]api.EnvironmentVariable{
								{
									Name:  convutils.Ref("first-worker-var"),
									Value: convutils.Ref("first"),
								},
								{
									Name:  convutils.Ref("second-worker-var"),
									Value: convutils.Ref("second"),
								},
							},
						},
					},
				},
			},
		}

		request, err := http.NewRequest("PUT", urlPolicyDistributed, makePayload(addBody))
		Expect(err).To(BeNil())
		request.Header = *headerForAdmin(1)

		response := httptest.NewRecorder()
		router.ServeHTTP(response, request.WithContext(tenantContextForTest(1)))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
		Expect(response.Body.String()).To(ContainSubstring("worker.environmentVariables.instances"))
		Expect(response.Body.String()).To(ContainSubstring("default value is missing for locked item 'third-worker-var'"))
		Expect(response.Body.String()).To(ContainSubstring("master.environmentVariables.instances"))
		Expect(response.Body.String()).To(ContainSubstring("default value is missing for locked item 'third-master-var'"))
		Expect(response.Body.String()).NotTo(ContainSubstring("'first-master-var'"))
		Expect(response.Body.String()).NotTo(ContainSubstring("'first-worker-var'"))
	})
})
