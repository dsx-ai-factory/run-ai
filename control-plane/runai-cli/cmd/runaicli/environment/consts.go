/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment

const (
	ProjectFlagUsage = "Specify the project. Higher-scope environments in the same tenant are included. Use 'runai project set <project>' to set the default."

	// Describe command
	DescribeShortMsg            = "describe an environment asset"
	FailedToDescribeEnvironment = "Failed to describe environment"
	DescribeExample             = `# Describe an environment asset with a default project
runai environment describe pytorch-env

# Describe an environment asset in a specific project
runai environment describe pytorch-env -p <project_name>

# Describe an environment asset with specific output format
runai environment describe pytorch-env -o json`

	// List command
	ListShortMsg                 = "list environment assets"
	FailedToListEnvironment      = "Failed to list environment assets"
	ListAllFlagUsage             = "Show resources from all projects"
	DefaultEnvironmentColumnsSet = "Default environment columns set to: %s\n"
	ListExample                  = `# List all environment assets
runai environment list -A

# List environment assets in a specific project
runai environment list -p <project_name>

# List environment assets with a specific output format
runai environment list --yaml`
)
