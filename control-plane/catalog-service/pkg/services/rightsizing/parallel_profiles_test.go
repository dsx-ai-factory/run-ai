// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
	"github.com/stretchr/testify/require"
)

func TestScalarMemoryFloorDoesNotEstablishParallelLayout(t *testing.T) {
	capacity, minimum, one := int64(80_000_000_000), 380.0, 1
	result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
		[]ngcmetadata.Profile{{ID: "tp1-floor", TensorParallelSize: &one, PipelineParallelSize: &one, MinimumVRAMPerDeviceGB: &minimum}}, 90)
	require.Equal(t, "no_recommendation", result.Status)
	require.Nil(t, result.Recommendation, "neither five nor eight GPUs is justified by this scalar TP1 floor")
}

func publishedParallelProfile(id string, minimum float64, tp, pp int) ngcmetadata.Profile {
	return ngcmetadata.Profile{ID: id, Precision: "bf16", InferenceBackend: "vllm", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &tp, PipelineParallelSize: &pp}
}

func TestParallelProfilesChooseFewestGPUsWithinMemoryAndNodeLimits(t *testing.T) {
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
			require.Equal(t, "no_recommendation", result.Status)
			require.Nil(t, result.Recommendation)
			continue
		}
		require.Equal(t, "recommended", result.Status)
		require.Equal(t, tc.profile, result.ProfileId)
		require.NotNil(t, result.Recommendation)
		require.Equal(t, tc.count, result.Recommendation.GpuCount)
		require.Equal(t, "gpu_count", result.Recommendation.AllocationType)
		require.Equal(t, capacity, result.Recommendation.GpuMemoryPerDeviceBytes)
		require.Equal(t, tc.count, result.TensorParallelSize)
		require.Equal(t, int64(1), result.PipelineParallelSize)
	}
}

func TestPublishedCountRetainsNonPowerOfTwoAndPipelineLayout(t *testing.T) {
	capacity := int64(80_000_000_000)
	result := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
		[]ngcmetadata.Profile{publishedParallelProfile("tp3-pp2", 30, 3, 2)}, 90)
	require.Equal(t, "recommended", result.Status)
	require.Equal(t, int64(6), result.Recommendation.GpuCount)
	require.Equal(t, int64(3), result.TensorParallelSize)
	require.Equal(t, int64(2), result.PipelineParallelSize)
}

func TestParallelSelectionFiltersFitBeforeHardwareTierThenUsesDeterministicTies(t *testing.T) {
	capacity := int64(80_000_000_000)
	hardware := gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 4}
	direct := publishedParallelProfile("direct", 70, 4, 1)
	direct.GPUDevice = "2330:10de"
	tested := publishedParallelProfile("tested", 30, 2, 1)
	tested.TestedGPUDevices = []string{"2330:10de"}
	generic := publishedParallelProfile("generic", 10, 2, 1)
	for _, profiles := range [][]ngcmetadata.Profile{{generic, tested, direct}, {direct, tested, generic}} {
		r := recommendForHardware(hardware, profiles, 90)
		require.Equal(t, "direct", r.ProfileId)
	}
	tooMany := direct
	eight := 8
	tooMany.TensorParallelSize = &eight
	r := recommendForHardware(hardware, []ngcmetadata.Profile{generic, tested, tooMany}, 90)
	require.Equal(t, "tested", r.ProfileId, "an oversized targeted profile cannot hide a fitting tested profile")
	tie := tested
	tie.ID = "a-tested"
	r = recommendForHardware(hardware, []ngcmetadata.Profile{tested, tie}, 90)
	require.Equal(t, "a-tested", r.ProfileId)
	smaller := tested
	smaller.ID = "smaller"
	minimum := 29.0
	smaller.MinimumVRAMPerDeviceGB = &minimum
	r = recommendForHardware(hardware, []ngcmetadata.Profile{tie, smaller}, 90)
	require.Equal(t, "smaller", r.ProfileId)
}

func TestFittingSingleDeviceBaselineKeepsExistingSharingPolicy(t *testing.T) {
	capacity := int64(80_000_000_000)
	direct := publishedParallelProfile("parallel-direct", 20, 2, 1)
	direct.GPUDevice = "2330:10de"
	for _, tc := range []struct {
		minimum float64
		kind    string
	}{{71.9, "gpu_memory"}, {72, "gpu_count"}, {80, "gpu_count"}} {
		r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{direct, publishedParallelProfile("one", tc.minimum, 1, 1)}, 90)
		require.Equal(t, "one", r.ProfileId)
		require.Equal(t, tc.kind, r.Recommendation.AllocationType)
		require.Equal(t, int64(1), r.Recommendation.GpuCount)
	}
}

func TestParallelLayoutsRequireExplicitPositiveNonOverflowingValues(t *testing.T) {
	capacity := int64(80_000_000_000)
	for _, tc := range []struct{ tp, pp int }{{0, 1}, {-1, 2}, {2, 0}, {2, -1}, {math.MaxInt, 2}} {
		r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8},
			[]ngcmetadata.Profile{publishedParallelProfile("bad-layout", 1, tc.tp, tc.pp)}, 90)
		require.Equal(t, "no_recommendation", r.Status)
		require.Nil(t, r.Recommendation)
	}
	absent := publishedParallelProfile("missing-pp", 20, 2, 1)
	absent.PipelineParallelSize = nil
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, []ngcmetadata.Profile{absent}, 90)
	require.Equal(t, "no_recommendation", r.Status)
}

func TestProfileMaximumGPUCountIsRespected(t *testing.T) {
	var profiles []ngcmetadata.Profile
	require.NoError(t, json.Unmarshal([]byte(`[
      {"profile_id":"invalid-two","tensor_parallel_size":2,"pipeline_parallel_size":1,"min_vram_per_device_gb":20,"max_gpu_count":1},
      {"profile_id":"valid-four","tensor_parallel_size":4,"pipeline_parallel_size":1,"min_vram_per_device_gb":30,"max_gpu_count":4}
    ]`), &profiles))
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, profiles, 90)
	require.Equal(t, "valid-four", r.ProfileId)
}

func TestUnknownNodeCountDoesNotProduceAllocation(t *testing.T) {
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity},
		[]ngcmetadata.Profile{publishedParallelProfile("one", 30, 1, 1)}, 90)
	require.Equal(t, "no_recommendation", r.Status)
	require.Contains(t, r.Reason, "single-node GPU count")
	require.Nil(t, r.Recommendation)
}

func TestExactLlama70BProfileFitsTwoH100sInsteadOfRescalingTP1(t *testing.T) {
	data, err := os.ReadFile("testdata/llama-3.1-70b-instruct-2.0.13.json")
	require.NoError(t, err)
	var metadata ngcmetadata.Metadata
	require.NoError(t, json.Unmarshal(data, &metadata))
	capacity := int64(80_000_000_000)
	r := recommendForHardware(gpuinventory.Hardware{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &capacity, MaxDevicesPerNode: 8}, metadata.Profiles, 90)
	require.Equal(t, "74dc099acf027ca44ed08bf7f182582f4f20bef8c9be768908a3a11f55a4803d", r.ProfileId)
	require.Equal(t, "fp8", r.Precision)
	require.Equal(t, 57.0, r.MinimumGpuMemoryPerDeviceGb)
	require.Equal(t, int64(2), r.Recommendation.GpuCount)
	require.Equal(t, int64(2), r.TensorParallelSize)
	require.Equal(t, int64(1), r.PipelineParallelSize)
}
