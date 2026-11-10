// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"encoding/json"
	"math"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = It("scalar memory floor does not establish parallel layout", func() {
	capacity, minimum, one := int64(80_000_000_000), 380.0, 1
	result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
		[]ngcmetadata.Profile{{ID: "tp1-floor", TensorParallelSize: &one, PipelineParallelSize: &one, MinimumVRAMPerDeviceGB: &minimum}}, 90)
	Expect(string(result.Status)).To(Equal("no_recommendation"))
	Expect(result.Recommendation).To(BeNil(), "neither five nor eight GPUs is justified by this scalar TP1 floor")
})

func publishedParallelProfile(id string, minimum float64, tp, pp int) ngcmetadata.Profile {
	return ngcmetadata.Profile{ID: id, Precision: "bf16", InferenceBackend: "vllm", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &tp, PipelineParallelSize: &pp}
}

var _ = It("parallel profiles choose fewest GPUs within memory and node limits", func() {
	profiles := []ngcmetadata.Profile{
		publishedParallelProfile("four", 31, 4, 1),
		publishedParallelProfile("eight", 18, 8, 1),
		publishedParallelProfile("two", 57, 2, 1),
	}
	for _, tc := range []struct {
		capacity, nodeCount int64
		profile             string
		count               int64
	}{
		{80, 8, "two", 2}, {40, 8, "four", 4}, {24, 8, "eight", 8},
		{16, 8, "", 0}, {80, 1, "", 0}, {40, 2, "", 0},
	} {
		capacity := tc.capacity * 1_000_000_000
		result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: tc.nodeCount}, profiles, 90)
		if tc.profile == "" {
			Expect(string(result.Status)).To(Equal("no_recommendation"))
			Expect(result.Recommendation).To(BeNil())
			continue
		}
		Expect(string(result.Status)).To(Equal("recommended"))
		Expect(result.ProfileId).To(Equal(tc.profile))
		Expect(result.Recommendation).To(Not(BeNil()))
		Expect(result.Recommendation.GpuCount).To(Equal(tc.count))
		Expect(string(result.Recommendation.AllocationType)).To(Equal("gpu_count"))
		Expect(result.Recommendation.GpuMemoryPerDeviceBytes).To(Equal(capacity))
		Expect(result.TensorParallelSize).To(Equal(tc.count))
		Expect(result.PipelineParallelSize).To(Equal(int64(1)))
	}
})

var _ = It("published count retains non power of two and pipeline layout", func() {
	capacity := int64(80_000_000_000)
	result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
		[]ngcmetadata.Profile{publishedParallelProfile("tp3-pp2", 30, 3, 2)}, 90)
	Expect(string(result.Status)).To(Equal("recommended"))
	Expect(result.Recommendation.GpuCount).To(Equal(int64(6)))
	Expect(result.TensorParallelSize).To(Equal(int64(3)))
	Expect(result.PipelineParallelSize).To(Equal(int64(2)))
})

var _ = It("ranks fitting generic parallel profiles by count, memory and stable ID", func() {
	capacity := int64(80_000_000_000)
	hardware := gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 4}
	targeted := publishedParallelProfile("targeted", 70, 4, 1)
	targeted.GPUDevice = "2330:10de"
	tested := publishedParallelProfile("tested", 30, 2, 1)
	tested.TestedGPUDevices = []string{"2330:10de"}
	generic := publishedParallelProfile("generic", 10, 2, 1)
	for _, profiles := range [][]ngcmetadata.Profile{{generic, tested, targeted}, {targeted, tested, generic}} {
		result := recommendForHardware(hardware, profiles, 90)
		Expect(result.ProfileId).To(Equal("generic"))
	}
	tooMany := publishedParallelProfile("too-many", 1, 8, 1)
	result := recommendForHardware(hardware, []ngcmetadata.Profile{generic, tested, tooMany}, 90)
	Expect(result.ProfileId).To(Equal("generic"))
	tie := generic
	tie.ID = "a-generic"
	result = recommendForHardware(hardware, []ngcmetadata.Profile{generic, tie}, 90)
	Expect(result.ProfileId).To(Equal("a-generic"))
	smaller := publishedParallelProfile("smaller", 9, 2, 1)
	result = recommendForHardware(hardware, []ngcmetadata.Profile{tie, smaller}, 90)
	Expect(result.ProfileId).To(Equal("smaller"))
	fewer := publishedParallelProfile("fewer", 70, 2, 1)
	more := publishedParallelProfile("more", 1, 4, 1)
	result = recommendForHardware(hardware, []ngcmetadata.Profile{more, fewer}, 90)
	Expect(result.ProfileId).To(Equal("fewer"))
})

var _ = It("fitting single device baseline keeps existing sharing policy", func() {
	capacity := int64(80_000_000_000)
	parallel := publishedParallelProfile("parallel", 20, 2, 1)
	for _, tc := range []struct {
		minimum float64
		kind    string
	}{{71.9, "gpu_memory"}, {72, "gpu_count"}, {80, "gpu_count"}} {
		r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{parallel, publishedParallelProfile("one", tc.minimum, 1, 1)}, 90)
		Expect(r.ProfileId).To(Equal("one"))
		Expect(string(r.Recommendation.AllocationType)).To(Equal(tc.kind))
		Expect(r.Recommendation.GpuCount).To(Equal(int64(1)))
	}
})

var _ = It("parallel layouts require explicit positive non overflowing values", func() {
	capacity := int64(80_000_000_000)
	for _, tc := range []struct{ tp, pp int }{{0, 1}, {-1, 2}, {2, 0}, {2, -1}, {math.MaxInt, 2}} {
		r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{publishedParallelProfile("bad-layout", 1, tc.tp, tc.pp)}, 90)
		Expect(string(r.Status)).To(Equal("no_recommendation"))
		Expect(r.Recommendation).To(BeNil())
	}
	absent := publishedParallelProfile("missing-pp", 20, 2, 1)
	absent.PipelineParallelSize = nil
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{absent}, 90)
	Expect(string(r.Status)).To(Equal("no_recommendation"))
})

var _ = It("profile maximum GPU count is respected", func() {
	var profiles []ngcmetadata.Profile
	Expect(json.Unmarshal([]byte(`[
      {"profile_id":"invalid-two","tensor_parallel_size":2,"pipeline_parallel_size":1,"min_vram_per_device_gb":20,"max_gpu_count":1},
      {"profile_id":"valid-four","tensor_parallel_size":4,"pipeline_parallel_size":1,"min_vram_per_device_gb":30,"max_gpu_count":4}
    ]`), &profiles)).To(BeNil())
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, profiles, 90)
	Expect(r.ProfileId).To(Equal("valid-four"))
})

var _ = It("unknown node count does not produce allocation", func() {
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity},
		[]ngcmetadata.Profile{publishedParallelProfile("one", 30, 1, 1)}, 90)
	Expect(string(r.Status)).To(Equal("no_recommendation"))
	Expect(r.Reason).To(ContainSubstring("single-node GPU count"))
	Expect(r.Recommendation).To(BeNil())
})

var _ = It("exact Llama 70B generic profile fits two devices without rescaling TP1", func() {
	data, err := os.ReadFile("testdata/llama-3.1-70b-instruct-2.0.13.json")
	Expect(err).To(BeNil())
	var metadata ngcmetadata.Metadata
	Expect(json.Unmarshal(data, &metadata)).To(BeNil())
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, metadata.Profiles, 90)
	Expect(r.ProfileId).To(Equal("313ac0ae8b3d63007c5c3af22b90e3371d99332aff7a8d19162c0a5dfaaeade5"))
	Expect(r.Precision).To(Equal("nvfp4"))
	Expect(r.MinimumGpuMemoryPerDeviceGb).To(BeNumerically("==", 43.0))
	Expect(r.Recommendation.GpuCount).To(Equal(int64(2)))
	Expect(r.TensorParallelSize).To(Equal(int64(2)))
	Expect(r.PipelineParallelSize).To(Equal(int64(1)))
})
