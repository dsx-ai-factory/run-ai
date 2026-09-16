// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import "testing"

func TestCreateClusterApiConfigUsesRequestedServer(t *testing.T) {
	const serverURL = "https://cluster.example.com/cluster-api"

	configuration, err := CreateClusterApiConfig(serverURL, "token")
	if err != nil {
		t.Fatalf("create configuration: %v", err)
	}
	if len(configuration.Servers) != 1 {
		t.Fatalf("expected one server, got %d", len(configuration.Servers))
	}
	if configuration.Servers[0].URL != serverURL {
		t.Fatalf("expected server URL %q, got %q", serverURL, configuration.Servers[0].URL)
	}
}
