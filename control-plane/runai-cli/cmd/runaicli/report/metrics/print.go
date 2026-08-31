/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"path/filepath"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/logreader"
	"github.com/spf13/cobra"
)

const (
	DefaultTailValue = 100
)

var (
	tail int
)

var output = &cobra.Command{
	Use:   "output",
	Short: consts.MetricsOutputShortMsg,
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		logReader, err := logreader.NewLogReader(filepath.Dir(getMetricsLogFilePath()), tail)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintReportMetrics, err)
		}
		lines, err := logReader.GetLogList()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToPrintReportMetrics, err)
		}

		for _, line := range lines {
			cmd.Println(line)
		}
		txn.BypassMetricReport()
		return nil
	},
}

func init() {
	Cmd.AddCommand(output)

	output.Flags().IntVar(&tail, "tail", DefaultTailValue, "number of tail metrics")
}
