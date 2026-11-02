/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	nrtv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NUMA node parser", func() {
	Context("when the NRT contains NUMA and non-NUMA zones", func() {
		It("uses allocatable as total and calculates allocated resources from available", func() {
			nrt := &nrtv1alpha2.NodeResourceTopology{
				ObjectMeta: metav1.ObjectMeta{Name: "worker-1"},
				Zones: nrtv1alpha2.ZoneList{
					{
						Name: "socket-0",
						Type: "Socket",
					},
					{
						Name: "node-1",
						Type: "Node",
						Resources: nrtv1alpha2.ResourceInfoList{
							resourceInfo("cpu", "24500m", "24000m", "21500m"),
							resourceInfo("memory", "64Gi", "63Gi", "60Gi"),
							resourceInfo("nvidia.com/gpu", "4", "4", "3"),
						},
					},
					{
						Name: "node-0",
						Type: "Node",
						Resources: nrtv1alpha2.ResourceInfoList{
							resourceInfo("cpu", "24", "23500m", "24"),
							resourceInfo("memory", "32Gi", "31Gi", "32Gi"),
							resourceInfo("amd.com/gpu", "2", "2", "1"),
						},
					},
				},
			}

			result := parsers.ParseNodeResourceTopology(nrt)

			Expect(result.Name).To(Equal("worker-1"))
			Expect(result.NumaNodes).To(HaveLen(2))

			node0 := result.NumaNodes[0]
			Expect(node0.Name).To(Equal("node-0"))
			Expect(node0.TotalCpuCores).To(Equal(23.5))
			Expect(node0.AllocatedCpuCores).To(BeZero())
			Expect(node0.TotalCpuMemoryBytes).To(Equal(int64(31 * 1024 * 1024 * 1024)))
			Expect(node0.AllocatedCpuMemoryBytes).To(BeZero())
			Expect(node0.TotalGpus).To(Equal(int32(2)))
			Expect(node0.AllocatedGpus).To(Equal(int32(1)))

			node1 := result.NumaNodes[1]
			Expect(node1.Name).To(Equal("node-1"))
			Expect(node1.TotalCpuCores).To(Equal(24.0))
			Expect(node1.AllocatedCpuCores).To(Equal(2.5))
			Expect(node1.TotalCpuMemoryBytes).To(Equal(int64(63 * 1024 * 1024 * 1024)))
			Expect(node1.AllocatedCpuMemoryBytes).To(Equal(int64(3 * 1024 * 1024 * 1024)))
			Expect(node1.TotalGpus).To(Equal(int32(4)))
			Expect(node1.AllocatedGpus).To(Equal(int32(1)))
		})
	})

	Context("when the NRT contains no NUMA zones", func() {
		It("returns an empty, non-nil NUMA node list", func() {
			result := parsers.ParseNodeResourceTopology(&nrtv1alpha2.NodeResourceTopology{
				ObjectMeta: metav1.ObjectMeta{Name: "worker-1"},
				Zones: nrtv1alpha2.ZoneList{{
					Name: "socket-0",
					Type: "Socket",
				}},
			})

			Expect(result.Name).To(Equal("worker-1"))
			Expect(result.NumaNodes).NotTo(BeNil())
			Expect(result.NumaNodes).To(BeEmpty())
		})
	})
})

func resourceInfo(name, capacity, allocatable, available string) nrtv1alpha2.ResourceInfo {
	return nrtv1alpha2.ResourceInfo{
		Name:        name,
		Capacity:    resource.MustParse(capacity),
		Allocatable: resource.MustParse(allocatable),
		Available:   resource.MustParse(available),
	}
}
