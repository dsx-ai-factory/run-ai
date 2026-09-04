/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

// Register wires the metrics (time-series consumption) tools.
func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerGetOrgUnitMetrics(r, client)
	registerGetClusterMetrics(r, client)
}
