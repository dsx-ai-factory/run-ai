/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package policies manages legacy workload policies and policies identified by
// Kubernetes Group, Version, and Kind (GVK).
package policies

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerGetWorkloadPolicyMapping(r, client)
	registerListWorkloadPolicies(r, client)
	registerManageWorkloadPolicy(r, client)
}
