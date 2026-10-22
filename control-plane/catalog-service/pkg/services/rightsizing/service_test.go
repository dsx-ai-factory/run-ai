/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rightsizing

import (
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/gpuinventory"
	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/clients/ngcmetadata"
)

type inventoryFunc func(context.Context, string) ([]gpuinventory.Hardware, error)

func (f inventoryFunc) Get(ctx context.Context, id string) ([]gpuinventory.Hardware, error) {
	return f(ctx, id)
}

type metadataFunc func(context.Context, string, string) (ngcmetadata.Metadata, error)

func (f metadataFunc) Get(ctx context.Context, image, tag string) (ngcmetadata.Metadata, error) {
	return f(ctx, image, tag)
}

func TestRecommendKeepsCapacitiesAndAppliesCutoff(t *testing.T) {
	small, large, minimum, one := int64(40_000_000_000), int64(80_000_000_000), 36.0, 1
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		return []gpuinventory.Hardware{{Type: "NVIDIA-A100", MemoryBytes: &small}, {Type: "NVIDIA-A100", MemoryBytes: &large}, {Type: "T4"}}, nil
	}), metadataFunc(func(_ context.Context, image, tag string) (ngcmetadata.Metadata, error) {
		if image != "nvcr.io/nim/nvidia/example" || tag != "2.1.4" {
			t.Fatalf("wrong image/tag: %s %s", image, tag)
		}
		return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{{ID: "profile", Precision: "bf16", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}}, nil
	}))
	result, err := s.Recommend(context.Background(), validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Results) != 3 {
		t.Fatalf("results=%+v", result.Results)
	}
	if result.Results[0].Recommendation.AllocationType != "gpu_count" || result.Results[1].Recommendation.AllocationType != "gpu_memory" ||
		result.Results[1].Recommendation.GpuMemoryPerDeviceBytes != 36_000_000_000 || result.Results[2].Status != "no_recommendation" {
		t.Fatalf("wrong allocations: %+v", result.Results)
	}
	gpu := "NVIDIA-A100"
	request := validRequest()
	request.GpuType = &gpu
	filtered, err := s.Recommend(context.Background(), request)
	if err != nil || len(filtered.Results) != 2 {
		t.Fatalf("filtered=%+v err=%v", filtered, err)
	}
}

func TestRecommendNeverUsesMultiDeviceMinimumAsTotal(t *testing.T) {
	capacity, minimum, tp, pp := int64(80_000_000_000), 35.0, 4, 1
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity}}, nil
	}),
		metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, error) {
			return ngcmetadata.Metadata{Profiles: []ngcmetadata.Profile{{ID: "tp4", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &tp, PipelineParallelSize: &pp}}}, nil
		}))
	result, err := s.Recommend(context.Background(), validRequest())
	if err != nil {
		t.Fatal(err)
	}
	if result.Results[0].Status != "no_recommendation" {
		t.Fatalf("multi-device profile treated as single GPU: %+v", result)
	}
}

func TestRecommendMissingMetadataAndUpstreamErrors(t *testing.T) {
	capacity := int64(80_000_000_000)
	inventory := inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) {
		return []gpuinventory.Hardware{{Type: "A100", MemoryBytes: &capacity}}, nil
	})
	s := New(inventory, metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, error) {
		return ngcmetadata.Metadata{}, nil
	}))
	result, err := s.Recommend(context.Background(), validRequest())
	if err != nil || result.Results[0].Status != "no_recommendation" {
		t.Fatalf("missing data=%+v %v", result, err)
	}
	failure := errors.New("upstream unavailable")
	s = New(inventory, metadataFunc(func(context.Context, string, string) (ngcmetadata.Metadata, error) {
		return ngcmetadata.Metadata{}, failure
	}))
	if _, err = s.Recommend(context.Background(), validRequest()); err == nil || err.Code != 503 {
		t.Fatalf("metadata failure hidden: %v", err)
	}
	s = New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) { return nil, failure }), nil)
	if _, err = s.Recommend(context.Background(), validRequest()); err == nil || err.Code != 503 {
		t.Fatalf("inventory failure hidden: %v", err)
	}
}

func TestRecommendRejectsInvalidContextBeforeLookup(t *testing.T) {
	s := New(nil, nil)
	request := validRequest()
	request.ClusterId = "invalid"
	if _, err := s.Recommend(context.Background(), request); err == nil || err.Code != 400 {
		t.Fatalf("invalid cluster accepted: %v", err)
	}
}

func validRequest() api.RightSizingRequest {
	return api.RightSizingRequest{ClusterId: "4b94efde-b05a-4118-a03a-1dedff81fa8e", Image: "nvcr.io/nim/nvidia/example", Tag: "2.1.4"}
}

func TestRejectsUnverifiedTargetedProfiles(t *testing.T) {
	capacity, minimum, one := int64(80_000_000_000), 36.0, 1
	for _, tc := range []struct{ name, gpu, device string }{
		{"display name only", "A100", ""}, {"PCI only", "", "20b2:10de"}, {"matching name different PCI", "A100", "ffff:10de"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity}, []ngcmetadata.Profile{{ID: "targeted", GPU: tc.gpu, GPUDevice: tc.device, MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}})
			if result.Status != "no_recommendation" || result.Recommendation != nil {
				t.Fatalf("unverified hardware recommended: %+v", result)
			}
		})
	}
}

func TestSingleDeviceBaselineProducesExplicitEstimate(t *testing.T) {
	capacity, minimum, one := int64(40_000_000_000), 81.0, 1
	result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity}, []ngcmetadata.Profile{{ID: "generic", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}})
	if result.Recommendation == nil || !result.Recommendation.Estimated || result.Recommendation.GpuCount != 3 || result.Recommendation.Warning == "" || result.ProfileId != "generic" {
		t.Fatalf("bad estimate: %+v", result)
	}
}

func TestProfileSelectionRequiresEvidenceAndIsDeterministic(t *testing.T) {
	capacity, one, minimum, larger := int64(80_000_000_000), 1, 16.0, 32.0
	baseline := ngcmetadata.Profile{ID: "b", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}
	for _, tc := range []struct {
		name     string
		profiles []ngcmetadata.Profile
		want     string
	}{
		{"missing memory", []ngcmetadata.Profile{{ID: "unknown", TensorParallelSize: &one, PipelineParallelSize: &one}}, ""},
		{"missing parallelism", []ngcmetadata.Profile{{ID: "unknown", MinimumVRAMPerDeviceGB: &minimum}}, ""},
		{"smallest memory then stable profile ID", []ngcmetadata.Profile{baseline, {ID: "larger", MinimumVRAMPerDeviceGB: &larger, TensorParallelSize: &one, PipelineParallelSize: &one}, {ID: "a", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one}}, "a"},
		{"tested devices are history", []ngcmetadata.Profile{{ID: "generic", MinimumVRAMPerDeviceGB: &minimum, TensorParallelSize: &one, PipelineParallelSize: &one, TestedGPUDevices: []string{"different-pci"}}}, "generic"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := recommendForHardware(gpuinventory.Hardware{Type: "A100", MemoryBytes: &capacity}, tc.profiles)
			if result.ProfileId != tc.want {
				t.Fatalf("profile=%q want=%q", result.ProfileId, tc.want)
			}
			if tc.want == "" && result.Recommendation != nil {
				t.Fatal("missing evidence produced allocation")
			}
		})
	}
}

func TestUnavailableRequestedGPUFailsBeforeMetadata(t *testing.T) {
	s := New(inventoryFunc(func(context.Context, string) ([]gpuinventory.Hardware, error) { return nil, nil }), nil)
	request := validRequest()
	missing := "missing"
	request.GpuType = &missing
	if _, err := s.Recommend(context.Background(), request); err == nil || err.Code != 400 {
		t.Fatalf("unknown type: %v", err)
	}
	request.GpuType = nil
	response, err := s.Recommend(context.Background(), request)
	if err != nil || response.Results == nil || len(response.Results) != 0 {
		t.Fatalf("empty inventory: %+v %v", response, err)
	}
}
