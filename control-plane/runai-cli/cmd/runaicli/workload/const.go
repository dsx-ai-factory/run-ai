/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

// submit
const (
	submitUse      = "submit [NAME]"
	submitShortMsg = "submit a workload from a yaml file"
	submitLongMsg  = `Submit a workload using a Kubernetes manifest file.

This command submits workloads directly from YAML manifests to the control plane.
It provides flexibility in managing workloads through declarative configuration.`

	submitExample = `# Submit a workload from a manifest file (uses manifest metadata.name or defaults to "workload")
runai workload submit --file workload.yaml --project my-project

# Submit with a specific name
runai workload submit <name> --file workload.yaml --project my-project

# Submit with name as prefix (generates random suffix)
runai workload submit --name-prefix my-workload --file workload.yaml --project my-project

# Submit with priority override
runai workload submit --file workload.yaml --project my-project --priority high

# Submit with category
runai workload submit --file workload.yaml --project my-project --category production

# Submit with MNNVL mode enabled
runai workload submit --file workload.yaml --project my-project --mnnvl Required`

	// Flag names
	flagFile       = "file"
	flagPriority   = "priority"
	flagCategory   = "category"
	flagNamePrefix = "name-prefix"
	flagMnnvl      = "mnnvl"

	// Flag usage descriptions
	flagFileUsage       = "Path to the workload manifest YAML file (required)"
	flagCategoryUsage   = "Set workload category"
	flagNamePrefixUsage = "Use the provided name as a prefix (generates random suffix). Mutually exclusive with [NAME] argument"
	flagMnnvlUsage      = "Enable MNNVL (Multi-Node NVLink) mode. Valid values: Required|None"

	// Success messages
	successWorkloadSubmitted = "Creating %s %s..."
)

// delete

const (
	FailedToBulkDeletePre           = "Bulk delete failed to execute"
	FailedToBulkDeleteSomeWorkloads = "At least one workload can't be deleted. You're either not authorized to delete them or they are not managed by Run:ai, delete them the way they were created."
	WorkloadsNotFound               = "The workloads in project '%s' were not found: %s"
	FailedMissingResponse           = "Unknown response value please try again"
	BulkDeleteSuccessNoFound        = "Successfully deleted all workloads that were found"
	BulkDeleteSuccess               = "Successfully deleted all workloads"
	SingleDeleteSuccess             = "Successfully deleted workload '%s'"

	BulkConfirmationMessage = "Are you sure you want to delete %d workloads?"
	BulkConfirmationAborted = "Workload delete aborted"

	BulkSpinnerMsg   = "Deleting Workloads..."
	SingleSpinnerMsg = "Deleting Workload '%s'..."

	AttributeNameYes  = "yes"
	AttributeUsageYes = "bypass confirmation dialog by answering yes"

	BulkDeleteExamples = `# Delete multiple workloads
runai workload delete -p proj1 workload01 workload02 workload03

# Delete list of workloads with PyTorch framework filter
runai workload delete -p proj1 --framework pytorch workload01 workload02 workload03

# Delete list of workloads with training type filter
runai workload delete -p proj1 --type training workload01 workload02 workload03

# Delete multiple workloads by bypassing confirmation
runai workload delete -p proj1 -y workload01 workload02 workload03`
)

const (
	maxDeleteThreshold = 100
	maxFilterSize      = 10
)

// list

const (
	statusAttribute = "status"
)

// exec
const (
	execShortMsg   = "exec into a workload"
	execExampleMsg = `# Exec into a workload using bash
runai workload exec <workload-name> -- bash

# Exec into a workload with a specific command
runai workload exec <workload-name> -- ls /tmp

# Exec into a specific pod in a workload
runai workload exec <workload-name> --pod=<pod-name> -- bash

# Exec into a specific container in a workload
runai workload exec <workload-name> --container=<container-name> -- bash

# Exec with an interactive TTY session
runai workload exec <workload-name> --stdin --tty -- bash

# Exec and wait up to 30 seconds for the workload to be ready
runai workload exec <workload-name> --wait-timeout=30s -- bash`
)

// logs
const (
	logsShortMsg   = "view logs of a workload"
	logsExampleMsg = `# Get logs for a workload
runai workload logs <workload-name>

# Get logs for a specific pod in a workload
runai workload logs <workload-name> --pod=<pod-name>

# Get logs for a specific container in a workload
runai workload logs <workload-name> --container=<container-name>

# Get the last 100 lines of logs
runai workload logs <workload-name> --tail=100

# Get logs with timestamps
runai workload logs <workload-name> --timestamps

# Follow the logs
runai workload logs <workload-name> --follow

# Get logs for the previous instance of the workload
runai workload logs <workload-name> --previous

# Limit the logs to 1024 bytes
runai workload logs <workload-name> --limit-bytes=1024

# Get logs since the last 5 minutes
runai workload logs <workload-name> --since=300s

# Get logs since a specific timestamp
runai workload logs <workload-name> --since-time=2023-05-30T10:00:00Z

# Wait up to 30 seconds for a workload to be ready for logs
runai workload logs <workload-name> --wait-timeout=30s`
)
