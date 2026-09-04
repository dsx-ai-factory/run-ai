/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaiclient"
	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/runaimcp"
)

type whoamiInput struct{}

func (whoamiInput) Validate() error { return nil }

type whoamiOutput struct {
	Subject   string `json:"subject,omitempty"    jsonschema:"The unique subject identifier (sub claim) of the authenticated principal."`
	Username  string `json:"username,omitempty"   jsonschema:"The username of the authenticated principal."`
	Email     string `json:"email,omitempty"      jsonschema:"The email address of the authenticated principal."`
	Name      string `json:"name,omitempty"       jsonschema:"The display name of the authenticated principal."`
	TenantID  string `json:"tenant_id,omitempty"  jsonschema:"The Run:ai tenant ID the principal belongs to."`
	TokenType string `json:"token_type,omitempty" jsonschema:"The type of the bearer token (e.g. Bearer or Application)."`
}

func registerWhoami(r *runaimcp.Registry, client *runaiclient.Client) {
	runaimcp.Register(r, runaimcp.Tool[whoamiInput, whoamiOutput]{
		Name: "whoami",
		Description: "Identify the currently authenticated Run:ai principal. Use this to answer " +
			"questions like \"who am I?\".",
		ReadOnly: true,
		Handler: func(ctx context.Context, _ whoamiInput) (whoamiOutput, error) {
			return whoamiRun(ctx, client)
		},
	})
}

func whoamiRun(ctx context.Context, client *runaiclient.Client) (whoamiOutput, error) {
	var out whoamiOutput

	token, err := client.Token(ctx)
	if err != nil {
		return out, fmt.Errorf("fetch bearer token: %w", runaimcp.ErrAuth)
	}
	if claims, err := decodeJWTClaims(token); err == nil {
		out.Subject = stringClaim(claims, "sub")
		out.Username = firstNonEmpty(
			stringClaim(claims, "preferred_username"),
			stringClaim(claims, "username"),
		)
		out.Email = stringClaim(claims, "email")
		out.Name = firstNonEmpty(
			stringClaim(claims, "name"),
			strings.TrimSpace(stringClaim(claims, "given_name")+" "+stringClaim(claims, "family_name")),
		)
		out.TenantID = firstNonEmpty(
			stringClaim(claims, "tenantId"),
			stringClaim(claims, "tenant_id"),
		)
		out.TokenType = firstNonEmpty(
			stringClaim(claims, "typ"),
			stringClaim(claims, "token_type"),
		)
	}

	return out, nil
}

func decodeJWTClaims(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("not a JWT")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}
	return claims, nil
}

func stringClaim(claims map[string]any, key string) string {
	v, ok := claims[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
