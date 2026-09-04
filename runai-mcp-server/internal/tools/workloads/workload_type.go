/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

const (
	workloadTypeTraining    = "Training"
	workloadTypeWorkspace   = "Workspace"
	workloadTypeDistributed = "Distributed"
	workloadTypeInference   = "Inference"
	workloadTypeDistribInf  = "DistributedInference"
)

func isInferenceType(t string) bool {
	return t == workloadTypeInference || t == workloadTypeDistribInf
}

func resolveWorkloadType(ctx context.Context, client *runaiclient.Client, workloadID string) (string, bool, error) {
	wl, httpResp, err := client.Workloads.WorkloadsAPI.GetWorkload(ctx, workloadID).Execute()
	if err == nil {
		return wl.Type, true, nil
	}
	if httpResp != nil && (httpResp.StatusCode == http.StatusNotFound || httpResp.StatusCode == http.StatusForbidden) {
		return "", false, fmt.Errorf("get workload %q: %w", workloadID, runaimcp.ErrNotFound)
	}
	zerolog.Ctx(ctx).Warn().Err(err).Str("workload_id", workloadID).Msg("workload type lookup failed, omitting type")
	return "", false, nil
}
