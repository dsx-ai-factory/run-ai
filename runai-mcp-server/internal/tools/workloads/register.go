/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListWorkloads(r, client)
	registerGetWorkloadStatus(r, client)
	registerGetWorkloadPods(r, client)
	registerGetWorkloadMetrics(r, client)
	registerGetWorkloadHistory(r, client)
	registerGetWorkloadSpec(r, client)
	registerGetWorkloadsSummary(r, client)
	registerGetNodePods(r, client)
	registerGetWorkloadEffectivePolicy(r, client)
	registerSubmitWorkload(r, client)
	registerManageWorkloadOperations(r, client)
}
