/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("ResourcesAndLimits tests", func() {
	var (
		ctx            context.Context
		resourceFields *field_groups.ResourceAllocationFields
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		resourceFields = &field_groups.ResourceAllocationFields{}
	})

	Context("Resource requirements and limits", func() {
		It("should set CPU and memory requests", func() {
			// Set CPU and memory requests
			resourceFields.Cpu = field_types.NewOptionalNumberField(ptr.To(2.5))
			resourceFields.Memory = field_types.NewOptionalQuantityField("2Gi", "")

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// No environment variables should be set for CPU/memory requests
			Expect(envVars).To(HaveLen(0))

			// Check CPU request
			cpuRequest, ok := result.Requests[corev1.ResourceCPU]
			Expect(ok).To(BeTrue())
			Expect(cpuRequest.MilliValue()).To(Equal(int64(2500))) // 2.5 CPU = 2500 millicores

			// Check memory request
			memRequest, ok := result.Requests[corev1.ResourceMemory]
			Expect(ok).To(BeTrue())
			expectedMem, err := resource.ParseQuantity("2Gi")
			Expect(err).NotTo(HaveOccurred())
			Expect(memRequest.Value()).To(Equal(expectedMem.Value()))
		})

		It("should set CPU and memory limits", func() {
			// Set CPU and memory limits
			resourceFields.CpuLimit = field_types.NewOptionalNumberField(ptr.To(3.5))
			resourceFields.MemoryLimit = field_types.NewOptionalQuantityField("4Gi", "")

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// No environment variables should be set for CPU/memory limits
			Expect(envVars).To(HaveLen(0))

			// Check CPU limit
			cpuLimit, ok := result.Limits[corev1.ResourceCPU]
			Expect(ok).To(BeTrue())
			Expect(cpuLimit.MilliValue()).To(Equal(int64(3500))) // 3.5 CPU = 3500 millicores

			// Check memory limit
			memLimit, ok := result.Limits[corev1.ResourceMemory]
			Expect(ok).To(BeTrue())
			expectedMem, err := resource.ParseQuantity("4Gi")
			Expect(err).NotTo(HaveOccurred())
			Expect(memLimit.Value()).To(Equal(expectedMem.Value()))
		})

		It("should handle GPU devices correctly", func() {
			// Set GPU devices
			resourceFields.GpuDevices = field_types.NewOptionalIntField(ptr.To(2))

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// No environment variables should be set for basic GPU devices without memory limits
			Expect(envVars).To(HaveLen(0))

			// Check GPU limit is set
			gpuLimit, ok := result.Limits[corev1.ResourceName("nvidia.com/gpu")]
			Expect(ok).To(BeTrue())
			Expect(gpuLimit.Value()).To(Equal(int64(2)))
		})

		It("should set GPU limits from gpu field (non-fractional)", func() {
			// Set GPU using the gpu field (non-fractional)
			resourceFields.Gpu = field_types.NewOptionalNumberField(ptr.To(3.0))

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// No environment variables should be set for GPU without limit/memory settings
			Expect(envVars).To(HaveLen(0))

			// Check GPU limit is set
			gpuLimit, ok := result.Limits[corev1.ResourceName("nvidia.com/gpu")]
			Expect(ok).To(BeTrue())
			Expect(gpuLimit.Value()).To(Equal(int64(3)))
		})

		It("should set extended resources", func() {
			// Create extended resource map
			extendedResources := map[string]field_types.ExtendedResourcesItemAndRules{
				"custom-resource": {
					Value: &field_types.ExtendedResourcesItemValue{
						Name:     ptr.To("custom.resource/type"),
						Quantity: ptr.To("2"),
					},
				},
			}
			resourceFields.ExtendedResources = field_types.NewExtendedResourcesField(extendedResources)

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			// No environment variables should be set for extended resources
			Expect(envVars).To(HaveLen(0))

			// Check extended resource is set in both requests and limits
			resourceName := corev1.ResourceName("custom.resource/type")

			extRequest, ok := result.Requests[resourceName]
			Expect(ok).To(BeTrue())
			Expect(extRequest.Value()).To(Equal(int64(2)))

			extLimit, ok := result.Limits[resourceName]
			Expect(ok).To(BeTrue())
			Expect(extLimit.Value()).To(Equal(int64(2)))
		})

		It("should set GPU environment variables when supported", func() {
			// Set GPU memory limit
			resourceFields.Gpu = field_types.NewOptionalNumberField(ptr.To(0.5))
			resourceFields.GpuLimit = field_types.NewOptionalNumberField(ptr.To(0.8))
			resourceFields.GpuMemory = field_types.NewOptionalQuantityField("4Gi", "")
			resourceFields.GpuMemoryLimit = field_types.NewOptionalQuantityField("8Gi", "")

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify results
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())

			// Two env vars should be set: one for GPU memory limit and one for GPU limit
			Expect(envVars).To(HaveLen(2))

			// Verify env vars
			var foundGpuMemoryLimit, foundGpuLimit bool
			for _, env := range envVars {
				if env.Name == constants.GpuMemoryLimitEnvVar {
					if env.Value == "8Gi" {
						foundGpuMemoryLimit = true
					} else if env.Value == "0.8" {
						foundGpuLimit = true
					}
				}
			}

			Expect(foundGpuMemoryLimit).To(BeTrue())
			Expect(foundGpuLimit).To(BeTrue())
		})

		It("should return error for invalid memory size", func() {
			// Set invalid memory size
			resourceFields.Memory = field_types.NewOptionalQuantityField("invalid", "")

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify error is returned
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid memory size"))
			Expect(result).To(BeNil())
			Expect(envVars).To(BeNil())
		})

		It("should return error for invalid memory limit", func() {
			// Set invalid memory limit
			resourceFields.MemoryLimit = field_types.NewOptionalQuantityField("invalid", "")

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify error is returned
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid memory limit"))
			Expect(result).To(BeNil())
			Expect(envVars).To(BeNil())
		})

		It("should return error for invalid extended resource quantity", func() {
			// Create extended resource map with invalid quantity
			extendedResources := map[string]field_types.ExtendedResourcesItemAndRules{
				"custom-resource": {
					Value: &field_types.ExtendedResourcesItemValue{
						Name:     ptr.To("custom.resource/type"),
						Quantity: ptr.To("invalid"),
					},
				},
			}
			resourceFields.ExtendedResources = field_types.NewExtendedResourcesField(extendedResources)

			// Call the function
			result, envVars, err := GetResourcesReqAndLimits(ctx, resourceFields)

			// Verify error is returned
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid quantity for extended resource"))
			Expect(result).To(BeNil())
			Expect(envVars).To(BeNil())
		})
	})
})
