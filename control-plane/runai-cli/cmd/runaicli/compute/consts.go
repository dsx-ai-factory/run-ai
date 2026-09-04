/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compute

const (
	ProjectFlagUsage = "Specify the project. Higher-scope compute-resources in the same tenant are included. Use runai project set <project> to set the default"

	// Describe command
	DescribeShortMsg        = "describe a compute resource asset"
	FailedToDescribeCompute = "Failed to describe compute"
	describeExample         = `# Describe a compute resource asset with a default project
runai compute describe compute-large

# Describe a compute resource asset in a specific project
runai compute describe compute-large -p <project_name>

# Describe a compute resource asset with specific output format
runai compute describe compute-large -o json`

	// List command
	ListShortMsg             = "list compute resource assets"
	FailedToListCompute      = "Failed to list compute assets"
	ListAllFlagUsage         = "Show resources from all projects"
	DefaultComputeColumnsSet = "Default compute columns set to: %s\n"
	DescribeExample          = `# List all compute resource assets
runai compute list -A

# List compute resource assets in a specific project
runai compute list -p <project_name>

# List compute resource assets with a specific output format
runai compute list --yaml`
)
