/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"os"
	"path/filepath"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// cmd represents the report metrics clear command
var cmd = &cobra.Command{
	Use:   "clear",
	Short: consts.MetricsClearShortMsg,
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		logFilePath := getMetricsLogFilePath()
		err := clearLogFolder(filepath.Dir(logFilePath))
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToClearReportMetrics, err)
		}
		txn.BypassMetricReport()
		return nil
	},
}

func clearLogFolder(logsPath string) error {
	return os.RemoveAll(logsPath)
}

func init() {
	Cmd.AddCommand(cmd)
}
