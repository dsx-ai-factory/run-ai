/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool

import (
	commAttr "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"
	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var (
	instance  = config.GetInstance()
	cmdHelper helper.Cmd

	txn *transaction.Txn
)

var Cmd = &cobra.Command{
	Use:     "nodepool",
	Short:   consts.NodePoolShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupResourcesID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		tokenStr, err := token.GetToken(txn.GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
		}
		instance.SetToken(tokenStr)

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}

		return nil

	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func init() {

}
