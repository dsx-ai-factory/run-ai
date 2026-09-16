/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	runaicluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

type EnsureCluster struct {
	ClusterName string
	Domain      string
	CACertPath  string
}

func (a EnsureCluster) Kind() string { return "runai.ensure_cluster" }

func (a EnsureCluster) Args() map[string]any {
	return map[string]any{
		"cluster": a.ClusterName,
		"domain":  a.Domain,
	}
}

func (a EnsureCluster) Describe() string {
	return fmt.Sprintf("ensure cluster %q is registered with control plane at %s", a.ClusterName, a.Domain)
}

func (a EnsureCluster) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	url := ControlPlaneURL(a.Domain)
	apiClient, err := deps.RunAIClientProvider(url, a.CACertPath)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	uuid, err := FindClusterUUID(ctx, apiClient, a.ClusterName)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if uuid != "" {
		return orchestrator2.ActionResult{}, nil
	}
	_, statusCode, err := apiClient.CreateCluster(ctx, runaicluster.ClusterCreationRequest{Name: a.ClusterName, Domain: &url})
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return orchestrator2.ActionResult{}, fmt.Errorf("create cluster %q: unexpected status %d", a.ClusterName, statusCode)
	}
	return orchestrator2.ActionResult{}, nil
}
