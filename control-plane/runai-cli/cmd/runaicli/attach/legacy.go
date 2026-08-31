/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attach"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/backward"
	commWorkload "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	workloadType string
	attachOpt    attach.Options
	instance     = config.GetInstance()
	txn          *transaction.Txn
)

// Deprecated: Use `runai workload attach` instead
var Cmd = &cobra.Command{
	Use:        "attach WORKLOAD_NAME",
	Short:      consts.AttachShortMsgDep,
	GroupID:    consts.GroupDeprecatedID,
	Args:       cobra.ExactArgs(1),
	Deprecated: consts.DepAttachMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := attach.Pre(cmd, args, &attachOpt)
		if err != nil {
			return err
		}
		clusterWorkload, err := backward.HandleUniqWorkloadBackwardCompatiblePreE(txn.GetCtx(),
			cmd, args[0], workloadType, attachOpt.Project, consts.AttachCommandError)
		if err != nil {
			return err
		}
		attachOpt.WorkloadType = clusterWorkload.Type
		attachOpt.WorkloadFramework = clusterWorkload.Framework
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return attach.Run(cmd, &attachOpt, txn)
	},
}

func init() {
	Cmd.Flags().StringVar(&workloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	attach.FlagOutputType(Cmd, &attachOpt)
	_ = Cmd.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
}
