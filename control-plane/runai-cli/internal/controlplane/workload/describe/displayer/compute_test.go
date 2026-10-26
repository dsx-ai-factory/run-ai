/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func TestBuilders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Builders Suite")
}

var _ = Describe("ComputeResourcesBuilder", func() {
	It("should build ComputeResources correctly", func() {
		resources := displayer.NewComputeResourcesBuilder().
			GPUComputeRequest(2).
			CPUCoresRequest(4).
			GPUMemoryRequest("8Gi").
			CPUMemoryRequest("16Gi").
			GPUComputeLimit(4).
			CPUCoresLimit(8).
			GPUMemoryLimit("16Gi").
			CPUMemoryLimit("32Gi").
			Build()

		Expect(resources.GPUCompute.Requested).To(Equal(2.0))
		Expect(resources.CPUCores.Requested).To(Equal(4.0))
		Expect(resources.GPUMemory.Requested).To(Equal("8Gi"))
		Expect(resources.CPUMemory.Requested).To(Equal("16Gi"))
		Expect(*resources.GPUCompute.Limit).To(Equal(4.0))
		Expect(*resources.CPUCores.Limit).To(Equal(8.0))
		Expect(resources.GPUMemory.Limit).To(Equal("16Gi"))
		Expect(resources.CPUMemory.Limit).To(Equal("32Gi"))
	})
})

var _ = Describe("AllocatedResourcesBuilder", func() {
	It("should build AllocatedResources correctly", func() {
		resources := displayer.NewAllocatedResourcesBuilder().
			GPUCompute(2).
			CPUCores(4).
			GPUMemory("8Gi").
			CPUMemory("16Gi").
			Build()

		Expect(resources.GPUCompute).To(Equal(2.0))
		Expect(resources.CPUCores).To(Equal(4.0))
		Expect(resources.GPUMemory).To(Equal("8Gi"))
		Expect(resources.CPUMemory).To(Equal("16Gi"))
	})
})

var _ = Describe("ComputeBuilder", func() {
	It("should build Compute correctly", func() {
		requestedResources := displayer.NewComputeResourcesBuilder().
			GPUComputeRequest(2).
			CPUCoresRequest(4).
			Build()

		allocatedResources := displayer.NewAllocatedResourcesBuilder().
			GPUCompute(2).
			CPUCores(4).
			Build()

		compute := displayer.NewComputeBuilder().
			RequestedResources(requestedResources).
			AllocatedResources(allocatedResources).
			RequestedNodePools([]string{"pool1", "pool2"}).
			CurrentNodePools([]string{"pool1"}).
			Build()

		Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(2.0))
		Expect(compute.AllocatedResources.GPUCompute).To(Equal(2.0))
		Expect(compute.RequestedNodePools).To(Equal([]string{"pool1", "pool2"}))
		Expect(compute.CurrentNodePools).To(Equal([]string{"pool1"}))
	})

	It("should set entire Compute with Set method", func() {
		original := displayer.NewComputeBuilder().
			RequestedNodePools([]string{"pool1"}).
			Build()

		compute := displayer.NewComputeBuilder().
			Set(original).
			Build()

		Expect(compute.RequestedNodePools).To(Equal([]string{"pool1"}))
	})
})

var _ = Describe("BuildComputeFromSpec", func() {
	DescribeTable("reads memory aliases without changing existing resource units",
		func(input string) {
			var spec openapi.WorkspaceSpecSpecAllOfCompute
			Expect(json.Unmarshal([]byte(input), &spec)).To(Succeed())
			compute := displayer.BuildComputeFromSpec(&spec)
			Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(0.5))
			Expect(compute.RequestedResources.GPUCompute.Limit).To(Equal(openapi.PtrFloat64(0.75)))
		},
		Entry("old response", `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`),
		Entry("new response", `{"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75}`),
		Entry("dual response", `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75,"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75}`),
		Entry("multiple devices", `{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.375}`),
	)

	DescribeTable("includes a mode only when returned by the API",
		func(input, expected string) {
			var spec openapi.WorkspaceSpecSpecAllOfCompute
			Expect(json.Unmarshal([]byte(input), &spec)).To(Succeed())
			compute := displayer.BuildComputeFromSpec(&spec)
			payload, err := json.Marshal(compute)
			Expect(err).NotTo(HaveOccurred())
			if expected == "" {
				Expect(compute.GPUComputeMode).To(BeNil())
				Expect(string(payload)).NotTo(ContainSubstring("gpuComputeMode"))
			} else {
				Expect(compute.GPUComputeMode).To(Equal(openapi.PtrString(expected)))
				Expect(string(payload)).To(ContainSubstring(`"gpuComputeMode":"` + expected + `"`))
			}
		},
		Entry("omitted", `{}`, ""),
		Entry("null", `{"gpuComputeMode":null}`, ""),
		Entry("time slicing", `{"gpuComputeMode":"time-slicing"}`, "time-slicing"),
		Entry("SM sharing", `{"gpuComputeMode":"sm-sharing"}`, "sm-sharing"),
	)

	It("maps every displayed request and limit field", func() {
		spec := openapi.NewWorkspaceSpecSpecAllOfCompute()
		spec.SetGpuPortionRequest(0.5)
		spec.SetGpuPortionLimit(0.75)
		spec.SetCpuCoreRequest(2)
		spec.SetCpuCoreLimit(4)
		spec.SetGpuMemoryRequest("8Gi")
		spec.SetGpuMemoryLimit("12Gi")
		spec.SetCpuMemoryRequest("16Gi")
		spec.SetCpuMemoryLimit("24Gi")

		compute := displayer.BuildComputeFromSpec(spec)

		Expect(compute).NotTo(BeNil())
		Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(0.5))
		Expect(compute.RequestedResources.GPUCompute.Limit).NotTo(BeNil())
		Expect(*compute.RequestedResources.GPUCompute.Limit).To(Equal(0.75))
		Expect(compute.RequestedResources.CPUCores.Requested).To(Equal(2.0))
		Expect(compute.RequestedResources.CPUCores.Limit).NotTo(BeNil())
		Expect(*compute.RequestedResources.CPUCores.Limit).To(Equal(4.0))
		Expect(compute.RequestedResources.GPUMemory.Requested).To(Equal("8Gi"))
		Expect(compute.RequestedResources.GPUMemory.Limit).To(Equal("12Gi"))
		Expect(compute.RequestedResources.CPUMemory.Requested).To(Equal("16Gi"))
		Expect(compute.RequestedResources.CPUMemory.Limit).To(Equal("24Gi"))
	})

	It("maps device-based GPU requests", func() {
		spec := openapi.NewWorkspaceSpecSpecAllOfCompute()
		spec.SetGpuDevicesRequest(2)

		compute := displayer.BuildComputeFromSpec(spec)

		Expect(compute).NotTo(BeNil())
		Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(2.0))
		Expect(compute.RequestedResources.GPUCompute.Limit).To(BeNil())
	})

	It("maps GPU portions to the total request across devices", func() {
		spec := openapi.NewWorkspaceSpecSpecAllOfCompute()
		spec.SetGpuDevicesRequest(2)
		spec.SetGpuPortionRequest(0.5)
		spec.SetGpuPortionLimit(0.75)

		compute := displayer.BuildComputeFromSpec(spec)

		Expect(compute).NotTo(BeNil())
		Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(1.0))
		Expect(compute.RequestedResources.GPUCompute.Limit).NotTo(BeNil())
		Expect(*compute.RequestedResources.GPUCompute.Limit).To(Equal(1.5))
	})

	It("preserves device requests when a legacy whole-number portion is also present", func() {
		spec := openapi.NewWorkspaceSpecSpecAllOfCompute()
		spec.SetGpuDevicesRequest(3)
		spec.SetGpuPortionRequest(2)

		compute := displayer.BuildComputeFromSpec(spec)

		Expect(compute).NotTo(BeNil())
		Expect(compute.RequestedResources.GPUCompute.Requested).To(Equal(3.0))
	})

	It("returns nil for a nil compute spec", func() {
		Expect(displayer.BuildComputeFromSpec(nil)).To(BeNil())
	})
})
