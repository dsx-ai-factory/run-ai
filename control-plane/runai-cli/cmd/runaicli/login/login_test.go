/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import "testing"

func TestListenAddress(t *testing.T) {
	for _, tc := range []struct {
		name string
		host string
		port int
		want string
	}{
		{name: "hostname", host: "localhost", port: 43121, want: "localhost:43121"},
		{name: "ipv4", host: "127.0.0.1", port: 43121, want: "127.0.0.1:43121"},
		{name: "ipv6", host: "::1", port: 43121, want: "[::1]:43121"},
		{name: "bracketed ipv6", host: "[::1]", port: 43121, want: "[::1]:43121"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := listenAddress(tc.host, tc.port); got != tc.want {
				t.Fatalf("listenAddress(%q, %d) = %q, want %q", tc.host, tc.port, got, tc.want)
			}
		})
	}
}
