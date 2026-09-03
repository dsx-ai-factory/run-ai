/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pf

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/backward"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/portforward"
	commWorkload "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cluster/portforward/options"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	workloadType string
	pfOpt        options.Options
	instance     = config.GetInstance()
	txn          *transaction.Txn
)

// clusters represents the clusters command
// Deprecated: Use `runai workload port-forward` instead
var Cmd = &cobra.Command{
	Use:        "port-forward WORKLOAD_NAME",
	Short:      consts.PortForwardShortMsgDep,
	GroupID:    consts.GroupDeprecatedID,
	Aliases:    []string{"pf"},
	Args:       cobra.ExactArgs(1),
	Deprecated: consts.DepPortForwardMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := portforward.Pre(cmd, args, &pfOpt)
		if err != nil {
			return err
		}
		clusterWorkload, err := backward.HandleUniqWorkloadBackwardCompatiblePreE(txn.GetCtx(),
			cmd, args[0], workloadType, pfOpt.Project, consts.FailedPfCommandError)
		if err != nil {
			return err
		}
		pfOpt.WorkloadType = clusterWorkload.Type
		pfOpt.WorkloadFramework = clusterWorkload.Framework
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return portforward.Run(cmd, &pfOpt)
	},
}

func init() {
	Cmd.Flags().StringVar(&workloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	portforward.FlagOutputType(Cmd, &pfOpt, api.UndefinedFramework)
	_ = Cmd.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
}
