/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

// manage_user_access is deliberately not exposed: it returns one-time credentials
// (password / client secret) in tool output. Guard against re-registering it.
func TestManageUserAccessNotRegistered(t *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{Name: "t", Version: "0"}, nil)
	r := runaimcp.New(srv, runaimcp.Options{Logger: zerolog.New(io.Discard)})
	Register(r, nil)

	names := r.ToolNames()
	assert.NotContains(t, names, "manage_user_access", "credential-returning tool must stay hidden")
	assert.Contains(t, names, "list_access_rules")
	assert.Contains(t, names, "list_roles")
}
