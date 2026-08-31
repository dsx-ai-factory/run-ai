/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"sort"

	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	clusterconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const numaNodeZoneType = "Node"

func ParseNodeResourceTopology(nrt *nrtv1alpha2.NodeResourceTopology) openapi.NodeNumaForSync {
	numaNodes := make([]openapi.NumaNodeDetails, 0, len(nrt.Zones))
	for _, zone := range nrt.Zones {
		if zone.Type != numaNodeZoneType {
			continue
		}
		numaNodes = append(numaNodes, parseNumaZone(zone))
	}
	sort.Slice(numaNodes, func(i, j int) bool {
		return numaNodes[i].Name < numaNodes[j].Name
	})
	return openapi.NodeNumaForSync{
		Name:      nrt.Name,
		NumaNodes: numaNodes,
	}
}

func parseNumaZone(zone nrtv1alpha2.Zone) openapi.NumaNodeDetails {
	details := openapi.NumaNodeDetails{
		Name: zone.Name,
	}
	for _, resourceInfo := range zone.Resources {
		switch resourceInfo.Name {
		case string(corev1.ResourceCPU):
			details.TotalCpuCores = cpuCores(resourceInfo.Allocatable)
			details.AllocatedCpuCores = nonNegativeFloat(
				details.TotalCpuCores - cpuCores(resourceInfo.Available),
			)
		case string(corev1.ResourceMemory):
			details.TotalCpuMemoryBytes = resourceInfo.Allocatable.Value()
			details.AllocatedCpuMemoryBytes = nonNegativeInt64(
				details.TotalCpuMemoryBytes - resourceInfo.Available.Value(),
			)
		case clusterconstants.NvidiaGpuResourceName, clusterconstants.AmdGpuResourceName:
			// A physical node's GPU count cannot approach the int32 limit used by the API.
			details.TotalGpus += int32(resourceInfo.Allocatable.Value()) //nolint:gosec
			details.AllocatedGpus += int32(nonNegativeInt64(             //nolint:gosec
				resourceInfo.Allocatable.Value() - resourceInfo.Available.Value(),
			))
		}
	}
	return details
}

func cpuCores(quantity resource.Quantity) float64 {
	// Kubernetes CPU quantities use cores as their base unit. MilliValue returns
	// milli-cores, so dividing by 1000 converts them to cores. Memory quantities
	// do not need this conversion because Quantity.Value already returns bytes.
	return float64(quantity.MilliValue()) / 1000
}

func nonNegativeFloat(value float64) float64 {
	if value < 0 {
		return 0
	}
	return value
}

func nonNegativeInt64(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
