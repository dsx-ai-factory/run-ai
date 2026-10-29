/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

// buildNodePoolPatchFromResources converts legacy per-nodepool ResourcesNullable into the new
// NodePoolResourcesPatchRequest format for a given orgUnit (department/project).
func buildNodePoolPatchFromResources(orgUnitId string, res api.ResourcesNullable) api.NodePoolResourcesPatchRequest {
	patch := api.NodePoolResourcesPatchRequest{
		OrgUnitId: orgUnitId,
	}
	if res.NodePool != nil {
		patch.NodePoolId = res.NodePool.Id
	}

	gpu := &api.ResourcesPerNodePoolNullableGpu{
		Deserved:        res.Gpu.Deserved,
		Limit:           res.Gpu.Limit,
		OverQuotaWeight: res.Gpu.OverQuotaWeight,
	}
	patch.Resources = &api.ResourcesPerNodePoolNullable{
		Gpu:    gpu,
		Cpu:    res.Cpu,
		Memory: res.Memory,
	}
	patch.Rank = res.Rank
	return patch
}

var _ = Describe("Batch patch resources API tests", Ordered, func() {
	clusterId := uuid.New().String()
	clusterId2 := uuid.New().String()
	ctx := SetContext(TestTenant, TestUser)

	var departmentApiService api.DepartmentsAPIServicer
	var projectApiService api.ProjectsAPIServicer
	var departmentId string
	var projectId string
	var initialDepartmentResources []api.Resources
	var initialProjectResources []api.Resources
	var departmentId2 string
	var projectId2 string

	defaultNodePoolName := "default"
	defaultNodePoolId := "1"
	defaultNodePoolId2 := "10"

	clusterServiceClient := mocks.IClusterClient{}
	clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
		Uuid:    clusterId,
		Version: *clustersOpenapi.NewNullableString(ptr.To("2.21")),
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     defaultNodePoolId,
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
			},
			{
				Name:                   "pool-2",
				Id:                     "2",
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
			},
		},
	}, nil)
	clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId2).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
		Nodepools: []clustersOpenapi.NodepoolMinimal{
			{
				Name:                   defaultNodePoolName,
				Id:                     defaultNodePoolId2,
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
			},
			{
				Name:                   "pool-2",
				Id:                     "20",
				Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
				SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
			},
		},
	}, nil)

	BeforeAll(func() {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Starting batch patch tests")
		CleanTables()

		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(clusterId, buildDisplayedCluster(clusterId, "2.21.0")).
			WithCluster(clusterId2, buildDisplayedCluster(clusterId2, "2.21.0")))
		checker.InitClusterCapabilitiesChecker("", getDefaultClustersCacheConfig(), "")

		departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)

		// Create a department
		createDepartmentRequest := getFullDepartmentCreateRequest("batch-dept", clusterId)
		depResp, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest)
		Expect(err).To(BeNil())
		Expect(depResp.Code).To(Equal(http.StatusCreated))
		createdDept := depResp.Body.(api.Department)
		departmentId = createdDept.Id
		initialDepartmentResources = createdDept.Resources

		// Create a project under the department
		projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
		projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)
		parentId := departmentId
		createProjectReq := api.ProjectCreationRequest{
			Name:      "batch-proj",
			ClusterId: clusterId,
			ParentId:  &parentId,
			Resources: createDefaultApiResources(true),
		}
		projResp, err := projectApiService.CreateProject(ctx, createProjectReq)
		Expect(err).To(BeNil())
		Expect(projResp.Code).To(Equal(http.StatusCreated))
		createdProj := projResp.Body.(api.Project)
		projectId = createdProj.Id
		initialProjectResources = createdProj.Resources

		// Create resources in a different cluster to validate cross-cluster batch
		createDepartmentRequest2 := api.DepartmentCreationRequest{
			Name:      "batch-dept-2",
			ClusterId: clusterId2,
			Resources: []api.ResourcesNullable{CreateNodePoolApiResources(true, defaultNodePoolId2, "default")},
		}
		depResp2, err := departmentApiService.CreateDepartment(ctx, createDepartmentRequest2)
		Expect(err).To(BeNil())
		Expect(depResp2.Code).To(Equal(http.StatusCreated), fmt.Sprintf("got body: %+v", depResp2.Body))
		departmentId2 = depResp2.Body.(api.Department).Id

		parentId2 := departmentId2
		createProjectReq2 := api.ProjectCreationRequest{
			Name:      "batch-proj-2",
			ClusterId: clusterId2,
			ParentId:  &parentId2,
			Resources: []api.ResourcesNullable{CreateNodePoolApiResources(true, defaultNodePoolId2, "default")},
		}
		projResp2, err := projectApiService.CreateProject(ctx, createProjectReq2)
		Expect(err).To(BeNil())
		Expect(projResp2.Code).To(Equal(http.StatusCreated), fmt.Sprintf("got body: %+v", projResp2.Body))
		projectId2 = projResp2.Body.(api.Project).Id
	})

	AfterAll(func() {
		CleanTables()
	})

	Context("Common services - nodepools cache", func() {
		It("uses cache and does not call cluster client when cached", func() {
			ctx := SetContext(TestTenant, TestUser)
			clusterId := uuid.New().String()
			cache := sync.Map{}

			// Prepare cached value
			cached := map[string]clustersOpenapi.NodepoolMinimal{
				"np-1": {
					Id:   "np-1",
					Name: "cached-pool",
				},
			}
			cache.Store(clusterId, cached)

			mockClient := mocks.IClusterClient{}
			// No expectation set for GetNodePoolsMinimalByCluster on purpose (should not be called)

			got, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, &mockClient, clusterId, &cache)
			Expect(apiErr).To(BeNil())
			Expect(got).To(Equal(cached))

			// Verify client was not called
			mockClient.AssertNotCalled(GinkgoT(), "GetNodePoolsMinimalByCluster", mock.Anything, clusterId)

			// Now clear cache and set expectation: should be called exactly once
			cache = sync.Map{}
			mockClient = mocks.IClusterClient{}
			mockClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Id:   "np-2",
						Name: "fetched-pool",
					},
				},
			}, nil).Once()

			got2, apiErr := services.GetNodePoolsForClusterFromServiceOrCache(ctx, &mockClient, clusterId, &cache)
			Expect(apiErr).To(BeNil())
			Expect(got2).To(HaveKey("np-2"))
			mockClient.AssertNumberOfCalls(GinkgoT(), "GetNodePoolsMinimalByCluster", 1)

			// Second call with cache populated should NOT call again
			_, apiErr = services.GetNodePoolsForClusterFromServiceOrCache(ctx, &mockClient, clusterId, &cache)
			Expect(apiErr).To(BeNil())
			mockClient.AssertNumberOfCalls(GinkgoT(), "GetNodePoolsMinimalByCluster", 1)
		})
	})

	Context("Departments batch patch", func() {
		It("returns 200 and mixed results when some succeed and some fail", func() {
			onePoolPatch := []api.ResourcesNullable{
				CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName),
			}
			onePoolPatch[0].Gpu.Deserved = ptr.To(float64(1.5))
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, onePoolPatch[0]),
					buildNodePoolPatchFromResources("999999", getUpdateResources()[0]),
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", resp.Body))

			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Succeeded)).To(Equal(1))
			Expect(len(body.Failed)).To(Equal(1))

			// Validate all relevant fields on succeeded entry
			succeeded := body.Succeeded[0]
			Expect(succeeded.Id).ToNot(BeEmpty())
			Expect(succeeded.OrgUnitId).To(Equal(departmentId))
			Expect(succeeded.OrgUnitName).To(Equal("batch-dept"))
			Expect(succeeded.ClusterId).To(Equal(clusterId))
			Expect(succeeded.NodePoolId).To(Equal(defaultNodePoolId))
			Expect(succeeded.NodePoolName).To(Equal(defaultNodePoolName))
			Expect(succeeded.Rank).To(Equal(services.DefaultResourceRank))
			Expect(succeeded.Resources.Gpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Deserved).To(Equal(1.5))

			// Validate failed entry matches the bad payload and carries proper error info
			failed := body.Failed[0]
			Expect(failed.Payload).To(Equal(req.PatchRequests[1]))
			Expect(failed.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(failed.Message).To(ContainSubstring("failed to get department node pool resources"))

			// Validate updated state using node pool resources view:
			// - default nodepool GPU deserved updated to 1.5
			// - second nodepool GPU deserved unchanged
			nodePoolResp, err := departmentApiService.GetDepartmentNodePoolResources(
				ctx,
				[]string{fmt.Sprintf("orgUnitId==%s", departmentId)},
				api.DEPARTMENTNODEPOOLRESOURCESFILTERSORTFIELDS_NODE_POOL_ID,
				"asc",
				0,
				0,
				"",
			)
			Expect(err).To(BeNil())
			Expect(nodePoolResp.Code).To(Equal(http.StatusOK))
			nodePools := nodePoolResp.Body.(api.GetDepartmentNodePoolResources200Response).NodePoolResources
			updatedById := map[string]api.DepartmentNodePoolResources{}
			for _, r := range nodePools {
				updatedById[r.NodePoolId] = r
			}
			initialById := map[string]api.Resources{}
			for _, r := range initialDepartmentResources {
				initialById[r.NodePool.Id] = r
			}
			Expect(*updatedById[defaultNodePoolId].Resources.Gpu.Deserved).To(Equal(1.5))
			Expect(*updatedById["2"].Resources.Gpu.Deserved).To(Equal(initialById["2"].Gpu.Deserved))
			Expect(updatedById[defaultNodePoolId].NodePoolName).To(Equal(defaultNodePoolName))
			Expect(updatedById["2"].NodePoolName).To(Equal("pool-2"))
		})

		It("returns 500 when all operations fail", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources("999999", getUpdateResources()[0]),
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))

			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Succeeded)).To(Equal(0))
			Expect(len(body.Failed)).To(Equal(1))
		})

		It("returns 400 for empty batch list", func() {
			req := api.BatchPatchNodePoolResourcesRequest{PatchRequests: []api.NodePoolResourcesPatchRequest{}}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("return 400 when invalid rank", func() {
			invalidRank := "Ultra"
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Rank:       ptr.To(invalidRank),
					},
				},
			}

			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(body.Failed[0].Message).To(ContainSubstring("unknown rank"))
		})

		It("returns 400 when duplicate departmentIds appear in batch", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(departmentId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("duplicate OrgUnitId//NodePoolId"))
		})

		It("patching only GPU keeps CPU/Memory/Rank unchanged for department", func() {
			// First, patch to populate CPU/Memory/Rank if they were nil
			newGpu := 3.0
			cpu := 4.0
			memory := 8192.0
			oqw := 3.0
			rank := "High"
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Rank:       ptr.To(rank),
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: &newGpu,
								Limit:    &newGpu,
							},
							Cpu: &api.ResourcesCpu{
								Deserved:        ptr.To(cpu),
								Limit:           ptr.To(cpu),
								OverQuotaWeight: ptr.To(oqw),
							},
							Memory: &api.ResourcesMemory{
								Deserved:        ptr.To(memory),
								Limit:           ptr.To(memory),
								OverQuotaWeight: ptr.To(oqw),
							},
						},
					},
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Now, Patch only GPU fields on that nodepool
			newGpu = 9.0
			req = api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: &newGpu,
								Limit:    &newGpu,
							},
						},
					},
				},
			}
			resp, err = departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(0)))
			Expect(len(body.Succeeded)).To(Equal(1))

			succeeded := body.Succeeded[0]
			Expect(succeeded.OrgUnitId).To(Equal(departmentId))
			Expect(succeeded.NodePoolId).To(Equal(defaultNodePoolId))

			// GPU updated
			Expect(succeeded.Resources.Gpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Deserved).To(Equal(newGpu))
			Expect(succeeded.Resources.Gpu.Limit).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Limit).To(Equal(newGpu))

			// CPU unchanged
			Expect(succeeded.Resources.Cpu).ToNot(BeNil())
			Expect(succeeded.Resources.Cpu.Deserved).To(Equal(ptr.To(cpu)))
			Expect(succeeded.Resources.Cpu.Limit).To(Equal(ptr.To(cpu)))
			Expect(succeeded.Resources.Cpu.OverQuotaWeight).To(Equal(ptr.To(oqw)))

			// Memory unchanged
			Expect(succeeded.Resources.Memory).ToNot(BeNil())
			Expect(succeeded.Resources.Memory.Deserved).To(Equal(ptr.To(memory)))
			Expect(succeeded.Resources.Memory.Limit).To(Equal(ptr.To(memory)))
			Expect(succeeded.Resources.Memory.OverQuotaWeight).To(Equal(ptr.To(oqw)))

			// Rank unchanged
			Expect(succeeded.Rank).To(Equal(rank))
		})

		It("patching only GPU OverQuotaWeight keeps Deserved/Limit unchanged for department", func() {
			// First, set initial values including OverQuotaWeight
			initialGpu := 5.0
			initialOqw := 2.0
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved:        &initialGpu,
								Limit:           &initialGpu,
								OverQuotaWeight: &initialOqw,
							},
						},
					},
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Verify initial state
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(*body.Succeeded[0].Resources.Gpu.Deserved).To(Equal(initialGpu))
			Expect(*body.Succeeded[0].Resources.Gpu.Limit).To(Equal(initialGpu))
			Expect(*body.Succeeded[0].Resources.Gpu.OverQuotaWeight).To(Equal(initialOqw))

			// Now patch ONLY OverQuotaWeight without providing Deserved
			newOqw := 15.0
			req = api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								OverQuotaWeight: &newOqw,
							},
						},
					},
				},
			}
			resp, err = departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body = resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(0)))
			Expect(len(body.Succeeded)).To(Equal(1))

			succeeded := body.Succeeded[0]
			Expect(succeeded.OrgUnitId).To(Equal(departmentId))
			Expect(succeeded.NodePoolId).To(Equal(defaultNodePoolId))

			// GPU OverQuotaWeight updated
			Expect(succeeded.Resources.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.OverQuotaWeight).To(Equal(newOqw))

			// GPU Deserved/Limit unchanged
			Expect(succeeded.Resources.Gpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Deserved).To(Equal(initialGpu))
			Expect(succeeded.Resources.Gpu.Limit).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Limit).To(Equal(initialGpu))
		})

		It("patching only CPU OverQuotaWeight keeps Deserved/Limit unchanged for department", func() {
			// First, set initial values including OverQuotaWeight
			initialCpu := 10.0
			initialOqw := 5.0
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Cpu: &api.ResourcesCpu{
								Deserved:        &initialCpu,
								Limit:           &initialCpu,
								OverQuotaWeight: &initialOqw,
							},
						},
					},
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Verify initial state
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(*body.Succeeded[0].Resources.Cpu.Deserved).To(Equal(initialCpu))
			Expect(*body.Succeeded[0].Resources.Cpu.Limit).To(Equal(initialCpu))
			Expect(*body.Succeeded[0].Resources.Cpu.OverQuotaWeight).To(Equal(initialOqw))

			// Now patch ONLY CPU OverQuotaWeight without providing Deserved
			newOqw := 12.0
			req = api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Cpu: &api.ResourcesCpu{
								OverQuotaWeight: &newOqw,
							},
						},
					},
				},
			}
			resp, err = departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body = resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(0)))

			succeeded := body.Succeeded[0]

			// CPU OverQuotaWeight updated
			Expect(succeeded.Resources.Cpu).ToNot(BeNil())
			Expect(succeeded.Resources.Cpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*succeeded.Resources.Cpu.OverQuotaWeight).To(Equal(newOqw))

			// CPU Deserved/Limit unchanged
			Expect(succeeded.Resources.Cpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Cpu.Deserved).To(Equal(initialCpu))
			Expect(succeeded.Resources.Cpu.Limit).ToNot(BeNil())
			Expect(*succeeded.Resources.Cpu.Limit).To(Equal(initialCpu))
		})

		It("patching only Memory OverQuotaWeight keeps Deserved/Limit unchanged for department", func() {
			// First, set initial values including OverQuotaWeight
			initialMemory := 4096.0
			initialOqw := 3.0
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Memory: &api.ResourcesMemory{
								Deserved:        &initialMemory,
								Limit:           &initialMemory,
								OverQuotaWeight: &initialOqw,
								Units:           api.UNITS_MIB,
							},
						},
					},
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Verify initial state
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(*body.Succeeded[0].Resources.Memory.Deserved).To(Equal(initialMemory))
			Expect(*body.Succeeded[0].Resources.Memory.Limit).To(Equal(initialMemory))
			Expect(*body.Succeeded[0].Resources.Memory.OverQuotaWeight).To(Equal(initialOqw))

			// Now patch ONLY Memory OverQuotaWeight without providing Deserved
			newOqw := 8.0
			req = api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  departmentId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Memory: &api.ResourcesMemory{
								OverQuotaWeight: &newOqw,
							},
						},
					},
				},
			}
			resp, err = departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body = resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(0)))

			succeeded := body.Succeeded[0]

			// Memory OverQuotaWeight updated
			Expect(succeeded.Resources.Memory).ToNot(BeNil())
			Expect(succeeded.Resources.Memory.OverQuotaWeight).ToNot(BeNil())
			Expect(*succeeded.Resources.Memory.OverQuotaWeight).To(Equal(newOqw))

			// Memory Deserved/Limit unchanged
			Expect(succeeded.Resources.Memory.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Memory.Deserved).To(Equal(initialMemory))
			Expect(succeeded.Resources.Memory.Limit).ToNot(BeNil())
			Expect(*succeeded.Resources.Memory.Limit).To(Equal(initialMemory))
		})
	})

	Context("Projects batch patch - limit quota over-subscription", func() {
		var (
			projectA, projectB, depId string
		)

		BeforeAll(func() {
			// Enable feature flag
			setLimitQuotaOverSubscription(true)
		})

		AfterAll(func() {
			setLimitQuotaOverSubscription(false)
		})

		BeforeEach(func() {
			// Create a new department with a small quota on default pool in clusterId
			createDeptReq := api.DepartmentCreationRequest{
				Name:      "batch-dept-oversub",
				ClusterId: clusterId,
				Resources: []api.ResourcesNullable{
					CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName),
				},
			}
			// Set department GPU quota to 2 on default pool
			createDeptReq.Resources[0].Gpu.Deserved = ptr.To(float64(2))
			createDeptReq.Resources[0].Gpu.Limit = ptr.To(float64(2))
			createDeptReq.Resources[0].Gpu.OverQuotaWeight = ptr.To(float64(2))

			depResp, err := departmentApiService.CreateDepartment(ctx, createDeptReq)
			Expect(err).To(BeNil())
			Expect(depResp.Code).To(Equal(http.StatusCreated))
			depId = depResp.Body.(api.Department).Id

			// Create two projects under the same department with initial GPU 0
			parentId := depId
			createProj := func(name string) string {
				req := api.ProjectCreationRequest{
					Name:      name,
					ClusterId: clusterId,
					ParentId:  &parentId,
					Resources: []api.ResourcesNullable{CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)},
				}
				// initial 0
				req.Resources[0].Gpu.Deserved = ptr.To(float64(0))
				req.Resources[0].Gpu.Limit = ptr.To(float64(0))
				resp, err := projectApiService.CreateProject(ctx, req)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusCreated), fmt.Sprintf("create project failed: %+v", resp.Body))
				return resp.Body.(api.Project).Id
			}

			projectA = createProj("oversub-proj-a")
			projectB = createProj("oversub-proj-b")
		})

		AfterEach(func() {
			_, err := projectApiService.DeleteProject(ctx, projectA, "", false)
			Expect(err).To(BeNil())
			_, err = projectApiService.DeleteProject(ctx, projectB, "", false)
			Expect(err).To(BeNil())
			_, err = departmentApiService.DeleteDepartment(ctx, depId)
		})

		It("passes when one request alone exceeds quota but combined batch stays within quota", func() {
			// Department quota GPU=2 on default pool, projectA=2, projectB=0 initially.
			// Seed current state: projectA=0, projectB=2 initially. (so A->2 alone would exceed if B stayed 2).
			seedReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(projectB, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			zero := 0.0
			two := 2.0
			// projectA -> 0
			seedReq.PatchRequests[0].Resources.Gpu.Deserved = &zero
			seedReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(0))
			seedReq.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(float64(0))
			// projectB -> 2
			seedReq.PatchRequests[1].Resources.Gpu.Deserved = &two
			seedReq.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(2))
			seedReq.PatchRequests[1].Resources.Gpu.OverQuotaWeight = ptr.To(float64(2))
			seedResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, seedReq)
			Expect(err).To(BeNil())
			Expect(seedResp.Code).To(Equal(http.StatusOK), fmt.Sprintf("seed resp: %+v", seedResp.Body))

			// Now batch: A -> 2, B -> 0; Combined stays at 2 => should pass
			batchReq := seedReq
			batchReq.PatchRequests[0].Resources.Gpu.Deserved = &two
			batchReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(2))
			batchReq.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(float64(2))
			batchReq.PatchRequests[1].Resources.Gpu.Deserved = &zero
			batchReq.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(0))
			batchReq.PatchRequests[1].Resources.Gpu.OverQuotaWeight = ptr.To(float64(0))

			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", resp.Body))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(2)))
			Expect(body.TotalFailed).To(Equal(int32(0)))

			// Validate updated desired values
			byProject := map[string]float64{}
			for _, s := range body.Succeeded {
				if s.NodePoolId != defaultNodePoolId {
					continue
				}
				byProject[s.OrgUnitId] = ptr.Deref(s.Resources.Gpu.Deserved, 0)
			}
			Expect(byProject[projectA]).To(Equal(2.0))
			Expect(byProject[projectB]).To(Equal(0.0))
		})

		It("one invalid patch fails in validation while another project in same department passes", func() {
			// Batch: projId1 invalid nodepool id, projId2 valid update
			batchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, "999999", "ghost")),
					buildNodePoolPatchFromResources(projectB, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			two := 2.0
			batchReq.PatchRequests[1].Resources.Gpu.Deserved = &two
			batchReq.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(2))

			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Succeeded)).To(Equal(1))
			Expect(len(body.Failed)).To(Equal(1))

			// invalid one has 400
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusBadRequest)))
			// valid one applied
			Expect(body.Succeeded[0].OrgUnitId).To(Equal(projectB))
			Expect(body.Succeeded[0].NodePoolId).To(Equal(defaultNodePoolId))
			Expect(ptr.Deref(body.Succeeded[0].Resources.Gpu.Deserved, 0)).To(Equal(2.0))
		})

		It("each alone succeeds, together fails due to department over-subscription", func() {
			// Single patch for A to GPU=2 should be OK
			soloReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			two := 2.0
			soloReq.PatchRequests[0].Resources.Gpu.Deserved = &two
			soloReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(2))
			soloReq.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(float64(2))
			soloResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, soloReq)
			Expect(err).To(BeNil())
			Expect(soloResp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", soloResp.Body))
			Expect(soloResp.Body.(api.BatchPatchProjectResourcesResponse).TotalSuccessful).To(Equal(int32(1)))

			// Revert A back to 0 so we can test both in the same batch
			revertA := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			zero := 0.0
			revertA.PatchRequests[0].Resources.Gpu.Deserved = &zero
			revertA.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(0))
			revertA.PatchRequests[0].Resources.Gpu.OverQuotaWeight = ptr.To(float64(0))

			revertResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, revertA)
			Expect(err).To(BeNil())
			Expect(revertResp.Code).To(Equal(http.StatusOK))

			// Together: both request GPU=2 -> combined 4 > dept 2 => fail both with 500
			batchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(projectB, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			for i := range batchReq.PatchRequests {
				batchReq.PatchRequests[i].Resources.Gpu.Deserved = &two
				batchReq.PatchRequests[i].Resources.Gpu.Limit = ptr.To(float64(2))
			}

			batchResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(batchResp.Code).To(Equal(http.StatusInternalServerError), fmt.Sprintf("got body: %+v", batchResp.Body))
			body := batchResp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(2)))
			Expect(len(body.Failed)).To(Equal(2))
			// error message should indicate exceeding department available quota
			for _, f := range body.Failed {
				Expect(f.Message).To(ContainSubstring("exceeds the department's available quota"))
			}
		})

		It("fails when requested deserved is below allocated non-preemptible usage", func() {
			// Create a department and a project
			createDeptReq := getFullDepartmentCreateRequest("dep-non-preempt", clusterId)
			depResp, err := departmentApiService.CreateDepartment(ctx, createDeptReq)
			Expect(err).To(BeNil())
			Expect(depResp.Code).To(Equal(http.StatusCreated), fmt.Sprintf("got body: %+v", depResp.Body))
			depId := depResp.Body.(api.Department).Id

			parentId := depId
			projReq := api.ProjectCreationRequest{
				Name:      "proj-non-preempt",
				ClusterId: clusterId,
				ParentId:  &parentId,
				Resources: []api.ResourcesNullable{CreateNodePoolApiResources(true, defaultNodePoolId, defaultNodePoolName)},
			}
			// initial desired 1 to match allocated we will set
			projReq.Resources[0].Gpu.Deserved = ptr.To(float64(1))
			projReq.Resources[0].Gpu.Limit = ptr.To(float64(1))
			projResp, err := projectApiService.CreateProject(ctx, projReq)
			Expect(err).To(BeNil())
			Expect(projResp.Code).To(Equal(http.StatusCreated))
			projId := projResp.Body.(api.Project).Id

			// Update project status to reflect allocated non-preemptible GPU = 1 on default pool
			statusReq := api.ProjectStatusUpdateRequest{
				Id:        projId,
				ClusterId: clusterId,
				Status: api.ProjectStatus{
					Phase: "Ready",
					NodePoolQuotaStatuses: []api.QuotaStatusNodePool{
						{
							NodePoolName: defaultNodePoolName,
							NodePoolId:   defaultNodePoolId,
							AllocatedNonPreemptible: api.QuotaStatusResource{
								Gpu:    1,
								Cpu:    0,
								Memory: 0,
							},
						},
					},
				},
			}
			updStatusResp, err := projectApiService.UpdateProjectStatus(ctx, statusReq)
			Expect(err).To(BeNil())
			Expect(updStatusResp.Code).To(Equal(http.StatusOK))

			// Batch patch to set desired GPUs to 0 -> under allocated non-preemptible => should fail
			batchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			zero := 0.0
			batchReq.PatchRequests[0].Resources.Gpu.Deserved = &zero
			batchReq.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(0))

			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Message).To(ContainSubstring("below the usage of non-preemptive workloads"))

			// add another patch request for the same department - should pass
			batchReq.PatchRequests = append(batchReq.PatchRequests, buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)))
			one := 1.0
			batchReq.PatchRequests[1].Resources.Gpu.Deserved = &one
			batchReq.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(1))
			batchReq.PatchRequests[1].Resources.Gpu.OverQuotaWeight = ptr.To(float64(1))
			resp, err = projectApiService.BatchPatchProjectNodePoolResources(ctx, batchReq)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			body = resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))

		})

		It("patches 100 projects across 2 departments in one batch; one department fails, the other succeeds", func() {
			// Helpers
			createDepartmentWithQuota := func(name string, quota float64) string {
				req := api.DepartmentCreationRequest{
					Name:      name,
					ClusterId: clusterId,
					Resources: []api.ResourcesNullable{CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)},
				}
				req.Resources[0].Gpu.Deserved = ptr.To(quota)
				req.Resources[0].Gpu.Limit = ptr.To(quota)
				req.Resources[0].Gpu.OverQuotaWeight = ptr.To(quota)
				resp, err := departmentApiService.CreateDepartment(ctx, req)
				Expect(err).To(BeNil())
				Expect(resp.Code).To(Equal(http.StatusCreated), fmt.Sprintf("create dept failed: %+v", resp.Body))
				return resp.Body.(api.Department).Id
			}
			createNProjects := func(depId string, n int) []string {
				parent := depId
				ids := make([]string, 0, n)
				for i := 0; i < n; i++ {
					req := api.ProjectCreationRequest{
						Name:      fmt.Sprintf("batch-proj-%s-%d", depId, i),
						ClusterId: clusterId,
						ParentId:  &parent,
						Resources: []api.ResourcesNullable{CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)},
					}
					req.Resources[0].Gpu.Deserved = ptr.To(float64(0))
					req.Resources[0].Gpu.Limit = ptr.To(float64(0))
					req.Resources[0].Gpu.OverQuotaWeight = ptr.To(float64(0))
					resp, err := projectApiService.CreateProject(ctx, req)
					Expect(err).To(BeNil())
					Expect(resp.Code).To(Equal(http.StatusCreated), fmt.Sprintf("create project failed: %+v", resp.Body))
					ids = append(ids, resp.Body.(api.Project).Id)
				}
				return ids
			}
			buildPatchRequests := func(projectIds []string, desired float64) []api.NodePoolResourcesPatchRequest {
				out := make([]api.NodePoolResourcesPatchRequest, 0, len(projectIds))
				for _, pid := range projectIds {
					patch := buildNodePoolPatchFromResources(pid, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName))
					patch.Resources.Gpu.Deserved = ptr.To(desired)
					patch.Resources.Gpu.Limit = ptr.To(desired)
					patch.Resources.Gpu.OverQuotaWeight = ptr.To(desired)
					out = append(out, patch)
				}
				return out
			}

			// Arrange: 2 departments, 50 projects each
			depPass := createDepartmentWithQuota("batch-large-pass", 100) // plenty of quota, 50x1 should pass
			depFail := createDepartmentWithQuota("batch-large-fail", 10)  // insufficient for 50x1; should fail
			projectsPass := createNProjects(depPass, 50)
			projectsFail := createNProjects(depFail, 50)

			// Build a single batch of 100 patch requests
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: append(buildPatchRequests(projectsPass, 1), buildPatchRequests(projectsFail, 1)...),
			}

			// Act
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			// Mixed results expected: one department succeeds, one fails
			Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", resp.Body))

			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(100)))
			Expect(body.TotalSuccessful).To(Equal(int32(50)))
			Expect(body.TotalFailed).To(Equal(int32(50)))
			Expect(len(body.Succeeded)).To(Equal(50))
			Expect(len(body.Failed)).To(Equal(50))

			// Assert that successes belong to depPass projects
			passSet := map[string]struct{}{}
			for _, id := range projectsPass {
				passSet[id] = struct{}{}
			}
			for _, s := range body.Succeeded {
				_, ok := passSet[s.OrgUnitId]
				if !ok {
					Fail("Succeeded contains project not in passing department: " + s.OrgUnitId)
				}
			}

			// Cleanup created resources
			for _, pid := range append(projectsPass, projectsFail...) {
				_, delErr := projectApiService.DeleteProject(ctx, pid, "", false)
				Expect(delErr).To(BeNil())
			}
			_, delDepErr := departmentApiService.DeleteDepartment(ctx, depPass)
			Expect(delDepErr).To(BeNil())
			_, delDepErr2 := departmentApiService.DeleteDepartment(ctx, depFail)
			Expect(delDepErr2).To(BeNil())
		})

		It("patch department should fail because causing over-subscription of projects", func() {
			// Arrange: set both projects to use GPU=1 on the default pool (total 2, equal to department quota)
			one := 1.0
			projectPatchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectA, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(projectB, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			for i := range projectPatchReq.PatchRequests {
				projectPatchReq.PatchRequests[i].Resources.Gpu.Deserved = &one
				projectPatchReq.PatchRequests[i].Resources.Gpu.Limit = ptr.To(one)
				projectPatchReq.PatchRequests[i].Resources.Gpu.OverQuotaWeight = ptr.To(one)
			}

			projResp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, projectPatchReq)
			Expect(err).To(BeNil())
			Expect(projResp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", projResp.Body))
			projBody := projResp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(projBody.TotalRequested).To(Equal(int32(2)))
			Expect(projBody.TotalSuccessful).To(Equal(int32(2)))
			Expect(projBody.TotalFailed).To(Equal(int32(0)))

			// Act: try to shrink department GPU quota on the same nodepool below the sum of projects' quotas
			newDeptGpu := 1.0 // < aggregated 2.0 from the two projects
			depPatchReq := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  depId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: &newDeptGpu,
								Limit:    ptr.To(newDeptGpu),
							},
						},
					},
				},
			}

			depResp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, depPatchReq)
			Expect(err).To(BeNil())

			// Assert: batch returns 500 with a single failed entry due to oversubscription
			Expect(depResp.Code).To(Equal(http.StatusInternalServerError), fmt.Sprintf("got body: %+v", depResp.Body))
			depBody := depResp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(depBody.TotalRequested).To(Equal(int32(1)))
			Expect(depBody.TotalSuccessful).To(Equal(int32(0)))
			Expect(depBody.TotalFailed).To(Equal(int32(1)))
			Expect(len(depBody.Succeeded)).To(Equal(0))
			Expect(len(depBody.Failed)).To(Equal(1))

			failed := depBody.Failed[0]
			Expect(failed.Payload).To(Equal(depPatchReq.PatchRequests[0]))
			Expect(failed.Code).To(Equal(int32(http.StatusBadRequest)))
			Expect(failed.Message).To(ContainSubstring("below the sum of department's projects quotas"))
		})
	})

	Context("Projects batch patch", func() {
		It("returns 200 and mixed results when some succeed and some fail", func() {
			onePoolPatch := []api.ResourcesNullable{
				CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName),
			}
			onePoolPatch[0].Gpu.Deserved = ptr.To(float64(1.5))
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, onePoolPatch[0]),
					buildNodePoolPatchFromResources("999999", getUpdateResources()[0]),
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Succeeded)).To(Equal(1))
			Expect(len(body.Failed)).To(Equal(1))

			// Validate all relevant fields on succeeded entry
			succeeded := body.Succeeded[0]
			Expect(succeeded.Id).ToNot(BeEmpty())
			Expect(succeeded.OrgUnitId).To(Equal(projectId))
			Expect(succeeded.OrgUnitName).To(Equal("batch-proj"))
			Expect(succeeded.ClusterId).To(Equal(clusterId))
			Expect(succeeded.NodePoolId).To(Equal(defaultNodePoolId))
			Expect(succeeded.NodePoolName).To(Equal(defaultNodePoolName))
			Expect(succeeded.Rank).To(Equal(services.DefaultResourceRank))
			Expect(succeeded.Resources.Gpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Deserved).To(Equal(1.5))

			// Validate failed entry matches the bad payload and carries proper error info
			failed := body.Failed[0]
			Expect(failed.Payload).To(Equal(req.PatchRequests[1]))
			Expect(failed.Code).To(Equal(int32(http.StatusNotFound)))
			Expect(failed.Message).To(ContainSubstring("failed to retrieve"))

			// Validate updated state via project node pool resources view
			nodePoolResp, err := projectApiService.GetProjectNodePoolResources(
				ctx,
				[]string{fmt.Sprintf("orgUnitId==%s", projectId)},
				api.PROJECTNODEPOOLRESOURCESFILTERSORTFIELDS_NODE_POOL_ID,
				"asc",
				0,
				0,
				"",
			)
			Expect(err).To(BeNil())
			Expect(nodePoolResp.Code).To(Equal(http.StatusOK))
			nodePools := nodePoolResp.Body.(api.GetProjectNodePoolResources200Response).NodePoolResources
			updatedById := map[string]api.ProjectNodePoolResources{}
			for _, r := range nodePools {
				updatedById[r.NodePoolId] = r
			}
			initialById := map[string]api.Resources{}
			for _, r := range initialProjectResources {
				initialById[r.NodePool.Id] = r
			}
			Expect(*updatedById[defaultNodePoolId].Resources.Gpu.Deserved).To(Equal(1.5))
			Expect(*updatedById["2"].Resources.Gpu.Deserved).To(Equal(initialById["2"].Gpu.Deserved))
			Expect(updatedById[defaultNodePoolId].NodePoolName).To(Equal(defaultNodePoolName))
			Expect(updatedById["2"].NodePoolName).To(Equal("pool-2"))
		})

		It("returns 500 when all operations fail", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources("999999", getUpdateResources()[0]),
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))

			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Succeeded)).To(Equal(0))
			Expect(len(body.Failed)).To(Equal(1))
		})

		It("returns 400 for empty batch list", func() {
			req := api.BatchPatchNodePoolResourcesRequest{PatchRequests: []api.NodePoolResourcesPatchRequest{}}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns 400 when duplicate projectIds appear in batch", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(projectId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			Expect(resp.Body.(api.Error).Message).To(ContainSubstring("duplicate OrgUnitId//NodePoolId"))
		})

		It("return 400 when invalid rank", func() {
			invalidRank := "Ultra"
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  projectId,
						NodePoolId: defaultNodePoolId,
						Rank:       ptr.To(invalidRank),
					},
				},
			}

			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(0)))
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Code).To(BeEquivalentTo(http.StatusBadRequest))
			Expect(body.Failed[0].Message).To(ContainSubstring("unknown rank"))
		})

		It("patching only GPU keeps CPU/Memory/Rank unchanged for project", func() {
			// First, patch to populate CPU/Memory/Rank if they were nil
			newGpu := 4.0
			cpu := 5.0
			memory := 4096.0
			oqw := 2.0
			rank := "High"
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  projectId,
						NodePoolId: defaultNodePoolId,
						Rank:       ptr.To(rank),
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: &newGpu,
								Limit:    &newGpu,
							},
							Cpu: &api.ResourcesCpu{
								Deserved:        ptr.To(cpu),
								Limit:           ptr.To(cpu),
								OverQuotaWeight: ptr.To(oqw),
							},
							Memory: &api.ResourcesMemory{
								Deserved:        ptr.To(memory),
								Limit:           ptr.To(memory),
								OverQuotaWeight: ptr.To(oqw),
							},
						},
					},
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			// Now, patch only GPU fields on that nodepool
			newGpu = 9.0
			req = api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  projectId,
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: &newGpu,
								Limit:    &newGpu,
							},
						},
					},
				},
			}
			resp, err = projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))

			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(1)))
			Expect(body.TotalSuccessful).To(Equal(int32(1)))
			Expect(body.TotalFailed).To(Equal(int32(0)))
			Expect(len(body.Succeeded)).To(Equal(1))

			succeeded := body.Succeeded[0]
			Expect(succeeded.OrgUnitId).To(Equal(projectId))
			Expect(succeeded.NodePoolId).To(Equal(defaultNodePoolId))

			// GPU updated
			Expect(succeeded.Resources.Gpu.Deserved).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Deserved).To(Equal(newGpu))
			Expect(succeeded.Resources.Gpu.Limit).ToNot(BeNil())
			Expect(*succeeded.Resources.Gpu.Limit).To(Equal(newGpu))

			// CPU unchanged
			Expect(succeeded.Resources.Cpu).ToNot(BeNil())
			Expect(succeeded.Resources.Cpu.Deserved).To(Equal(ptr.To(cpu)))
			Expect(succeeded.Resources.Cpu.Limit).To(Equal(ptr.To(cpu)))
			Expect(succeeded.Resources.Cpu.OverQuotaWeight).To(Equal(ptr.To(oqw)))

			// Memory unchanged
			Expect(succeeded.Resources.Memory).ToNot(BeNil())
			Expect(succeeded.Resources.Memory.Deserved).To(Equal(ptr.To(memory)))
			Expect(succeeded.Resources.Memory.Limit).To(Equal(ptr.To(memory)))
			Expect(succeeded.Resources.Memory.OverQuotaWeight).To(Equal(ptr.To(oqw)))

			// Rank unchanged
			Expect(succeeded.Rank).To(Equal(rank))
		})
	})

	Context("Cross-cluster batch patch", func() {
		It("patches two departments from different clusters", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
					buildNodePoolPatchFromResources(departmentId2, CreateNodePoolApiResources(false, defaultNodePoolId2, defaultNodePoolName)),
				},
			}
			thirty := 30.0
			forty := 40.0
			req.PatchRequests[0].Resources.Gpu.Deserved = &thirty
			req.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(30))
			req.PatchRequests[1].Resources.Gpu.Deserved = &forty
			req.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(40))

			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", resp.Body))
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(2)))
			Expect(len(body.Succeeded)).To(Equal(2))
			for _, success := range body.Succeeded {
				if success.OrgUnitId == departmentId {
					Expect(ptr.Deref(success.Resources.Gpu.Deserved, 0)).To(Equal(30.0))
				} else if success.OrgUnitId == departmentId2 {
					Expect(ptr.Deref(success.Resources.Gpu.Deserved, 0)).To(Equal(40.0))
				} else {
					Fail("Unexpected department id in succeeded results: " + success.OrgUnitId)
				}
			}
		})

		It("patches two projects from different clusters", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, CreateNodePoolApiResources(false, "2", "pool-2")),
					buildNodePoolPatchFromResources(projectId2, CreateNodePoolApiResources(false, "20", "pool-2")),
				},
			}
			fifteen := 15.0
			twenty := 20.0
			req.PatchRequests[0].Resources.Gpu.Deserved = &fifteen
			req.PatchRequests[0].Resources.Gpu.Limit = ptr.To(float64(15))
			req.PatchRequests[1].Resources.Gpu.Deserved = &twenty
			req.PatchRequests[1].Resources.Gpu.Limit = ptr.To(float64(20))

			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK), fmt.Sprintf("got body: %+v", resp.Body))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalRequested).To(Equal(int32(2)))
			Expect(body.TotalSuccessful).To(Equal(int32(2)))
			Expect(len(body.Succeeded)).To(Equal(2))
			for _, success := range body.Succeeded {
				if success.OrgUnitId == projectId && success.NodePoolId == "2" {
					Expect(ptr.Deref(success.Resources.Gpu.Deserved, 0)).To(Equal(15.0))
				} else if success.OrgUnitId == projectId2 && success.NodePoolId == "20" {
					Expect(ptr.Deref(success.Resources.Gpu.Deserved, 0)).To(Equal(20.0))
				} else {
					Fail("Unexpected project id/nodepool id in succeeded results: " + success.OrgUnitId + "/" + success.NodePoolId)
				}
			}
		})
	})

	Context("Permission scope failures", func() {
		It("department exists but not permitted in ctx", func() {
			limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Departments: []string{departmentId2}, // exclude departmentId
			})
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(limitedCtx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Payload).To(Equal(req.PatchRequests[0]))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusNotFound)))
			Expect(body.Failed[0].Message).To(ContainSubstring("failed to get department node pool resources"))
		})

		It("project exists but not permitted in ctx", func() {
			limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
				Projects: []string{projectId2}, // exclude projectId
			})
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName)),
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(limitedCtx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Payload).To(Equal(req.PatchRequests[0]))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusNotFound)))
			Expect(body.Failed[0].Message).To(ContainSubstring("failed to retrieve"))
		})
	})

	Context("Invalid resources in patch requests", func() {
		It("department patch includes nodepool that doesn't exist", func() {
			invalid := []api.ResourcesNullable{CreateNodePoolApiResources(false, "999999", "ghost")}
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(departmentId, invalid[0]),
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusNotFound)))
		})

		It("project patch includes nodepool that doesn't exist", func() {
			invalid := []api.ResourcesNullable{CreateNodePoolApiResources(false, "999999", "ghost")}
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					buildNodePoolPatchFromResources(projectId, invalid[0]),
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(len(body.Failed)).To(Equal(1))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusBadRequest)))
		})
	})

	Context("Invalid id format in batch requests", func() {
		It("department id not a number", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  "abc",
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: ptr.To(float64(1)),
							},
						},
					},
				},
			}
			resp, err := departmentApiService.BatchPatchDepartmentNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchDepartmentResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusBadRequest)))
		})

		It("project id not a number", func() {
			req := api.BatchPatchNodePoolResourcesRequest{
				PatchRequests: []api.NodePoolResourcesPatchRequest{
					{
						OrgUnitId:  "abc",
						NodePoolId: defaultNodePoolId,
						Resources: &api.ResourcesPerNodePoolNullable{
							Gpu: &api.ResourcesPerNodePoolNullableGpu{
								Deserved: ptr.To(float64(1)),
							},
						},
					},
				},
			}
			resp, err := projectApiService.BatchPatchProjectNodePoolResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusInternalServerError))
			body := resp.Body.(api.BatchPatchProjectResourcesResponse)
			Expect(body.TotalFailed).To(Equal(int32(1)))
			Expect(body.Failed[0].Code).To(Equal(int32(http.StatusBadRequest)))
		})
	})

	Context("PatchDepartmentResources (single)", func() {
		It("patches department resources successfully", func() {
			updateResources := []api.ResourcesNullable{
				CreateNodePoolApiResources(false, defaultNodePoolId, defaultNodePoolName),
			}
			updateResources[0].Gpu.Deserved = ptr.To(float64(1.25))
			resp, err := departmentApiService.PatchDepartmentResources(ctx, departmentId, updateResources)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusOK))
			res := resp.Body.([]api.Resources)
			Expect(len(res)).To(Equal(2))
			byId := map[string]api.Resources{}
			for _, r := range res {
				byId[r.NodePool.Id] = r
			}
			Expect(byId[defaultNodePoolId].Gpu.Deserved).To(Equal(1.25))
			initialById := map[string]api.Resources{}
			for _, r := range initialDepartmentResources {
				initialById[r.NodePool.Id] = r
			}
			Expect(byId["2"].Gpu.Deserved).To(Equal(initialById["2"].Gpu.Deserved))
		})

		It("returns 400 for invalid department id format", func() {
			updateResources := getUpdateResources()
			resp, err := departmentApiService.PatchDepartmentResources(ctx, "invalid-id", updateResources)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
		})
	})
})
