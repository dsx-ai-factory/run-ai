/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/report/metrics"
	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "report",
	Short:   consts.ReportShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		txn := instance.GetTxn()
		cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		tokenStr, err := token.GetToken(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
		}
		instance.SetToken(tokenStr)
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func init() {
	Cmd.AddCommand(metrics.Cmd)
}
