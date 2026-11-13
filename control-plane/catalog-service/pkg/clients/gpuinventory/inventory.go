/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package gpuinventory consumes existing node-pool hardware metadata. It does
// not discover devices, evaluate project quota, or report free capacity.
package gpuinventory

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	clusterapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/google/uuid"
)

type ClusterClient interface {
	GetClusterMinimalDataById(context.Context, string) (*clusterapi.ClustersMinimalData, error)
	GetNodePoolsMinimalByCluster(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error)
}

type Hardware struct {
	Type string
	// MemoryBytes is nil when nodes do not report a usable per-device capacity.
	MemoryBytes *int64
	// Largest reported count on one node for this exact type/capacity; zero is unknown.
	MaxDevicesPerNode int64
}

type Client struct {
	cluster        ClusterClient
	requestTimeout time.Duration
}

func New(cluster ClusterClient, requestTimeout time.Duration) *Client {
	return &Client{
		cluster:        cluster,
		requestTimeout: requestTimeout,
	}
}

func (c *Client) Get(ctx context.Context, clusterID string) ([]Hardware, *api.Error) {
	if err := uuid.Validate(clusterID); err != nil {
		return nil, api.NewError(http.StatusBadRequest, "Invalid cluster ID", err)
	}
	ctx, cancel := context.WithTimeout(ctx, c.requestTimeout)
	defer cancel()
	// An empty filtered list cannot establish cluster existence. Use the
	// existing authorized minimal lookup before reading node-pool inventory.
	cluster, err := c.cluster.GetClusterMinimalDataById(ctx, clusterID)
	if err != nil {
		return nil, inventoryError(err)
	}
	if cluster == nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "Cluster lookup returned no response", nil)
	}
	pools, err := c.cluster.GetNodePoolsMinimalByCluster(ctx, clusterID)
	if err != nil {
		return nil, inventoryError(err)
	}
	if pools == nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "cluster GPU inventory returned no response", nil)
	}
	type hardwareKey struct {
		gpuType string
		memory  int64
	}
	indices := make(map[hardwareKey]int)
	result := make([]Hardware, 0)
	for _, pool := range pools.Nodepools {
		for _, gpu := range pool.GpuTypes {
			name := strings.TrimSpace(gpu.Type)
			if gpu.TotalDevices < 0 || (gpu.TotalDevices > 0 && name == "") {
				return nil, api.NewError(http.StatusServiceUnavailable, "cluster GPU inventory contains an invalid device count or missing GPU type", nil)
			}
			var maxPerNode int64
			if gpu.MaxDevicesPerNode != nil {
				maxPerNode = *gpu.MaxDevicesPerNode
				if maxPerNode <= 0 || maxPerNode > int64(gpu.TotalDevices) {
					return nil, api.NewError(http.StatusServiceUnavailable, "cluster GPU inventory contains an invalid per-node device count", nil)
				}
			}
			if gpu.TotalDevices == 0 {
				continue
			}
			var memory int64
			if gpu.GpuMemoryBytes != nil && *gpu.GpuMemoryBytes > 0 {
				memory = *gpu.GpuMemoryBytes
			}
			hardwareIdentity := hardwareKey{
				gpuType: name,
				memory:  memory,
			}
			if index, exists := indices[hardwareIdentity]; exists {
				// A pool or page is not a node. Summing these counts would invent
				// a single-node layout from GPUs installed on different nodes.
				result[index].MaxDevicesPerNode = max(result[index].MaxDevicesPerNode, maxPerNode)
				continue
			}
			indices[hardwareIdentity] = len(result)
			hardware := Hardware{
				Type:              name,
				MaxDevicesPerNode: maxPerNode,
			}
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

// Preserve caller-facing access and existence errors. Dependency failures
// remain unavailable; an upstream HTTP status is not a sizing outcome.
func inventoryError(err error) *api.Error {
	status := http.StatusServiceUnavailable
	var upstream *clusterapi.Error
	if errors.As(err, &upstream) {
		switch upstream.Code {
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			status = int(upstream.Code)
		}
	}
	return &api.Error{Code: int32(status), Message: "Cluster GPU inventory is not available", Details: err.Error()}
}
