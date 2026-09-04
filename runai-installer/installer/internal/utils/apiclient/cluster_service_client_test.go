/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClusterInstallationInfoResponseSupportsBothCredentials(t *testing.T) {
	tests := []struct {
		name           string
		response       string
		bootstrapToken string
		clientSecret   string
	}{
		{
			name:           "bootstrap token",
			response:       `{"installationStr":"","repositoryName":"runai","chartRepoURL":"","bootstrapToken":"token"}`,
			bootstrapToken: "token",
		},
		{
			name:         "client secret",
			response:     `{"installationStr":"","repositoryName":"runai","chartRepoURL":"","clientSecret":"secret"}`,
			clientSecret: "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var installInfo cluster.ClusterInstallationInfoResponse
			require.NoError(t, json.Unmarshal([]byte(tt.response), &installInfo))
			assert.Equal(t, tt.bootstrapToken, installInfo.BootstrapToken)
			assert.Equal(t, tt.clientSecret, installInfo.ClientSecret)
		})
	}
}
