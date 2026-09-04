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
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/spf13/cobra"
)

const (
	DefaultFileSize   = 10
	DefaultFileNumber = 30
	DefaultFileAge    = 14
)

var (
	enable     attributes.EnableFlag
	reportType attributes.ReportType
	size       int
	files      int
	age        int
)

// metrics represents the report metrics config command
var metrics = &cobra.Command{
	Use:   "config",
	Short: consts.MetricsConfigShortMsg,
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if cap(args) == 0 {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		if enable.String() != "" {
			err := SetMetricsEnable(enable.IsEnabled())
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetReportMetrics, err)
			}
		}

		if reportType != "" {
			err := SetReportType(reportType, size, files, age)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetReportMetrics, err)
			}
		}
		return nil
	},
}

func getTxnConfig() *config.Txn {
	txnConfig, _ := instance.GetTxnConfig()
	if txnConfig == nil {
		txnConfig = &config.Txn{}
	}
	return txnConfig
}

func SetMetricsEnable(enable bool) error {
	txnConfig := getTxnConfig()
	txnConfig.Enabled = &enable
	return instance.SetTxnConfig(txnConfig)
}

func SetReportType(reportType attributes.ReportType, size int, files int, age int) error {
	logsPath := getMetricsLogFilePath()
	err := os.MkdirAll(filepath.Dir(logsPath), os.ModePerm)
	if err != nil {
		return err
	}
	txnConfig := getTxnConfig()
	txnConfig.Report, err = config.NewReport(reportType, logsPath, files, size, age)
	if err != nil {
		return err
	}
	return instance.SetTxnConfig(txnConfig)
}

func init() {
	Cmd.AddCommand(metrics)

	metrics.Flags().Var(&enable, "metrics", "metrics enable flag (enabled|disabled)")
	metrics.Flags().Var(&reportType, "type", "report generated type (none|logger|local)")
	metrics.Flags().IntVar(&size, "size", DefaultFileSize, "metrics max file size")
	metrics.Flags().IntVar(&files, "files", DefaultFileNumber, "metrics max file number")
	metrics.Flags().IntVar(&age, "age", DefaultFileAge, "metrics max file age")
}
