/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package access

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

type listRolesInput struct {
	Limit  int32 `json:"limit,omitempty"  jsonschema:"Maximum number of roles to return. Defaults to 50, capped at 200."`
	Offset int32 `json:"offset,omitempty" jsonschema:"Pagination cursor from a previous response's nextOffset. Defaults to 0."`
}

func (in listRolesInput) Validate() error {
	return common.ValidatePagination(in.Limit, in.Offset)
}

type roleInfo struct {
	ID          int32  `json:"id"                  jsonschema:"Role ID — pass as roleId to manage_user_access."`
	Name        string `json:"name"                jsonschema:"Role name — pass as roleName to manage_user_access."`
	Description string `json:"description"         jsonschema:"Human-readable description of what the role grants."`
	ScopeType   string `json:"scopeType,omitempty" jsonschema:"Scope level the role is defined for, when the platform sets one."`
	Deprecated  bool   `json:"deprecated"          jsonschema:"True when the role is deprecated and should not be used for new grants."`
}

type listRolesOutput struct {
	Roles      []roleInfo `json:"roles"                jsonschema:"Roles available to grant. Always present; empty array when none match."`
	HasMore    bool       `json:"hasMore"              jsonschema:"True when more roles exist beyond the returned page; pass nextOffset back as offset."`
	NextOffset int32      `json:"nextOffset,omitempty" jsonschema:"Offset to pass on the next call to fetch more roles. Only set when hasMore is true."`
}

func registerListRoles(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[listRolesInput, listRolesOutput]{
		Name: "list_roles",
		Description: "List the roles that can be granted via manage_user_access — id, name, description, and whether " +
			"the role is deprecated. Call this to discover a valid roleName/roleId before creating an access-rule. Does " +
			"NOT show who holds a role (use list_access_rules for that).",
		ReadOnly: true,
		Handler: func(ctx context.Context, in listRolesInput) (listRolesOutput, error) {
			return listRolesRun(ctx, client, in)
		},
	})
}

func listRolesRun(ctx context.Context, client *runaiclient.Client, in listRolesInput) (listRolesOutput, error) {
	resp, httpResp, err := client.Authorization.RolesAPI.GetRoles(ctx).Execute()
	if err != nil {
		return listRolesOutput{}, fmt.Errorf("list roles: %w", runaimcp.MapHTTPError(httpResp, err))
	}
	roles := []roleInfo{}
	if resp != nil {
		for _, role := range resp.Roles {
			ri := roleInfo{
				ID:          role.Id,
				Name:        role.Name,
				Description: role.Description,
				Deprecated:  role.Deprecated != nil && *role.Deprecated,
			}
			if role.ScopeType != nil {
				ri.ScopeType = *role.ScopeType
			}
			roles = append(roles, ri)
		}
	}
	page, hasMore, next := common.Paginate(roles, in.Offset, common.EffectiveLimit(in.Limit))
	return listRolesOutput{Roles: page, HasMore: hasMore, NextOffset: next}, nil
}
