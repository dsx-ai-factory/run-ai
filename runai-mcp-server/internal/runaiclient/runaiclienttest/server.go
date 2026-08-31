/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaiclienttest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth/mock"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
)

func NewServer(t *testing.T) (*runaiclient.Client, *http.ServeMux) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	client := runaiclient.New(srv.URL, &mock.FakeProvider{TokenValue: "test-token"})
	return client, mux
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// PermScopesPath is the permitted-scopes endpoint the RBAC probe (common.PermittedScopeForAction) hits.
const PermScopesPath = "/api/v1/authorization/permitted-scopes"

// PermTenantHandler serves permitted-scopes granting the listed actions at tenant-wide (system)
// reach and denying the rest — for tests of tenant-scoped tools.
func PermTenantHandler(grant ...string) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, a := range grant {
		allowed[a] = true
	}
	scope := func(action string) map[string]any { return map[string]any{"system": allowed[action]} }
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]any{
			"create": scope("create"), "read": scope("read"),
			"update": scope("update"), "delete": scope("delete"),
		})
	}
}

func BoolPtr(b bool) *bool { return &b }
