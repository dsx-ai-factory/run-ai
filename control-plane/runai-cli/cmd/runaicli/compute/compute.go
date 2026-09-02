/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compute

import (
	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
	txn       *transaction.Txn
)

var Cmd = &cobra.Command{
	Use:     "compute",
	Short:   "Compute resource asset management",
	GroupID: consts.GroupAssetsID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initInstance(cmd)
	},
}

func initInstance(cmd *cobra.Command) error {
	instance = config.GetInstance()

	txn = instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterUrl := instance.GetConfigValue(config.ClusterDomain)

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
