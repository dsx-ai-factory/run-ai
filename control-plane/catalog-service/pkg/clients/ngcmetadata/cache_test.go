// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package ngcmetadata

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/stretchr/testify/require"
)

func enableMetadataCache(t *testing.T) {
	t.Helper()
	cachetesting.EnableCache()
	require.True(t, cache.IsInitialized())
	t.Cleanup(cachetesting.DisableCache)
}

func TestMetadataCacheReusesProfilesAcrossClients(t *testing.T) {
	enableMetadataCache(t)
	calls := 0
	httpClient := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		body := fmt.Sprintf(`{"profiles":[{"profile_id":"upstream-%d","min_vram_per_device_gb":27}]}`, calls)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}
	first, err := New(httpClient, 15*time.Minute).Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	require.NoError(t, err)
	second, err := New(httpClient, 15*time.Minute).Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	require.NoError(t, err)
	require.Equal(t, "upstream-1", first.Profiles[0].ID)
	require.Equal(t, "upstream-1", second.Profiles[0].ID, "second client must reuse the cached public metadata")
	require.Equal(t, 1, calls)
}

// Removing the image or tag from the key would reuse another release's profiles.
func TestMetadataCacheSeparatesImagesAndExactTags(t *testing.T) {
	enableMetadataCache(t)
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
			require.NoError(t, err)
			require.Equal(t, tc.id, result.Profiles[0].ID)
		}
	}
	require.Equal(t, 3, calls)
}

func TestMetadataCacheExpiresAndDoesNotServeStaleDataOnFailure(t *testing.T) {
	enableMetadataCache(t)
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
	require.NoError(t, err)
	require.Equal(t, "revision-1", first.Profiles[0].ID)
	cached, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	require.NoError(t, err)
	require.Equal(t, "revision-1", cached.Profiles[0].ID)
	time.Sleep(1100 * time.Millisecond)
	_, err = client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	require.ErrorContains(t, err, "503")
	updated, err := client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	require.NoError(t, err)
	require.Equal(t, "revision-3", updated.Profiles[0].ID)
	require.Equal(t, 3, calls)
}

func TestMetadataCacheSkipsMissingProfilesAndFailures(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			enableMetadataCache(t)
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
			require.Equal(t, tc.wantError, err != nil)
			require.Empty(t, first.Profiles)
			for range 2 {
				result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
				require.NoError(t, err)
				require.Equal(t, "newly-published", result.Profiles[0].ID)
			}
			require.Equal(t, 2, calls)
		})
	}
}

func TestMetadataCacheFailuresFallBackToNGC(t *testing.T) {
	for _, state := range []string{"disabled", "registration unavailable", "read and write unavailable"} {
		t.Run(state, func(t *testing.T) {
			enableMetadataCache(t)
			if state == "disabled" {
				cachetesting.DisableCache()
			}
			if state == "registration unavailable" {
				require.NoError(t, cache.Close())
			}
			calls := 0
			client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":"fresh-%d"}]}`, calls)), nil
			})}, 15*time.Minute)
			if state == "read and write unavailable" {
				require.NoError(t, cache.Close())
			}
			for i := 1; i <= 2; i++ {
				result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
				require.NoError(t, err)
				require.Equal(t, fmt.Sprintf("fresh-%d", i), result.Profiles[0].ID)
			}
		})
	}
}

func TestMetadataCachePreservesLargeProfilesAndIsolatesCallerMutation(t *testing.T) {
	enableMetadataCache(t)
	id := strings.Repeat("a", 8192) // Exceeds the shared test cache's default record limit.
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return metadataResponse(200, fmt.Sprintf(`{"profiles":[{"profile_id":%q,"min_vram_per_device_gb":27,"tested_gpu_devices":["2330:10de"]}]}`, id)), nil
	})}, 15*time.Minute)
	result, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	require.NoError(t, err)
	result.Profiles[0].ID = "mutated"
	*result.Profiles[0].MinimumVRAMPerDeviceGB = 99
	result.Profiles[0].TestedGPUDevices[0] = "wrong-device"
	cached, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "2.1.4")
	require.NoError(t, err)
	require.Equal(t, id, cached.Profiles[0].ID)
	require.Equal(t, 27.0, *cached.Profiles[0].MinimumVRAMPerDeviceGB)
	require.Equal(t, []string{"2330:10de"}, cached.Profiles[0].TestedGPUDevices)
	require.Equal(t, 1, calls)
}

func TestMetadataCacheHonorsCanceledRequestOnHit(t *testing.T) {
	enableMetadataCache(t)
	calls := 0
	client := New(&http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return metadataResponse(200, `{"profiles":[{"profile_id":"cached"}]}`), nil
	})}, 15*time.Minute)
	_, err := client.Get(context.Background(), "nvcr.io/nim/nvidia/example", "latest")
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = client.Get(ctx, "nvcr.io/nim/nvidia/example", "latest")
	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, calls)
}

func metadataResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}
