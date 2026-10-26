/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/ptr"
)

func createDefaultApiV1NodePoolResources(fullResources bool, nodePools []api.ResourcesNodePool) []api.ResourcesNullable {
	gpuLimit := float64(2)
	gpuDeserved := float64(2)
	gpuOqWeight := float64(2)
	resources := []api.ResourcesNullable{}
	for _, nodePool := range nodePools {
		npResources := api.ResourcesNullable{
			NodePool: &nodePool,
			Gpu: api.ResourcesNullableGpu{
				Deserved:        ptr.To(gpuDeserved),
				Limit:           &gpuLimit,
				OverQuotaWeight: &gpuOqWeight,
			},
		}
		if fullResources {
			limit := float64(3)
			deserved := float64(3)
			oqWeight := float64(2)
			npResources.Cpu = &api.ResourcesCpu{
				Limit:           &limit,
				Deserved:        &deserved,
				OverQuotaWeight: &oqWeight,
			}
			npResources.Memory = &api.ResourcesMemory{
				Limit:           &limit,
				Deserved:        &deserved,
				OverQuotaWeight: &oqWeight,
			}
		}
		resources = append(resources, npResources)
	}
	return resources
}

func createProjectName() string {
	return fmt.Sprint("proj-", uuid.New())
}

func createProjName() string {
	return fmt.Sprint("p-", rand.String(5))
}

func createDepName() string {
	return fmt.Sprint("d-", rand.String(5))
}

func getDefaultClustersCacheConfig() *checker.ClustersCacheConfig {
	return &checker.ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}
}

func buildDisplayedCluster(clusterId string, version string) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: clusterId,
	}
	result.Version.Set(&version)
	return result
}

func validateRank(requestedRank *string, actualRank *string, fieldName string) {
	p1, _ := services.ParseRank(requestedRank)
	p2, _ := services.ParseRank(actualRank)
	ExpectWithOffset(3, p1).To(Equal(p2), fmt.Sprintf("mismatch in %s", fieldName))
}

func validateResourcesAndNullable(requestedResources api.ResourcesNullable, actualResources api.Resources) {
	Expect(requestedResources.Cpu).To(Equal(actualResources.Cpu))
	Expect(ptr.Deref(requestedResources.Gpu.Deserved, 0)).To(Equal(actualResources.Gpu.Deserved))
	Expect(requestedResources.Gpu.Limit).To(Equal(actualResources.Gpu.Limit))
	Expect(requestedResources.Gpu.OverQuotaWeight).To(Equal(actualResources.Gpu.OverQuotaWeight))
	Expect(requestedResources.NodePool).To(Equal(actualResources.NodePool))
	Expect(requestedResources.Memory).To(Equal(actualResources.Memory))
	validateRank(requestedResources.Rank, actualResources.Rank, "rank")
}

func validateResources(resources1 api.Resources, resources2 api.Resources) {
	Expect(resources1.Cpu).To(Equal(resources2.Cpu))
	Expect(resources1.Gpu).To(Equal(resources2.Gpu))
	Expect(resources1.NodePool).To(Equal(resources2.NodePool))
	Expect(resources1.Memory).To(Equal(resources2.Memory))
	validateRank(resources1.Rank, resources2.Rank, "rank")
}

func findResourceByNodePoolId(resourcesSlice []api.Resources, nodePoolId string) *api.Resources {
	for _, resources := range resourcesSlice {
		if nodePoolId == resources.NodePool.Id {
			return &resources // Item found, return index and true
		}
	}
	return nil // Item not found, return nil
}

func validateResourcesArrayToNullable(requestedResources []api.ResourcesNullable, actualResources []api.Resources) {
	Expect(len(requestedResources)).To(Equal(len(actualResources)))
	for _, resource1 := range requestedResources {
		resource2 := findResourceByNodePoolId(actualResources, resource1.NodePool.Id)
		Expect(resource2).NotTo(BeNil())
		validateResourcesAndNullable(resource1, *resource2)
	}
}

func validateResourcesArrayTo(resources1 []api.Resources, resources2 []api.Resources) {
	Expect(len(resources1)).To(Equal(len(resources2)))
	for _, resource1 := range resources1 {
		resource2 := findResourceByNodePoolId(resources2, resource1.NodePool.Id)
		Expect(resource2).NotTo(BeNil())
		validateResources(resource1, *resource2)
	}
}

func NpSchedCapabilitiesDefault(cpuQuota, weights bool) *openapi.SchedulingCapabilities {
	return npSchedCapabilities(cpuQuota, weights, true, true)
}

func npSchedCapabilities(cpuQuota, weights bool, quota, ranks bool) *openapi.SchedulingCapabilities {
	return &openapi.SchedulingCapabilities{
		AdjustableWeights: weights,
		CpuQuota:          cpuQuota,

		Quota: quota,
		Ranks: ranks,
	}
}
