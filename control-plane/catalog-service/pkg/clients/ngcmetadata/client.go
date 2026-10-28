/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package ngcmetadata reads published NIM execution requirements. Memory is per
// device for a specific profile, not the total memory needed by an arbitrary model.
package ngcmetadata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	metadataURL      = "https://api.ngc.nvidia.com/v2/nim/metadata"
	maxResponseBytes = 16 << 20
)

var (
	imagePattern = regexp.MustCompile(`^nvcr\.io/nim/[a-z0-9]+(?:[._-][a-z0-9]+)*/[a-z0-9]+(?:[._-][a-z0-9]+)*$`)
	tagPattern   = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`)
)

type Profile struct {
	ID                     string   `json:"profile_id"`
	GPU                    string   `json:"gpu,omitempty"`
	GPUDevice              string   `json:"gpu_device,omitempty"`
	TestedGPUDevices       []string `json:"tested_gpu_devices,omitempty"`
	Precision              string   `json:"precision,omitempty"`
	InferenceBackend       string   `json:"inference_backend,omitempty"`
	TensorParallelSize     *int     `json:"tensor_parallel_size,omitempty"`
	PipelineParallelSize   *int     `json:"pipeline_parallel_size,omitempty"`
	MaxGPUCount            *int64   `json:"max_gpu_count,omitempty"`
	MinimumVRAMPerDeviceGB *float64 `json:"min_vram_per_device_gb,omitempty"`
}

type Metadata struct {
	Profiles []Profile `json:"profiles"`
}

type Client struct {
	httpClient     *http.Client
	cacheAvailable bool
}

// New accepts the service's timeout-configured HTTP client. No registry key or
// caller JWT is needed for this public metadata endpoint.
func New(httpClient *http.Client, cacheTTL time.Duration) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	// Use the existing CP cache backend, with a separate store so catalog
	// image/tag browsing retains its own expiry and record-size policy.
	err := cache.RegisterStore[Metadata](options.Ttl(cacheTTL), options.MaxRecordSizeKB(maxResponseBytes/1024))
	if err != nil {
		logging.Warnf("Failed to register NGC metadata cache; fetching directly: %v", err)
	}
	return &Client{httpClient: httpClient, cacheAvailable: err == nil}
}

// Validate checks the image and separate tag before constructing an NGC query.
// Tags, including latest, are passed verbatim: registry updates do not imply
// that NGC has republished metadata for that tag.
func Validate(image, tag string) error {
	if !imagePattern.MatchString(image) {
		return fmt.Errorf("image must be an untagged nvcr.io/nim/<organization>/<model> reference")
	}
	if !tagPattern.MatchString(tag) {
		return fmt.Errorf("tag must be a nonempty container image tag")
	}
	return nil
}

// Get returns an empty profile list when metadata is not published (404).
// Transport, rate-limit, server, and malformed-response failures remain errors.
func (c *Client) Get(ctx context.Context, image, tag string) (Metadata, error) {
	if err := Validate(image, tag); err != nil {
		return Metadata{}, err
	}
	if err := ctx.Err(); err != nil {
		return Metadata{}, err
	}
	// Metadata is public and independent of cluster/user identity. Keep the
	// exact image and tag (including latest) in the versioned cache key.
	// v2 includes max_gpu_count, which older cached projections discarded.
	cacheKey := "ngc_metadata_v2:" + image + ":" + tag
	if c.cacheAvailable {
		cached, err := cache.Get[Metadata](ctx, cacheKey)
		if err == nil && cached != nil {
			return *cached, nil
		}
	}
	query := url.Values{
		"resourceId": {strings.TrimPrefix(image, "nvcr.io/")},
		"tag":        {tag},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataURL+"?"+query.Encode(), nil)
	if err != nil {
		return Metadata{}, fmt.Errorf("create NGC metadata request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Metadata{}, fmt.Errorf("retrieve NGC metadata: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return Metadata{Profiles: []Profile{}}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return Metadata{}, fmt.Errorf("NGC metadata returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return Metadata{}, fmt.Errorf("read NGC metadata: %w", err)
	}
	if len(body) > maxResponseBytes {
		return Metadata{}, fmt.Errorf("NGC metadata response exceeds size limit")
	}
	var metadata *Metadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return Metadata{}, fmt.Errorf("decode NGC metadata: %w", err)
	}
	if metadata == nil {
		return Metadata{}, fmt.Errorf("NGC metadata response must be an object")
	}
	// NGC permits release-only metadata; profiles is an optional section.
	if metadata.Profiles == nil {
		metadata.Profiles = []Profile{}
	}
	// Empty/absent profiles and 404s remain uncached: metadata publication
	// can lag the container release. Dependency errors also never enter cache.
	if c.cacheAvailable && len(metadata.Profiles) > 0 {
		if err := cache.Set[Metadata](ctx, cacheKey, *metadata); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set NGC metadata cache: %v", err)
		}
	}
	return *metadata, nil
}
