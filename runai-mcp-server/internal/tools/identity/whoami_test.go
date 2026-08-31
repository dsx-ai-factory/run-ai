/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth/mock"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWhoamiRun(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		want    whoamiOutput
		wantErr error
	}{
		{
			name: "full_claims",
			token: jwtWithClaims(t, map[string]any{
				"sub":                "user-123",
				"preferred_username": "alice",
				"email":              "alice@example.com",
				"name":               "Alice Example",
				"tenantId":           "tenant-1",
				"typ":                "Bearer",
			}),
			want: whoamiOutput{
				Subject:   "user-123",
				Username:  "alice",
				Email:     "alice@example.com",
				Name:      "Alice Example",
				TenantID:  "tenant-1",
				TokenType: "Bearer",
			},
		},
		{
			name: "fallback_claims",
			token: jwtWithClaims(t, map[string]any{
				"sub":         "user-456",
				"username":    "bob",
				"given_name":  "Bob",
				"family_name": "Builder",
				"tenant_id":   "tenant-2",
				"token_type":  "Application",
			}),
			want: whoamiOutput{
				Subject:   "user-456",
				Username:  "bob",
				Name:      "Bob Builder",
				TenantID:  "tenant-2",
				TokenType: "Application",
			},
		},
		{
			name:  "non_jwt_token_yields_empty_output",
			token: "not-a-jwt",
			want:  whoamiOutput{},
		},
		{
			name:  "malformed_payload_yields_empty_output",
			token: "header.@@@not-base64@@@.sig",
			want:  whoamiOutput{},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := runaiclient.New("http://unused", &mock.FakeProvider{TokenValue: tc.token})
			got, err := whoamiRun(context.Background(), c)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWhoamiRun_TokenError(t *testing.T) {
	c := runaiclient.New("http://unused", &mock.FakeProvider{TokenErr: errors.New("no token on request")})
	_, err := whoamiRun(context.Background(), c)
	assert.ErrorIs(t, err, runaimcp.ErrAuth)
}

func jwtWithClaims(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	require.NoError(t, err, "marshal claims")
	return fmt.Sprintf("header.%s.sig", base64.RawURLEncoding.EncodeToString(payload))
}
