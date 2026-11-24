/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

func gpuDevice(name, deviceType, product, memory string) resourcev1.Device {
	device := resourcev1.Device{
		Name:       name,
		Attributes: map[resourcev1.QualifiedName]resourcev1.DeviceAttribute{},
		Capacity:   map[resourcev1.QualifiedName]resourcev1.DeviceCapacity{},
	}
	if deviceType != "" {
		device.Attributes["type"] = resourcev1.DeviceAttribute{StringValue: ptr.To(deviceType)}
	}
	if product != "" {
		device.Attributes["productName"] = resourcev1.DeviceAttribute{StringValue: ptr.To(product)}
	}
	if memory != "" {
		device.Capacity["memory"] = resourcev1.DeviceCapacity{Value: resource.MustParse(memory)}
	}
	return device
}

func resourceSlice(driver, pool string, generation int64, devices ...resourcev1.Device) resourcev1.ResourceSlice {
	return resourcev1.ResourceSlice{
		Spec: resourcev1.ResourceSliceSpec{
			Driver:   driver,
			NodeName: ptr.To("node-1"),
			Pool:     resourcev1.ResourcePool{Name: pool, Generation: generation},
			Devices:  devices,
		},
	}
}

func TestNvidiaGpuNodeInfoFromResourceSlices(t *testing.T) {
	const a100 = "NVIDIA A100-SXM4-80GB"

	tests := []struct {
		name     string
		slices   []resourcev1.ResourceSlice
		expected *NvidiaGpuNodeInfo
	}{
		{
			name:     "no slices",
			expected: nil,
		},
		{
			name: "whole GPUs across two slices of one pool",
			slices: []resourcev1.ResourceSlice{
				resourceSlice(NvidiaDraDriverName, "node-1", 1,
					gpuDevice("gpu-0", "gpu", a100, "80Gi"), gpuDevice("gpu-1", "gpu", a100, "80Gi")),
				resourceSlice(NvidiaDraDriverName, "node-1", 1,
					gpuDevice("gpu-2", "gpu", a100, "80Gi"), gpuDevice("gpu-3", "gpu", a100, "80Gi")),
			},
			expected: &NvidiaGpuNodeInfo{Product: "NVIDIA-A100-SXM4-80GB", Count: 4, MemoryBytes: 80 * 1024 * 1024 * 1024},
		},
		{
			name: "other drivers are ignored",
			slices: []resourcev1.ResourceSlice{
				resourceSlice("example.com/other", "node-1", 1, gpuDevice("dev-0", "gpu", "Other", "1Gi")),
			},
			expected: nil,
		},
		{
			name: "MIG devices are not counted as whole GPUs",
			slices: []resourcev1.ResourceSlice{
				resourceSlice(NvidiaDraDriverName, "node-1", 1,
					gpuDevice("gpu-0", "gpu", a100, "80Gi"), gpuDevice("gpu-1-mig-0", "mig", a100, "10Gi")),
			},
			expected: &NvidiaGpuNodeInfo{Product: "NVIDIA-A100-SXM4-80GB", Count: 1, MemoryBytes: 80 * 1024 * 1024 * 1024},
		},
		{
			name: "only the newest generation of a pool counts",
			slices: []resourcev1.ResourceSlice{
				resourceSlice(NvidiaDraDriverName, "node-1", 1,
					gpuDevice("gpu-0", "gpu", a100, "80Gi"), gpuDevice("gpu-1", "gpu", a100, "80Gi")),
				resourceSlice(NvidiaDraDriverName, "node-1", 2, gpuDevice("gpu-0", "gpu", a100, "80Gi")),
			},
			expected: &NvidiaGpuNodeInfo{Product: "NVIDIA-A100-SXM4-80GB", Count: 1, MemoryBytes: 80 * 1024 * 1024 * 1024},
		},
		{
			name: "device without type attribute counts as a whole GPU",
			slices: []resourcev1.ResourceSlice{
				resourceSlice(NvidiaDraDriverName, "node-1", 1, gpuDevice("gpu-0", "", a100, "")),
			},
			expected: &NvidiaGpuNodeInfo{Product: "NVIDIA-A100-SXM4-80GB", Count: 1},
		},
		{
			name: "memory and product are optional",
			slices: []resourcev1.ResourceSlice{
				resourceSlice(NvidiaDraDriverName, "node-1", 1, gpuDevice("gpu-0", "gpu", "", "")),
			},
			expected: &NvidiaGpuNodeInfo{Count: 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NvidiaGpuNodeInfoFromResourceSlices(tc.slices))
		})
	}
}

func TestNvidiaGpuNodeInfoFromResourceSlicesAcceptsQualifiedAttributeNames(t *testing.T) {
	device := resourcev1.Device{
		Name: "gpu-0",
		Attributes: map[resourcev1.QualifiedName]resourcev1.DeviceAttribute{
			"gpu.nvidia.com/type":        {StringValue: ptr.To("gpu")},
			"gpu.nvidia.com/productName": {StringValue: ptr.To("NVIDIA H100 80GB HBM3")},
		},
	}

	info := NvidiaGpuNodeInfoFromResourceSlices([]resourcev1.ResourceSlice{
		resourceSlice(NvidiaDraDriverName, "node-1", 1, device),
	})

	assert.Equal(t, &NvidiaGpuNodeInfo{Product: "NVIDIA-H100-80GB-HBM3", Count: 1}, info)
}
