/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
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

var Cmd = &cobra.Command{
	Use:     "my-credential",
	Short:   ShortMsg,
	Long:    "Manage personal credentials associated with the current user.",
	GroupID: consts.GroupUserManagementID,
	Args:    cobra.NoArgs,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initInstance(cmd)
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}

func initInstance(cmd *cobra.Command) error {
	instance = config.GetInstance()

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	tokenStr, err := token.GetToken(instance.GetTxn().GetCtx())
	if err != nil {
		return err
	}
	instance.SetToken(tokenStr)

	return nil
}
