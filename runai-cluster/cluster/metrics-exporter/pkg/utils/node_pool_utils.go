/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import runaiv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

func GetNodePoolNameFromLabels(labels map[string]string) string {
	nodePoolName, found := labels[runaiv1alpha1.NodePoolNameLabel]
	if !found || nodePoolName == "" {
		nodePoolName = runaiv1alpha1.DefaultNodePoolName
	}

	return nodePoolName
}
