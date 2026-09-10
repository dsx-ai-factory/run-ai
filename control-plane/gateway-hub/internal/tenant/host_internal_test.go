/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant

import "testing"

func TestExtractTenantNameFromHost(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     string
	}{
		{
			name:     "subdomain with domain and port",
			hostname: "my-tenant.example.com:8080",
			want:     "my-tenant",
		},
		{
			name:     "subdomain with domain",
			hostname: "my-tenant.example.com",
			want:     "my-tenant",
		},
		{
			name:     "bare hostname with port",
			hostname: "localhost:8080",
			want:     "localhost",
		},
		{
			name:     "bare hostname without port",
			hostname: "localhost",
			want:     "localhost",
		},
		{
			name:     "empty hostname",
			hostname: "",
			want:     "",
		},
		{
			name:     "multiple subdomains",
			hostname: "tenant.multi-tenant.example.com",
			want:     "tenant",
		},
		{
			name:     "IP address with port",
			hostname: "192.168.1.1:8080",
			want:     "192",
		},
		{
			name:     "subdomain with trailing port",
			hostname: "my-tenant.run.ai:443",
			want:     "my-tenant",
		},
		{
			name:     "subdomain with staging",
			hostname: "consoletest1.mock-q7z3k.example.com",
			want:     "consoletest1",
		},
		{
			name:     "run.ai domain",
			hostname: "consoletest1.run.ai",
			want:     "consoletest1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTenantNameFromHost(tt.hostname)
			if got != tt.want {
				t.Errorf("extractTenantNameFromHost(%q) = %q, want %q", tt.hostname, got, tt.want)
			}
		})
	}
}
