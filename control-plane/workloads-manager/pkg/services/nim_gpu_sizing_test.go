// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package services

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NIM automatic GPU sizing", func() {
	var service *NimServicesImpl
	var projectID string
	var calls int
	var response *catalog.RightSizingResponse
	var dependencyError *api.Error
	BeforeEach(func() {
		randoms := testutils.NewRandomValues()
		projectID = randoms.ProjectIdStr
		org_unit.SetMock(org_unit.NewSimpleMock().WithRandoms(randoms))
		calls = 0
		dependencyError = nil
		response = &catalog.RightSizingResponse{}
		Expect(json.Unmarshal([]byte(`{"results":[{"gpuType":"NVIDIA-H100-80GB-HBM3","status":"recommended","profileId":"fp8-profile","gpuMemoryCapacityBytes":80000000000,"recommendation":{"allocationType":"gpu_memory","gpuCount":1,"gpuMemoryPerDeviceBytes":27000000000,"estimated":false}}]}`), response)).To(Succeed())
		service = &NimServicesImpl{recommendGPU: func(ctx context.Context, request catalog.RightSizingRequest) (*catalog.RightSizingResponse, *api.Error) {
			calls++
			Expect(request.ClusterId).To(Equal(randoms.ClusterId))
			Expect(request.Image).To(Equal("nvcr.io/nim/meta/llama-3.1-8b-instruct"))
			Expect(request.Tag).To(Equal("1.8.0"))
			Expect(request.GetGpuType()).To(Equal("NVIDIA-H100-80GB-HBM3"))
			return response, dependencyError
		}}
	})
	newSpec := func(extra string) api.NimServiceSpec {
		var spec api.NimServiceSpec
		Expect(json.Unmarshal([]byte(`{"gpuType":"NVIDIA-H100-80GB-HBM3","image":"nvcr.io/nim/meta/llama-3.1-8b-instruct:1.8.0"`+extra+`}`), &spec)).To(Succeed())
		return spec
	}
	DescribeTable("does not size an existing allocation", func(compute string) {
		spec := newSpec(`,"compute":` + compute)
		before, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())
		changed, failure := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(failure).To(BeNil())
		Expect(changed).To(BeFalse())
		Expect(calls).To(BeZero())
		after, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(MatchJSON(before))
	},
		Entry("device count", `{"gpuDevicesRequest":2}`),
		Entry("memory", `{"gpuMemoryRequest":"27G"}`),
		Entry("fraction", `{"gpuMemoryPortionRequest":0.5}`),
		Entry("legacy fraction", `{"gpuPortionRequest":0.5}`),
		Entry("explicit zero must be validated rather than replaced", `{"gpuDevicesRequest":0}`),
		Entry("limit alone must be validated rather than replaced", `{"gpuMemoryLimit":"30G"}`),
	)
	It("keeps legacy submission independent of recommendations", func() {
		spec := newSpec("")
		spec.GpuType = nil
		changed, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).To(BeNil())
		Expect(changed).To(BeFalse())
		Expect(calls).To(BeZero())
	})
	It("applies published memory and pins the runtime profile while preserving CPU resources", func() {
		spec := newSpec(`,"compute":{"cpuCoreRequest":2}`)
		changed, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).To(BeNil())
		Expect(changed).To(BeTrue())
		Expect(*spec.Compute.GpuMemoryRequest).To(Equal("27000000000"))
		Expect(*spec.Compute.GpuDevicesRequest).To(Equal(int32(1)))
		Expect(*spec.Compute.CpuCoreRequest).To(Equal(float64(2)))
		Expect(*spec.EnvironmentVariables).To(ContainElement(api.EnvironmentVariable{Name: convutils.Ref("NIM_MODEL_PROFILE"), Value: convutils.Ref("fp8-profile")}))
	})
	It("pins the cache profile to the same profile used for sizing", func() {
		spec := newSpec(`,"modelStore":{"nimCache":{"name":"cache"}}`)
		_, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).To(BeNil())
		Expect(*spec.ModelStore.NimCache.Profile).To(Equal("fp8-profile"))
	})
	DescribeTable("fails before creating a workload when sizing cannot be applied", func(extra, results string, code int) {
		spec := newSpec(extra)
		if results != "" {
			Expect(json.Unmarshal([]byte(results), response)).To(Succeed())
		}
		_, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(code))
	},
		Entry("no matching result", "", `{"results":[]}`, 400),
		Entry("no recommendation", "", `{"results":[{"gpuType":"NVIDIA-H100-80GB-HBM3","status":"no_recommendation","reason":"No eligible profile"}]}`, 400),
		Entry("estimated multiple GPUs", "", `{"results":[{"gpuType":"NVIDIA-H100-80GB-HBM3","status":"recommended","profileId":"fp8-profile","recommendation":{"allocationType":"gpu_count","gpuCount":2,"estimated":true}}]}`, 400),
		Entry("conflicting profile env", `,"environmentVariables":[{"name":"NIM_MODEL_PROFILE","value":"other"}]`, "", 400),
		Entry("conflicting cache profile", `,"modelStore":{"nimCache":{"name":"cache","profile":"other"}}`, "", 400),
		Entry("multi-node sizing is separate", `,"multiNode":{"workers":2}`, "", 400),
	)
	It("preserves dependency errors", func() {
		dependencyError = api.NewError(http.StatusServiceUnavailable, "catalog unavailable", nil)
		spec := newSpec("")
		_, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).To(Equal(dependencyError))
	})
	It("accepts validated whole-GPU counts returned by the recommendation API", func() {
		response.Results[0].Recommendation.Set(catalog.NewRightSizingRecommendation("gpu_count", 2, false))
		spec := newSpec("")
		_, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).To(BeNil())
		Expect(*spec.Compute.GpuDevicesRequest).To(Equal(int32(2)))
		Expect(spec.Compute.GpuMemoryRequest).To(BeNil())
	})
	It("rejects different profiles for capacities sharing the same product label", func() {
		other := response.Results[0]
		other.ProfileId = convutils.Ref("different-profile")
		response.Results = append(response.Results, other)
		spec := newSpec("")
		_, err := service.resolveNimGPUAllocation(context.Background(), projectID, &spec)
		Expect(err).NotTo(BeNil())
		Expect(int(err.Code)).To(Equal(400))
		Expect(spec.Compute).To(BeNil())
	})
})
