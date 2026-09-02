/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services/departments"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
	"k8s.io/utils/ptr"
)

var _ = Describe("Department mapper test", Ordered, func() {
	Context("test department mapper, cpu ft off, oq ft off", func() {
		cpuFt := false
		oqFt := false
		var hasJobs bool

		It("map create / update request without jobs", func() {
			hasJobs = false
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map create / update request with jobs", func() {
			hasJobs = true
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map model to response without jobs", func() {
			hasJobs = false
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			hasJobs = true
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test department mapper, cpu ft on, oq ft off", func() {
		cpuFt := true
		oqFt := false
		var hasJobs bool

		It("map create / update request without jobs", func() {
			hasJobs = false
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map create / update request with jobs", func() {
			hasJobs = true
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map model to response without jobs", func() {
			hasJobs = false
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			hasJobs = true
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test department mapper, cpu ft off, oq ft on", func() {
		cpuFt := false
		oqFt := true
		var hasJobs bool

		It("map create / update request without jobs", func() {
			hasJobs = false
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map create / update request with jobs", func() {
			hasJobs = true
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map model to response without jobs", func() {
			hasJobs = false
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			hasJobs = true
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("test department mapper, cpu ft on, oq ft on", func() {
		cpuFt := true
		oqFt := true
		var hasJobs bool

		It("map create / update request without jobs", func() {
			hasJobs = false
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map create / update request with jobs", func() {
			hasJobs = true
			request := createDepartmentRequest(cpuFt, hasJobs)
			nodePools := CreateNodePools(cpuFt, oqFt)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(mappedReq)
			unitData := mappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)

			updateRequest := updateDepartmentRequest(cpuFt, hasJobs)
			updateMappedReq, err := departments.MapUpdateRequestToModel(nil, updateRequest, &mappedReq, nodePools)
			Expect(err).To(BeNil())

			assertDepartmentDefaults(updateMappedReq)
			unitData = updateMappedReq.OrgUnitData
			// AssertModelJobsAndNodeTypes(unitData, hasJobs)
			AssertModelDataResources(unitData, cpuFt, oqFt, true)
		})

		It("map model to response without jobs", func() {
			hasJobs = false
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})

		It("map model to response with jobs", func() {
			hasJobs = true
			department := createDepartment(hasJobs, cpuFt, oqFt)
			nodePools := CreateNodePools(cpuFt, oqFt)
			api := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			assertApiDepartmentDefaults(api)
			// AssertApiJobsAndNodeTypes(api.SchedulingRules, api.NodeTypes, hasJobs)
			AssertApiResources(api.Resources, oqFt, cpuFt, false)
		})
	})

	Context("per-nodepool flags behavior", func() {
		It("maps create/update with mixed cpu/oq flags per nodepool", func() {
			// pool1: cpu on, oq off; pool2: cpu off, oq on
			nodePools := CreateNodePoolsCustom(true, false, false, true)
			request := createDepartmentRequest(true, false)
			request.Resources = CreateRequestResourcesPerPool(true, false)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			resources := mappedReq.OrgUnitData.Resources
			Expect(len(resources)).To(Equal(2))
			pool1 := resources[0]
			pool2 := resources[1]

			// pool1: CPU present, OQ disabled => weights equal deserved
			Expect(pool1.CpuDeserved).ToNot(BeNil())
			Expect(pool1.CpuOverQuotaWeight).ToNot(BeNil())
			Expect(*pool1.CpuOverQuotaWeight).To(Equal(*pool1.CpuDeserved))
			Expect(pool1.MemoryDeserved).ToNot(BeNil())
			Expect(pool1.MemoryOverQuotaWeight).ToNot(BeNil())
			Expect(*pool1.MemoryOverQuotaWeight).To(Equal(*pool1.MemoryDeserved))
			Expect(pool1.GpuOverQuotaWeight).ToNot(BeNil())
			Expect(*pool1.GpuOverQuotaWeight).To(Equal(*pool1.GpuDeserved))

			// pool2: CPU absent, OQ enabled => gpu weight equals requested weight
			Expect(pool2.CpuDeserved).To(BeNil())
			Expect(pool2.CpuOverQuotaWeight).To(BeNil())
			Expect(pool2.MemoryDeserved).To(BeNil())
			Expect(pool2.MemoryOverQuotaWeight).To(BeNil())
			Expect(pool2.GpuOverQuotaWeight).ToNot(BeNil())
			Expect(*pool2.GpuOverQuotaWeight).To(Equal(oqWeight2))

			// update path should also work with the same nodePools
			updateReq := updateDepartmentRequest(true, false)
			updateReq.Resources = CreateRequestResourcesPerPool(true, false)
			_, err = departments.MapUpdateRequestToModel(nil, updateReq, &mappedReq, nodePools)
			Expect(err).To(BeNil())
		})

		It("maps model to api with mixed per-nodepool oq flags", func() {
			// pool1: oq off, pool2: oq on
			nodePools := CreateNodePoolsCustom(true, false, false, true)
			request := createDepartmentRequest(true, false)
			request.Resources = CreateRequestResourcesPerPool(true, false)
			mappedReq, err := departments.MapCreateRequestToModel(nil, request, nodePools)
			Expect(err).To(BeNil())

			apiDep := departments.MapModelToResponse(mappedReq, models.SlimVerbosity, nodePools)
			Expect(len(apiDep.Resources)).To(Equal(2))
			apiPool1 := apiDep.Resources[0]
			apiPool2 := apiDep.Resources[1]

			// pool1: oq disabled => weights equal deserved
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

	Context("ranks-mode minimal over-quota-weight floor when adjustable weights off", func() {
		It("floors GPU, CPU, and Memory overQuotaWeight to MinOverQuotaWeight when deserved is 0", func() {
			awFlag := false
			cpuFlag := true
			cluster, _ := uuid.Parse(clusterId)
			department := &models.Department{
				TenantId:      tenantId,
				ClusterId:     cluster,
				Name:          departmentName,
				OrgUnitDataID: uuid.New(),
				OrgUnitData: models.OrgUnitData{
					Resources: []models.Resources{
						createModelNodePoolResources(nodePoolName, nodePoolId, awFlag, cpuFlag, 0, -1, oqWeight),
					},
				},
			}
			nodePools := CreateNodePools(cpuFlag, awFlag)

			apiDep := departments.MapModelToResponse(*department, models.SlimVerbosity, nodePools)
			Expect(len(apiDep.Resources)).To(Equal(1))
			pool := apiDep.Resources[0]

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

func createDepartment(override bool, cpuFlag bool, awFlag bool) *models.Department {
	cluster, _ := uuid.Parse(clusterId)
	nodePools := pq.StringArray{nodePoolName, nodePool2Name}
	resourcesList := createModelResources(awFlag, cpuFlag)
	department := models.Department{
		TenantId:      tenantId,
		ClusterId:     cluster,
		Name:          departmentName,
		OrgUnitDataID: uuid.New(),
		OrgUnitData: models.OrgUnitData{
			Resources:           resourcesList,
			DefaultNodePoolList: &nodePools,
		},
		ProjectChildren: nil,
	}
	if override {
		limit := jobsLimit
		department.OrgUnitData.InteractiveJobTimeLimit = &limit
		department.OrgUnitData.InteractiveJobMaxIdleDuration = &limit
		department.OrgUnitData.TrainingJobMaxIdleDuration = &limit
		department.OrgUnitData.TrainingJobTimeLimit = &limit
		department.OrgUnitData.TrainingNodeTypes = &[]models.NodeType{{
			IntBaseModel: models.IntBaseModel{
				ID: 0,
			},
			Name:      "nodeType",
			ClusterId: cluster,
			TenantId:  tenantId,
		}}
	}
	return &department
}

func createModelResources(awFlag bool, cpuFlag bool) []models.Resources {
	resources1 := createModelNodePoolResources(nodePoolName, nodePoolId, awFlag, cpuFlag, deserved, limit, oqWeight)
	resources2 := createModelNodePoolResources(nodePool2Name, nodePool2Id, awFlag, cpuFlag, deserved2, limit2, oqWeight2)
	resourcesList := []models.Resources{resources1, resources2}
	return resourcesList
}

func createModelNodePoolResources(nodepoolName string, nodepoolId string, awFlag bool, cpuFlag bool, deserved float64, limit float64, oqWeight float64) models.Resources {
	if cpuFlag {
		resources := models.Resources{
			BaseModel:     dao_utils.BaseModel{},
			NodePoolId:    nodepoolId,
			NodePoolName:  nodepoolName,
			OrgUnitDataID: uuid.New(),
		}
		resources.GpuDeserved = ptr.To(deserved)
		resources.GpuLimit = ptr.To(limit)
		resources.CpuDeserved = ptr.To(deserved)
		resources.CpuLimit = ptr.To(limit)
		resources.MemoryDeserved = ptr.To(deserved)
		resources.MemoryLimit = ptr.To(limit)
		if awFlag {
			resources.GpuOverQuotaWeight = ptr.To(oqWeight)
			resources.CpuOverQuotaWeight = ptr.To(oqWeight)
			resources.MemoryOverQuotaWeight = ptr.To(oqWeight)
		} else {
			resources.GpuOverQuotaWeight = ptr.To(deserved)
			resources.CpuOverQuotaWeight = ptr.To(deserved)
			resources.MemoryOverQuotaWeight = ptr.To(deserved)
		}
		return resources
	} else {
		resources := models.Resources{
			BaseModel:     dao_utils.BaseModel{},
			NodePoolId:    nodepoolId,
			NodePoolName:  nodepoolName,
			OrgUnitDataID: uuid.New(),
		}
		resources.GpuDeserved = ptr.To(deserved)
		resources.GpuLimit = ptr.To(limit)
		if awFlag {
			resources.GpuOverQuotaWeight = ptr.To(oqWeight)
		} else {
			resources.GpuOverQuotaWeight = ptr.To(float64(2))
		}
		return resources
	}
}

func assertDepartmentDefaults(requestToModel models.Department) {
	Expect(requestToModel.Name).To(Equal(departmentName))
	Expect(requestToModel.ClusterId.String()).To(Equal(clusterId))
}

func assertApiDepartmentDefaults(requestToApi api.Department) {
	Expect(requestToApi.Name).To(Equal(departmentName))
	Expect(requestToApi.ClusterId).To(Equal(clusterId))
	// Expect(*requestToApi.DefaultNodePools).To(Equal([]string{nodePoolName, nodePool2Name}))
}

func createDepartmentRequest(cpuFt bool, withRules bool) api.DepartmentCreationRequest {
	resources := CreateRequestResources(cpuFt)
	// jobsConfig := CreateRequestJobsConfig(withRules)
	// nodeTypes := CreateRequestNodeType(withRules)
	// nodePools := []string{nodePoolName, nodePool2Name}
	request := api.DepartmentCreationRequest{
		Name: departmentName,
		// DefaultNodePools: &nodePools,
		// ParentId:           nil,
		ClusterId: clusterId,
		// SchedulingRules: jobsConfig,
		// NodeTypes:       nodeTypes,
		Resources: resources,
	}
	return request
}

func updateDepartmentRequest(cpuFt bool, withRules bool) api.DepartmentUpdateRequest {
	resources := CreateRequestResources(cpuFt)
	// jobsConfig := CreateRequestJobsConfig(withRules)
	// nodePools := []string{nodePoolName, nodePool2Name}
	// nodeTypes := CreateRequestNodeType(withRules)
	request := api.DepartmentUpdateRequest{
		// SchedulingRules:  jobsConfig,
		// DefaultNodePools: &nodePools,
		// NodeTypes:        nodeTypes,
		Resources: resources,
	}
	return request
}
