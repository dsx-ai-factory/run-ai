/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

type ClusterInstallationData struct {
	StackName          string
	EksClusterName     string
	TenantName         string
	ControlPlaneUrl    string
	InstallerAppId     string
	InstallerAppSecret string
	ClusterId          string
	ClusterUrl         string
}
