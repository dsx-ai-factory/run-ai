/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ngcmetadata

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

var _ = It("metadata preserves published profile and exact tag", func() {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		Expect(r.URL.Host).To(Equal("api.ngc.nvidia.com"))
		Expect(r.URL.Path).To(Equal("/v2/nim/metadata"))
		Expect(r.URL.Query().Get("resourceId")).To(Equal("nim/nvidia/example"))
		Expect(r.URL.Query().Get("tag")).To(Equal("2.1.4"))
		Expect(r.Header.Get("Authorization")).To(BeEmpty())
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"profiles":[{"profile_id":"published-profile","gpu":"NVIDIA-B200","precision":"bf16","inference_backend":"vllm","tensor_parallel_size":2,"pipeline_parallel_size":1,"max_gpu_count":4,"min_vram_per_device_gb":35,"gpu_device":"2901:10de","tested_gpu_devices":["2901:10de"]}]}`))}, nil
	})}, 15*time.Minute)
	metadata, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	Expect(err).To(BeNil())
	Expect(metadata.Profiles).To(HaveLen(1))
	p := metadata.Profiles[0]
	Expect(p.ID).To(Equal("published-profile"))
	Expect(p.MinimumVRAMPerDeviceGB).NotTo(BeNil())
	Expect(*p.MinimumVRAMPerDeviceGB).To(BeNumerically("==", 35))
	Expect(p.TensorParallelSize).NotTo(BeNil())
	Expect(*p.TensorParallelSize).To(BeNumerically("==", 2))
	Expect(p.PipelineParallelSize).NotTo(BeNil())
	Expect(*p.PipelineParallelSize).To(BeNumerically("==", 1))
	Expect(p.GPU).To(Equal("NVIDIA-B200"))
	Expect(p.GPUDevice).To(Equal("2901:10de"))
	Expect(p.Precision).To(Equal("bf16"))
	Expect(p.InferenceBackend).To(Equal("vllm"))
	Expect(p.TestedGPUDevices).To(HaveLen(1))
	Expect(p.MaxGPUCount).NotTo(BeNil())
	Expect(*p.MaxGPUCount).To(BeNumerically("==", 4))
})

var _ = It("tag validation explains length and character constraints", func() {
	image := "nvcr.io/nim/nvidia/example"
	{
		err := Validate(image, strings.Repeat("a", 128))
		Expect(err).To(BeNil())
	}
	for _, tag := range []string{strings.Repeat("a", 129), "-bad", "a&b", ""} {
		err := Validate(image, tag)
		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("128"))
		Expect(err.Error()).To(ContainSubstring("start"))
	}
})

var _ = It("oversized metadata reports the response limit", func() {
	client := New(&http.Client{Transport: transportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(strings.Repeat(" ", maxResponseBytes+1))),
		}, nil
	})}, 15*time.Minute)
	_, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/oversized", "1.0")
	Expect(err).NotTo(BeNil())
	Expect(err.Details).To(ContainSubstring(strconv.Itoa(maxResponseBytes)))
})

var _ = Describe("metadata missing differs from upstream failure", func() {
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
		It(tt.name, func() {
			client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body))}, nil
			})}, 15*time.Minute)
			result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
			Expect((err != nil)).To(Equal(tt.wantError))
			if !tt.wantError {
				Expect(result.Profiles).To(HaveLen(0))
			}
		})
	}
})

var _ = It("metadata does not invent missing memory or parallelism", func() {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"profiles":[{"profile_id":"missing"},{"profile_id":"zero","min_vram_per_device_gb":0}]}`))}, nil
	})}, 15*time.Minute)
	result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(BeNil())
	Expect(result.Profiles[0].MinimumVRAMPerDeviceGB).To(BeNil())
	Expect(result.Profiles[0].TensorParallelSize).To(BeNil())
	Expect(result.Profiles[1].MinimumVRAMPerDeviceGB).NotTo(BeNil())
	Expect(*result.Profiles[1].MinimumVRAMPerDeviceGB).To(BeNumerically("==", 0))
})

var _ = It("metadata rejects invalid image or tag before network", func() {
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		Fail("invalid input reached NGC")
		return nil, nil
	})}, 15*time.Minute)
	for _, tt := range []struct{ image, tag string }{
		{"https://attacker.invalid/model", "latest"},
		{"nvcr.io/other/model", "latest"},
		{"nvcr.io/nim/nvidia/example:latest", "latest"},
		{"nvcr.io/nim/nvidia/../example", "latest"},
		{"nvcr.io/nim/nvidia/example", ""},
		{"nvcr.io/nim/nvidia/example", "a&tag=b"},
	} {
		{
			_, err := client.Get(context.Background(), tt.image, tt.tag)
			Expect(err).NotTo(BeNil())
		}
	}
})

var _ = It("metadata propagates cancellation", func() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return nil, r.Context().Err()
	})}, 15*time.Minute)
	_, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).NotTo(BeNil())
	Expect(err.Details).To(ContainSubstring(context.Canceled.Error()))
})
