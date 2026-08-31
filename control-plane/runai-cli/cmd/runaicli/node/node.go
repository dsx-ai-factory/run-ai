/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	commAttr "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "node",
	Short:   consts.NodeShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupResourcesID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initNodeInstance(cmd)
		if err != nil {
			// err is already formatted
			return err
		}

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}

		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func init() {

}

func initNodeInstance(cmd *cobra.Command) error {
	txn := instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)

	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	tokenStr, err := auth.GetToken(txn.GetCtx())
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
	}
	instance.SetToken(tokenStr)

	return nil
}
