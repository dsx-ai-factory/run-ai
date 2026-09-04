/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clusterstatecalculator

type ClusterState = string

const (
	ClusterStateWaitingToConnect     ClusterState = "WaitingToConnect"
	ClusterStateConnected            ClusterState = "Connected"
	ClusterStateDisconnected         ClusterState = "Disconnected"
	ClusterStateMissingPrerequisites ClusterState = "MissingPrerequisites"
	ClusterStateServiceIssues        ClusterState = "ServiceIssues"
	ClusterStateUnknown              ClusterState = "Unknown"
)
