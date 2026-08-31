/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

// Logs strings
const (
	JobName                 = "name"
	Pod                     = "pod"
	Tail                    = "tail"
	Container               = "container"
	ContainerShort          = "c"
	Timestamps              = "timestamps"
	Follow                  = "follow"
	FollowShort             = "f"
	FileShort               = "f"
	Previous                = "previous"
	LimitBytes              = "limit-bytes"
	Since                   = "since"
	SinceTime               = "since-time"
	WaitTimeout             = "wait-timeout"
	WorkloadType            = "type"
	WorkloadFramework       = "framework"
	DeletedWorkload         = "deleted"
	WorkloadTypeUsage       = "The type of workload. Valid values: training, workspace, inference."
	FailedToGetWorkloadType = "workload type is invalid must be one of: training, workspace"
	InvalidFrameworkType    = "workload framework is invalid must be one of: %s"

	JobNameAttributeExplanation       = "The name of the workload. Must match the name used at submission."
	PodIdAttributeExplanation         = "The pod ID. If not specified, the first pod will be used."
	TailAttributeExplanation          = "The number of log lines to display from the end of the logs. Use -1 to show all lines."
	ContainerAttributeExplanation     = "The name of the container within the pod."
	TimestampsAttributeExplanation    = "Include timestamps in log output."
	FollowAttributeExplanation        = "Stream logs in real time."
	PreviousAttributeExplanation      = "Show logs from the previous instance of the pod, if it was restarted."
	LimitBytesAttributeExplanation    = "Maximum bytes of logs to return. Defaults to no limit (e.g, 20, 2000)."
	SinceSecondsAttributeExplanation  = "Show logs newer than a relative duration (e.g., 5s, 2m, 3h)."
	SinceTimeAttributeExplanation     = "Show logs generated after a specific RFC3339 timestamp (e.g. 2025-05-25T14:30:00Z)."
	WaitTimeoutAttributeExplanation   = "Timeout while waiting for the workload to become ready for log streaming (e.g., 5s, 2m, 3h)."
	LogsCommandError                  = "logs command failed"
	LogsInvalidArgument               = "invalid argument"
	ErrorWorkloadNameIsRequired       = "workload name is missing, please use 'runai workload <command> <workload-name> [FLAGS]'"
	SinceTimeAndSecondsError          = "at most one of `sinceTime` or `since` may be specified"
	SinceSecondsError                 = "--since must be greater than 0"
	TailLinesError                    = "--tail must be greater than or equal to -1"
	SinceTimeError                    = "--sinceTime must be in RFC3339 format (e.g. 2025-05-25T14:30:00Z)."
	LimitBytesError                   = "--limit-bytes must be greater than 0"
	FailedToPrintLogs                 = "failed to print logs"
	GenericStreamLogsError            = "got error while trying to stream workload: %w"
	WorkloadLogsStreamerNotReadyError = "workload is not ready to stream: %s"
	WorkloadLogsStreamerNotFoundError = "one of requested resource was not found: %v"
	FailedLogsCommand                 = "logs command failed"
	ContentLengthInvalid              = "invalid response received: Content length must be greater than 0, but got %d"
	InvalidNegativeValue              = "value cannot be negative"
)

// Exec strings
const (
	ExecCommandError    = "exec command failed"
	ExecInvalidArgument = "invalid argument"

	FailedToPrintExecOutput = "failed to exec output"

	AttributeNameStdin  = "stdin"
	AttributeUsageStdin = "Pass stdin to the container"
	AttributeNameTTY    = "tty"
	AttributeUsageTTY   = "Stdin is a TTY"
)

// Attach strings
const (
	AttachCommandError    = "attach command failed"
	AttachInvalidArgument = "invalid argument"

	FailedToPrintAttachOutput = "failed to exec output"
)

// describe consts
const (
	DescribeOutputUsage          = "Output format: table, json, or yaml. Default is table."
	DescribeGeneralUsage         = "Show general resource information."
	DescribeComputeUsage         = "Show compute resources information (CPU, GPU, memory, etc.)."
	DescribeStorageUsage         = "Show storage information (PVCs, ConfigMaps, Secrets, etc.)."
	DescribePodsUsage            = "Show information about pods associated with the resource."
	DescribeEventsUsage          = "Show recent events for the resource."
	DescribeNetworksUsage        = "Show network-related information."
	DescribeContainersUsage      = "Include container-level details in pod information."
	DescribePodLimitUsage        = "Limit the number of pods displayed. Use -1 for no limit."
	DescribeEventLimitUsage      = "Limit the number of events displayed. Use -1 for no limit."
	DescribeEventsSortOrderFlag  = "sortEvents"
	DescribeEventsSortOrderUsage = "Sort events in ascending or descending order. Valid values: asc, desc."
)

// port-forward consts
const (
	PortForwardShortMsg             = "port forward management"
	FailedPfCommandError            = "port-forward command failed"
	StartingPortForward             = "port-forward started, opening ports %s\n"
	ClosingPortForward              = "port-forward terminating, closing port %s\n"
	PodIdAttributePortForward       = "Workload pod ID for port-forward, default: distributed(master) otherwise(random)"
	PodIdAttributePortForwardMaster = "Workload pod ID for port-forward, default: master"
	PodIdAttributePortForwardRandom = "Workload pod ID for port-forward, default: random"
)
