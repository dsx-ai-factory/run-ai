/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async_events

type ProjectDeleted struct {
	ClusterId   string `json:"cluster_id"`
	ProjectId   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	// A flag that indicates if the receiver of the event should delete cluster sources objects, or leave them orphans behind
	LeaveOrphans bool `json:"leave_orphans"`
}

func NewProjectDeleted(
	clusterId string,
	projectId string,
	projectName string,
	leaveOrphans bool,
) *ProjectDeleted {
	return &ProjectDeleted{
		ClusterId:    clusterId,
		ProjectId:    projectId,
		ProjectName:  projectName,
		LeaveOrphans: leaveOrphans,
	}
}
