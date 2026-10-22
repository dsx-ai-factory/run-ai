/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package gpuinventory consumes existing node-pool hardware metadata. It does
// not discover devices, evaluate project quota, or report free capacity.
package gpuinventory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	clusterapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/google/uuid"
)

type NodePoolsClient interface {
	GetNodePoolsMinimalByCluster(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error)
}

type Hardware struct {
	Type string
	// MemoryBytes is nil when nodes do not report a usable per-device capacity.
	MemoryBytes *int64
}

type Client struct {
	nodePools      NodePoolsClient
	requestTimeout time.Duration
}

func New(nodePools NodePoolsClient, requestTimeout time.Duration) *Client {
	return &Client{nodePools: nodePools, requestTimeout: requestTimeout}
}

func (c *Client) Get(ctx context.Context, clusterID string) ([]Hardware, error) {
	if err := uuid.Validate(clusterID); err != nil {
		return nil, fmt.Errorf("invalid cluster ID: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	pools, err := c.nodePools.GetNodePoolsMinimalByCluster(ctx, clusterID)
	if err != nil {
		return nil, fmt.Errorf("retrieve cluster GPU inventory: %w", err)
	}
	if pools == nil {
		return nil, fmt.Errorf("cluster GPU inventory returned no response")
	}
	type key struct {
		gpuType string
		memory  int64
	}
	seen := make(map[key]bool)
	result := make([]Hardware, 0)
	for _, pool := range pools.Nodepools {
		for _, gpu := range pool.GpuTypes {
			name := strings.TrimSpace(gpu.Type)
			if gpu.TotalDevices < 0 || (gpu.TotalDevices > 0 && name == "") {
				return nil, fmt.Errorf("cluster GPU inventory contains an invalid device count or missing GPU type")
			}
			if gpu.TotalDevices == 0 {
				continue
			}
			var memory int64
			if gpu.GpuMemoryBytes != nil && *gpu.GpuMemoryBytes > 0 {
				memory = *gpu.GpuMemoryBytes
			}
			k := key{name, memory}
			if seen[k] {
				continue
			}
			seen[k] = true
			hardware := Hardware{Type: name}
			if memory > 0 {
				hardware.MemoryBytes = &memory
			}
			result = append(result, hardware)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Type != result[j].Type {
			return result[i].Type < result[j].Type
		}
		if result[i].MemoryBytes == nil {
			return result[j].MemoryBytes != nil
		}
		if result[j].MemoryBytes == nil {
			return false
		}
		return *result[i].MemoryBytes < *result[j].MemoryBytes
	})
	return result, nil
}
