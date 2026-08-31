/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package hooks

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/stretchr/testify/assert"
)

func TestShouldPerformControlPlaneAssistedInstallationOrUpgrade(t *testing.T) {
	base := ClusterHookParams{
		ClusterDetails: ClusterDetails{
			UID: "test-uid",
			URL: "https://cluster.example.com",
		},
	}

	tests := []struct {
		name     string
		details  clients.ControlPlaneDetails
		expected bool
	}{
		{
			name: "bootstrap token alone is sufficient",
			details: clients.ControlPlaneDetails{
				URL:            "https://cp.example.com",
				BootstrapToken: "my-token",
			},
			expected: true,
		},
		{
			name: "client secret alone is sufficient",
			details: clients.ControlPlaneDetails{
				URL:          "https://cp.example.com",
				ClientSecret: "my-secret",
			},
			expected: true,
		},
		{
			name: "both token and secret returns true",
			details: clients.ControlPlaneDetails{
				URL:            "https://cp.example.com",
				ClientSecret:   "my-secret",
				BootstrapToken: "my-token",
			},
			expected: true,
		},
		{
			name: "neither token nor secret returns false",
			details: clients.ControlPlaneDetails{
				URL: "https://cp.example.com",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := base
			params.ControlPlaneDetails = tt.details
			assert.Equal(t, tt.expected, params.ShouldPerformControlPlaneAssistedInstallationOrUpgrade())
		})
	}
}
