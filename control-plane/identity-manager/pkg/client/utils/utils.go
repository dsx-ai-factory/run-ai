/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

const (
	ClientTypeAttr = "runai.client_type"
	ClusterIdAttr  = "runai.cluster_id"
)

func GetClusterIdAttr() string {
	return ClusterIdAttr
}

func GetClientTypeAttr() string {
	return ClientTypeAttr
}
