/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package step

import "github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"

func ClusterAuthValue(installInfo *cluster.ClusterInstallationInfoResponse) string {
	if installInfo.BootstrapToken != "" {
		return "controlPlane.bootstrapToken=" + installInfo.BootstrapToken
	}
	return "controlPlane.clientSecret=" + installInfo.ClientSecret
}
