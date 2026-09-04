/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package factory

import (
	"errors"

	commAttr "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func WorkloadsPersistentPreRunE(cmd *cobra.Command, _ []string) error {
	err := InstanceInit(cmd)
	if err != nil {
		return err
	}

	instance := config.GetInstance()

	if instance.GetOptions().Quiet {
		if err := commAttr.EnableQuietMode(cmd); err != nil {
			return err
		}
	}

	return nil
}

func WorkloadsPreRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	}
	return nil
}

func WorkloadsCmdInit(cmd *cobra.Command, framework api.Framework, workloadType api.WorkloadType) {
	cmdHelper := helper.Cmd{}
	commandFactory := NewCommands(config.GetInstance(), cmdHelper)
	commands, err := commandFactory.GetAllSupportedCommands(framework, workloadType)
	if err != nil {
		return
	}
	for _, command := range commands {
		_ = command.RegisterFlags()
		cmd.AddCommand(command.GetCommand())
	}
}

func InstanceInit(cmd *cobra.Command) error {
	instance := config.GetInstance()
	txn := instance.GetTxn()
	if txn == nil {
		return errors.New("transaction is nil")
	}
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterUrl := instance.GetConfigValue(config.ClusterDomain)

	cmdHelper := helper.Cmd{}
	cmdHelper.AddValidation(validators.NewCpValidator(cpUrl))
	cmdHelper.AddValidation(validators.NewClusterValidator(clusterUrl))

	if msg, err := cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}

	tokenStr, err := token.GetToken(txn.GetCtx())
	if err != nil {
		return err
	}
	instance.SetToken(tokenStr)

	return nil
}
