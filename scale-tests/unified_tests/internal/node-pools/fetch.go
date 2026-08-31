/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node_pools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	clusterclient "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

// NodepoolInfo holds the fields callers need from a nodepool response.
type NodepoolInfo struct {
	Name string
	ID   string
	// Phase is status.phase as reported by the cluster service, e.g. "Ready".
	// Empty when the API omits it.
	Phase string
}

// FetchNodepools reads the full nodepool list via raw HTTP body parsing.
// The generated client fails to unmarshal GetNodepools responses because the API
// returns empty status.placementStrategy.cpu/gpu fields; we bypass it and read the
// body directly with io.ReadAll to avoid the truncation risk of a single Read call.
func FetchNodepools(ctx context.Context, api clusterclient.NodePoolsAPI) ([]NodepoolInfo, error) {
	_, resp, _ := api.GetNodepools(ctx).Execute()
	if resp == nil || resp.StatusCode != 200 {
		code := 0
		if resp != nil {
			code = resp.StatusCode
		}
		return nil, fmt.Errorf("GetNodepools returned status %d", code)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read nodepools response: %w", err)
	}

	var parsed struct {
		Nodepools []struct {
			Name  string `json:"name"`
			ID    string `json:"id"`
			Phase string `json:"phase"`
		} `json:"nodepools"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse nodepools response: %w", err)
	}

	result := make([]NodepoolInfo, len(parsed.Nodepools))
	for i, np := range parsed.Nodepools {
		result[i] = NodepoolInfo{Name: np.Name, ID: np.ID, Phase: np.Phase}
	}
	return result, nil
}
