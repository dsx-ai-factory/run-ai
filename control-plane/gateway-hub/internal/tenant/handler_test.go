/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/gateway-hub/internal/tenant"
)

const headerTenantId = "X-Runai-Tenant-Id"

func TestMain(m *testing.M) {
	os.Setenv("SERVICE_NAME", "gateway-hub-test")
	os.Setenv("SECRETS_MOUNT_PATH", ".")
	os.Exit(m.Run())
}

// tenantsManagerServer fakes the tenants-manager resolve endpoint: known
// tenant names resolve to an id, "boom" returns 500, anything else 404.
func tenantsManagerServer(t *testing.T, tenants map[string]int32, calls *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/internal/api/v1/tenants/resolve" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		name := r.URL.Query().Get("tenantName")
		if name == "boom" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id, ok := tenants[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp, _ := json.Marshal(map[string]any{"id": id, "name": name})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(resp)
	}))
}

type tenantFlowCase struct {
	name string

	forwardedHost string
	host          string

	wantTenantId string // "" means the header must be absent
	wantCalls    int32
}

func TestTenantHandler_Flow(t *testing.T) {
	tenants := map[string]int32{
		"my-tenant":    42,
		"consoletest1": 7,
		"localhost":    3,
	}

	cases := []tenantFlowCase{
		{
			name:          "tenant resolved from forwarded host",
			forwardedHost: "my-tenant.example.com",
			wantTenantId:  "42",
			wantCalls:     1,
		},
		{
			name:          "port stripped from forwarded host",
			forwardedHost: "my-tenant.run.ai:443",
			wantTenantId:  "42",
			wantCalls:     1,
		},
		{
			name:          "forwarded host takes precedence over host",
			forwardedHost: "consoletest1.staging.run.ai",
			host:          "my-tenant.example.com",
			wantTenantId:  "7",
			wantCalls:     1,
		},
		{
			name:         "falls back to host header",
			host:         "my-tenant.example.com",
			wantTenantId: "42",
			wantCalls:    1,
		},
		{
			name:          "bare hostname resolved as-is",
			forwardedHost: "localhost:8080",
			wantTenantId:  "3",
			wantCalls:     1,
		},
		{
			name:      "empty host skips tenants-manager",
			wantCalls: 0,
		},
		{
			name:          "unknown tenant passes through without header",
			forwardedHost: "unknown.example.com",
			wantCalls:     1,
		},
		{
			name:          "tenants-manager error passes through without header",
			forwardedHost: "boom.example.com",
			wantCalls:     1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var calls atomic.Int32
			srv := tenantsManagerServer(t, tenants, &calls)
			defer srv.Close()

			handler := tenant.NewHandler(srv.URL)
			req := httptest.NewRequest(http.MethodGet, "/internal/plugins/extract-tenant-id", nil)
			req.Host = tc.host
			if tc.forwardedHost != "" {
				req.Header.Set("X-Forwarded-Host", tc.forwardedHost)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get(headerTenantId); got != tc.wantTenantId {
				t.Errorf("%s = %q, want %q", headerTenantId, got, tc.wantTenantId)
			}
			if calls.Load() != tc.wantCalls {
				t.Errorf("tenants-manager calls = %d, want %d", calls.Load(), tc.wantCalls)
			}
		})
	}
}

func TestTenantHandler_TenantsManagerUnreachable(t *testing.T) {
	var calls atomic.Int32
	srv := tenantsManagerServer(t, nil, &calls)
	srv.Close() // unreachable

	handler := tenant.NewHandler(srv.URL)
	req := httptest.NewRequest(http.MethodGet, "/internal/plugins/extract-tenant-id", nil)
	req.Header.Set("X-Forwarded-Host", "my-tenant.example.com")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(headerTenantId); got != "" {
		t.Errorf("%s = %q, want absent", headerTenantId, got)
	}
}

func TestTenantHandler_TenantNameIsQueryEscaped(t *testing.T) {
	received := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.URL.Query().Get("tenantName")
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	handler := tenant.NewHandler(srv.URL)
	req := httptest.NewRequest(http.MethodGet, "/internal/plugins/extract-tenant-id", nil)
	req.Header.Set("X-Forwarded-Host", "a&b=c.example.com")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	select {
	case got := <-received:
		if got != "a&b=c" {
			t.Errorf("tenantName = %q, want %q", got, "a&b=c")
		}
	default:
		t.Fatal("tenants-manager was not called")
	}
}
