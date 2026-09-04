/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/ui"
	"github.com/spf13/cobra"
)

var (
	interactive      attributes.EnableFlag
	instance         = config.GetInstance()
	settingsInstance = settings.GetInstance()
	txn              *transaction.Txn
)

// Config represents the config command
var Config = &cobra.Command{
	Use:     "config",
	Short:   consts.ConfigShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		cmdHelper := helper.NewCmd()
		options := instance.GetOptions()
		if !cmdHelper.CheckInteractive(txn, interactive, options.Interactive) {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		err := ui.ConfigurationInteractive()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ConfigInteractiveExecutionError, err)
		}
		return nil
	},
}

func init() {
	Config.Flags().Var(&interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = Config.RegisterFlagCompletionFunc(consts.Interactive, attributes.EnableFlagCompletion)
}
