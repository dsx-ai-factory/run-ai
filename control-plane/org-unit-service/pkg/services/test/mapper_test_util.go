/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"strconv"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	"k8s.io/utils/ptr"
)

const (
	departmentName   = "dep1"
	projectName      = "proj1"
	projectNamespace = "proj1-ns"
	parentId         = "123123123"
	parentName       = "dep-123123123"
	tenantId         = "1"
	subjectId        = "a@a"
	clusterId        = "7a15a2f0-89fb-45f3-ba75-b42125391234"
	nodePoolId       = "1"
	nodePoolName     = "default"
	nodePool2Id      = "2"
	nodePool2Name    = "pool2"
	limit            = float64(3)
	deserved         = float64(3)
	oqWeight         = float64(3)
	limit2           = float64(-1)
	deserved2        = float64(-1)
	oqWeight2        = float64(-1)
	oqDefaultWeight  = float64(2)
	jobsLimit        = int32(2000)
	nodeTypeId       = "0"
	nodeType         = "nodeType"
	phase            = "Ready"
	statusNamespace  = "runai-proj1"
	statusMessage    = ""
)

func AssertApiJobsAndNodeTypes(schedulingRules *api.SchedulingRules, nodeTypes api.NodeTypesPerWorkload, override bool) {
	if !override {
		Expect(schedulingRules).To(BeNil())
		Expect(len(nodeTypes.Training)).To(Equal(0))
		Expect(len(nodeTypes.Workspace)).To(Equal(0))
	} else {
		Expect(*schedulingRules.InteractiveJobMaxIdleDurationSeconds).To(Equal(jobsLimit))
		Expect(*schedulingRules.InteractiveJobTimeLimitSeconds).To(Equal(jobsLimit))
		Expect(schedulingRules.InteractiveJobPreemptIdleDurationSeconds).To(BeNil())
		Expect(*schedulingRules.TrainingJobMaxIdleDurationSeconds).To(Equal(jobsLimit))
		Expect(*schedulingRules.TrainingJobTimeLimitSeconds).To(Equal(jobsLimit))
		Expect(len(nodeTypes.Training)).To(Equal(1))
		Expect(len(nodeTypes.Workspace)).To(Equal(0))
		Expect(nodeTypes.Training[0]).To(Equal(nodeTypeId))
	}
}

func AssertModelJobsAndNodeTypes(unitData models.OrgUnitData, override bool) {
	if !override {
		Expect(unitData.InteractiveJobMaxIdleDuration).To(BeNil())
		Expect(unitData.InteractiveJobTimeLimit).To(BeNil())
		Expect(unitData.InteractiveJobPreemptIdleDuration).To(BeNil())
		Expect(unitData.TrainingJobMaxIdleDuration).To(BeNil())
		Expect(unitData.TrainingJobTimeLimit).To(BeNil())
		Expect(unitData.WorkspaceNodeTypes).To(BeNil())
		Expect(unitData.TrainingNodeTypes).To(BeNil())
	} else {
		Expect(*unitData.InteractiveJobMaxIdleDuration).To(Equal(jobsLimit))
		Expect(*unitData.InteractiveJobTimeLimit).To(Equal(jobsLimit))
		Expect(unitData.InteractiveJobPreemptIdleDuration).To(BeNil())
		Expect(*unitData.TrainingJobMaxIdleDuration).To(Equal(jobsLimit))
		Expect(*unitData.TrainingJobTimeLimit).To(Equal(jobsLimit))
		Expect(unitData.WorkspaceNodeTypes).To(BeNil())
		Expect(unitData.TrainingNodeTypes).NotTo(BeNil())
		Expect(len(*unitData.TrainingNodeTypes)).To(Equal(1))
		Expect(strconv.FormatInt((*unitData.TrainingNodeTypes)[0].ID, 10)).To(Equal(nodeTypeId))
	}
}

func AssertModelDataResources(unitData models.OrgUnitData, cpuFlag bool, awFlag bool, isDepartment bool) {
	resources := unitData.Resources
	Expect(len(resources)).To(Equal(2))
	defaultResources := resources[0]
	pool1Resources := resources[1]
	validateNodePoolResources(defaultResources, awFlag, cpuFlag, deserved, limit, oqWeight, isDepartment)
	validateNodePoolResources(pool1Resources, awFlag, cpuFlag, deserved2, limit2, oqWeight2, isDepartment)
}

func AssertApiResources(resources []api.Resources, awFlag bool, cpuFlag bool, overrideDeserved bool) {
	Expect(len(resources)).To(Equal(2))
	defaultResources := resources[0]
	validateApiNodepoolResources(defaultResources, awFlag, cpuFlag, deserved, limit, oqWeight, overrideDeserved)
	poolResources := resources[1]
	validateApiNodepoolResources(poolResources, awFlag, cpuFlag, deserved2, limit2, oqWeight2, overrideDeserved)
}

func validateApiNodepoolResources(resources api.Resources, awFlag bool, cpuFlag bool, deserved float64, limit float64, oqWeight float64, overrideDeserved bool) {
	if overrideDeserved {
		projectsDeserved := float64(-1)
		Expect(resources.Gpu.Deserved).To(Equal(projectsDeserved))
		if awFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(expectedOQWhenAwOff(projectsDeserved)))
		}
	} else {
		Expect(resources.Gpu.Deserved).To(Equal(deserved))
		Expect(*resources.Gpu.Limit).To(Equal(limit))
		if awFlag {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Gpu.OverQuotaWeight).To(Equal(expectedOQWhenAwOff(deserved)))
		}
	}
	if cpuFlag {
		Expect(*resources.Cpu.Deserved).To(Equal(deserved))
		Expect(*resources.Cpu.Limit).To(Equal(limit))
		Expect(*resources.Memory.Deserved).To(Equal(deserved))
		Expect(*resources.Memory.Limit).To(Equal(limit))
		if awFlag {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(oqWeight))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(oqWeight))
		} else {
			Expect(*resources.Cpu.OverQuotaWeight).To(Equal(expectedOQWhenAwOff(deserved)))
			Expect(*resources.Memory.OverQuotaWeight).To(Equal(expectedOQWhenAwOff(deserved)))
		}
	} else {
		Expect(resources.Cpu).To(BeNil())
		Expect(resources.Memory).To(BeNil())
	}
}

// expectedOQWhenAwOff mirrors the read-side floor in services.ApiResourceFromModel:
// when adjustable weights are off, OverQuotaWeight equals deserved, except deserved == 0
// which is floored to services.MinOverQuotaWeight so the scheduler can still fairshare.
func expectedOQWhenAwOff(deserved float64) float64 {
	if deserved == 0 {
		return services.MinOverQuotaWeight
	}
	return deserved
}

func validateNodePoolResources(resources models.Resources, awFlag bool, cpuFlag bool, deserved float64, limit float64, oqWeight float64, isDepartment bool) {
	Expect(*resources.GpuDeserved).To(Equal(deserved))
	Expect(*resources.GpuLimit).To(Equal(limit))
	if awFlag {
		Expect(*resources.GpuOverQuotaWeight).To(Equal(oqWeight))
	} else {
		if isDepartment {
			Expect(*resources.GpuOverQuotaWeight).To(Equal(deserved))
		} else {
			Expect(*resources.GpuOverQuotaWeight).To(Equal(oqDefaultWeight))
		}
	}
	if cpuFlag {
		Expect(*resources.CpuDeserved).To(Equal(deserved))
		Expect(*resources.CpuLimit).To(Equal(limit))
		Expect(*resources.MemoryDeserved).To(Equal(deserved))
		Expect(*resources.MemoryLimit).To(Equal(limit))
		if awFlag {
			Expect(*resources.CpuOverQuotaWeight).To(Equal(oqWeight))
			Expect(*resources.MemoryOverQuotaWeight).To(Equal(oqWeight))
		} else {
			if isDepartment {
				Expect(*resources.CpuOverQuotaWeight).To(Equal(deserved))
				Expect(*resources.MemoryOverQuotaWeight).To(Equal(deserved))
			} else {
				Expect(*resources.CpuOverQuotaWeight).To(Equal(oqDefaultWeight))
				Expect(*resources.MemoryOverQuotaWeight).To(Equal(oqDefaultWeight))
			}
		}
	} else {
		Expect(resources.CpuDeserved).To(BeNil())
		Expect(resources.CpuLimit).To(BeNil())
		Expect(resources.CpuOverQuotaWeight).To(BeNil())
		Expect(resources.MemoryDeserved).To(BeNil())
		Expect(resources.MemoryLimit).To(BeNil())
		Expect(resources.MemoryOverQuotaWeight).To(BeNil())
	}
}

func AssertModelDataNoOverride(unitData models.OrgUnitData) {
	Expect(unitData.InteractiveJobMaxIdleDuration).To(BeNil())
	Expect(unitData.InteractiveJobTimeLimit).To(BeNil())
	Expect(unitData.InteractiveJobPreemptIdleDuration).To(BeNil())
	Expect(unitData.TrainingJobMaxIdleDuration).To(BeNil())
	Expect(unitData.TrainingJobTimeLimit).To(BeNil())
	resources := (unitData.Resources)[0]
	Expect(*resources.GpuDeserved).To(Equal(deserved))
	Expect(*resources.GpuLimit).To(Equal(limit))
	Expect(*resources.GpuOverQuotaWeight).To(Equal(oqWeight))
	Expect(resources.CpuDeserved).To(BeNil())
	Expect(resources.CpuLimit).To(BeNil())
	Expect(resources.CpuOverQuotaWeight).To(BeNil())
	Expect(resources.MemoryDeserved).To(BeNil())
	Expect(resources.MemoryLimit).To(BeNil())
	Expect(resources.MemoryOverQuotaWeight).To(BeNil())
}

func CreateRequestResources(withCpu bool) []api.ResourcesNullable {
	reqLimit := limit
	reqDeserved := deserved
	reqOqWeight := oqWeight
	req2Limit := limit2
	req2Deserved := deserved2
	req2OqWeight := oqWeight2
	resources := []api.ResourcesNullable{{
		NodePool: &api.ResourcesNodePool{
			Id:   nodePoolId,
			Name: nodePoolName,
		},
		Gpu: api.ResourcesNullableGpu{
			Limit:           &reqLimit,
			Deserved:        ptr.To(reqDeserved),
			OverQuotaWeight: &reqOqWeight,
		},
	}, {
		NodePool: &api.ResourcesNodePool{
			Id:   nodePool2Id,
			Name: nodePool2Name,
		},
		Gpu: api.ResourcesNullableGpu{
			Limit:           &req2Limit,
			Deserved:        ptr.To(req2Deserved),
			OverQuotaWeight: &req2OqWeight,
		},
	}}
	if withCpu {
		resources[0].Cpu = &api.ResourcesCpu{
			Limit:           &reqLimit,
			Deserved:        &reqDeserved,
			OverQuotaWeight: &reqOqWeight,
		}
		resources[0].Memory = &api.ResourcesMemory{
			Limit:           &reqLimit,
			Deserved:        &reqDeserved,
			OverQuotaWeight: &reqOqWeight,
		}
		resources[1].Cpu = &api.ResourcesCpu{
			Limit:           &req2Limit,
			Deserved:        &req2Deserved,
			OverQuotaWeight: &req2OqWeight,
		}
		resources[1].Memory = &api.ResourcesMemory{
			Limit:           &req2Limit,
			Deserved:        &req2Deserved,
			OverQuotaWeight: &req2OqWeight,
		}
	}
	return resources
}

func CreateRequestJobsConfig(withRules bool) *api.SchedulingRules {
	if withRules {
		reqJobsLimit := jobsLimit
		return &api.SchedulingRules{
			InteractiveJobTimeLimitSeconds:       &reqJobsLimit,
			InteractiveJobMaxIdleDurationSeconds: &reqJobsLimit,
			TrainingJobMaxIdleDurationSeconds:    &reqJobsLimit,
			TrainingJobTimeLimitSeconds:          &reqJobsLimit,
		}
	}
	return nil
}

func CreateRequestNodeType(nodeTypeDefined bool) api.NodeTypesPerWorkload {
	if nodeTypeDefined {
		return api.NodeTypesPerWorkload{
			Training:  []string{nodeTypeId},
			Workspace: nil,
		}
	}
	return api.NodeTypesPerWorkload{}
}

// CreateNodePools builds a minimal nodePools map keyed by id, with per-nodepool
// scheduling capabilities reflecting test flags.
func CreateNodePools(cpuFlag bool, awFlag bool) map[string]openapi.NodepoolMinimal {
	np1 := openapi.NodepoolMinimal{
		Id:   nodePoolId,
		Name: nodePoolName,
	}
	np2 := openapi.NodepoolMinimal{
		Id:   nodePool2Id,
		Name: nodePool2Name,
	}
	np1.SchedulingCapabilities = &openapi.SchedulingCapabilities{
		AdjustableWeights: awFlag,
		CpuQuota:          cpuFlag,
		// Other fields (Quota, Ranks) are irrelevant for these tests
	}
	np2.SchedulingCapabilities = &openapi.SchedulingCapabilities{
		AdjustableWeights: awFlag,
		CpuQuota:          cpuFlag,
	}

	return map[string]openapi.NodepoolMinimal{
		nodePoolId:  np1,
		nodePool2Id: np2,
	}
}

// CreateNodePoolsCustom builds a nodePools map with per-nodepool flags.
func CreateNodePoolsCustom(cpu1 bool, oq1 bool, cpu2 bool, oq2 bool) map[string]openapi.NodepoolMinimal {
	np1 := openapi.NodepoolMinimal{
		Id:   nodePoolId,
		Name: nodePoolName,
	}
	np2 := openapi.NodepoolMinimal{
		Id:   nodePool2Id,
		Name: nodePool2Name,
	}
	np1.SchedulingCapabilities = &openapi.SchedulingCapabilities{
		AdjustableWeights: oq1,
		CpuQuota:          cpu1,
	}
	np2.SchedulingCapabilities = &openapi.SchedulingCapabilities{
		AdjustableWeights: oq2,
		CpuQuota:          cpu2,
	}
	return map[string]openapi.NodepoolMinimal{
		nodePoolId:  np1,
		nodePool2Id: np2,
	}
}

// CreateRequestResourcesPerPool builds two resources entries, controlling CPU/Memory presence per nodepool.
func CreateRequestResourcesPerPool(cpuForPool1 bool, cpuForPool2 bool) []api.ResourcesNullable {
	reqLimit := limit
	reqDeserved := deserved
	reqOqWeight := oqWeight
	req2Limit := limit2
	req2Deserved := deserved2
	req2OqWeight := oqWeight2
	resources := []api.ResourcesNullable{{
		NodePool: &api.ResourcesNodePool{
			Id:   nodePoolId,
			Name: nodePoolName,
		},
		Gpu: api.ResourcesNullableGpu{
			Limit:           &reqLimit,
			Deserved:        ptr.To(reqDeserved),
			OverQuotaWeight: &reqOqWeight,
		},
	}, {
		NodePool: &api.ResourcesNodePool{
			Id:   nodePool2Id,
			Name: nodePool2Name,
		},
		Gpu: api.ResourcesNullableGpu{
			Limit:           &req2Limit,
			Deserved:        ptr.To(req2Deserved),
			OverQuotaWeight: &req2OqWeight,
		},
	}}
	if cpuForPool1 {
		resources[0].Cpu = &api.ResourcesCpu{
			Limit:           &reqLimit,
			Deserved:        &reqDeserved,
			OverQuotaWeight: &reqOqWeight,
		}
		resources[0].Memory = &api.ResourcesMemory{
			Limit:           &reqLimit,
			Deserved:        &reqDeserved,
			OverQuotaWeight: &reqOqWeight,
		}
	}
	if cpuForPool2 {
		resources[1].Cpu = &api.ResourcesCpu{
			Limit:           &req2Limit,
			Deserved:        &req2Deserved,
			OverQuotaWeight: &req2OqWeight,
		}
		resources[1].Memory = &api.ResourcesMemory{
			Limit:           &req2Limit,
			Deserved:        &req2Deserved,
			OverQuotaWeight: &req2OqWeight,
		}
	}
	return resources
}
