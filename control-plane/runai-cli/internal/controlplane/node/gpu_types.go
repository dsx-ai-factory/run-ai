// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package node

import (
	"context"
	"sort"
)

// GPUTypes returns the exact product labels reported by the selected cluster.
// Inventory alone is sufficient; completion does not need telemetry.
func (bns *BackendNodeService) GPUTypes(ctx context.Context, clusterID string) ([]string, error) {
	nodes, err := FetchAllNodes(ctx, bns.nodeClient, clusterID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, node := range nodes {
		if info := node.GpuInfo.Get(); info != nil && info.GetGpuType() != "" {
			seen[info.GetGpuType()] = true
		}
	}
	labels := make([]string, 0, len(seen))
	for label := range seen {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels, nil
}
