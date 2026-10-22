/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ngcmetadata

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMetadataPreservesPublishedProfileAndExactTag(t *testing.T) {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "api.ngc.nvidia.com" || r.URL.Path != "/v2/nim/metadata" ||
			r.URL.Query().Get("resourceId") != "nim/nvidia/example" || r.URL.Query().Get("tag") != "2.1.4" {
			t.Fatalf("unexpected metadata request: %s", r.URL)
		}
		if r.Header.Get("Authorization") != "" {
			t.Fatal("public NGC lookup must not forward caller credentials")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"profiles":[{"profile_id":"published-profile","gpu":"NVIDIA-B200","precision":"bf16","inference_backend":"vllm","tensor_parallel_size":2,"pipeline_parallel_size":1,"min_vram_per_device_gb":35,"gpu_device":"2901:10de","tested_gpu_devices":["2901:10de"]}]}`))}, nil
	})})
	metadata, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	if err != nil {
		t.Fatal(err)
	}
	if len(metadata.Profiles) != 1 {
		t.Fatalf("profiles = %d", len(metadata.Profiles))
	}
	p := metadata.Profiles[0]
	if p.ID != "published-profile" || p.MinimumVRAMPerDeviceGB == nil || *p.MinimumVRAMPerDeviceGB != 35 ||
		p.TensorParallelSize == nil || *p.TensorParallelSize != 2 || p.PipelineParallelSize == nil || *p.PipelineParallelSize != 1 ||
		p.GPU != "NVIDIA-B200" || p.GPUDevice != "2901:10de" || p.Precision != "bf16" ||
		p.InferenceBackend != "vllm" || len(p.TestedGPUDevices) != 1 {
		t.Fatalf("profile information lost: %+v", p)
	}
}

func TestMetadataMissingDiffersFromUpstreamFailure(t *testing.T) {
	for _, tt := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"not published", 404, `{"message":"not found"}`, false},
		{"no profiles", 200, `{"profiles":[]}`, false},
		{"rate limited", 429, `{"message":"rate limited"}`, true},
		{"upstream failure", 503, `{"message":"unavailable"}`, true},
		{"malformed JSON", 200, `{`, true},
		{"release metadata without profiles", 200, `{"release_details":{"container_version":"2.1.4"}}`, false},
		{"empty published metadata", 200, `{}`, false},
		{"invalid envelope", 200, `[]`, true},
		{"null response", 200, `null`, true},
		{"trailing response", 200, `{"profiles":[]} {}`, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})})
			result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v; want error %v", err, tt.wantError)
			}
			if !tt.wantError && len(result.Profiles) != 0 {
				t.Fatalf("expected missing metadata, got %+v", result)
			}
		})
	}
}

func TestMetadataDoesNotInventMissingMemoryOrParallelism(t *testing.T) {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"profiles":[{"profile_id":"missing"},{"profile_id":"zero","min_vram_per_device_gb":0}]}`))}, nil
	})})
	result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
	if err != nil {
		t.Fatal(err)
	}
	if result.Profiles[0].MinimumVRAMPerDeviceGB != nil || result.Profiles[0].TensorParallelSize != nil ||
		result.Profiles[1].MinimumVRAMPerDeviceGB == nil || *result.Profiles[1].MinimumVRAMPerDeviceGB != 0 {
		t.Fatalf("missing metadata was defaulted: %+v", result)
	}
}

func TestMetadataRejectsInvalidImageOrTagBeforeNetwork(t *testing.T) {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid input reached NGC")
		return nil, nil
	})})
	for _, tt := range []struct{ image, tag string }{
		{"https://attacker.invalid/model", "latest"},
		{"nvcr.io/other/model", "latest"},
		{"nvcr.io/nim/nvidia/example:latest", "latest"},
		{"nvcr.io/nim/nvidia/../example", "latest"},
		{"nvcr.io/nim/nvidia/example", ""},
		{"nvcr.io/nim/nvidia/example", "a&tag=b"},
	} {
		if _, err := client.Get(context.Background(), tt.image, tt.tag); err == nil {
			t.Errorf("accepted image %q tag %q", tt.image, tt.tag)
		}
	}
}

func TestMetadataPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})})
	_, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}
