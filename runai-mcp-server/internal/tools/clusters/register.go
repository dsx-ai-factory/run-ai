/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusters

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

// Register wires the cluster-service-backed tools.
func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListClusters(r, client)
	registerListNodePools(r, client)
	registerManageNodePool(r, client)
	registerGetClusterPhysicalInventory(r, client)
	registerGetClusterInfrastructureHealth(r, client)
}
