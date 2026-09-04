/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package diagnostics

// diagnostics command
const (
	// cobra
	DiagnosticsUse      = "diagnostics"
	DiagnosticsAlias    = "diag"
	DiagnosticsShortMsg = "system diagnostics and troubleshooting tools"
)

// collect-logs command
const (
	// cobra
	CollectLogsUse      = "collect-logs"
	CollectLogsAlias    = "logs"
	CollectLogsShortMsg = "collect diagnostic logs from the Kubernetes cluster for troubleshooting or sharing with NVIDIA Run:ai Support"
	CollectLogsExample  = `# Collect logs from all default namespaces
runai diagnostics collect-logs

# Collect logs to a specific output directory
runai diagnostics collect-logs --output-dir /tmp/logs

# Collect logs without previous logs from pods
runai diagnostics collect-logs --no-previous

# Collect logs from specific namespaces
runai diagnostics collect-logs --namespaces runai,runai-backend
`

	// flags
	FlagNameOutputDir    = "output-dir"
	FlagDefaultOutputDir = "./"
	FlagHelpOutputDir    = "Directory to output the archive file"

	FlagNameNoPrevious    = "no-previous"
	FlagDefaultNoPrevious = false
	FlagHelpNoPrevious    = "Do not collect previous logs from pods"

	FlagNameNamespaces = "namespaces"
	FlagTypeNamespaces = "strings"
	FlagHelpNamespaces = "Comma-separated list of namespaces to collect logs from"

	// errors and warnings
	FailedToValidateOutputDir      = "failed to validate output directory"
	FailedToConfigureFileLogging   = "failed to configure file logging"
	FailedToInitializeArchive      = "failed to initialize archive"
	FailedToCollectLogs            = "failed to collect diagnostic logs"
	K8SCommunicationFailed         = "failed to communicate with the Kubernetes cluster. make sure the cluster is running, the network is connected, and credentials are valid"
	ClusterValidationFailed        = "failed to validate cluster identity in kubernetes cluster"
	ClusterNotValid                = "cluster in configuration differs from the Kubernetes current-context cluster"
	NamespacesNotExist             = "namespaces do not exist in cluster"
	NamespacesNotPermitted         = "namespaces are not permitted in cluster"
	NoValidNamespaces              = "no valid namespaces found to collect logs from. Verify your permissions and try again"
	FailedToBuildNamespacesSummary = "failed to build namespaces summary"

	// file name formats
	TarOutputFileNameFormat    = "runai-logs-%s.tar"
	LoggerOutputFileNameFormat = "cli-output-%s.log"

	// timestamp format
	TimestampFormat = "2006_01_02_15_04_05" // YYYY_MM_DD_HH_MM_SS

	// static file names
	ArchiveLoggerOutputFileName = "cli-output.log"

	// spinner messages
	CollectLogsSpinnerMsg = "Collecting diagnostic logs..."
)
