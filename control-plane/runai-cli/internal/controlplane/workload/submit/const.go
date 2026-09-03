/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

const (
	// Error messages
	errorFailedToReadManifest   = "failed to read manifest file: %w"
	errorFailedToParseYAML      = "failed to parse YAML manifest: %w"
	errorFailedToGetProjectID   = "failed to get project ID: %w"
	errorFailedToCreateClient   = "failed to create workload client: %w"
	errorFailedToSubmitWorkload = "failed to submit workload: %w"
	errorProjectRequired        = "Error: project is missing. value is empty.\n\nYou can set it in one of the following ways:\n " +
		" • Run:  runai project set <PROJECT_NAME>\n  • Use the CLI flag:  -p <PROJECT_NAME>\n" +
		"  • Or, specify the 'namespace' field in your YAML manifest"
)
