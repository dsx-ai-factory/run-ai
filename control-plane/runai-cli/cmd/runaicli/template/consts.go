/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

const (
	// Template command messages
	ShortMsg         = "workload template management"
	ProjectFlagUsage = "Specify the project. Higher-scope templates in the same tenant are included. Use 'runai project set <project>' to set the default."

	// Describe command messages
	DescribeShortMsg         = "describe a workload template"
	FailedToDescribeTemplate = "Failed to describe template"
	DescribeExample          = `# Describe a template with a default project
runai template describe my-template

# Describe a template in a specific project
runai template describe my-template -p my-project

# Describe a template with specific output format
runai template describe my-template -o json`

	// List command messages
	FailedToListTemplates = "Template list failed"
	ListShortMsg          = "list all workload templates"
	AllFlagUsage          = "Show templates from all scopes"
	ListExample           = `# List all templates from all scopes
runai template list -A

# List templates in a specific project (includes higher-scope templates)
runai template list -p <project-name>

# List templates with pagination
runai template list --page-size 50 --max-items 100

# List templates with a specific output format
runai template list --yaml

# List templates with custom columns
runai template list --columns name,scopeType,workloadType,createdAt`
)
