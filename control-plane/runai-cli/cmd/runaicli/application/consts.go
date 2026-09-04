/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

const (
	// Application command messages
	AIApplicationShortMsg          = "ai application management"
	FailedToListAIApplication      = "applications list failed"
	DefaultAIApplicationColumnsSet = "default application columns set to: %s\n"

	// List command messages
	ListShortMsg     = "List AI applications"
	ListAllFlagUsage = "Show resources from all projects"
	ProjectFlagUsage = "Specify the project for the command to use. Defaults to the project set in the context, if any. Use 'runai project set <project>' to set the default."
	ListExample      = `# List all AI applications from all projects
runai application list -A

# List AI applications in a specific project
runai application list -p <project_name>

# List AI applications with a specific output format
runai application list --yaml

# List AI applications with custom columns
runai application list --columns name,status,gpucomputereq`

	// Delete command messages
	DeleteShortMsg              = "Delete an AI application"
	FailedToDeleteAIApplication = "application delete failed"
	ForceFlagUsage              = "Skip confirmation prompt"
	YesFlagUsage                = "Alias for --force"
	DeleteConfirmMsg            = "Are you sure you want to delete application %q in project %q?\nThis will remove the application and all its associated resources."
	DeleteAbortedMsg            = "Application delete aborted"
	DeleteSuccessMsg            = "Application %q deleted successfully."
	ErrAppNotFound              = "application %q not found in project %q"
	ErrAppBeingDeleted          = "application %q is already being deleted"
	DeleteExample               = `# Delete an application (with confirmation prompt)
runai application delete my-grafana -p team-alpha

# Delete without confirmation
runai application delete my-grafana -p team-alpha --force`

	// Submit command messages
	SubmitShortMsg              = "Submit an AI application"
	FailedToSubmitAIApplication = "failed to create application"
	ChartURLFlagUsage           = "URL to a packaged Helm chart (.tgz)"
	ValuesFlagUsage             = "Path to a YAML values file for Helm chart overrides"
	SetFlagUsage                = "Inline key=value overrides (nested keys supported), merged on top of --values"
	CredentialFlagUsage         = "NGC API key credential asset name for private chart URLs"
	SubmitSuccessMsg            = "Creating application %[1]s...\nTo track the application's status, run 'runai application describe %[1]s -p %s'"
	SubmitExample               = `# Basic: submit from a chart URL
runai application submit my-grafana -p team-alpha \
  --chart-url https://charts.bitnami.com/bitnami/grafana-11.0.0.tgz

# With value overrides
runai application submit my-grafana -p team-alpha \
  --chart-url https://charts.bitnami.com/bitnami/grafana-11.0.0.tgz \
  --values ./my-values.yaml \
  --set replicaCount=2 \
  --set persistence.enabled=false`

	// application phase values as reported by the API
	desiredPhaseDeleted  = "deleted"
	actionStatusDeleting = "deleting"

	attributeNameYes        = "yes"
	attributeNameChartURL   = "chart-url"
	attributeNameValues     = "values"
	attributeNameSet        = "set"
	attributeNameCredential = "credential"
)
