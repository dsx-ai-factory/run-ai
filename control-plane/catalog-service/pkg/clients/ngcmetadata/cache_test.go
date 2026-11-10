// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ngcmetadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func enableMetadataCache() {
	GinkgoHelper()
	cachetesting.EnableCache()
	Expect(cache.IsInitialized()).To(BeTrue())
	DeferCleanup(cachetesting.DisableCache)
}

var _ = It("metadata cache reuses profiles across clients", func() {
	enableMetadataCache()
	calls := 0
	httpClient := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		body := fmt.Sprintf(`{"profiles":[{"profile_id":"upstream-%d","min_vram_per_device_gb":27}]}`, calls)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	first, err := New(httpClient, 15*time.Minute).Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	Expect(err).To(BeNil())
	second, err := New(httpClient, 15*time.Minute).Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	Expect(err).To(BeNil())
	Expect(first.Profiles[0].ID).To(Equal("upstream-1"))
	Expect(second.Profiles[0].ID).To(Equal("upstream-1"), "second client must reuse the cached public metadata")
	Expect(calls).To(BeNumerically("==", 1))
})

// Removing the image or tag from the key would reuse another release's profiles.
var _ = It("metadata cache separates images and exact tags", func() {
	enableMetadataCache()
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		id := r.URL.Query().Get("resourceId") + "@" + r.URL.Query().Get("tag")
		return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":%q}]}`, id)), nil
	})}, 15*time.Minute)
	for _, tc := range []struct{ image, tag, id string }{
		{"nvcr.io/nim/nvidia/example", "2.1.4", "nim/nvidia/example@2.1.4"},
		{"nvcr.io/nim/nvidia/example", "latest", "nim/nvidia/example@latest"},
		{"nvcr.io/nim/meta/example", "2.1.4", "nim/meta/example@2.1.4"},
	} {
		for range 2 {
			result, err := client.Get(context.Background(), tc.image, tc.tag)
			Expect(err).To(BeNil())
			Expect(result.Profiles[0].ID).To(Equal(tc.id))
		}
	}
	Expect(calls).To(BeNumerically("==", 3))
})

var _ = It("metadata cache expires and does not serve stale data on failure", func() {
	enableMetadataCache()
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if calls == 2 {
			return metadataResponse(503, `{}`), nil
		}
		return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":"revision-%d"}]}`, calls)), nil
	})}, time.Second)
	ctx := context.Background()
	first, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(BeNil())
	Expect(first.Profiles[0].ID).To(Equal("revision-1"))
	cached, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(BeNil())
	Expect(cached.Profiles[0].ID).To(Equal("revision-1"))
	time.Sleep(1100 * time.Millisecond)
	_, err = client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(MatchError(ContainSubstring("503")))
	updated, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(BeNil())
	Expect(updated.Profiles[0].ID).To(Equal("revision-3"))
	Expect(calls).To(BeNumerically("==", 3))
})

var _ = Describe("metadata cache skips missing profiles and failures", func() {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"not published", 404, `{}`, false},
		{"empty profiles", 200, `{"profiles":[]}`, false},
		{"release-only metadata", 200, `{"release_details":{}}`, false},
		{"rate limited", 429, `{}`, true},
		{"server error", 503, `{}`, true},
		{"malformed", 200, `{`, true},
		{"transport error", 0, "", true},
	} {
		It(tc.name, func() {
			enableMetadataCache()
			calls := 0
			client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					if tc.status == 0 {
						return nil, fmt.Errorf("upstream unavailable")
					}
					return metadataResponse(tc.status, tc.body), nil
				}
				return metadataResponse(200, `{"profiles":[{"profile_id":"newly-published"}]}`), nil
			})}, 15*time.Minute)
			first, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
			Expect(err != nil).To(Equal(tc.wantError))
			Expect(first.Profiles).To(BeEmpty())
			for range 2 {
				result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
				Expect(err).To(BeNil())
				Expect(result.Profiles[0].ID).To(Equal("newly-published"))
			}
			Expect(calls).To(BeNumerically("==", 2))
		})
	}
})

var _ = Describe("metadata cache failures fall back to NGC", func() {
	for _, state := range []string{"disabled", "registration unavailable", "read and write unavailable"} {
		It(state, func() {
			enableMetadataCache()
			if state == "disabled" {
				cachetesting.DisableCache()
			}
			if state == "registration unavailable" {
				Expect(cache.Close()).To(BeNil())
			}
			calls := 0
			client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":"fresh-%d"}]}`, calls)), nil
			})}, 15*time.Minute)
			if state == "read and write unavailable" {
				Expect(cache.Close()).To(BeNil())
			}
			for i := 1; i <= 2; i++ {
				result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
				Expect(err).To(BeNil())
				Expect(result.Profiles[0].ID).To(Equal(fmt.Sprintf("fresh-%d", i)))
			}
		})
	}
})

var _ = It("metadata cache preserves large profiles and isolates caller mutation", func() {
	enableMetadataCache()
	id := strings.Repeat("a", 8192) // Exceeds the shared test cache's default record limit.
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":%q,"min_vram_per_device_gb":27,"tested_gpu_devices":["2330:10de"]}]}`, id)), nil
	})}, 15*time.Minute)
	result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	Expect(err).To(BeNil())
	result.Profiles[0].ID = "mutated"
	*result.Profiles[0].MinimumVRAMPerDeviceGB = 99
	result.Profiles[0].TestedGPUDevices[0] = "wrong-device"
	cached, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	Expect(err).To(BeNil())
	Expect(cached.Profiles[0].ID).To(Equal(id))
	Expect(*cached.Profiles[0].MinimumVRAMPerDeviceGB).To(BeNumerically("==", 27.0))
	Expect(cached.Profiles[0].TestedGPUDevices).To(Equal([]string{"2330:10de"}))
	Expect(calls).To(BeNumerically("==", 1))
})

var _ = It("metadata cache honors canceled request on hit", func() {
	enableMetadataCache()
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return metadataResponse(200, `{"profiles":[{"profile_id":"cached"}]}`), nil
	})}, 15*time.Minute)
	_, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(BeNil())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	Expect(err).To(Not(BeNil()))
	Expect(err.Details).To(ContainSubstring(context.Canceled.Error()))
	Expect(calls).To(BeNumerically("==", 1))
})

func metadataResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}
