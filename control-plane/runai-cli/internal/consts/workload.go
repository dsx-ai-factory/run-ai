/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

// Workspace consts
const (
	WorkspaceShortMsg = "workspace management"

	FailedToSubmit = "failed to submit"
	FailedToUpdate = "failed to update"
	InvalidFlags   = "Invalid flags"

	SuccessSubmit  = "Creating %s %s...\n%s"
	SuccessUpdate  = "Updating %s %s...\n%s"
	SuccessDelete  = "Deleting %s %s...\n%s"
	SuccessResume  = "Resuming %s %s...\n%s"
	SuccessSuspend = "Stopping %s %s...\n%s"

	FollowingWorkspace = "To track the workload's status, run 'runai workspace describe %s'"
	FollowingTraining  = "To track the workload's status, run 'runai training standard describe %s'"
)

// Workload consts
const (
	WorkloadShortMsg                = "workload management"
	FailedToListWorkload            = "Workloads list failed"
	FailedToDescribeWorkload        = "Describe workload failed"
	TimerDescribeWorkload           = "describe workload"
	NoWorkloadWithNameFound         = "workload not found"
	ErrorMoreThanOneWithNameFound   = "multiple workloads found with workload name"
	PreFoundMultipleWorkloadMessage = "found multiple workloads with same name:"
	ErrorWorkloadPodsListAPI        = "workload pods list API call failed, status code error: %s"
	ErrorWorkloadEventsListAPI      = "workload events list API call failed, status code error: %s"
	DefaultWorkloadColumnsSet       = "Default workload columns set to: %s\n"
	DefaultClusterColumnsSet        = "Default cluster columns set to: %s\n"
	DefaultWProjectColumnsSet       = "Default project columns set to: %s\n"
	DefaultNodeColumnsSet           = "Default node columns set to: %s\n"

	AttributeAllUsage         = "Show resources from all projects"
	AttributeDeletedWorkloads = "Return only resources that have been deleted."
	AttributeTypeUsage        = "Filter by workload type. Valid values: training, workspace, inference."
	AttributeFrameworkUsage   = "Filter by workload framework (TensorFlow, PyTorch, MPI, XGBoost, JAX)."
	AttributeStateUsage       = "Filter by workload state (e.g., Pending, Running, Completed, Failed, Stopped, Initializing)."

	AttributeUuidUsage  = "The unique identifier (UUID) of the resource, as returned by the API."
	AttributeImageUsage = "The container image to use for the workload."
)

// Workload type strings
const (
	WorkloadTypeShortMsg          = "workload type management"
	FailedToListWorkloadTypes     = "failed to list workload types"
	FailedToDescribeWorkloadType  = "failed to describe workload type"
	DefaultWorkloadTypeColumnsSet = "default workload type columns set to: %s\n"
	ExternalOnlyUsage             = "Excludes native workload types and returns only NVIDIA Run:ai supported workload types and user-defined workload types."
	ExternalOnly                  = "external-only"
)
