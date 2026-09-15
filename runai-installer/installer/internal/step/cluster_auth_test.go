/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import (
	"testing"

	cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/stretchr/testify/assert"
)

func TestClusterAuthValue(t *testing.T) {
	bootstrapTokenInfo := cluster.NewClusterInstallationInfoResponse("", "", "", "token")
	bootstrapTokenInfo.SetClientSecret("secret")
	assert.Equal(t,
		"controlPlane.bootstrapToken=token",
		ClusterAuthValue(bootstrapTokenInfo),
	)
	clientSecretInfo := cluster.NewClusterInstallationInfoResponse("", "", "", "")
	clientSecretInfo.SetClientSecret("secret")
	assert.Equal(t,
		"controlPlane.clientSecret=secret",
		ClusterAuthValue(clientSecretInfo),
	)
}
