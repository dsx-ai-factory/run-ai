/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func countWorkloads(ctx context.Context, client *runaiclient.Client, filters []string, search string, deleted bool) (int64, error) {
	req := client.Workloads.WorkloadsAPI.CountWorkloads(ctx).Precise(true)
	if deleted {
		req = req.Deleted(true)
	}
	if len(filters) > 0 {
		req = req.FilterBy(filters)
	}
	if search != "" {
		req = req.Search(search)
	}

	resp, httpResp, err := req.Execute()
	if err != nil {
		return 0, runaimcp.MapHTTPError(httpResp, err)
	}
	if resp == nil {
		return 0, fmt.Errorf("nil count response from upstream: %w", runaimcp.ErrUpstream)
	}
	if resp.Count < 0 {
		return 0, fmt.Errorf("negative workload count from upstream: %w", runaimcp.ErrUpstream)
	}
	return resp.Count, nil
}
