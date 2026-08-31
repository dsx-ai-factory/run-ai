/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async_events

type ClusterDelete struct {
	ClusterId string `json:"cluster_id"`
}

func NewClusterDelete(clusterId string) *ClusterDelete {
	return &ClusterDelete{
		ClusterId: clusterId,
	}
}
