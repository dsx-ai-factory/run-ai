// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package nodepool

import (
	"context"
	"sort"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/clients"
)

// GetGPUTypes uses minimal inventory so inference users do not need nodes/read.
func GetGPUTypes(ctx context.Context, domain, token, clusterID string) ([]string, error) {
	client, err := clients.NewClusterService(domain, token)
	if err != nil {
		return nil, err
	}
	pools, err := NewBackendNodepoolService(client, nil).GetNodepoolMinimalList(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, pool := range pools.Nodepools {
		for _, gpu := range pool.GpuTypes {
			if gpu.Type != "" {
				seen[gpu.Type] = true
			}
		}
	}
	labels := make([]string, 0, len(seen))
	for label := range seen {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	return labels, nil
}
