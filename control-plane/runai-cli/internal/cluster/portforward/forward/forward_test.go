/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package forward

import "testing"

func TestGetAddressUrl(t *testing.T) {
	for _, tc := range []struct {
		name    string
		address string
		port    int32
		want    string
	}{
		{name: "ipv4 loopback", address: "127.0.0.1", port: 8080, want: "127.0.0.1:8080"},
		{name: "hostname", address: "localhost", port: 18082, want: "localhost:18082"},
		{name: "ipv6 loopback", address: "::1", port: 18082, want: "[::1]:18082"},
		{name: "ipv6 any", address: "::", port: 80, want: "[::]:80"},
		{name: "bracketed ipv6 loopback", address: "[::1]", port: 18082, want: "[::1]:18082"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := getAddressUrl(tc.address, tc.port); got != tc.want {
				t.Fatalf("getAddressUrl(%q, %d) = %q, want %q", tc.address, tc.port, got, tc.want)
			}
		})
	}
}
