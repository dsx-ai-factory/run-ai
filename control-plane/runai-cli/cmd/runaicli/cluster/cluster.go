/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/table"

	commAttr "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/spf13/cobra"
)

var (
	interactive attributes.EnableFlag
	cmdHelper   helper.Cmd
	instance    = config.GetInstance()
	txn         *transaction.Txn
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "cluster",
	Short:   consts.ClusterShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupResourcesID,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
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

		if instance.GetOptions().Quiet {
			if err := commAttr.EnableQuietMode(cmd); err != nil {
				return err
			}
		}
		return nil
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		options := instance.GetOptions()
		if !cmdHelper.CheckInteractive(txn, interactive, options.Interactive) {
			cmd.HelpFunc()(cmd, args)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		token := instance.GetToken()
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)

		handler, err := cluster.GetClusterList(txn, cpUrl, token)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, err)
		}

		selectedCluster, err := table.Selection(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, err)
		}

		if selectedCluster == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, fmt.Errorf("selection returned empty"))
		}

		selectedClusterName, exist := (*selectedCluster)["Name"]
		if !exist {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, fmt.Errorf("cluster Name not found"))
		}
		selectedClusterDomain, exist := (*selectedCluster)["URL"]
		if !exist {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, fmt.Errorf("cluster URL not found"))
		}

		err = fmt.Errorf("cluster Name %s, not found in list", selectedClusterName)
		clusterList := handler.GetList()
		for i, cluster := range clusterList {
			if cluster.Domain == selectedClusterDomain && cluster.Name == selectedClusterName {
				selected := clusterList[i]
				err = instance.SetCluster(selected.Uuid, selected.Name, selected.Domain)
				break
			}
		}

		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListClusters, err)
		}
		cmd.Println(fmt.Sprintf(auth.SetClusterProperties, instance.GetConfigValue(config.ClusterName)))
		return nil
	},
}

func init() {
	Cmd.Flags().Var(&interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = Cmd.RegisterFlagCompletionFunc(consts.Interactive, attributes.EnableFlagCompletion)
}
