/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

const (
	// Field names
	FieldNode          = "Node"
	FieldCompletedAt   = "Completed At"
	FieldNodePool      = "Node Pool"
	FieldContainers    = "Containers"
	FieldRole          = "Role"
	FieldRunningPods   = "Running Pods"
	FieldRequestedPods = "Requested Pods"

	// Column widths
	PodColumnWidth = 15

	// Section title
	TitlePods = "Pods"

	// Role constants
	RoleLeader = "Leader"
	RoleWorker = "Worker"
)
