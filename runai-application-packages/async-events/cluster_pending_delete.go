/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async_events

type ClusterPendingDelete struct {
	ClusterId string `json:"cluster_id"`
}

func NewClusterPendingDelete(clusterId string) *ClusterPendingDelete {
	return &ClusterPendingDelete{
		ClusterId: clusterId,
	}
}
