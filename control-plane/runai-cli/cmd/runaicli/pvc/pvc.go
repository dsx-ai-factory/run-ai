/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	token "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	ShortMsg = "PVC management"
)

var (
	cmdHelper helper.Cmd
	instance  = config.GetInstance()
	txn       *transaction.Txn
)

// Cmd represents the pvc command
var Cmd = &cobra.Command{
	Use:     "pvc",
	Short:   ShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupResourcesID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := initPVCInstance(cmd)
		if err != nil {
			return err
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

// todo: combine with other init functions
func initPVCInstance(cmd *cobra.Command) error {
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
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ReLoginMsg, err)
	}
	instance.SetToken(tokenStr)

	return nil
}
