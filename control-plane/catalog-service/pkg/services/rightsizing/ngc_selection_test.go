// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package rightsizing

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

func TestExactLlamaTagSelectsPrecisionPerHardware(t *testing.T) {
	data, err := os.ReadFile("testdata/llama-3.1-8b-instruct-2.0.13.json")
	if err != nil {
		t.Fatal(err)
	}
	var metadata ngcmetadata.Metadata
	if err = json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	capacity := int64(80_000_000_000)
	for _, tc := range []struct {
		gpu, precision, profile string
		minimum                 float64
	}{
		{"NVIDIA-A100-SXM4-40GB", "bf16", "092ed4213624e774d24cdaf84e3b6222839bab2008a21d3c214ab46626366f90", 33},
		{"NVIDIA-A100-SXM4-80GB", "bf16", "092ed4213624e774d24cdaf84e3b6222839bab2008a21d3c214ab46626366f90", 33},
		{"NVIDIA-H100-80GB-HBM3", "fp8", "4d9fed9cb9fa4c5694b793a6421b62077b812e4e0326a10032b948cc9ddc0712", 27},
		{"NVIDIA-B200", "nvfp4", "20739fe3e660205a0a64339c06bf2aa5c90f4935dd5e547b2b22d96f9c91fbc4", 24},
	} {
		t.Run(tc.gpu, func(t *testing.T) {
			r := recommendForHardware(gpuinventory.Hardware{Type: tc.gpu, MemoryBytes: &capacity}, metadata.Profiles, 90)
			if r.Status != "recommended" || r.ProfileId != tc.profile || r.Precision != tc.precision || r.InferenceBackend != "vllm" || r.MinimumGpuMemoryPerDeviceGb != tc.minimum {
				t.Fatalf("wrong profile: %+v", r)
			}
			if tc.gpu == "NVIDIA-B200" && !strings.Contains(r.Recommendation.Warning, "GPU-agnostic") {
				t.Fatalf("ambiguous B200 identity must not claim tested match: %+v", r)
			}
		})
	}
}

func sizingProfile(id string, minimum float64) ngcmetadata.Profile {
	one := 1
	return ngcmetadata.Profile{ID: id, MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}
}

func TestNGCTargetingTiersAndFit(t *testing.T) {
	direct := sizingProfile("direct", 36)
	direct.GPUDevice = "20b2:10de"
	tested := sizingProfile("tested", 30)
	tested.GPU = "another-name"
	tested.TestedGPUDevices = []string{"20b2:10de"}
	generic := sizingProfile("generic", 20)
	genericEmpty := generic
	genericEmpty.TestedGPUDevices = []string{}
	foreign := sizingProfile("foreign", 10)
	foreign.GPUDevice = "2330:10de"
	named := sizingProfile("named", 10)
	named.GPU = "NVIDIA-A100-SXM4-80GB"
	oversized := sizingProfile("oversized", 81)
	oversized.GPUDevice = "20b2:10de"
	noMinimum := direct
	noMinimum.MinimumVRAMPerDeviceGB = nil
	capacity := int64(80_000_000_000)
	for _, tc := range []struct {
		name, gpu, want string
		profiles        []ngcmetadata.Profile
	}{
		{"direct with empty tested list wins", "NVIDIA-A100-SXM4-80GB", "direct", []ngcmetadata.Profile{generic, tested, direct}},
		{"tested wins over generic", "NVIDIA-A100-SXM4-80GB", "tested", []ngcmetadata.Profile{generic, tested}},
		{"generic with empty tested list", "NVIDIA-A100-SXM4-80GB", "generic", []ngcmetadata.Profile{genericEmpty}},
		{"generic with no tested list", "NVIDIA-A100-SXM4-80GB", "generic", []ngcmetadata.Profile{generic}},
		{"unknown PCI still permits generic", "unknown-GPU", "generic", []ngcmetadata.Profile{direct, generic}},
		{"ambiguous PCI still permits generic", "NVIDIA-B200", "generic", []ngcmetadata.Profile{direct, generic}},
		{"no matching target", "NVIDIA-A100-SXM4-80GB", "", []ngcmetadata.Profile{foreign}},
		{"display name alone cannot match", "NVIDIA-A100-SXM4-80GB", "", []ngcmetadata.Profile{named}},
		{"fitting fallback survives oversized direct", "NVIDIA-A100-SXM4-80GB", "generic", []ngcmetadata.Profile{oversized, generic}},
		{"missing minimum permits tested fallback", "NVIDIA-A100-SXM4-80GB", "tested", []ngcmetadata.Profile{noMinimum, tested}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := recommendForHardware(gpuinventory.Hardware{Type: tc.gpu, MemoryBytes: &capacity}, tc.profiles, 90)
			if r.ProfileId != tc.want {
				t.Fatalf("profile=%q want=%q; result=%+v", r.ProfileId, tc.want, r)
			}
			if tc.want == "" && (r.Status != "no_recommendation" || r.Reason == "" || r.Recommendation != nil) {
				t.Fatalf("missing recommendation reason: %+v", r)
			}
		})
	}
}

func TestMixedInventoryAndExplicitGPUType(t *testing.T) {
	small, large := int64(40_000_000_000), int64(80_000_000_000)
	a := sizingProfile("a100", 33)
	a.GPUDevice = "20b2:10de"
	a.Precision = "bf16"
	h := sizingProfile("h100", 27)
	h.GPUDevice = "2330:10de"
	h.Precision = "fp8"
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		return []gpuinventory.Hardware{
			{Type: "NVIDIA-A100-SXM4-80GB", MemoryBytes: &small}, {Type: "NVIDIA-A100-SXM4-80GB", MemoryBytes: &large},
			{Type: "NVIDIA-H100-80GB-HBM3", MemoryBytes: &large}, {Type: "unmatched", MemoryBytes: &large}}, nil
	}),
		metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, error) {
			return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{a, h}}, nil
		}), 90)
	request := validRequest()
	r, err := s.Recommend(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Results) != 4 || r.Results[0].ProfileId != "a100" || r.Results[1].ProfileId != "a100" || r.Results[2].ProfileId != "h100" || r.Results[3].Status != "no_recommendation" {
		t.Fatalf("mixed inventory: %+v", r)
	}
	gpu := "NVIDIA-A100-SXM4-80GB"
	request.GpuType = &gpu
	r, err = s.Recommend(context.Background(), request)
	if err != nil || len(r.Results) != 2 || r.Results[0].ProfileId != "a100" || r.Results[1].ProfileId != "a100" {
		t.Fatalf("explicit type: %+v error=%v", r, err)
	}
}

func TestRecommendationAllocationBoundaries(t *testing.T) {
	capacity := int64(100_000_000_000)
	for _, tc := range []struct {
		name          string
		minimum       float64
		cutoff        int64
		allocation    string
		count, memory int64
		estimated     bool
	}{
		{"below 90 percent", 89.999999999, 90, "gpu_memory", 1, 89_999_999_999, false},
		{"exactly 90 percent", 90, 90, "gpu_count", 1, 100_000_000_000, false},
		{"above 90 percent", 91, 90, "gpu_count", 1, 100_000_000_000, false},
		{"exactly capacity", 100, 90, "gpu_count", 1, 100_000_000_000, false},
		{"above capacity stays estimated", 101, 90, "gpu_count", 2, 0, true},
		{"configured 80 percent", 85, 80, "gpu_count", 1, 100_000_000_000, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := recommendForHardware(gpuinventory.Hardware{Type: "unknown", MemoryBytes: &capacity}, []ngcmetadata.Profile{sizingProfile("generic", tc.minimum)}, tc.cutoff)
			a := r.Recommendation
			if r.Status != "recommended" || r.MinimumGpuMemoryPerDeviceGb != tc.minimum || a == nil || a.AllocationType != tc.allocation || a.GpuCount != tc.count || a.GpuMemoryPerDeviceBytes != tc.memory || a.Estimated != tc.estimated {
				t.Fatalf("allocation=%+v result=%+v", a, r)
			}
		})
	}
}
