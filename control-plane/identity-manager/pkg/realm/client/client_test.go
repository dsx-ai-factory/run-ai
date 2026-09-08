/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"reflect"
	"testing"
)

func TestGetDefaultRedirectUris(t *testing.T) {
	testCases := []struct {
		name            string
		tenantUrl       string
		appHost         string
		oldRedirectUris *[]string
		expected        *[]string
	}{
		{
			name:            "no old redirect uris",
			tenantUrl:       "https://tenant.run.ai",
			appHost:         "https://mock-q7z3k.example.com",
			oldRedirectUris: nil,
			expected: &[]string{
				"https://tenant.run.ai/login/callback",
				"https://tenant.run.ai/login/callback?runai_flow=invite",
				"https://tenant.run.ai/oauth-code",
				"https://tenant.run.ai",
				"https://mock-q7z3k.example.com/login/callback",
				"https://mock-q7z3k.example.com/oauth-code",
				"http://localhost",
			},
		},
		{
			name:      "with old redirect uris: broad wildcards and bare appHost are stripped, specific paths kept",
			tenantUrl: "https://tenant.run.ai",
			appHost:   "https://mock-q7z3k.example.com",
			oldRedirectUris: &[]string{
				"https://tenant.run.ai/*",
				"https://tenant.run.ai/login/callback/*",
				"https://abcd.tenant.run.ai",
				"https://mock-q7z3k.example.com",
				"https://mock-q7z3k.example.com/*",
				"https://mock-q7z3k.example.com/login/callback",
				"https://mock-q7z3k.example.com/login/callback/*",
				"https://mock-q7z3k.example.com/oauth-code",
				"http://localhost:*",
			},
			expected: &[]string{
				"https://tenant.run.ai/login/callback",
				"https://tenant.run.ai/login/callback?runai_flow=invite",
				"https://tenant.run.ai/oauth-code",
				"https://tenant.run.ai",
				"https://mock-q7z3k.example.com/login/callback",
				"https://mock-q7z3k.example.com/oauth-code",
				"http://localhost",
				"https://abcd.tenant.run.ai",
			},
		},
		{
			name:      "filters out legacy grafana redirect uris and broad wildcards",
			tenantUrl: "https://tenant.run.ai",
			appHost:   "https://mock-q7z3k.example.com",
			oldRedirectUris: &[]string{
				"https://tenant.run.ai/*",
				"https://grafana-tenant.run.ai/login/generic_oauth",
				"https://abcd.tenant.run.ai",
				"https://other-sso.run.ai/login/generic_oauth",
				"http://localhost:*",
			},
			expected: &[]string{
				"https://tenant.run.ai/login/callback",
				"https://tenant.run.ai/login/callback?runai_flow=invite",
				"https://tenant.run.ai/oauth-code",
				"https://tenant.run.ai",
				"https://mock-q7z3k.example.com/login/callback",
				"https://mock-q7z3k.example.com/oauth-code",
				"http://localhost",
				"https://abcd.tenant.run.ai",
				"https://other-sso.run.ai/login/generic_oauth",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := getDefaultRedirectUris(tc.tenantUrl, tc.appHost, tc.oldRedirectUris)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
