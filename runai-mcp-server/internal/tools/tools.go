/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tools

import (
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/access"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/aiapplications"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/clusters"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/credentials"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/identity"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/metrics"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/orgunits"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/resources"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/settings"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/templates"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/workloads"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	identity.Register(r, client)
	access.Register(r, client)
	aiapplications.Register(r, client)
	clusters.Register(r, client)
	credentials.Register(r, client)
	metrics.Register(r, client)
	orgunits.Register(r, client)
	resources.Register(r, client)
	settings.Register(r, client)
	templates.Register(r, client)
	workloads.Register(r, client)
}
