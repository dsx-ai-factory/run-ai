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
	"os"
	"time"

	clientsclusterservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/clusterservice"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/clusterservice"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/authorization"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/dsx-ai-factory/run-ai/control-plane/workloads/mocks/pkg/clients/backend"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/assets"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/datavolumes"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/testing_crud"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/mock"
)

var allWorkspaceTypes = []TableEntry{
	Entry(Workspace, api.WORKLOADTYPE_WORKSPACE),
	Entry(Training, api.WORKLOADTYPE_TRAINING),
	Entry(Distributed, api.WORKLOADTYPE_DISTRIBUTED),
	Entry(Inference, api.WORKLOADTYPE_INFERENCE),
}

var fixedTime = time.Date(2020, time.July, 1, 0, 0, 0, 0, time.UTC)

var nonDistributedWorkspaceTypes = []TableEntry{
	Entry(Workspace, api.WORKLOADTYPE_WORKSPACE),
	Entry(Training, api.WORKLOADTYPE_TRAINING),
	Entry(Inference, api.WORKLOADTYPE_INFERENCE),
}

var _ = Describe("unified service validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues

	var authMock *testing_crud.RandomsBasedAuthMockClient

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		policyMockUpdateTime := time.Now()
		policies.SetMock(policies.SimpleMockClientFactory{
			SimpleMockClient: policies.SimpleMockClient{
				EffectiveUpdatedAt: policyMockUpdateTime,
			},
		})

		authMock = testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		authorization.SetMock(authMock)

		orgunit.SetMock(authMock.OrgUnitMock())

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedCluster(randoms.ClusterId, testing.TestClusterVersionNativeSubmission, convutils.Ref(true))))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, GetProjectResponse(*randoms)))
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().WithDefaultNodePool(randoms.ClusterId))
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	Context("GPU memory compatibility for copy and edit", func() {
		DescribeTable("preserves both aliases after creating a workspace", func(prefix string) {
			request := api.WorkspaceCreationRequest{Name: "gpu-copy", ProjectId: randoms.ProjectIdStr, ClusterId: randoms.ClusterId}
			Expect(json.Unmarshal([]byte(fmt.Sprintf(`{"image":"ubuntu","compute":{"gpuDevicesRequest":1,"gpuRequestType":"portion","%sRequest":0.5,"%sLimit":0.5}}`, prefix, prefix)), &request.Spec)).To(Succeed())
			settings := &bc_mocks.TenantSettingsClient{}
			settings.On("GetGpuResourceOptimization", mock.Anything).Return(true, nil).Once()
			created, err := CreateWorkspace(ctx, settings, request)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Code).To(Equal(http.StatusAccepted), "%v", created.Body)
			workload := created.Body.(workloadsApi.SupersetWorkload)
			req := unifiedReqByRandoms(api.WORKLOADTYPE_WORKSPACE, randoms)
			req.WorkloadId = &workload.WorkloadId
			response, err := NewUnifiedAPIService().GetUnifiedWorkload(ctx, api.UnifiedWorkloadRequest(*req))
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusOK))
			result := response.Body.(*workloadsApi.UnifiedWorkload)
			compute := result.Spec.Get().Worker.Get().Compute.Get()
			Expect(compute.GetGpuPortionRequest()).To(Equal(0.5))
			Expect(compute.GetGpuMemoryPortionRequest()).To(Equal(0.5))
			Expect(compute.GetGpuPortionLimit()).To(Equal(0.5))
			Expect(compute.GetGpuMemoryPortionLimit()).To(Equal(0.5))
		}, Entry("legacy request", "gpuPortion"), Entry("new request", "gpuMemoryPortion"))

		for _, workloadType := range []api.WorkloadType{api.WORKLOADTYPE_WORKSPACE, api.WORKLOADTYPE_TRAINING, api.WORKLOADTYPE_INFERENCE, api.WORKLOADTYPE_DISTRIBUTED, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE} {
			DescribeTable("preserves snapshot aliases for "+string(workloadType), func(prefix string, value float64) {
				snapshot := workloadV2SnapshotForTest(uuid.New(), workloadType, randoms)
				computeJSON := fmt.Sprintf(`{"compute":{"gpuDevicesRequest":1,"gpuRequestType":"portion","%sRequest":%v,"%sLimit":0.5}}`, prefix, value, prefix)
				payload := computeJSON
				if workloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE {
					payload = fmt.Sprintf(`{"leader":%s,"worker":%s}`, computeJSON, computeJSON)
				}
				Expect(json.Unmarshal([]byte(payload), snapshot.Spec)).To(Succeed())
				if workloadType == api.WORKLOADTYPE_DISTRIBUTED {
					snapshot.MasterSpec = &workloadsApi.MasterSpec{}
					Expect(json.Unmarshal([]byte(computeJSON), snapshot.MasterSpec)).To(Succeed())
				}
				Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())
				req := unifiedReqByRandoms(workloadType, randoms)
				req.WorkloadId = convutils.Ref(snapshot.ID.String())
				result, err := req.Exec(ctx)
				Expect(err).To(BeNil())
				spec := result.Spec.Get()
				computes := []*workloadsApi.WorkspaceSpecSpecAllOfCompute{spec.Worker.Get().Compute.Get()}
				if workloadType == api.WORKLOADTYPE_DISTRIBUTED {
					computes = append(computes, spec.Master.Get().Compute.Get())
				}
				if workloadType == api.WORKLOADTYPE_DISTRIBUTED_INFERENCE {
					computes = []*workloadsApi.WorkspaceSpecSpecAllOfCompute{spec.Worker.Get().Leader.Get().Compute.Get(), spec.Worker.Get().Worker.Get().Compute.Get()}
				}
				for _, compute := range computes {
					Expect(compute).NotTo(BeNil())
					Expect(compute.GetGpuPortionRequest()).To(Equal(value))
					if value > 1 {
						Expect(compute.GpuMemoryPortionRequest.IsSet()).To(BeFalse())
					} else {
						Expect(compute.GpuMemoryPortionRequest.Get()).NotTo(BeNil())
						Expect(compute.GetGpuMemoryPortionRequest()).To(Equal(value))
					}
					Expect(compute.GetGpuPortionLimit()).To(Equal(0.5))
					Expect(compute.GetGpuMemoryPortionLimit()).To(Equal(0.5))
				}
			}, Entry("legacy fraction", "gpuPortion", 0.5), Entry("new fraction", "gpuMemoryPortion", 0.5), Entry("explicit zero", "gpuPortion", 0.0), Entry("legacy whole GPU count", "gpuPortion", 4.0))
		}
	})

	Context("no policy defined", func() {
		DescribeTable("returns empty response", func(workloadType api.WorkloadType) {
			result, err := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(err).To(BeNil())
			Expect(result.Spec.Get()).To(BeNil())
			Expect(result.PolicyInfo.Get()).To(BeNil())
		}, allWorkspaceTypes)
	})

	Context("policy with defaults only - Workspace", func() {
		It("returns correct response", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with defaults only - Training", func() {
		It("returns correct response", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_TRAINING

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with defaults only - Inference", func() {
		It("returns correct response", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_INFERENCE

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with defaults only - Distributed", func() {
		It("returns correct response", func() {
			fileName := "defaults_only_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with defaults and rules - Workspace", func() {
		It("returns correct response", func() {
			fileName := "defaults_and_rules"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with defaults and rules - need to clear non existent node pools", func() {
		It("returns correct response", func() {
			fileName := "defaults_node_pools"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			projectInfo := GetProjectResponse(*randoms)
			projectInfo.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
				Workspace: []string{"1", "2"},
				Names: map[string]string{
					"1": "node-1",
					"2": "node-2",
				},
			}
			projectInfo.Effective.DefaultNodePools = []string{
				"node-pool-3",
				"node-pool-4",
			}
			projectInfo.Resources = []orgUnitApi.Resources{
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-1"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-2"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-3"),
					}),
				},
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, projectInfo))
			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fmt.Sprintf("%s_nodes_policy_fixed", fileName), workloadType)
		})
	})

	Context("policy with defaults and rules (dist) - need to clear non existent node pools", func() {
		It("returns correct response", func() {
			fileName := "defaults_node_pools_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			projectInfo := GetProjectResponse(*randoms)
			projectInfo.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
				Workspace: []string{"1", "2"},
				Names: map[string]string{
					"1": "node-1",
					"2": "node-2",
				},
			}
			projectInfo.Effective.DefaultNodePools = []string{
				"node-pool-3",
				"node-pool-4",
			}
			projectInfo.Resources = []orgUnitApi.Resources{
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-1"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-2"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-3"),
					}),
				},
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, projectInfo))
			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fmt.Sprintf("%s_nodes_policy_dist_fixed", fileName), workloadType)
		})
	})

	Context("policy with defaults and rules - node type and node pools from the project", func() {
		It("returns correct response", func() {
			fileName := "defaults_and_rules"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			projectInfo := GetProjectResponse(*randoms)
			projectInfo.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
				Workspace: []string{"1", "2"},
				Names: map[string]string{
					"1": "node-1",
					"2": "node-2",
				},
			}
			projectInfo.Effective.DefaultNodePools = []string{
				"node-pool-1",
				"node-pool-2",
			}
			projectInfo.Resources = []orgUnitApi.Resources{
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-1"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-2"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-3"),
					}),
				},
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, projectInfo))
			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fmt.Sprintf("%s_nodes_policy", fileName), workloadType)
		})
	})

	Context("policy with defaults and rules - node type and node pools from the project, disributed", func() {
		It("returns correct response", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			projectInfo := GetProjectResponse(*randoms)
			projectInfo.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
				Names: map[string]string{
					"1": "node-1",
					"2": "node-2",
				},
			}
			projectInfo.Effective.DefaultNodePools = []string{
				"node-pool-1",
				"node-pool-2",
			}
			projectInfo.Resources = []orgUnitApi.Resources{
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-1"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-2"),
					}),
				},
				{
					NodePool: *orgUnitApi.NewNullableResourcesNodePool(&orgUnitApi.ResourcesNodePool{
						Name: convutils.Ref("node-pool-3"),
					}),
				},
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, projectInfo))

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, "no_policy_nodes_pools_dist", workloadType)
		})
	})

	Context("policy with defaults and rules - distributed - policy have defaults also", func() {
		It("returns correct response", func() {
			fileName := "defaults_and_rules_dist_nodes"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			projectInfo := GetProjectResponse(*randoms)
			projectInfo.NodeTypes = &orgUnitApi.NodeTypesPerWorkload{
				Training: []string{"1", "2"},
				Names: map[string]string{
					"1": "node-1-from-project",
					"2": "node-2-from-project",
				},
			}
			projectInfo.Effective.DefaultNodePools = []string{
				"node-pool-1-from-project",
				"node-pool-2-from-project",
			}

			org_unit.SetMock(org_unit.NewSimpleMock().WithProject(randoms.ProjectIdStr, projectInfo))
			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fmt.Sprintf("%s_nodes_policy", fileName), workloadType)
		})
	})

	Context("policy with defaults and rules - Distributed", func() {
		It("returns correct response", func() {
			fileName := "defaults_and_rules_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			loadPolicyForTest(fileName, workloadType)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName, workloadType)
		})
	})

	Context("policy with assets, missing clusterId", func() {
		It("error", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(uuid.NewString())

			_, execErr := req.Exec(ctx)
			Expect(int(execErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(execErr.Message).To(ContainSubstring("clusterId is mandatory"))
		})
	})

	Context("policy with assets, asset does not exist", func() {
		It("error", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			assets.SetMock(assets.NewSimpleMock())

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(uuid.NewString())
			req.ClusterId = &randoms.ClusterId

			_, execErr := req.Exec(ctx)
			Expect(int(execErr.Code)).To(Equal(http.StatusNotFound))
			Expect(execErr.Message).To(ContainSubstring("asset not found"))
		})
	})

	Context("policy with assets, asset does not exist", func() {
		It("error", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			assets.SetMock(assets.NewSimpleMock())

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(uuid.NewString())
			req.ClusterId = &randoms.ClusterId

			_, execErr := req.Exec(ctx)
			Expect(int(execErr.Code)).To(Equal(http.StatusNotFound))
			Expect(execErr.Message).To(ContainSubstring("asset not found"))
		})
	})

	Context("policy with assets, asset exists", func() {
		It("should join the policy and asset together", func() {
			fileName := "defaults_and_rules"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			assetId := uuid.NewString()

			assets.SetMock(assets.NewSimpleMock().WithMapping(workloadType, assetId, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("image_from_asset")),
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("env_name_from_asset")),
						Value: *assetsApi.NewNullableString(convutils.Ref("env_value_from_asset")),
					},
				},
			}))

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(assetId)
			req.ClusterId = &randoms.ClusterId

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_asset", workloadType)
		})
	})

	Context("policy with two assets", func() {
		It("should join the policy and assets together", func() {
			fileName := "defaults_and_rules"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			envAssetId := uuid.NewString()
			computeAssetId := uuid.NewString()

			assetsMock := assets.NewSimpleMock()

			assetsMock.WithMapping(workloadType, envAssetId, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("image_from_asset")),
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("env_name_from_asset")),
						Value: *assetsApi.NewNullableString(convutils.Ref("env_value_from_asset")),
					},
				},
			})

			assetsMock.WithMapping(workloadType, computeAssetId, &assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(&assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					LargeShmRequest: *assetsApi.NewNullableBool(convutils.Ref(true)),
				}),
			})

			assets.SetMock(assetsMock)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetIds = []string{envAssetId, computeAssetId}
			req.ClusterId = &randoms.ClusterId

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_two_asset", workloadType)
		})
	})

	Context("policy with assets, asset exists - distributed", func() {
		It("should join the policy and asset together", func() {
			fileName := "defaults_and_rules_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			loadPolicyForTest(fileName, workloadType)

			assetId := uuid.NewString()

			assets.SetMock(assets.NewSimpleMock().WithMapping(workloadType, assetId, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("image_from_asset")),
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("env_name_from_asset")),
						Value: *assetsApi.NewNullableString(convutils.Ref("env_value_from_asset")),
					},
				},
			}))

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(assetId)
			req.ClusterId = &randoms.ClusterId

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_asset", workloadType)
		})
	})

	Context("policy with workload and asset id", func() {
		It("error", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(uuid.NewString())
			req.WorkloadId = convutils.Ref(uuid.NewString())

			_, execErr := req.Exec(ctx)
			Expect(int(execErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(execErr.Message).To(ContainSubstring("mutual exclusive"))
		})
	})

	Context("template and workload id", func() {
		It("error", func() {
			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(uuid.NewString())
			req.TemplateId = convutils.Ref(uuid.NewString())

			_, execErr := req.Exec(ctx)
			Expect(int(execErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(execErr.Message).To(ContainSubstring("mutual exclusive"))
		})
	})

	Context("policy with workload - workspace", func() {
		It("should give unified info", func() {
			request := api.WorkspaceCreationRequest{
				Name:      "ws-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("from-spec"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec"),
							Value: convutils.Ref("env-value-from-spec"),
						},
					},
				},
			}

			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateWorkspace(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			workspace := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(workspace.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_workspace", workloadType)
		})
	})

	Context("policy with V2 workload snapshot", func() {
		It("falls back to the V2 workload snapshot when crud_workloads has no row", func() {
			workloadType := api.WORKLOADTYPE_TRAINING
			workloadId := uuid.New()
			snapshot := workloadV2SnapshotForTest(workloadId, workloadType, randoms)
			snapshot.Spec.Image = *workloadsApi.NewNullableString(convutils.Ref("snapshot-image"))

			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(workloadId.String())

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())
			Expect(result.Meta).NotTo(BeNil())
			Expect(result.Meta.Name).To(Equal(snapshot.Name))
			Expect(result.Spec.Get().Worker.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Image.Get()).To(Equal("snapshot-image"))
		})

		It("preserves distributed master and masterSameAsWorker semantics from the snapshot", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED
			workloadId := uuid.New()
			masterSameAsWorker := false
			snapshot := workloadV2SnapshotForTest(workloadId, workloadType, randoms)
			snapshot.Spec.Image = *workloadsApi.NewNullableString(convutils.Ref("worker-image"))
			snapshot.MasterSpec = &workloadsApi.MasterSpec{
				Image: *workloadsApi.NewNullableString(convutils.Ref("master-image")),
			}
			snapshot.MasterSameAsWorker = &masterSameAsWorker

			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(workloadId.String())

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Image.Get()).To(Equal("worker-image"))
			Expect(result.Spec.Get().Master.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Master.Get().Image.Get()).To(Equal("master-image"))
			Expect(result.Spec.Get().MasterSpecSameAsWorker.Get()).To(BeNil())
		})

		It("does not expose snapshot-only rows to legacy sync", func() {
			workloadType := api.WORKLOADTYPE_TRAINING
			snapshot := workloadV2SnapshotForTest(uuid.New(), workloadType, randoms)
			Expect(dao_crud.DbUpsertWorkloadV2Snapshot(ctx, snapshot)).To(BeNil())

			updates, dbErr := dao_crud.DbListUpdates(ctx, randoms.TenantId, string(workloadType), randoms.ClusterId, 0, 10)
			Expect(dbErr).To(BeNil())
			Expect(updates).To(BeEmpty())
		})
	})

	Context("policy with template - workspace", func() {
		It("should give unified info", func() {
			request := api.WorkspaceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "ts-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.WorkspaceTemplateSpec{
					Image: convutils.Ref("from-spec"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec"),
							Value: convutils.Ref("env-value-from-spec"),
						},
					},
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								ExistingPvc: convutils.Ref(true),
								ClaimName:   convutils.Ref("aaa"),
							},
						},
					},
				},
			}

			fileName := "defaults_only"
			workloadType := api.WORKLOADTYPE_WORKSPACE

			loadPolicyForTest(fileName, workloadType)

			ts := TemplatesService{}
			createResponse, _ := ts.CreateWorkspaceTemplate(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			template := createResponse.Body.(api.SupersetTemplate)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.TemplateId = convutils.Ref(template.Meta.Id)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_ws_template", workloadType)
		})
	})

	Context("policy with workload - distributed", func() {
		It("should give unified info", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image: convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
				},
				MasterSpec: &api.MasterSpec{
					Image: convutils.Ref("image-from-spec-master"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-master"),
							Value: convutils.Ref("env-value-from-spec-master"),
						},
					},
				},
			}

			fileName := "defaults_only_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			distributed := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(distributed.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_dist_w_m", workloadType)
		})
	})

	Context("policy with workload + same as master - distributed", func() {
		It("should give unified info", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-same-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image: convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			fileName := "defaults_only_dist"
			workloadType := api.WORKLOADTYPE_DISTRIBUTED

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			distributed := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(distributed.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_dist_same_as", workloadType)
		})
	})

	Context("policy with workload - distributed same as worker", func() {
		It("should give unified info", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					Image: convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
				},
				MasterSpecSameAsWorker: convutils.Ref(true),
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED
			fileName := "defaults_worker"

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			distributed := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(distributed.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_dist_same_as", workloadType)
		})
	})

	Context("policy with workload - distributed no master", func() {
		It("should give unified info", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					CreateHomeDir: convutils.Ref(true),
					WorkingDir:    convutils.Ref("/working/dir"),
					Image:         convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
				},
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED
			fileName := "defaults_and_rules_dist"

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			distributed := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(distributed.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_dist_w_only", workloadType)
		})

		It("should fail when trying to add PVC despite canAdd=false", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					CreateHomeDir: convutils.Ref(true),
					WorkingDir:    convutils.Ref("/working/dir"),
					Image:         convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								Path:        convutils.Ref("/new/pvc/path"),
								ExistingPvc: convutils.Ref(false),
							},
						},
					},
				},
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED
			fileName := "defaults_and_rules_dist"

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited adding new items"))
		})

		It("should fail when trying to add PVC despite canAdd=false", func() {
			request := api.DistributedCreationRequest{
				Name:      "ds-2",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedSpecSpec{
					CreateHomeDir: convutils.Ref(true),
					WorkingDir:    convutils.Ref("/working/dir"),
					Image:         convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("env-name-from-spec-worker"),
							Value: convutils.Ref("env-value-from-spec-worker"),
						},
					},
					DistributedFramework: convutils.Ref(api.DISTRIBUTEDFRAMEWORK_TF),
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						Pvc: &[]api.PvcInstance{
							{
								Path:        convutils.Ref("/new/pvc/path"),
								ExistingPvc: convutils.Ref(false),
							},
						},
					},
				},
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED
			fileName := "defaults_and_rules_dist"

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateDistributed(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited adding new items"))
		})
	})

	Context("policy with template violation - training with canEdit=false", func() {
		It("should fail when template violates canEdit rule", func() {
			// Create a template that changes the image field
			request := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "template-violates-policy",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("image-from-template"), // This violates policy with canEdit=false
				},
			}

			workloadType := api.WORKLOADTYPE_TRAINING
			fileName := "image_can_edit_false" // Policy with canEdit=false for image field

			loadPolicyForTest(fileName, workloadType)

			ts := TemplatesService{}
			createResponse, _ := ts.CreateTrainingTemplate(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			template := createResponse.Body.(api.SupersetTemplate)

			// Create a training workload using this template
			trainingRequest := api.TrainingCreationRequest{
				Name:       "training-with-bad-template",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(template.Meta.Id),
			}

			createWorkloadResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, trainingRequest)
			Expect(createWorkloadResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createWorkloadResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited changing the value"))
		})
	})

	Context("policy with locked environment variable", func() {
		It("should fail when trying to modify a locked environment variable", func() {
			// Create a training workload that tries to modify a locked environment variable
			request := api.TrainingCreationRequest{
				Name:      "training-with-locked-env",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.TrainingSpecSpec{
					Image: convutils.Ref("ubuntu"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("LOCKED_VAR"),
							Value: convutils.Ref("different-value"), // Trying to change locked value
						},
					},
				},
			}

			workloadType := api.WORKLOADTYPE_TRAINING
			fileName := "env_locked_policy"

			loadPolicyForTest(fileName, workloadType)

			createResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited modifying item"))
		})

		It("should fail when template tries to modify a locked environment variable", func() {
			// Create a template that tries to modify a locked environment variable
			templateRequest := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "template-with-locked-env",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("LOCKED_VAR"),
							Value: convutils.Ref("template-value"), // Template trying to change locked value
						},
					},
				},
			}

			workloadType := api.WORKLOADTYPE_TRAINING
			fileName := "env_locked_policy"

			loadPolicyForTest(fileName, workloadType)

			ts := TemplatesService{}
			createTemplateResponse, _ := ts.CreateTrainingTemplate(ctx, templateRequest)
			Expect(createTemplateResponse.Code).To(Equal(http.StatusOK))

			template := createTemplateResponse.Body.(api.SupersetTemplate)

			// Create a training workload using this template
			trainingRequest := api.TrainingCreationRequest{
				Name:       "training-using-bad-template",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(template.Meta.Id),
			}

			createWorkloadResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, trainingRequest)
			Expect(createWorkloadResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createWorkloadResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited modifying item"))
		})
	})

	Context("policy with template violation - training with canEdit=false", func() {
		It("should fail when template violates canEdit rule", func() {
			// Create a template that changes the image field
			request := api.TrainingTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "template-violates-policy",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.TrainingTemplateSpec{
					Image: convutils.Ref("image-from-template"), // This violates policy with canEdit=false
				},
			}

			workloadType := api.WORKLOADTYPE_TRAINING
			fileName := "image_can_edit_false" // Policy with canEdit=false for image field

			loadPolicyForTest(fileName, workloadType)

			ts := TemplatesService{}
			createResponse, _ := ts.CreateTrainingTemplate(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			template := createResponse.Body.(api.SupersetTemplate)

			// Create a training workload using this template
			trainingRequest := api.TrainingCreationRequest{
				Name:       "training-with-bad-template",
				ProjectId:  randoms.ProjectIdStr,
				ClusterId:  randoms.ClusterId,
				TemplateId: convutils.Ref(template.Meta.Id),
			}

			createWorkloadResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, trainingRequest)
			Expect(createWorkloadResponse.Code).To(Equal(http.StatusBadRequest))
			payload := createWorkloadResponse.Body.(*workloadsApi.SubmissionErrorV2)
			Expect(len(payload.ComplianceIssues)).To(BeNumerically(">", 0))
			Expect(payload.ComplianceIssues[0].Details).To(ContainSubstring("administrator prohibited changing the value"))
		})
	})

	Context("unified request without projectId", func() {
		It("should return unified info without project info", func() {
			workloadType := api.WORKLOADTYPE_WORKSPACE
			fileName := "defaults_only"

			loadPolicyForTest(fileName, workloadType)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.ProjectId = "" // Clear project ID

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())
			Expect(result.ProjectInfo.Get()).To(BeNil())
			Expect(result.PolicyInfo.Get()).To(BeNil())
		})
	})

	It("should succeed when workload does not modify locked annotations or labels (nil values)", func() {
		// BUG: Create a training workload that does NOT modify locked annotations or labels
		// by not sending them at all (nil values). This should succeed but currently fails.
		// The workload should inherit the defaults from the policy without triggering
		// "prohibited modifying item" errors.
		request := api.TrainingCreationRequest{
			Name:      "training-without-locked-fields",
			ProjectId: randoms.ProjectIdStr,
			ClusterId: randoms.ClusterId,
			Spec: api.TrainingSpecSpec{
				Image: convutils.Ref("ubuntu"),
				// Not specifying annotations or labels - should use defaults from policy
				Annotations: nil,
				Labels:      nil,
			},
		}

		workloadType := api.WORKLOADTYPE_TRAINING
		fileName := "annotations_labels_locked_policy"

		loadPolicyForTest(fileName, workloadType)

		createResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, request)
		// This should succeed since we're not trying to modify the locked fields
		Expect(createResponse.Code).To(Equal(http.StatusAccepted))
	})

	It("should succeed when workload does not modify locked annotations or labels (empty arrays)", func() {
		// BUG: Create a training workload that sends empty arrays for annotations/labels
		// This should succeed but currently fails with compliance errors.
		// Empty arrays mean "don't change anything, use policy defaults"
		emptyAnnotations := []api.Annotation{}
		emptyLabels := []api.Label{}

		request := api.TrainingCreationRequest{
			Name:      "training-with-empty-fields",
			ProjectId: randoms.ProjectIdStr,
			ClusterId: randoms.ClusterId,
			Spec: api.TrainingSpecSpec{
				Image:       convutils.Ref("ubuntu"),
				Annotations: &emptyAnnotations,
				Labels:      &emptyLabels,
			},
		}

		workloadType := api.WORKLOADTYPE_TRAINING
		fileName := "annotations_labels_locked_policy"

		loadPolicyForTest(fileName, workloadType)

		createResponse, _ := CreateTraining(ctx, &bc_mocks.TenantSettingsClient{}, request)
		// This should succeed since we're not trying to modify the locked fields
		Expect(createResponse.Code).To(Equal(http.StatusAccepted))
	})

	Context("exclude default instances", func() {
		It("should exclude all excluded and replace replaced instances with their up to date instance", func() {
			request := api.WorkspaceCreationRequest{
				Name:      "ws-1",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.WorkspaceSpecSpec{
					Image: convutils.Ref("image-from-spec-worker"),
					EnvironmentVariables: &[]api.EnvironmentVariable{
						{
							Name:    convutils.Ref("env-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("env-3"),
							Value: convutils.Ref("modified-value-3"),
						},
					},
					Annotations: &[]api.Annotation{
						{
							Name:    convutils.Ref("ann-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("ann-3"),
							Value: convutils.Ref("modified-ann-3"),
						},
					},
					Labels: &[]api.Label{
						{
							Name:    convutils.Ref("label-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:  convutils.Ref("label-3"),
							Value: convutils.Ref("modified-label-3"),
						},
					},
					Ports: &[]api.Port{
						{
							Name:    convutils.Ref("port-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name:      convutils.Ref("port-3"),
							Container: convutils.Ref(int32(19083)),
						},
					},
					ExposedUrls: &[]api.ExposedUrl{
						{
							Name:    convutils.Ref("exposed-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name: convutils.Ref("exposed-3"),
							Url:  convutils.Ref("http://exposed.com/modified-url/3"),
						},
					},
					RelatedUrls: &[]api.RelatedUrl{
						{
							Name:    convutils.Ref("related-1"),
							Exclude: convutils.Ref(true),
						},
						{
							Name: convutils.Ref("related-3"),
							Url:  convutils.Ref("http://related.com/modified-url/3"),
						},
					},
					Compute: &api.WorkspaceSpecSpecAllOfCompute{
						ExtendedResources: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("aaa/1"),
								Exclude:  convutils.Ref(true),
							},
							{
								Resource: convutils.Ref("aaa/3"),
								Quantity: convutils.Ref("3333M"),
							},
						},
					},
					Storage: &api.WorkspaceSpecSpecAllOfStorage{
						S3: &[]api.S3Instance{
							{
								Name:    convutils.Ref("s3-1"),
								Exclude: convutils.Ref(true),
							},
							{
								Name: convutils.Ref("s3-3"),
								Path: convutils.Ref("/s3/modified-path-3"),
							},
						},
						HostPath: &[]api.HostPathInstance{
							{
								Name:    convutils.Ref("hp-1"),
								Exclude: convutils.Ref(true),
							},
							{
								Name:      convutils.Ref("hp-3"),
								Path:      convutils.Ref("/hp/modified-path-3"),
								MountPath: convutils.Ref("/hp/modified-mount-path-3"),
							},
						},
						Git: &[]api.GitInstance{
							{
								Name:    convutils.Ref("git-1"),
								Exclude: convutils.Ref(true),
							},
							{
								Name: convutils.Ref("git-3"),
								Path: convutils.Ref("/git/modified-path-3"),
							},
						},
						DataVolume: &[]api.DataVolumeInstance{
							{
								Id:      convutils.Ref("1234-0001"),
								Exclude: convutils.Ref(true),
							},
							{
								Id:        convutils.Ref("1234-0003"),
								MountPath: convutils.Ref("/dv/modified-path-3"),
							},
						},
						Pvc: &[]api.PvcInstance{
							{
								Name:    convutils.Ref("pvc-1"),
								Exclude: convutils.Ref(true),
							},
							{
								Name:        convutils.Ref("pvc-3"),
								ExistingPvc: convutils.Ref(true),
								Path:        convutils.Ref("/path/modified-path-3"),
								ClaimName:   convutils.Ref("modified-claim-3"),
							},
						},
						Nfs: &[]api.NfsInstance{
							{
								Name:    convutils.Ref("nfs-1"),
								Exclude: convutils.Ref(true),
							},
							{
								Name:   convutils.Ref("nfs-3"),
								Path:   convutils.Ref("/path/modified-path-1"),
								Server: convutils.Ref("modified-server-3"),
							},
						},
						ConfigMapVolume: &[]api.ConfigMapInstance{
							{
								Name:      convutils.Ref("cm-1"),
								ConfigMap: convutils.Ref("cm-1"),
								Exclude:   convutils.Ref(true),
							},
							{
								Name:      convutils.Ref("cm-3"),
								ConfigMap: convutils.Ref("cm-3"),
								MountPath: convutils.Ref("/cm/modified-path-3"),
							},
						},
					},
				},
			}

			workloadType := api.WORKLOADTYPE_WORKSPACE
			fileName := "defaults_for_instances"

			loadPolicyForTest(fileName, workloadType)

			// Mocking data volumes since the policy have default data volume instances
			dataVolumesMock := datavolumes.NewSimpleMock().WithDataVolumes([]openapi.DatavolumeNoSharedScopes{
				{
					Id: "1234-0001",
				},
				{
					Id: "1234-0002",
				},
				{
					Id: "1234-0003",
				},
			})
			datavolumes.SetMock(dataVolumesMock)

			createResponse, _ := CreateWorkspace(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			distributed := createResponse.Body.(workloadsApi.SupersetWorkload)

			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(distributed.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			validateVsExpected(result, fileName+"_with_exclude_ws", workloadType)
		})
	})

	Context("policy with defaults only - DistributedInference", func() {
		It("returns correct response in spec.worker", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up DistributedInference policy mock directly
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(3))),
							ServingPort: *policyApi.NewNullableDistributedInferenceServingPort(&policyApi.DistributedInferenceServingPort{
								Port: *policyApi.NewNullableInt32(convutils.Ref(int32(8080))),
							}),
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image: *policyApi.NewNullableString(convutils.Ref("leader-image:v1")),
								Compute: *policyApi.NewNullableSupersetDefaultsAllOfCompute(&policyApi.SupersetDefaultsAllOfCompute{
									CpuCoreLimit: *policyApi.NewNullableFloat64(convutils.Ref(float64(2))),
								}),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image: *policyApi.NewNullableString(convutils.Ref("worker-image:v1")),
								Compute: *policyApi.NewNullableSupersetDefaultsAllOfCompute(&policyApi.SupersetDefaultsAllOfCompute{
									CpuCoreLimit: *policyApi.NewNullableFloat64(convutils.Ref(float64(4))),
								}),
							}),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			// Validate the response structure
			Expect(result).NotTo(BeNil())
			Expect(result.ProjectInfo.Get()).NotTo(BeNil())
			Expect(result.ProjectInfo.Get().Name).To(Equal(convutils.Ref("project-1")))

			// Verify policy info structure (in WORKER)
			Expect(result.PolicyInfo.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get().Defaults.Get()).NotTo(BeNil())

			// Verify spec structure (in WORKER)
			Expect(result.Spec.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Replicas.Get()).To(Equal(int32(2)))
			Expect(*result.Spec.Get().Worker.Get().Workers.Get()).To(Equal(int32(3)))
			Expect(*result.Spec.Get().Worker.Get().ServingPort.Get().Port.Get()).To(Equal(int32(8080)))
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("leader-image:v1"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("worker-image:v1"))
		})
	})

	Context("policy with workload - DistributedInference", func() {
		It("should give unified info when copying existing workload", func() {
			// Create a DistributedInference workload first
			request := api.DistributedInferenceCreationRequest{
				Name:      "dinf-to-copy",
				ProjectId: randoms.ProjectIdStr,
				ClusterId: randoms.ClusterId,
				Spec: api.DistributedInferenceSpecSpec{
					ServingPort: &api.DistributedInferenceServingPort{
						Port:     convutils.Ref(int32(9090)),
						Protocol: convutils.Ref(api.DISTRIBUTEDINFERENCESERVINGPORTPROTOCOL_HTTP),
					},
					Workers:       convutils.Ref(int32(4)),
					Replicas:      convutils.Ref(int32(2)),
					StartupPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED),
					RestartPolicy: convutils.Ref(api.DISTRIBUTEDINFERENCERESTARTPOLICY_RECREATE_GROUP_ON_POD_RESTART),
					Leader: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("my-leader-image:v2"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerSpec{
						Image: convutils.Ref("my-worker-image:v2"),
					},
				},
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up policy mock
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			// Create the workload
			createResponse, _ := CreateDistributedInference(ctx, &bc_mocks.TenantSettingsClient{}, request)
			Expect(createResponse.Code).To(Equal(http.StatusAccepted))

			createdWorkload := createResponse.Body.(workloadsApi.SupersetWorkload)

			// Now call unified API with workloadId (simulating "Copy" button click)
			req := unifiedReqByRandoms(workloadType, randoms)
			req.WorkloadId = convutils.Ref(createdWorkload.WorkloadId)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify data is in spec.WORKER
			Expect(result.Spec.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())

			// Verify workload values are in the response (not just policy defaults)
			Expect(*result.Spec.Get().Worker.Get().Replicas.Get()).To(Equal(int32(2)))
			Expect(*result.Spec.Get().Worker.Get().Workers.Get()).To(Equal(int32(4)))
			Expect(*result.Spec.Get().Worker.Get().ServingPort.Get().Port.Get()).To(Equal(int32(9090)))
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("my-leader-image:v2"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("my-worker-image:v2"))

			// Verify policy info is also in WORKER
			Expect(result.PolicyInfo.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get()).NotTo(BeNil())
		})
	})

	Context("policy with defaults and rules - DistributedInference", func() {
		It("returns correct response with both defaults and rules", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up policy mock with both defaults and rules
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image: *policyApi.NewNullableString(convutils.Ref("default-leader-image:v1")),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image: *policyApi.NewNullableString(convutils.Ref("default-worker-image:v1")),
							}),
						}),
						Rules: *policyApi.NewNullableDistributedInferenceRulesV2(&policyApi.DistributedInferenceRulesV2{
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(true)),
								}),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(false)),
								}),
							}),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify structure in WORKER
			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get()).NotTo(BeNil())

			// Check defaults
			Expect(*result.Spec.Get().Worker.Get().Replicas.Get()).To(Equal(int32(1)))
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("default-leader-image:v1"))

			// Check rules in policyInfo (complex structure, just checking existence)
			Expect(result.PolicyInfo.Get().Worker.Get().Rules.Get()).NotTo(BeNil())
		})
	})

	Context("policy defaults structure validation - DistributedInference", func() {
		It("should preserve leader and worker subsections in policyInfo.worker.defaults and policyInfo.worker.rules", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up policy with leader and worker subsections in BOTH defaults and rules
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(3))),
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image:   *policyApi.NewNullableString(convutils.Ref("policy-leader-image:v1")),
								Command: *policyApi.NewNullableString(convutils.Ref("ubuntu")),
								Args:    *policyApi.NewNullableString(convutils.Ref("puki")),
								Compute: *policyApi.NewNullableSupersetDefaultsAllOfCompute(&policyApi.SupersetDefaultsAllOfCompute{
									CpuCoreLimit: *policyApi.NewNullableFloat64(convutils.Ref(float64(2))),
								}),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image:   *policyApi.NewNullableString(convutils.Ref("policy-worker-image:v1")),
								Command: *policyApi.NewNullableString(convutils.Ref("alpine")),
								Args:    *policyApi.NewNullableString(convutils.Ref("worker-args")),
								Compute: *policyApi.NewNullableSupersetDefaultsAllOfCompute(&policyApi.SupersetDefaultsAllOfCompute{
									CpuCoreLimit: *policyApi.NewNullableFloat64(convutils.Ref(float64(4))),
								}),
							}),
						}),
						Rules: *policyApi.NewNullableDistributedInferenceRulesV2(&policyApi.DistributedInferenceRulesV2{
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(true)),
								}),
								Command: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(false)),
								}),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(false)),
								}),
								Args: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(true)),
								}),
							}),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			req := unifiedReqByRandoms(workloadType, randoms)
			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify policyInfo structure exists
			Expect(result.PolicyInfo.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get().Defaults.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get().Rules.Get()).NotTo(BeNil())

			policyDefaults := result.PolicyInfo.Get().Worker.Get().Defaults.Get()
			policyDefaultsJson, jsonErr := json.Marshal(policyDefaults)
			Expect(jsonErr).To(BeNil())

			var policyDefaultsMap map[string]interface{}
			mapErr := json.Unmarshal(policyDefaultsJson, &policyDefaultsMap)
			Expect(mapErr).To(BeNil())

			keys := make([]string, 0, len(policyDefaultsMap))
			for k := range policyDefaultsMap {
				keys = append(keys, k)
			}

			if replicasVal, hasReplicas := policyDefaultsMap["replicas"]; hasReplicas {
				Expect(replicasVal).To(Equal(float64(2)))
			}
			if workersVal, hasWorkers := policyDefaultsMap["workers"]; hasWorkers {
				Expect(workersVal).To(Equal(float64(3)))
			}

			leaderMap, leaderExists := policyDefaultsMap["leader"]
			Expect(leaderExists).To(BeTrue(), "policyInfo.worker.defaults should contain 'leader' subsection. Available keys: %v", keys)
			Expect(leaderMap).NotTo(BeNil())

			leaderMapTyped := leaderMap.(map[string]interface{})
			Expect(leaderMapTyped["image"]).To(Equal("policy-leader-image:v1"))
			Expect(leaderMapTyped["command"]).To(Equal("ubuntu"))
			Expect(leaderMapTyped["args"]).To(Equal("puki"))
			computeMap, computeExists := leaderMapTyped["compute"]
			Expect(computeExists).To(BeTrue())
			computeMapTyped := computeMap.(map[string]interface{})
			Expect(computeMapTyped["cpuCoreLimit"]).To(Equal(float64(2)))

			workerMap, workerExists := policyDefaultsMap["worker"]
			Expect(workerExists).To(BeTrue(), "policyInfo.worker.defaults should contain 'worker' subsection. Available keys: %v", keys)
			Expect(workerMap).NotTo(BeNil())

			workerMapTyped := workerMap.(map[string]interface{})
			Expect(workerMapTyped["image"]).To(Equal("policy-worker-image:v1"))
			Expect(workerMapTyped["command"]).To(Equal("alpine"))
			Expect(workerMapTyped["args"]).To(Equal("worker-args"))
			workerComputeMap, workerComputeExists := workerMapTyped["compute"]
			Expect(workerComputeExists).To(BeTrue())
			workerComputeMapTyped := workerComputeMap.(map[string]interface{})
			Expect(workerComputeMapTyped["cpuCoreLimit"]).To(Equal(float64(4)))

			policyRules := result.PolicyInfo.Get().Worker.Get().Rules.Get()
			policyRulesJson, rulesJsonErr := json.Marshal(policyRules)
			Expect(rulesJsonErr).To(BeNil())

			var policyRulesMap map[string]interface{}
			rulesMapErr := json.Unmarshal(policyRulesJson, &policyRulesMap)
			Expect(rulesMapErr).To(BeNil())

			rulesLeaderMap, rulesLeaderExists := policyRulesMap["leader"]
			Expect(rulesLeaderExists).To(BeTrue(), "policyInfo.worker.rules should contain 'leader' subsection")
			Expect(rulesLeaderMap).NotTo(BeNil())

			rulesLeaderMapTyped := rulesLeaderMap.(map[string]interface{})
			leaderImageRule, leaderImageExists := rulesLeaderMapTyped["image"]
			Expect(leaderImageExists).To(BeTrue())
			leaderImageRuleTyped := leaderImageRule.(map[string]interface{})
			Expect(leaderImageRuleTyped["canEdit"]).To(Equal(true))
			leaderCommandRule, leaderCommandExists := rulesLeaderMapTyped["command"]
			Expect(leaderCommandExists).To(BeTrue())
			leaderCommandRuleTyped := leaderCommandRule.(map[string]interface{})
			Expect(leaderCommandRuleTyped["canEdit"]).To(Equal(false))

			rulesWorkerMap, rulesWorkerExists := policyRulesMap["worker"]
			Expect(rulesWorkerExists).To(BeTrue(), "policyInfo.worker.rules should contain 'worker' subsection")
			Expect(rulesWorkerMap).NotTo(BeNil())

			rulesWorkerMapTyped := rulesWorkerMap.(map[string]interface{})
			workerImageRule, workerImageExists := rulesWorkerMapTyped["image"]
			Expect(workerImageExists).To(BeTrue())
			workerImageRuleTyped := workerImageRule.(map[string]interface{})
			Expect(workerImageRuleTyped["canEdit"]).To(Equal(false))
			workerArgsRule, workerArgsExists := rulesWorkerMapTyped["args"]
			Expect(workerArgsExists).To(BeTrue())
			workerArgsRuleTyped := workerArgsRule.(map[string]interface{})
			Expect(workerArgsRuleTyped["canEdit"]).To(Equal(true))

			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get().Leader.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get().Worker.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("policy-leader-image:v1"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("policy-worker-image:v1"))
		})
	})

	Context("policy with assets - DistributedInference", func() {
		It("should populate leader and worker subsections with asset specs", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up policy with defaults
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Command: *policyApi.NewNullableString(convutils.Ref("/bin/bash")),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Command: *policyApi.NewNullableString(convutils.Ref("/bin/sh")),
							}),
						}),
						Rules: *policyApi.NewNullableDistributedInferenceRulesV2(&policyApi.DistributedInferenceRulesV2{
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(true)),
								}),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerRulesV2(&policyApi.DistributedInferenceLeaderWorkerRulesV2{
								Image: *policyApi.NewNullableStringRules(&policyApi.StringRules{
									CanEdit: *policyApi.NewNullableBool(convutils.Ref(true)),
								}),
							}),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			// Create environment asset mock
			envAssetId := uuid.NewString()
			assetsMock := assets.NewSimpleMock()

			assetsMock.WithMapping(workloadType, envAssetId, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("ubuntu:22.04")),
				EnvironmentVariables: []assetsApi.EnvironmentVariable{
					{
						Name:  *assetsApi.NewNullableString(convutils.Ref("TEST_ENV")),
						Value: *assetsApi.NewNullableString(convutils.Ref("test_value")),
					},
				},
			})

			assets.SetMock(assetsMock)

			// Make unified request with asset
			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetId = convutils.Ref(envAssetId)
			req.ClusterId = &randoms.ClusterId

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())
			Expect(result).NotTo(BeNil())
			Expect(result.Spec.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())

			// Verify leader and worker subsections are populated with asset specs
			Expect(result.Spec.Get().Worker.Get().Leader.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get().Worker.Get()).NotTo(BeNil())

			Expect(result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("ubuntu:22.04"))

			Expect(result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("ubuntu:22.04"))

			Expect(result.Spec.Get().Worker.Get().Leader.Get().EnvironmentVariables).NotTo(BeNil())
			Expect(len(result.Spec.Get().Worker.Get().Leader.Get().EnvironmentVariables)).To(BeNumerically(">", 0))

			Expect(result.Spec.Get().Worker.Get().Worker.Get().EnvironmentVariables).NotTo(BeNil())
			Expect(len(result.Spec.Get().Worker.Get().Worker.Get().EnvironmentVariables)).To(BeNumerically(">", 0))

			// Verify policy info is populated
			Expect(result.PolicyInfo.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get().Defaults.Get()).NotTo(BeNil())
			Expect(result.PolicyInfo.Get().Worker.Get().Rules.Get()).NotTo(BeNil())
		})

		It("should handle multiple assets (environment + compute) for DistributedInference", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Set up policy
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			// Create environment and compute asset mocks
			envAssetId := uuid.NewString()
			computeAssetId := uuid.NewString()

			assetsMock := assets.NewSimpleMock()

			assetsMock.WithMapping(workloadType, envAssetId, &assetsApi.SupersetSpec{
				Image: *assetsApi.NewNullableString(convutils.Ref("pytorch:2.0")),
			})

			assetsMock.WithMapping(workloadType, computeAssetId, &assetsApi.SupersetSpec{
				Compute: *assetsApi.NewNullableDistributedInferenceLeaderWorkerSpecCompute(&assetsApi.DistributedInferenceLeaderWorkerSpecCompute{
					GpuDevicesRequest: *assetsApi.NewNullableInt32(convutils.Ref(int32(2))),
					CpuCoreRequest:    *assetsApi.NewNullableFloat64(convutils.Ref(float64(4.0))),
				}),
			})

			assets.SetMock(assetsMock)

			// Make unified request with multiple assets
			req := unifiedReqByRandoms(workloadType, randoms)
			req.AssetIds = []string{envAssetId, computeAssetId}
			req.ClusterId = &randoms.ClusterId

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify both image and compute merged into leader subsection
			Expect(result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("pytorch:2.0"))
			Expect(result.Spec.Get().Worker.Get().Leader.Get().Compute.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Compute.Get().GpuDevicesRequest.Get()).To(Equal(int32(2)))

			// Verify both image and compute merged into worker subsection
			Expect(result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("pytorch:2.0"))
			Expect(result.Spec.Get().Worker.Get().Worker.Get().Compute.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Compute.Get().GpuDevicesRequest.Get()).To(Equal(int32(2)))

			// Verify container fields are NOT duplicated at parent level
			Expect(result.Spec.Get().Worker.Get().Image.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().ImagePullPolicy.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Command.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Args.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Compute.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Security.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().EnvironmentVariables).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Storage.Get()).To(BeNil())
		})

		It("should not duplicate container fields at parent level for DistributedInference", func() {
			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))),
							Workers:  *policyApi.NewNullableInt32(convutils.Ref(int32(2))),
							Leader: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image:   *policyApi.NewNullableString(convutils.Ref("leader-img:v1")),
								Command: *policyApi.NewNullableString(convutils.Ref("leader-cmd")),
								Args:    *policyApi.NewNullableString(convutils.Ref("leader-args")),
							}),
							Worker: *policyApi.NewNullableDistributedInferenceLeaderWorkerDefaultsV2(&policyApi.DistributedInferenceLeaderWorkerDefaultsV2{
								Image:   *policyApi.NewNullableString(convutils.Ref("worker-img:v1")),
								Command: *policyApi.NewNullableString(convutils.Ref("worker-cmd")),
								Args:    *policyApi.NewNullableString(convutils.Ref("worker-args")),
							}),
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			result, execErr := unifiedReqByRandoms(workloadType, randoms).Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify fields exist ONLY inside leader subsection
			Expect(result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("leader-img:v1"))
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Command.Get()).To(Equal("leader-cmd"))
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Args.Get()).To(Equal("leader-args"))

			// Verify fields exist ONLY inside worker subsection
			Expect(result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).NotTo(BeNil())
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("worker-img:v1"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Command.Get()).To(Equal("worker-cmd"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Args.Get()).To(Equal("worker-args"))

			// Verify container fields are NOT at parent level (no duplicates)
			Expect(result.Spec.Get().Worker.Get().Image.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().ImagePullPolicy.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Command.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Args.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Compute.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Security.Get()).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().EnvironmentVariables).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Annotations).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Labels).To(BeNil())
			Expect(result.Spec.Get().Worker.Get().Storage.Get()).To(BeNil())
		})
	})

	Context("policy with template - DistributedInference", func() {
		It("should give unified info with template values", func() {
			request := api.DistributedInferenceTemplateCreationRequest{
				Meta: api.TemplateCreationMetaFields{
					Name:      "dinf-template-1",
					ScopeType: api.SCOPETYPE_PROJECT,
					ScopeId:   convutils.Ref(randoms.ProjectIdStr),
				},
				Spec: api.DistributedInferenceTemplateSpec{
					Replicas: convutils.Ref(int32(5)),
					Workers:  convutils.Ref(int32(4)),
					Leader: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("template-leader:v1"),
					},
					Worker: &api.DistributedInferenceLeaderWorkerTemplateSpec{
						Image: convutils.Ref("template-worker:v1"),
					},
				},
			}

			workloadType := api.WORKLOADTYPE_DISTRIBUTED_INFERENCE

			// Mock policy
			policyMock := policies.SimpleMockClientFactory{
				SimpleMockClient: policies.SimpleMockClient{
					EffectiveDistributedInferencePolicyV2: &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{
						Defaults: *policyApi.NewNullableDistributedInferenceDefaultsV2(&policyApi.DistributedInferenceDefaultsV2{
							Replicas: *policyApi.NewNullableInt32(convutils.Ref(int32(1))), // Policy says 1
						}),
					},
				},
			}
			policies.SetMock(policyMock)

			// Create Template
			ts := TemplatesService{}
			createResponse, _ := ts.CreateDistributedInferenceTemplate(ctx, request)
			Expect(createResponse.Code).To(Equal(http.StatusOK))

			template := createResponse.Body.(api.SupersetTemplate)

			// Call Unified API with Template ID
			req := unifiedReqByRandoms(workloadType, randoms)
			req.TemplateId = convutils.Ref(template.Meta.Id)

			result, execErr := req.Exec(ctx)
			Expect(execErr).To(BeNil())

			// Verify template values override policy defaults
			Expect(result.Spec.Get()).NotTo(BeNil())
			Expect(result.Spec.Get().Worker.Get()).NotTo(BeNil())

			// Template has 5 replicas (overrides policy 1)
			Expect(*result.Spec.Get().Worker.Get().Replicas.Get()).To(Equal(int32(5)))
			// Template has 4 workers
			Expect(*result.Spec.Get().Worker.Get().Workers.Get()).To(Equal(int32(4)))
			// Template images
			Expect(*result.Spec.Get().Worker.Get().Leader.Get().Image.Get()).To(Equal("template-leader:v1"))
			Expect(*result.Spec.Get().Worker.Get().Worker.Get().Image.Get()).To(Equal("template-worker:v1"))
		})
	})

})

func unifiedReqByRandoms(workloadType api.WorkloadType, randoms *testutils.RandomValues) *UnifiedWorkloadRequest {
	return &UnifiedWorkloadRequest{
		WorkloadType: workloadType,
		ProjectId:    randoms.ProjectIdStr,
	}
}

func workloadV2SnapshotForTest(workloadId uuid.UUID, workloadType api.WorkloadType, randoms *testutils.RandomValues) *models_crud.WorkloadV2Snapshot {
	return &models_crud.WorkloadV2Snapshot{
		ID:            workloadId,
		CreatedAt:     fixedTime,
		CreatedBy:     testing.TestUser,
		UpdatedAt:     fixedTime,
		UpdatedBy:     testing.TestUser,
		TenantId:      int32(randoms.TenantId),
		ClusterId:     randoms.ClusterId,
		DepartmentId:  randoms.DepartmentId,
		ProjectId:     randoms.ProjectIdStr,
		WorkloadType:  string(workloadType),
		RequestedName: "snapshot-workload",
		Name:          "snapshot-workload",
		Namespace:     "runai-snapshot",
		Spec: &workloadsApi.SupersetSpec{
			Image: *workloadsApi.NewNullableString(convutils.Ref("ubuntu")),
		},
	}
}

func loadPolicyForTest(fileName string, workloadType api.WorkloadType) {
	payload, err := testutils.LoadInputFile(fileName, ".")
	Expect(err).To(BeNil())

	policyMock := policies.SimpleMockClientFactory{}

	payloadAsMap := make(map[string]interface{})
	yamlErr := yaml.Unmarshal(payload, &payloadAsMap)
	Expect(yamlErr).To(BeNil())

	switch workloadType {
	case Workspace:
		policyMock.EffectiveWorkspacePolicyV2 = &policyApi.WorkspacePolicyDefaultsAndRulesV2{}
		jsonErr := json.Unmarshal(gojsonq.ToBytes(payloadAsMap), policyMock.EffectiveWorkspacePolicyV2)
		Expect(jsonErr).To(BeNil())
	case Training:
		policyMock.EffectiveTrainingPolicyV2 = &policyApi.TrainingPolicyDefaultsAndRulesV2{}
		jsonErr := json.Unmarshal(gojsonq.ToBytes(payloadAsMap), policyMock.EffectiveTrainingPolicyV2)
		Expect(jsonErr).To(BeNil())
	case Distributed:
		policyMock.EffectiveDistributedPolicyV2 = &policyApi.DistributedPolicyDefaultsAndRulesV2{}
		jsonErr := json.Unmarshal(gojsonq.ToBytes(payloadAsMap), policyMock.EffectiveDistributedPolicyV2)
		Expect(jsonErr).To(BeNil())
	case Inference:
		policyMock.EffectiveInferencePolicyV2 = &policyApi.InferencePolicyDefaultsAndRulesV2{}
		jsonErr := json.Unmarshal(gojsonq.ToBytes(payloadAsMap), policyMock.EffectiveInferencePolicyV2)
		Expect(jsonErr).To(BeNil())
	case DistributedInference:
		policyMock.EffectiveDistributedInferencePolicyV2 = &policyApi.DistributedInferencePolicyDefaultsAndRulesV2{}
		jsonErr := json.Unmarshal(gojsonq.ToBytes(payloadAsMap), policyMock.EffectiveDistributedInferencePolicyV2)
		Expect(jsonErr).To(BeNil())
	default:
		assert.Fail(GinkgoT(), fmt.Sprintf("unsupported workload type %s", workloadType))
	}

	policies.SetMock(policyMock)
}

func validateVsExpected(actualResponse *workloadsApi.UnifiedWorkload, fileName string, workloadType api.WorkloadType) {
	// need to clean fields that change b/w test to avoid flakiness
	if actualResponse.Meta != nil {
		actualResponse.Meta.ClusterId = "cluster-id"
		actualResponse.Meta.DepartmentId = convutils.Ref("department-id")
		actualResponse.Meta.ProjectId = "project-id"
		actualResponse.Meta.CreatedAt = fixedTime
	}

	actualYaml, yamlErr := yaml.Marshal(actualResponse)
	Expect(yamlErr).To(BeNil())

	expectedYaml := testutils.LoadExpectedFile(fmt.Sprintf("%s_%s", fileName, workloadType), actualYaml, ".")

	isOk := assert.Equal(GinkgoT(), string(expectedYaml), string(actualYaml), fileName)
	if !isOk {
		tmpPath := fmt.Sprintf("%s/%s.yaml", os.TempDir(), fileName)
		osError := os.WriteFile(tmpPath, actualYaml, 0o755)
		if osError != nil {
			assert.Fail(GinkgoT(), fmt.Sprintf("failed to backup the result to %s: %v", tmpPath, osError.Error()))
		} else {
			assert.Fail(GinkgoT(), fmt.Sprintf("actual result backed up to %s", tmpPath))
		}
	}
}
