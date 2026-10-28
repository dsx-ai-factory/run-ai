/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/access"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/aiapplications"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/clusters"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/computeresources"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/credentials"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/environments"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/identity"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/orgunits"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/policies"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/resources"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/settings"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/templates"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/workloads"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	identity.Register(r, client)
	access.Register(r, client)
	aiapplications.Register(r, client)
	clusters.Register(r, client)
	computeresources.Register(r, client)
	credentials.Register(r, client)
	environments.Register(r, client)
	metrics.Register(r, client)
	orgunits.Register(r, client)
	policies.Register(r, client)
	resources.Register(r, client)
	settings.Register(r, client)
	templates.Register(r, client)
	workloads.Register(r, client)
}
