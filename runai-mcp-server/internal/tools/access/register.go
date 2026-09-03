/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

func Register(r *runaimcp.Registry, client *runaiclient.Client) {
	registerListRoles(r, client)
	registerListAccessRules(r, client)
	// manage_user_access is intentionally not registered for now: creating a user or
	// service-account returns a one-time credential (password / client secret) in the
	// tool output, which would land in the model context and the transcript. Keep the
	// code; re-expose the tool once the credential is delivered out-of-band.
	_ = registerManageUserAccess
}
