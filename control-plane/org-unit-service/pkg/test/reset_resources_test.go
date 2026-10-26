/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"
	"net/http"
	"time"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	clustersOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Org unit tests", Ordered, func() {
	ctx := SetContext(TestTenant, TestUser)

	var (
		projectApiService    api.ProjectsAPIServicer
		departmentApiService api.DepartmentsAPIServicer
	)

	Context("test ProjectsResetResources DepartmentsResetResources", func() {
		var (
			clusterServiceClient mocks.IClusterClient
			clusterId            string
			otherClusterId       string
			nodePoolName2        string
			nodePoolId2          string

			depA, depB       api.Department
			projA, projB     api.Project
			depOtherCluster  api.Department
			projOtherCluster api.Project

			localDepartmentIds []string
			localProjectIds    []string
		)

		BeforeAll(func() {
			clusterId = uuid.New().String()
			otherClusterId = uuid.New().String()
			nodePoolName2 = "pool2"
			nodePoolId2 = "20"

			clusterServiceClient = mocks.IClusterClient{}
			// cluster version supports priorities/weights
			clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
				Uuid:    clusterId,
				Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
			}, nil)
			// Nodepools listing for creation flows
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     fmt.Sprint(defaultNodePoolId),
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
					{
						Name:                   nodePoolName2,
						Id:                     nodePoolId2,
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil) // Nodepools listing for creation flows
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:  defaultNodePoolName,
						Id:    fmt.Sprint(defaultNodePoolId),
						Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					},
					{
						Name:  nodePoolName2,
						Id:    nodePoolId2,
						Phase: ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
					},
				},
			}, nil)
			// Nodepools metadata for validation helper (validateNodepoolsClusterMatching)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName, ClusterId: clusterId},
					{Id: nodePoolId2, Name: nodePoolName2, ClusterId: clusterId},
				},
			}, nil)
			// For other cluster: only default nodepool present
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, otherClusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{
						Name:                   defaultNodePoolName,
						Id:                     "10101010", // different nodepoolId
						Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
						SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
					},
				},
			}, nil)
			clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, otherClusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
				Nodepools: []clustersOpenapi.NodepoolMinimal{
					{Id: "10101010", Name: defaultNodePoolName, ClusterId: otherClusterId},
				},
			}, nil)

			// Bind services
			projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
			departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
			projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

			// Create Department A with full resources on default np, GPU-only on np2
			departmentCreateReqA := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			defaultNodepoolResourcesA := CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			defaultNodepoolResourcesA.Gpu.OverQuotaWeight = ptr.To(float64(3))
			defaultNodepoolResourcesA.Cpu.OverQuotaWeight = ptr.To(float64(4))
			defaultNodepoolResourcesA.Memory.OverQuotaWeight = ptr.To(float64(5))
			defaultNodepoolResourcesA.Rank = ptr.To("High")
			nodepool2GpuOnlyResourcesA := CreateNodePoolApiResources(true, nodePoolId2, nodePoolName2)
			nodepool2GpuOnlyResourcesA.Rank = ptr.To("VeryLow")
			departmentCreateReqA.Resources = []api.ResourcesNullable{defaultNodepoolResourcesA, nodepool2GpuOnlyResourcesA}
			createDepartmentAResponse, err := departmentApiService.CreateDepartment(ctx, departmentCreateReqA)
			Expect(err).To(BeNil())
			Expect(createDepartmentAResponse.Code).To(Equal(http.StatusCreated), fmt.Sprint("failed creating dep A: ", createDepartmentAResponse.Body))
			depA = createDepartmentAResponse.Body.(api.Department)
			localDepartmentIds = append(localDepartmentIds, depA.Id)

			// Create Department B with GPU-only default, GPU with weight on np2
			departmentCreateReqB := getDepartmentCreateRequest(createDepartmentName(), clusterId)
			defaultNodepoolGpuOnlyResourcesB := CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			defaultNodepoolGpuOnlyResourcesB.Rank = ptr.To("Low")
			nodepool2GpuWithWeightResourcesB := CreateNodePoolApiResources(true, nodePoolId2, nodePoolName2)
			nodepool2GpuWithWeightResourcesB.Gpu.OverQuotaWeight = ptr.To(float64(7))
			departmentCreateReqB.Resources = []api.ResourcesNullable{defaultNodepoolGpuOnlyResourcesB, nodepool2GpuWithWeightResourcesB}
			createDepartmentBResponse, err := departmentApiService.CreateDepartment(ctx, departmentCreateReqB)
			Expect(err).To(BeNil())
			Expect(createDepartmentBResponse.Code).To(Equal(http.StatusCreated))
			depB = createDepartmentBResponse.Body.(api.Department)
			localDepartmentIds = append(localDepartmentIds, depB.Id)

			// Create Project under depA with CPU+MEM on default, GPU-only on np2
			projectCreateReqA := getProjectCreateRequest(createProjectName(), "", depA.Id, clusterId)
			defaultNodepoolProjectResourcesA := CreateNodePoolApiResources(true, fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			defaultNodepoolProjectResourcesA.Gpu.OverQuotaWeight = ptr.To(float64(6))
			defaultNodepoolProjectResourcesA.Cpu.OverQuotaWeight = ptr.To(float64(5))
			defaultNodepoolProjectResourcesA.Memory.OverQuotaWeight = ptr.To(float64(4))
			defaultNodepoolProjectResourcesA.Rank = ptr.To("VeryHigh")
			nodepool2ProjectResourcesA := CreateNodePoolApiResourcesNoQuota(nodePoolId2, nodePoolName2)
			nodepool2ProjectResourcesA.Rank = ptr.To("Low")
			projectCreateReqA.Resources = []api.ResourcesNullable{defaultNodepoolProjectResourcesA, nodepool2ProjectResourcesA}
			createProjectAResponse, err := projectApiService.CreateProject(ctx, projectCreateReqA)
			Expect(err).To(BeNil())
			Expect(createProjectAResponse.Code).To(Equal(http.StatusCreated))
			projA = createProjectAResponse.Body.(api.Project)
			localProjectIds = append(localProjectIds, projA.Id)

			// Create Project under depB with GPU-only default and custom GPU weight on np2
			projectCreateReqB := getProjectCreateRequest(createProjectName(), "", depB.Id, clusterId)
			defaultNodepoolProjectResourcesB := CreateNodePoolApiResourcesNoQuota(fmt.Sprint(defaultNodePoolId), defaultNodePoolName)
			nodepool2ProjectResourcesB := CreateNodePoolApiResourcesNoQuota(nodePoolId2, nodePoolName2)
			nodepool2ProjectResourcesB.Gpu.OverQuotaWeight = ptr.To(float64(9))
			projectCreateReqB.Resources = []api.ResourcesNullable{defaultNodepoolProjectResourcesB, nodepool2ProjectResourcesB}
			createProjectBResponse, err := projectApiService.CreateProject(ctx, projectCreateReqB)
			Expect(err).To(BeNil())
			Expect(createProjectBResponse.Code).To(Equal(http.StatusCreated))
			projB = createProjectBResponse.Body.(api.Project)
			localProjectIds = append(localProjectIds, projB.Id)

			// Other cluster: create department and project to validate no cross-cluster impact
			departmentCreateReqOther := getDepartmentCreateRequest(createDepartmentName(), otherClusterId)
			departmentCreateReqOther.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResourcesNoQuota("10101010", defaultNodePoolName),
			}
			createDepartmentOtherResponse, err := departmentApiService.CreateDepartment(ctx, departmentCreateReqOther)
			Expect(err).To(BeNil())
			Expect(createDepartmentOtherResponse.Code).To(Equal(http.StatusCreated))
			depOtherCluster = createDepartmentOtherResponse.Body.(api.Department)
			localDepartmentIds = append(localDepartmentIds, depOtherCluster.Id)

			projectCreateReqOther := getProjectCreateRequest(createProjectName(), "", depOtherCluster.Id, otherClusterId)
			projectCreateReqOther.Resources = []api.ResourcesNullable{
				CreateNodePoolApiResources(true, "10101010", defaultNodePoolName),
			}

			createProjectOtherResponse, err := projectApiService.CreateProject(ctx, projectCreateReqOther)
			Expect(err).To(BeNil())
			Expect(createProjectOtherResponse.Code).To(Equal(http.StatusCreated))
			projOtherCluster = createProjectOtherResponse.Body.(api.Project)
			localProjectIds = append(localProjectIds, projOtherCluster.Id)
		})

		AfterAll(func() {
			for _, id := range localProjectIds {
				_, _ = projectApiService.DeleteProject(ctx, id, api.DELETIONTYPES_ORPHAN, true)
			}
			localProjectIds = []string{}
			for _, id := range localDepartmentIds {
				_, _ = departmentApiService.DeleteDepartment(ctx, id)
			}
			localDepartmentIds = []string{}
		})

		It("DepartmentsResetResources QUOTAS_TO_DEFAULT affects only requested nodepool across departments", func() {
			req := api.DepartmentsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
			}
			res, err := departmentApiService.DepartmentsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// Validate depA default np quotas zeroed, np2 unchanged
			getA, _ := departmentApiService.GetDepartment(ctx, depA.Id)
			gotA := getA.Body.(api.Department)
			var depADefaultRes, depANp2Res api.Resources
			for _, r := range gotA.Resources {
				if r.NodePool.Name == defaultNodePoolName {
					depADefaultRes = r
				}
				if r.NodePool.Name == nodePoolName2 {
					depANp2Res = r
				}
			}
			Expect(depADefaultRes.Gpu.Deserved).To(Equal(float64(0)))
			Expect(depADefaultRes.Gpu.Limit).To(BeNil())
			Expect(depADefaultRes.Cpu).ToNot(BeNil())
			Expect(depADefaultRes.Cpu.Deserved).To(Equal(ptr.To(-1.0)))
			Expect(depADefaultRes.Cpu.Limit).To(Equal(ptr.To(-1.0)))
			Expect(depADefaultRes.Memory).ToNot(BeNil())
			Expect(depADefaultRes.Memory.Deserved).To(Equal(ptr.To(-1.0)))
			Expect(depADefaultRes.Memory.Limit).To(Equal(ptr.To(-1.0)))
			// np2 untouched
			Expect(depANp2Res.Cpu).ToNot(BeNil())
			Expect(depANp2Res.Cpu.Deserved).To(Equal(ptr.To(FullResourcesDeservedGpu)))
			Expect(depANp2Res.Cpu.Limit).To(Equal(ptr.To(FullResourcesDeservedGpu)))
			Expect(depANp2Res.Memory).ToNot(BeNil())
			Expect(depANp2Res.Memory.Deserved).To(Equal(ptr.To(FullResourcesDeservedGpu)))
			Expect(depANp2Res.Memory.Limit).To(Equal(ptr.To(FullResourcesDeservedGpu)))
			// Validate depB similarly
			getB, _ := departmentApiService.GetDepartment(ctx, depB.Id)
			gotB := getB.Body.(api.Department)
			var depBDefaultRes, depBNp2Res api.Resources
			for _, r := range gotB.Resources {
				if r.NodePool.Name == defaultNodePoolName {
					depBDefaultRes = r
				}
				if r.NodePool.Name == nodePoolName2 {
					depBNp2Res = r
				}
			}
			Expect(depBDefaultRes.Gpu.Deserved).To(Equal(float64(0)))
			Expect(depBDefaultRes.Gpu.Limit).To(BeNil())
			Expect(depBDefaultRes.Cpu).ToNot(BeNil())
			Expect(depBDefaultRes.Cpu.Deserved).To(Equal(ptr.To(-1.0)))
			Expect(depBDefaultRes.Cpu.Limit).To(Equal(ptr.To(-1.0)))
			Expect(depBDefaultRes.Memory).ToNot(BeNil())
			Expect(depBDefaultRes.Memory.Deserved).To(Equal(ptr.To(-1.0)))
			Expect(depBDefaultRes.Memory.Limit).To(Equal(ptr.To(-1.0)))
			// np2 untouched
			Expect(depBNp2Res.Gpu.OverQuotaWeight).ToNot(BeNil())
			Expect(*depBNp2Res.Gpu.OverQuotaWeight).To(Equal(float64(7)))

			// Validate via list across cluster scope
			depListRes, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(depListRes.Code).To(Equal(http.StatusOK))
			for _, d := range depListRes.Body.(api.GetDepartments200Response).Departments {
				var defaultRes api.Resources
				for _, r := range d.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						defaultRes = r
					}
				}
				Expect(defaultRes.Gpu.Deserved).To(Equal(float64(0)))
			}

			// Other cluster not affected
			otherDepRes, _ := departmentApiService.GetDepartment(ctx, depOtherCluster.Id)
			for _, r := range otherDepRes.Body.(api.Department).Resources {
				if r.NodePool.Name == defaultNodePoolName {
					Expect(r.Gpu.Deserved).To(Equal(float64(0)))
				}
			}
		})

		It("DepartmentsResetResources RANKS_TO_DEFAULT sets default priority only on target nodepool", func() {
			// set depA to have non-default priorities on both nodepools
			response, err := departmentApiService.GetDepartment(ctx, depA.Id)
			Expect(err).To(BeNil())
			depABefore := response.Body.(api.Department)
			departmentUpdateReq := getDepartmentUpdateRequest(depABefore)
			for i := range departmentUpdateReq.Resources {
				departmentUpdateReq.Resources[i].Rank = ptr.To("VeryLow")
			}
			updRes, err := departmentApiService.UpdateDepartment(ctx, depA.Id, departmentUpdateReq)
			Expect(err).To(BeNil())
			Expect(updRes.Code).To(Equal(http.StatusOK))
			var beforeResourcesNpDefault, beforeResourcesNp2 api.Resources
			for _, br := range depABefore.Resources {
				if br.NodePool.Name == defaultNodePoolName {
					beforeResourcesNpDefault = br
				} else if br.NodePool.Name == nodePoolName2 {
					beforeResourcesNp2 = br
				}
			}

			req := api.DepartmentsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_RANKS_TO_DEFAULT,
			}
			res, err := departmentApiService.DepartmentsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// Validate all departments default priority reset to default, for depA - np2 stays "VeryLow"
			response, _ = departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			for _, d := range response.Body.(api.GetDepartments200Response).Departments {
				for _, r := range d.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						Expect(r.Rank).ToNot(BeNil())
						Expect(*r.Rank).To(Equal("MediumLow"))
					}
					if d.Id == depA.Id {
						if r.NodePool.Name == nodePoolName2 {
							Expect(r.Rank).ToNot(BeNil())
							Expect(*r.Rank).To(Equal("VeryLow"))
						}
						// also validate quota and weights unchanged - compare with depABefore
						if r.NodePool.Name == defaultNodePoolName {
							Expect(r.Gpu).To(Equal(beforeResourcesNpDefault.Gpu))
							Expect(r.Cpu).To(Equal(beforeResourcesNpDefault.Cpu))
							Expect(r.Memory).To(Equal(beforeResourcesNpDefault.Memory))
						} else if r.NodePool.Name == nodePoolName2 {
							Expect(r.Gpu).To(Equal(beforeResourcesNp2.Gpu))
							Expect(r.Cpu).To(Equal(beforeResourcesNp2.Cpu))
							Expect(r.Memory).To(Equal(beforeResourcesNp2.Memory))
						}
					}
				}
			}

			getA, _ := departmentApiService.GetDepartment(ctx, depA.Id)
			gotA := getA.Body.(api.Department)
			var depADefaultRes, depANp2Res api.Resources
			for _, r := range gotA.Resources {
				if r.NodePool.Name == defaultNodePoolName {
					depADefaultRes = r
				}
				if r.NodePool.Name == nodePoolName2 {
					depANp2Res = r
				}
			}
			Expect(depADefaultRes.Rank).ToNot(BeNil())
			Expect(*depADefaultRes.Rank).To(Equal("MediumLow"))
			Expect(depANp2Res.Rank).ToNot(BeNil())
			Expect(*depANp2Res.Rank).To(Equal("VeryLow"))
		})

		It("DepartmentsResetResources WEIGHTS_TO_DEFAULT sets correct weights on target nodepool", func() {
			// set weights that are different from default on all departments default nodepool
			response, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			for _, dep := range response.Body.(api.GetDepartments200Response).Departments {
				departmentUpdateReq := getDepartmentUpdateRequestFromListResponse(dep)
				for i := range departmentUpdateReq.Resources {
					if departmentUpdateReq.Resources[i].NodePool.Name == defaultNodePoolName {
						departmentUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(6))
						departmentUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(12))
						departmentUpdateReq.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(16))
						if departmentUpdateReq.Resources[i].Cpu != nil {
							departmentUpdateReq.Resources[i].Cpu.Deserved = nil
							departmentUpdateReq.Resources[i].Cpu.Limit = nil
							departmentUpdateReq.Resources[i].Cpu.OverQuotaWeight = ptr.To(float64(15))
						} else {
							departmentUpdateReq.Resources[i].Cpu = &api.ResourcesCpu{Deserved: nil, Limit: nil, OverQuotaWeight: ptr.To(float64(15))}
						}
						if departmentUpdateReq.Resources[i].Memory != nil {
							departmentUpdateReq.Resources[i].Memory.Deserved = ptr.To(float64(4))
							departmentUpdateReq.Resources[i].Memory.Limit = ptr.To(float64(4))
							departmentUpdateReq.Resources[i].Memory.OverQuotaWeight = ptr.To(float64(15))
						} else {
							departmentUpdateReq.Resources[i].Memory = &api.ResourcesMemory{Deserved: ptr.To(float64(4)), Limit: ptr.To(float64(4)), OverQuotaWeight: ptr.To(float64(15))}
						}
					}
				}
				updRes, err := departmentApiService.UpdateDepartment(ctx, dep.Id, departmentUpdateReq)
				Expect(err).To(BeNil())
				Expect(updRes.Code).To(Equal(http.StatusOK))
			}

			req := api.DepartmentsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
			}
			res, err := departmentApiService.DepartmentsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			depListRes, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(depListRes.Code).To(Equal(http.StatusOK))
			for _, d := range depListRes.Body.(api.GetDepartments200Response).Departments {
				for _, r := range d.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						// quota unchanged, weight = deserved
						Expect(r.Gpu.Deserved).To(Equal(float64(6)))
						Expect(r.Gpu.Limit).To(Equal(ptr.To(float64(12))))
						Expect(r.Gpu.OverQuotaWeight).ToNot(BeNil())
						Expect(*r.Gpu.OverQuotaWeight).To(Equal(float64(2)))

						Expect(r.Cpu).ToNot(BeNil())
						Expect(r.Cpu.Deserved).To(Equal(ptr.To(-1.0)))
						Expect(r.Cpu.Limit).To(Equal(ptr.To(-1.0)))
						Expect(r.Cpu.OverQuotaWeight).To(Equal(ptr.To(float64(2))))
						Expect(r.Memory).ToNot(BeNil())
						Expect(r.Memory.Deserved).To(Equal(ptr.To(float64(4))))
						Expect(r.Memory.Limit).To(Equal(ptr.To(float64(4))))
						Expect(r.Memory.OverQuotaWeight).To(Equal(ptr.To(float64(2))))
					}
				}
			}
		})

		It("DepartmentsResetResources WEIGHTS_TO_DESERVED sets weights equal to deserved on target nodepool", func() {
			// set weights that are different from deserved on all departments node pool 2
			response, err := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))
			for _, dep := range response.Body.(api.GetDepartments200Response).Departments {
				departmentUpdateReq := getDepartmentUpdateRequestFromListResponse(dep)
				for i := range departmentUpdateReq.Resources {
					if departmentUpdateReq.Resources[i].NodePool.Id == nodePoolId2 {
						departmentUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(5))
						departmentUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(10))
						departmentUpdateReq.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(15))
					}
				}
				updRes, err := departmentApiService.UpdateDepartment(ctx, dep.Id, departmentUpdateReq)
				Expect(err).To(BeNil())
				Expect(updRes.Code).To(Equal(http.StatusOK))
			}

			req := api.DepartmentsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{nodePoolId2},
				Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DESERVED,
			}
			res, err := departmentApiService.DepartmentsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			depListRes, _ := departmentApiService.GetDepartments(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", "", 0, 0)
			for _, d := range depListRes.Body.(api.GetDepartments200Response).Departments {
				for _, r := range d.Resources {
					if r.NodePool.Id == nodePoolId2 {
						// quota unchanged, weight = deserved
						Expect(r.Gpu.Deserved).To(Equal(float64(5)))
						Expect(r.Gpu.Limit).To(Equal(ptr.To(float64(10))))
						Expect(r.Gpu.OverQuotaWeight).ToNot(BeNil())
						Expect(*r.Gpu.OverQuotaWeight).To(Equal(r.Gpu.Deserved))
					}
				}
			}
		})

		It("ProjectsResetResources QUOTAS_TO_DEFAULT with ParentId affects only that department's projects and nodepool", func() {
			// first, update all projects in cluster to have non-zero gpu&cpu&memory quotas and non-default weights on default np to validate the reset
			projListBefore, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			for _, p := range projListBefore.Body.(api.GetProjects200Response).Projects {
				projectUpdateReq := getProjectUpdateRequest(p)
				for i := range projectUpdateReq.Resources {
					if projectUpdateReq.Resources[i].NodePool.Name == defaultNodePoolName {
						projectUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(3))
						projectUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(3))
						projectUpdateReq.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(5))
						projectUpdateReq.Resources[i].Cpu = &api.ResourcesCpu{
							Deserved:        ptr.To(float64(2)),
							Limit:           ptr.To(float64(4)),
							OverQuotaWeight: ptr.To(float64(6)),
						}
						projectUpdateReq.Resources[i].Memory = &api.ResourcesMemory{
							Deserved:        ptr.To(float64(4)),
							Limit:           ptr.To(float64(8)),
							OverQuotaWeight: ptr.To(float64(12)),
						}
					}
				}
				updateProjectResponse, err := projectApiService.UpdateProject(ctx, p.Id, projectUpdateReq)
				Expect(err).To(BeNil())
				Expect(updateProjectResponse.Code).To(Equal(http.StatusOK))
			}

			// Now perform reset on depA only

			req := api.ProjectsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
				ParentId:    depA.Id,
			}
			res, err := projectApiService.ProjectsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// Validate with list across cluster; only depA's projects changed
			projListRes, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(projListRes.Code).To(Equal(http.StatusOK))
			for _, p := range projListRes.Body.(api.GetProjects200Response).Projects {
				var defaultRes api.Resources
				for _, r := range p.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						defaultRes = r
					}
				}
				if p.ParentId == depA.Id {
					// deserved & limit - nil, OQ unchanged
					Expect(defaultRes.Gpu.Deserved).To(Equal(float64(0)))
					Expect(defaultRes.Gpu.Limit).To(BeNil())
					Expect(defaultRes.Gpu.OverQuotaWeight).To(Equal(ptr.To(float64(5))))
					Expect(defaultRes.Cpu).ToNot(BeNil())
					Expect(defaultRes.Cpu.Deserved).To(Equal(ptr.To(-1.0)))
					Expect(defaultRes.Cpu.Limit).To(Equal(ptr.To(-1.0)))
					Expect(defaultRes.Cpu.OverQuotaWeight).To(Equal(ptr.To(float64(6))))
					Expect(defaultRes.Memory).ToNot(BeNil())
					Expect(defaultRes.Memory.Deserved).To(Equal(ptr.To(-1.0)))
					Expect(defaultRes.Memory.Limit).To(Equal(ptr.To(-1.0)))
					Expect(defaultRes.Memory.OverQuotaWeight).To(Equal(ptr.To(float64(12))))
				} else {
					// unchanged on different parent
					Expect(defaultRes.Gpu.Deserved).To(Equal(float64(3)))
					Expect(defaultRes.Gpu.Limit).To(Equal(ptr.To(float64(3))))
					Expect(defaultRes.Cpu).ToNot(BeNil())
					Expect(defaultRes.Cpu.Deserved).To(Equal(ptr.To(float64(2))))
					Expect(defaultRes.Cpu.Limit).To(Equal(ptr.To(float64(4))))
					Expect(defaultRes.Memory).ToNot(BeNil())
					Expect(defaultRes.Memory.Deserved).To(Equal(ptr.To(float64(4))))
					Expect(defaultRes.Memory.Limit).To(Equal(ptr.To(float64(8))))
				}
			}

			// other cluster project unchanged
			getOtherProj, _ := projectApiService.GetProject(ctx, projOtherCluster.Id)
			for _, r := range getOtherProj.Body.(api.Project).Resources {
				if r.NodePool.Name == defaultNodePoolName {
					Expect(r.Gpu.Deserved).To(Equal(FullResourcesDeservedGpu))
					Expect(r.Gpu.Limit).To(Equal(ptr.To(FullResourcesDeservedGpu)))
					Expect(r.Cpu).ToNot(BeNil())
					Expect(r.Cpu.Deserved).To(Equal(ptr.To(FullResourcesDeservedGpu)))
					Expect(r.Memory).ToNot(BeNil())
					Expect(r.Memory.Limit).To(Equal(ptr.To(FullResourcesDeservedGpu)))
				}
			}
		})

		It("ProjectsResetResources RANKS_TO_DEFAULT sets default priority on target nodepool", func() {
			// first, update projA to have non-default priority on node pool 2
			getProjectAResponse, _ := projectApiService.GetProject(ctx, projA.Id)
			projectABefore := getProjectAResponse.Body.(api.Project)
			projectUpdateReq := getProjectUpdateRequest(projectABefore)
			for i := range projectUpdateReq.Resources {
				if projectUpdateReq.Resources[i].NodePool.Name == nodePoolName2 {
					projectUpdateReq.Resources[i].Rank = ptr.To("High")
				}
			}
			updateProjectAResponse, err := projectApiService.UpdateProject(ctx, projA.Id, projectUpdateReq)
			Expect(err).To(BeNil())
			Expect(updateProjectAResponse.Code).To(Equal(http.StatusOK))

			// Now reset priorities on nodepool2

			req := api.ProjectsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{nodePoolId2},
				Action:      api.RESOURCESRESETACTION_RANKS_TO_DEFAULT,
			}
			res, err := projectApiService.ProjectsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			projListRes, _ := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			for _, p := range projListRes.Body.(api.GetProjects200Response).Projects {
				for _, r := range p.Resources {
					if r.NodePool.Id == nodePoolId2 {
						Expect(r.Rank).ToNot(BeNil())
						Expect(*r.Rank).To(Equal("MediumLow"))
					}
				}
			}
		})

		It("ProjectsResetResources WEIGHTS_TO_DEFAULT sets oq weights to defaults where applicable", func() {
			// first, update all projects to have non-default oq weights on default np
			projListBefore, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			for _, p := range projListBefore.Body.(api.GetProjects200Response).Projects {
				projectUpdateReq := getProjectUpdateRequest(p)
				for i := range projectUpdateReq.Resources {
					if projectUpdateReq.Resources[i].NodePool.Name == defaultNodePoolName {
						projectUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(4))
						projectUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(4))
						projectUpdateReq.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(15))
						if projectUpdateReq.Resources[i].Cpu != nil {
							projectUpdateReq.Resources[i].Cpu.Deserved = nil
							projectUpdateReq.Resources[i].Cpu.Limit = nil
							projectUpdateReq.Resources[i].Cpu.OverQuotaWeight = ptr.To(float64(15))
						} else {
							projectUpdateReq.Resources[i].Cpu = &api.ResourcesCpu{Deserved: nil, Limit: nil, OverQuotaWeight: ptr.To(float64(15))}
						}
						if projectUpdateReq.Resources[i].Memory != nil {
							projectUpdateReq.Resources[i].Memory.Deserved = ptr.To(float64(4))
							projectUpdateReq.Resources[i].Memory.Limit = ptr.To(float64(4))
							projectUpdateReq.Resources[i].Memory.OverQuotaWeight = ptr.To(float64(15))
						} else {
							projectUpdateReq.Resources[i].Memory = &api.ResourcesMemory{Deserved: ptr.To(float64(4)), Limit: ptr.To(float64(4)), OverQuotaWeight: ptr.To(float64(15))}
						}
					}
				}
				updateProjectResponse, err := projectApiService.UpdateProject(ctx, p.Id, projectUpdateReq)
				Expect(err).To(BeNil())
				Expect(updateProjectResponse.Code).To(Equal(http.StatusOK))
			}

			// Now perform reset on projects for nodepool default

			req := api.ProjectsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
			}
			res, err := projectApiService.ProjectsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// validate for all projects
			response, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			Expect(response.Code).To(Equal(http.StatusOK))

			for _, p := range response.Body.(api.GetProjects200Response).Projects {
				for _, r := range p.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						// weight always defaulted to 2, cpu/memory unchanged, nil remains nil
						Expect(r.Gpu.OverQuotaWeight).ToNot(BeNil())
						Expect(*r.Gpu.OverQuotaWeight).To(Equal(float64(2)))
						Expect(r.Gpu.Deserved).To(Equal(float64(4)))
						Expect(r.Gpu.Limit).To(Equal(ptr.To(float64(4))))
						// quota was nil after quotas_to_zero
						Expect(r.Cpu).ToNot(BeNil())
						Expect(r.Cpu.Deserved).To(Equal(ptr.To(-1.0)))
						Expect(r.Cpu.Limit).To(Equal(ptr.To(-1.0)))
						Expect(r.Cpu.OverQuotaWeight).To(Equal(ptr.To(float64(2))))
						Expect(r.Memory).ToNot(BeNil())
						Expect(r.Memory.Deserved).To(Equal(ptr.To(float64(4))))
						Expect(r.Memory.Limit).To(Equal(ptr.To(float64(4))))
						Expect(r.Memory.OverQuotaWeight).To(Equal(ptr.To(float64(2))))
					}
				}
			}
		})

		It("ProjectsResetResources WEIGHTS_TO_DESERVED sets oq weights equal to deserved values", func() {
			// First, modify projB default GPU deserved to a distinct number to assert mapping
			getProjectBResponse, _ := projectApiService.GetProject(ctx, projB.Id)
			projectBBefore := getProjectBResponse.Body.(api.Project)
			projectUpdateReq := getProjectUpdateRequest(projectBBefore)
			for i := range projectUpdateReq.Resources {
				if projectUpdateReq.Resources[i].NodePool.Name == defaultNodePoolName {
					projectUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(3))
				}
			}
			updateProjectBResponse, err := projectApiService.UpdateProject(ctx, projB.Id, projectUpdateReq)
			Expect(err).To(BeNil())
			Expect(updateProjectBResponse.Code).To(Equal(http.StatusOK))

			req := api.ProjectsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DESERVED,
			}
			res, err := projectApiService.ProjectsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			projListRes, _ := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			for _, p := range projListRes.Body.(api.GetProjects200Response).Projects {
				for _, r := range p.Resources {
					if r.NodePool.Name == defaultNodePoolName {
						Expect(r.Gpu.OverQuotaWeight).ToNot(BeNil())
						Expect(*r.Gpu.OverQuotaWeight).To(Equal(r.Gpu.Deserved))
					}
				}
			}
		})

		It("validate agent request GetSyncProjects after resets", func() {
			// first, update all projects to have non-default oq weights on default np
			projListBefore, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
			Expect(err).To(BeNil())
			for _, p := range projListBefore.Body.(api.GetProjects200Response).Projects {
				projectUpdateReq := getProjectUpdateRequest(p)
				for i := range projectUpdateReq.Resources {
					if projectUpdateReq.Resources[i].NodePool.Name == defaultNodePoolName {
						projectUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(4))
						projectUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(4))
						projectUpdateReq.Resources[i].Gpu.OverQuotaWeight = ptr.To(float64(15))
						if projectUpdateReq.Resources[i].Cpu != nil {
							projectUpdateReq.Resources[i].Cpu.Deserved = nil
							projectUpdateReq.Resources[i].Cpu.Limit = nil
							projectUpdateReq.Resources[i].Cpu.OverQuotaWeight = ptr.To(float64(15))
						} else {
							projectUpdateReq.Resources[i].Cpu = &api.ResourcesCpu{Deserved: nil, Limit: nil, OverQuotaWeight: ptr.To(float64(15))}
						}
						if projectUpdateReq.Resources[i].Memory != nil {
							projectUpdateReq.Resources[i].Memory.Deserved = ptr.To(float64(4))
							projectUpdateReq.Resources[i].Memory.Limit = ptr.To(float64(4))
							projectUpdateReq.Resources[i].Memory.OverQuotaWeight = ptr.To(float64(15))
						} else {
							projectUpdateReq.Resources[i].Memory = &api.ResourcesMemory{Deserved: ptr.To(float64(4)), Limit: ptr.To(float64(4)), OverQuotaWeight: ptr.To(float64(15))}
						}
					}
				}
				updateProjectResponse, err := projectApiService.UpdateProject(ctx, p.Id, projectUpdateReq)
				Expect(err).To(BeNil())
				Expect(updateProjectResponse.Code).To(Equal(http.StatusOK))

				// update status - else all the project will always return... (WHERE ... status_last_updated_time is null OR ....)
				updateStatusRequest := api.ProjectStatusUpdateRequest{
					Id:        p.Id,
					ClusterId: clusterId,
					Status: api.ProjectStatus{
						LastUpdatedTime: ptr.To(time.Now()),
						Phase:           api.PROJECTPHASE_READY,
					},
				}
				updateStatusRes, err := projectApiService.UpdateProjectStatus(ctx, updateStatusRequest)
				Expect(err).To(BeNil())
				Expect(updateStatusRes.Code).To(Equal(http.StatusOK))
			}

			// get projects sync request
			getResponse, err := projectApiService.GetSyncProjects(ctx, clusterId, time.Time{}, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects := getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(2))

			timeBefore := time.Now()
			time.Sleep(time.Duration(2) * time.Millisecond)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBefore, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(0)) // no updates since last check

			// Now perform reset on projects for nodepool default - department A
			req := api.ProjectsResourcesResetRequest{
				ClusterId:   clusterId,
				NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
				ParentId:    depA.Id,
				Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
			}
			res, err := projectApiService.ProjectsResetResources(ctx, req)
			Expect(err).To(BeNil())
			Expect(res.Code).To(Equal(http.StatusNoContent))

			// get projects sync request again with timeBefore
			time.Sleep(time.Duration(2) * time.Millisecond)
			getResponse, err = projectApiService.GetSyncProjects(ctx, clusterId, timeBefore, 0, 0, true)
			Expect(err).To(BeNil())
			Expect(getResponse.Code).To(Equal(http.StatusOK))
			syncProjects = getResponse.Body.(api.GetSyncProjects200Response).Projects
			Expect(len(syncProjects)).To(Equal(1)) // only depA's project updated
			Expect(syncProjects[0].Id).To(Equal(projA.Id))
		})

		Context("some tests with 3 nodepools", func() {
			var nodePool3Name, nodePool3Id string

			BeforeAll(func() {
				nodePool3Name = "pool-3"
				nodePool3Id = "33"

				clusterServiceClient = mocks.IClusterClient{}
				// cluster version supports priorities/weights
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					Uuid:    clusterId,
					Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				}, nil)
				// Nodepools listing for creation flows
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
						{
							Name:                   nodePoolName2,
							Id:                     nodePoolId2,
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
						{
							Name:                   nodePool3Name,
							Id:                     nodePool3Id,
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
					},
				}, nil)
				// Nodepools metadata for validation helper (validateNodepoolsClusterMatching)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName, ClusterId: clusterId},
						{Id: nodePoolId2, Name: nodePoolName2, ClusterId: clusterId},
						{Id: nodePool3Id, Name: nodePool3Name, ClusterId: clusterId},
					},
				}, nil)

				// Bind services
				projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
				departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

				notificationsService := services.NewNotificationsApiService(&clusterServiceClient)
				updated, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
					Id:        nodePool3Id,
					Name:      nodePool3Name,
					ClusterId: clusterId,
					Operation: services.CreatedOperation,
				})
				Expect(err).To(BeNil())
				Expect(updated.Code).To(Equal(http.StatusOK))
			})

			AfterAll(func() {
				// revert to previous state with 2 nodepools
				clusterServiceClient = mocks.IClusterClient{}
				// cluster version supports priorities/weights
				clusterServiceClient.On("GetClusterMinimalDataById", mock.Anything, mock.Anything).Return(&clustersOpenapi.ClustersMinimalData{
					Uuid:    clusterId,
					Version: *clustersOpenapi.NewNullableString(ptr.To("2.22")),
				}, nil)
				// Nodepools listing for creation flows
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     fmt.Sprint(defaultNodePoolId),
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
						{
							Name:                   nodePoolName2,
							Id:                     nodePoolId2,
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
					},
				}, nil)
				// Nodepools metadata for validation helper (validateNodepoolsClusterMatching)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, clusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{Id: fmt.Sprint(defaultNodePoolId), Name: defaultNodePoolName, ClusterId: clusterId},
						{Id: nodePoolId2, Name: nodePoolName2, ClusterId: clusterId},
					},
				}, nil)
				// For other cluster: only default nodepool present
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, otherClusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{
							Name:                   defaultNodePoolName,
							Id:                     "10101010", // different nodepoolId
							Phase:                  ptr.To(clustersOpenapi.NODEPOOLPHASE_READY),
							SchedulingCapabilities: NpSchedCapabilitiesDefault(true, true),
						},
					},
				}, nil)
				clusterServiceClient.On("GetNodePoolsMinimalByCluster", mock.Anything, otherClusterId).Return(&clustersOpenapi.GetNodepoolsMinimal200Response{
					Nodepools: []clustersOpenapi.NodepoolMinimal{
						{Id: "10101010", Name: defaultNodePoolName, ClusterId: otherClusterId},
					},
				}, nil)

				// Bind services
				projectDeletionService := services.NewProjectDeletionService(&clusterServiceClient, nil, nil)
				departmentApiService = departments.NewDepartmentApiService(&clusterServiceClient, nil)
				projectApiService = projects.NewProjectApiService(projectDeletionService, &clusterServiceClient, nil)

				notificationsService := services.NewNotificationsApiService(&clusterServiceClient)
				updated, err := notificationsService.NodePoolUpdated(ctx, false, api.NodePool{
					Id:        nodePool3Id,
					Name:      nodePool3Name,
					ClusterId: clusterId,
					Operation: services.DeletedOperation,
				})
				Expect(err).To(BeNil())
				Expect(updated.Code).To(Equal(http.StatusOK))
			})

			It("ProjectsResetResources QUOTAS_TO_DEFAULT with 2 nodepools out of 3 - validate only those affected", func() {
				// first, update all projects in cluster to have non-zero gpu&cpu&memory quotas and non-default weights on all nodepools to validate the reset
				projListBefore, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
				Expect(err).To(BeNil())
				for _, p := range projListBefore.Body.(api.GetProjects200Response).Projects {
					projectUpdateReq := getProjectUpdateRequest(p)
					for i := range projectUpdateReq.Resources {
						projectUpdateReq.Resources[i].Gpu.Deserved = ptr.To(float64(3))
						projectUpdateReq.Resources[i].Gpu.Limit = ptr.To(float64(3))
						projectUpdateReq.Resources[i].Cpu = &api.ResourcesCpu{
							Deserved: ptr.To(float64(2)),
							Limit:    ptr.To(float64(4)),
						}
						projectUpdateReq.Resources[i].Memory = &api.ResourcesMemory{
							Deserved: ptr.To(float64(4)),
							Limit:    ptr.To(float64(8)),
						}
					}
					updateProjectResponse, err := projectApiService.UpdateProject(ctx, p.Id, projectUpdateReq)
					Expect(err).To(BeNil())
					Expect(updateProjectResponse.Code).To(Equal(http.StatusOK))
				}

				// now, perform reset on 2 of the 3 nodepools
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId), nodePoolId2},
					Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNoContent))

				// validate for all projects - the 2 nodepools affected, and the third one remains unchanged
				response, err := projectApiService.GetProjects(ctx, []string{fmt.Sprintf("clusterId==%s", clusterId)}, "", "", 0, 0, "")
				Expect(err).To(BeNil())
				Expect(response.Code).To(Equal(http.StatusOK))

				for _, p := range response.Body.(api.GetProjects200Response).Projects {
					for _, r := range p.Resources {
						if r.NodePool.Id == fmt.Sprint(defaultNodePoolId) || r.NodePool.Id == nodePoolId2 {
							// deserved & limit - nil
							Expect(r.Gpu.Deserved).To(Equal(float64(0)))
							Expect(r.Gpu.Limit).To(BeNil())
							Expect(r.Cpu).To(BeNil())
							Expect(r.Memory).To(BeNil())
						} else {
							// unchanged on third nodepool
							Expect(r.Gpu.Deserved).To(Equal(float64(3)))
							Expect(r.Gpu.Limit).To(Equal(ptr.To(float64(3))))
							Expect(r.Cpu).ToNot(BeNil())
							Expect(r.Cpu.Deserved).To(Equal(ptr.To(float64(2))))
							Expect(r.Cpu.Limit).To(Equal(ptr.To(float64(4))))
							Expect(r.Memory).ToNot(BeNil())
							Expect(r.Memory.Deserved).To(Equal(ptr.To(float64(4))))
							Expect(r.Memory.Limit).To(Equal(ptr.To(float64(8))))
						}
					}
				}
			})
		})

		Context("test invalid requests", func() {
			It("ProjectsResetResources returns 400 for invalid cluster id", func() {
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   "bad-uuid",
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("invalid UUID"))
			})

			It("DepartmentsResetResources returns 400 for invalid cluster id", func() {
				req := api.DepartmentsResourcesResetRequest{
					ClusterId:   "bad-uuid2",
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_QUOTAS_TO_DEFAULT,
				}
				res, err := departmentApiService.DepartmentsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("invalid UUID"))
			})

			It("ProjectsResetResources returns 400 for nodepool mismatch", func() {
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{"8888"},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("do not match"))
			})

			It("DepartmentsResetResources returns 400 for nodepool mismatch", func() {
				req := api.DepartmentsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{"99999"},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := departmentApiService.DepartmentsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("do not match"))
			})

			It("ProjectsResetResources returns 400 for non-numeric ParentId", func() {
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DESERVED,
					ParentId:    "abc",
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("failed to parse parent id"))
			})

			It("ProjectsResetResources returns 400 when ParentId and ClusterId mismatch", func() {
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId, // different from depC's cluster
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_RANKS_TO_DEFAULT,
					ParentId:    depOtherCluster.Id,
				}
				res, err := projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("do not match"))

				req.ParentId = "98989898" // non-existent department in the cluster
				res, err = projectApiService.ProjectsResetResources(ctx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusBadRequest))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("do not match"))
			})
		})

		Context("test permission scopes - validateResetResourcesPermitted", func() {
			It("DepartmentsResetResources returns 403 when unpermitted", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Departments: []string{depA.Id}, // not all departments in the cluster
				})
				req := api.DepartmentsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := departmentApiService.DepartmentsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("not authorized"))
			})

			It("DepartmentsResetResources returns 403 when unpermitted - with permission for all projects, that's not good enough", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Projects: []string{projA.Id, projB.Id}, // all projects in cluster, but missing depB
				})
				req := api.DepartmentsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := departmentApiService.DepartmentsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("not authorized"))
			})

			It("DepartmentsResetResources permitted when permissions for all departments separately", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Departments: []string{depA.Id, depB.Id}, // all departments in the cluster
				})
				req := api.DepartmentsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := departmentApiService.DepartmentsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("ProjectsResetResources returns 403 when unpermitted - in department scope", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					// no permission to depA projects
					Departments: []string{depB.Id},
					// permission for only one of the 2 projects under depA
					Projects: []string{projB.Id},
				})
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
					ParentId:    depA.Id,
				}
				res, err := projectApiService.ProjectsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
				Expect(fmt.Sprint(res.Body)).To(ContainSubstring("not authorized"))
			})

			It("ProjectsResetResources permitted in department scope when having permissions for all projects in dep separately", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					Projects: []string{projB.Id, projA.Id}, // all projects under depA
				})
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
					ParentId:    depA.Id,
				}
				res, err := projectApiService.ProjectsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})

			It("ProjectsResetResources returns 403 when unpermitted - in cluster scope (parentId empty)", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					// only permitted to projects under depA (subset of cluster)
					Departments: []string{depA.Id},
				})
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusForbidden))
			})

			It("Validate permitted with permitted scopes of all departments in cluster separately", func() {
				limitedCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, &authorization_api.PermittedScopes{
					// permitted scopes being al departments separately - should pass
					Departments: []string{depA.Id, depB.Id},
				})
				req := api.ProjectsResourcesResetRequest{
					ClusterId:   clusterId,
					NodePoolIds: []string{fmt.Sprint(defaultNodePoolId)},
					Action:      api.RESOURCESRESETACTION_WEIGHTS_TO_DEFAULT,
				}
				res, err := projectApiService.ProjectsResetResources(limitedCtx, req)
				Expect(err).To(BeNil())
				Expect(res.Code).To(Equal(http.StatusNoContent))
			})
		})
	})
})
