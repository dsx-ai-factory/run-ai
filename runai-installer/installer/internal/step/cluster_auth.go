/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import cluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

func ClusterAuthValue(installInfo *cluster.ClusterInstallationInfoResponse) string {
	return ClusterAuthValueFromCredentials(installInfo.BootstrapToken, installInfo.GetClientSecret())
}

func ClusterAuthValueFromCredentials(bootstrapToken, clientSecret string) string {
	if bootstrapToken != "" {
		return "controlPlane.bootstrapToken=" + bootstrapToken
	}
	return "controlPlane.clientSecret=" + clientSecret
}
