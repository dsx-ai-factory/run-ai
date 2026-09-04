/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/projects"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project mapper test", Ordered, func() {
	Context("test project mapper, cpu ft off, oq ft off", func() {
		cpuFt := false
		oqFt := false
		var override bool

		It("map create / update request without override", func() {
			override = false
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map create / update request with override", func() {
			override = true
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft on, oq ft off", func() {
		cpuFt := true
		oqFt := false
		var override bool

		It("map create / update request without override", func() {
			override = false
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		Context("per-nodepool flags behavior", func() {
			It("maps create/update with mixed cpu/oq flags per nodepool", func() {
				// pool1: cpu on, oq off; pool2: cpu off, oq on
				nodePools := CreateNodePoolsCustom(true, false, false, true)
				request := createProjectRequest(true, false)
				request.Resources = CreateRequestResourcesPerPool(true, false)
				mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
				Expect(err).To(BeNil())

				resources := mappedReq.OrgUnitData.Resources
				Expect(len(resources)).To(Equal(2))
				pool1 := resources[0]
				pool2 := resources[1]

				// pool1: CPU present, OQ disabled => weights equal default (2) for project level
				Expect(pool1.CpuDeserved).ToNot(BeNil())
				Expect(pool1.CpuOverQuotaWeight).ToNot(BeNil())
				Expect(*pool1.CpuOverQuotaWeight).To(Equal(oqDefaultWeight))
				Expect(pool1.MemoryDeserved).ToNot(BeNil())
				Expect(pool1.MemoryOverQuotaWeight).ToNot(BeNil())
				Expect(*pool1.MemoryOverQuotaWeight).To(Equal(oqDefaultWeight))
				Expect(pool1.GpuOverQuotaWeight).ToNot(BeNil())
				Expect(*pool1.GpuOverQuotaWeight).To(Equal(oqDefaultWeight))

				// pool2: CPU absent, OQ enabled => gpu weight equals requested weight
				Expect(pool2.CpuDeserved).To(BeNil())
				Expect(pool2.CpuOverQuotaWeight).To(BeNil())
				Expect(pool2.MemoryDeserved).To(BeNil())
				Expect(pool2.MemoryOverQuotaWeight).To(BeNil())
				Expect(pool2.GpuOverQuotaWeight).ToNot(BeNil())
				Expect(*pool2.GpuOverQuotaWeight).To(Equal(oqWeight2))

				// update path
				updateReq := updateProjectRequest(true, false)
				updateReq.Resources = CreateRequestResourcesPerPool(true, false)
				_, err = projects.MapUpdateRequestToModel(nil, updateReq, &mappedReq, nodePools, nil)
				Expect(err).To(BeNil())
			})

			It("maps model to api with mixed per-nodepool oq flags", func() {
				nodePools := CreateNodePoolsCustom(true, false, false, true)
				request := createProjectRequest(true, false)
				request.Resources = CreateRequestResourcesPerPool(true, false)
				mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
				Expect(err).To(BeNil())

				apiProj := projects.MapModelToResponse(context.Background(), mappedReq, nodePools)
				Expect(len(apiProj.Resources)).To(Equal(2))
				apiPool1 := apiProj.Resources[0]
				apiPool2 := apiProj.Resources[1]

				// pool1: oq disabled => weights equal deserved in API
				Expect(apiPool1.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*apiPool1.Gpu.OverQuotaWeight).To(Equal(apiPool1.Gpu.Deserved))
				Expect(apiPool1.Cpu).ToNot(BeNil())
				Expect(apiPool1.Cpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*apiPool1.Cpu.OverQuotaWeight).To(Equal(*apiPool1.Cpu.Deserved))
				Expect(apiPool1.Memory).ToNot(BeNil())
				Expect(apiPool1.Memory.OverQuotaWeight).ToNot(BeNil())
				Expect(*apiPool1.Memory.OverQuotaWeight).To(Equal(*apiPool1.Memory.Deserved))

				// pool2: oq enabled and no CPU present
				Expect(apiPool2.Cpu).To(BeNil())
				Expect(apiPool2.Memory).To(BeNil())
				Expect(apiPool2.Gpu.OverQuotaWeight).ToNot(BeNil())
				Expect(*apiPool2.Gpu.OverQuotaWeight).To(Equal(oqWeight2))
			})
		})
		It("map create / update request with override", func() {
			override = true
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft off, oq ft on", func() {
		cpuFt := false
		oqFt := true
		var override bool

		It("map create / update request without override", func() {
			override = false
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map create / update request with override", func() {
			override = true
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test project mapper, cpu ft on, oq ft on", func() {
		cpuFt := true
		oqFt := true
		var override bool

		It("map create / update request without override", func() {
			override = false
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map create / update request with override", func() {
			override = true
			request := createProjectRequest(cpuFt, override)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := projects.MapCreateRequestToModel(nil, request, models.Department{Name: parentName}, nodePools)
			Expect(err).To(BeNil())

			assertProjectDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)

			updateRequest := updateProjectRequest(cpuFt, override)
			updateMappedReq, err := projects.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools, nil)
			Expect(err).To(BeNil())

			assertProjectDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			AssertModelJobsAndNodeTypes(unitData, override)
			AssertModelDataResources(unitData, cpuFt, oqFt, false)
		})

		It("map model to response without jobs", func() {
			override = false
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			override = true
			project := createProject(override, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := projects.MapModelToResponse(context.Background(), *project, nodePools)
			assertApiProjectDefaults(api)
			AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, override)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("ranks-mode minimal over-quota-weight floor when adjustable weights off", func() {
		It("floors GPU, CPU, and Memory overQuotaWeight to MinOverQuotaWeight when deserved is 0", func() {
			awFlag := false
			cpuFlag := true
			cluster, _ := uuid.Parse(clusterId)
			nodePoolList := pq.StringArray{nodePoolName}
			project := &models.Project{
				TenantId:      tenantId,
				ClusterId:     cluster,
				Name:          projectName,
				OrgUnitDataID: uuid.New(),
				OrgUnitData: models.OrgUnitData{
					Resources: []models.Resources{
						createModelNodePoolResources(nodePoolName, nodePoolId, awFlag, cpuFlag, 0, -1, oqWeight),
					},
					DefaultNodePoolList: &nodePoolList,
					EffectiveData:       models.EffectiveData{DefaultNodePoolListEffective: &nodePoolList},
				},
			}
			nodePools := CreateNodePools(cpuFlag, awFlag)

			apiProj := projects.MapModelToResponse(context.Background(), *project, nodePools)
			Expect(len(apiProj.Resources)).To(Equal(1))
			pool := apiProj.Resources[0]

			Expect(pool.Gpu.Deserved).To(Equal(float64(0)))
			Expect(*pool.Gpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(pool.Cpu).ToNot(BeNil())
			Expect(*pool.Cpu.Deserved).To(Equal(float64(0)))
			Expect(*pool.Cpu.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
			Expect(pool.Memory).ToNot(BeNil())
			Expect(*pool.Memory.Deserved).To(Equal(float64(0)))
			Expect(*pool.Memory.OverQuotaWeight).To(Equal(services.MinOverQuotaWeight))
		})
	})
})

func assertApiProjectDefaults(requestToApi api.Project) {
	//jsonStatus,	fullStatus := v1.RawStatus{
	//	Status:        status,
	//	QuotaStatuses: nil,
	//}
	//_ := json.Marshal(fullStatus)
	Expect(requestToApi.Name).To(Equal(projectName))
	Expect(requestToApi.ClusterId).To(Equal(clusterId))
	Expect(*requestToApi.DefaultNodePools).To(Equal([]string{nodePoolName, nodePool2Name}))
	Expect(string(requestToApi.Status.Phase)).To(Equal(phase))
	Expect(requestToApi.ParentId).To(Equal(parentId))
}

func assertProjectDefaults(requestToModel models.Project) {
	Expect(requestToModel.Name).To(Equal(projectName))
	Expect(requestToModel.ClusterId.String()).To(Equal(clusterId))
}

func createProjectRequest(cpuFt bool, withOverride bool) api.ProjectCreationRequest {
	resources := CreateRequestResources(cpuFt)
	jobsConfig := CreateRequestJobsConfig(withOverride)
	nodeTypes := CreateRequestNodeType(withOverride)
	nodePools := []string{nodePoolName, nodePool2Name}
	reqParentId := parentId
	request := api.ProjectCreationRequest{
		Name:               projectName,
		RequestedNamespace: projectNamespace,
		DefaultNodePools:   &nodePools,
		ParentId:           &reqParentId,
		ClusterId:          clusterId,
		SchedulingRules:    jobsConfig,
		NodeTypes:          nodeTypes,
		Resources:          resources,
	}
	return request
}

func updateProjectRequest(cpuFt bool, withOverride bool) api.ProjectUpdateRequest {
	resources := CreateRequestResources(cpuFt)
	jobsConfig := CreateRequestJobsConfig(withOverride)
	nodeTypes := CreateRequestNodeType(withOverride)
	nodePools := []string{nodePoolName, nodePool2Name}
	request := api.ProjectUpdateRequest{
		DefaultNodePools: &nodePools,
		SchedulingRules:  jobsConfig,
		NodeTypes:        nodeTypes,
		Resources:        resources,
	}
	return request
}

func createProject(override bool, cpuFlag bool, awFlag bool) *models.Project {
	parent, _ := strconv.ParseInt(parentId, 10, 64)
	cluster, _ := uuid.Parse(clusterId)
	nodePools := pq.StringArray{nodePoolName, nodePool2Name}
	resourcesList := createModelResources(awFlag, cpuFlag)

	project := models.Project{
		TenantId:      tenantId,
		ClusterId:     cluster,
		Name:          projectName,
		OrgUnitDataID: uuid.New(),
		OrgUnitData: models.OrgUnitData{
			Resources:           resourcesList,
			DefaultNodePoolList: &nodePools,
			EffectiveData: models.EffectiveData{
				DefaultNodePoolListEffective: &nodePools,
			},
		},
		ParentID: &parent,
		ProjectStatus: models.ProjectStatus{
			Phase:           ptr.To(phase),
			StatusNamespace: ptr.To("runai-proj1"),
			Conditions: []models.Condition{
				{
					Type:               "NamespaceReady",
					Status:             "True",
					LastProbeTime:      nil,
					LastTransitionTime: ptr.To(time.Date(2024, 2, 15, 14, 58, 19, 0, time.UTC)),
				},
			},
		},
	}

	if override {
		project.OrgUnitData.InteractiveJobTimeLimit = ptr.To(jobsLimit)
		project.OrgUnitData.InteractiveJobMaxIdleDuration = ptr.To(jobsLimit)
		project.OrgUnitData.TrainingJobMaxIdleDuration = ptr.To(jobsLimit)
		project.OrgUnitData.TrainingJobTimeLimit = ptr.To(jobsLimit)
		project.OrgUnitData.InteractiveJobTimeLimitEffective = ptr.To(jobsLimit)
		project.OrgUnitData.InteractiveJobMaxIdleDurationEffective = ptr.To(jobsLimit)
		project.OrgUnitData.TrainingJobMaxIdleDurationEffective = ptr.To(jobsLimit)
		project.OrgUnitData.TrainingJobTimeLimitEffective = ptr.To(jobsLimit)
		project.OrgUnitData.TrainingNodeTypes = &[]models.NodeType{{
			IntBaseModel: models.IntBaseModel{
				ID: 0,
			},
			Name:      "nodeType",
			ClusterId: cluster,
			TenantId:  tenantId,
		}}
		project.OrgUnitData.TrainingNodeTypesEffective = &[]models.NodeType{{
			IntBaseModel: models.IntBaseModel{
				ID: 0,
			},
			Name:      "nodeType",
			ClusterId: cluster,
			TenantId:  tenantId,
		}}
	}
	return &project
}
