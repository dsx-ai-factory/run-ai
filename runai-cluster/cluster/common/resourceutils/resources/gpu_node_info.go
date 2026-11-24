/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"strings"

	resourcev1 "k8s.io/api/resource/v1"
)

const (
	// NvidiaDraDriverName is the DRA driver that publishes NVIDIA GPUs as ResourceSlices.
	NvidiaDraDriverName = "gpu.nvidia.com"

	nvidiaDraDeviceTypeAttribute = "type"
	nvidiaDraGpuDeviceType       = "gpu"
	nvidiaDraProductAttribute    = "productName"
	nvidiaDraMemoryCapacity      = "memory"
)

// NvidiaGpuNodeInfo is the GPU inventory of a node as published by the NVIDIA DRA driver. It carries the same facts
// that GPU Feature Discovery puts on node labels (nvidia.com/gpu.product, nvidia.com/gpu.count and
// nvidia.com/gpu.memory), for nodes where GFD does not run, such as GPU Operator GPUCluster (DRA-only) mode.
type NvidiaGpuNodeInfo struct {
	Product     string
	Count       int
	MemoryBytes int64
}

// NvidiaGpuNodeInfoFromResourceSlices derives the whole-GPU inventory of one node from its ResourceSlices. The slices
// must already be limited to that node. Slices of other drivers are ignored, as are devices that are not whole GPUs
// (for example MIG devices) and slices from superseded generations of a pool. It returns nil when no GPU is found.
func NvidiaGpuNodeInfoFromResourceSlices(slices []resourcev1.ResourceSlice) *NvidiaGpuNodeInfo {
	newestGeneration := map[string]int64{}
	for i := range slices {
		spec := &slices[i].Spec
		if spec.Driver != NvidiaDraDriverName {
			continue
		}
		if generation, found := newestGeneration[spec.Pool.Name]; !found || spec.Pool.Generation > generation {
			newestGeneration[spec.Pool.Name] = spec.Pool.Generation
		}
	}

	info := &NvidiaGpuNodeInfo{}
	for i := range slices {
		spec := &slices[i].Spec
		if spec.Driver != NvidiaDraDriverName || spec.Pool.Generation != newestGeneration[spec.Pool.Name] {
			continue
		}

		for j := range spec.Devices {
			device := &spec.Devices[j]
			if !isNvidiaWholeGpu(device) {
				continue
			}

			info.Count++
			if info.Product == "" {
				// GFD writes the product label with dashes in place of spaces, so match that to keep GPU types
				// comparable between nodes that have the label and nodes that only have ResourceSlices.
				info.Product = strings.Join(strings.Fields(deviceStringAttribute(device, nvidiaDraProductAttribute)), "-")
			}
			if info.MemoryBytes == 0 {
				if memory, found := device.Capacity[resourcev1.QualifiedName(nvidiaDraMemoryCapacity)]; found {
					info.MemoryBytes = memory.Value.Value()
				}
			}
		}
	}

	if info.Count == 0 {
		return nil
	}
	return info
}

// A device without a type attribute is treated as a whole GPU, since only partitions are explicitly typed otherwise.
func isNvidiaWholeGpu(device *resourcev1.Device) bool {
	deviceType := deviceStringAttribute(device, nvidiaDraDeviceTypeAttribute)
	return deviceType == "" || strings.EqualFold(deviceType, nvidiaDraGpuDeviceType)
}

// Attributes are published unqualified by the NVIDIA driver, but a driver-domain-qualified name is equivalent.
func deviceStringAttribute(device *resourcev1.Device, name string) string {
	for _, key := range []string{name, NvidiaDraDriverName + "/" + name} {
		attribute, found := device.Attributes[resourcev1.QualifiedName(key)]
		if found && attribute.StringValue != nil {
			return *attribute.StringValue
		}
	}
	return ""
}
