/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

func TestMatchProjectAccelerator(t *testing.T) {
	const vendor = "example.org/gpu"
	const key = "hardware.example.org/model"
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"exact match", `{"items":[{"resourceName":"example.org/gpu","devices":[{"key":"hardware.example.org/model","value":"X2"}]}]}`, true},
		{"no matching node", `{"items":[]}`, false},
		{"response containing a different vendor", `{"items":[{"resourceName":"other.example/gpu","devices":[{"key":"hardware.example.org/model","value":"X2"}]}]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/api/v1/org-unit/projects/42/accelerators" {
					t.Errorf("unexpected inventory request: %s %s", r.Method, r.URL.Path)
				}
				query := r.URL.Query()
				if query.Get("resourceName") != vendor || query.Get("deviceLabelKey") != key || query.Get("gpuType") != "X2" {
					t.Errorf("unexpected query: %v", query)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			config := openapi.NewConfiguration()
			config.Servers = openapi.ServerConfigurations{{URL: server.URL}}
			c := OrgUnitClient{ApiClient: openapi.NewAPIClient(config)}
			got, err := c.MatchProjectAccelerator(context.Background(), "42", vendor, key, "X2")
			if err != nil || got != tc.want {
				t.Fatalf("match=%t, error=%v; want %t", got, err, tc.want)
			}
		})
	}
}
