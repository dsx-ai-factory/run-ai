/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckClusterReachableKubernetesVersion(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expected    orchestrator2.Status
		reason      string
		remediation string
	}{
		{name: "minimum", version: "v1.30.0", expected: orchestrator2.StatusOK},
		{name: "newer", version: "v1.31.2", expected: orchestrator2.StatusOK},
		{name: "vendor suffix", version: "v1.30.0-eks-123456", expected: orchestrator2.StatusOK},
		{
			name:        "too old",
			version:     "v1.29.9",
			expected:    orchestrator2.StatusFail,
			reason:      "Kubernetes v1.29.9 is unsupported; minimum required version is 1.30.0",
			remediation: "upgrade the cluster to Kubernetes 1.30.0 or newer",
		},
		{
			name:        "invalid",
			version:     "not-a-version",
			expected:    orchestrator2.StatusFail,
			reason:      "could not parse Kubernetes server version",
			remediation: "verify the cluster reports a valid semantic Kubernetes version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := CheckClusterReachable{}
			result, err := action.Execute(t.Context(), orchestrator2.Deps{
				ClusterProbe: &harness.FakeClusterProbe{Version: tt.version},
			})

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result.Status)
			assert.Contains(t, result.Reason, tt.reason)
			assert.Equal(t, tt.remediation, result.Remediation)
			assert.Equal(t, tt.version, result.Data["kubernetesVersion"])
		})
	}
}
