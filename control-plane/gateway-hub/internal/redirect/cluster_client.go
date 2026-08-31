/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redirect

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const clusterURLCacheTTL = 5 * time.Minute

type ClusterURLProvider interface {
	GetClusterURL(ctx context.Context, clusterUUID string) (string, error)
}

type clusterURLEntry struct{ URL string }

var cacheInitOnce sync.Once

func ensureCacheReady() {
	cacheInitOnce.Do(func() {
		if err := cache.Init(nil); err != nil {
			logging.Errorf("failed to initialize cache: %v", err)
			return
		}
		if err := cache.RegisterStore[clusterURLEntry](options.Ttl(clusterURLCacheTTL)); err != nil {
			logging.Errorf("failed to register cluster URL cache store: %v", err)
		}
	})
}

type clusterServiceClient struct {
	api openapi.ClustersAPI
}

func newClusterServiceClient(baseURL string) *clusterServiceClient {
	cfg := openapi.NewConfiguration()
	cfg.Host = strings.TrimPrefix(strings.TrimPrefix(baseURL, "https://"), "http://")
	cfg.Scheme = "http"
	cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	return &clusterServiceClient{
		api: openapi.NewAPIClient(cfg).ClustersAPI,
	}
}

func (c *clusterServiceClient) GetClusterURL(ctx context.Context, clusterUUID string) (string, error) {
	ensureCacheReady()

	if entry, err := cache.Get[clusterURLEntry](ctx, clusterUUID); err == nil && entry != nil {
		return entry.URL, nil
	}

	resp, _, err := c.api.GetClusterUrlExecute(c.api.GetClusterUrl(ctx, clusterUUID))
	if err != nil {
		return "", fmt.Errorf("cluster-service lookup for %q: %w", clusterUUID, err)
	}

	clusterURL := resp.GetUrl()
	if clusterURL == "" {
		return "", fmt.Errorf("cluster-service returned empty URL for cluster %q — domain not yet registered", clusterUUID)
	}
	if err := cache.Set(ctx, clusterUUID, clusterURLEntry{URL: clusterURL}); err != nil {
		logging.Errorf("failed to cache cluster URL for %q: %v", clusterUUID, err)
	}
	return clusterURL, nil
}
