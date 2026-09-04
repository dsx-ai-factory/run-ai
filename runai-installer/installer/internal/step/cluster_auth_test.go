/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/stretchr/testify/assert"
)

func TestClusterAuthValue(t *testing.T) {
	assert.Equal(t,
		"controlPlane.bootstrapToken=token",
		ClusterAuthValue(&cluster.ClusterInstallationInfoResponse{
			BootstrapToken: "token",
			ClientSecret:   "secret",
		}),
	)
	assert.Equal(t,
		"controlPlane.clientSecret=secret",
		ClusterAuthValue(&cluster.ClusterInstallationInfoResponse{
			ClientSecret: "secret",
		}),
	)
}
