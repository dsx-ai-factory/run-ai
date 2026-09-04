/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/backward"
	commWorkload "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	workloadType string
	txn          *transaction.Txn
	instance     = config.GetInstance()
)

// Deprecated: Use `runai workload describe` instead
var job = &cobra.Command{
	Use:        "job JOB_NAME",
	Short:      consts.DescribeJobShortMsgDep,
	Aliases:    []string{"jobs"},
	Args:       cobra.ExactArgs(1),
	Deprecated: consts.DepDescribeJobMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},

	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		err = workload.DescribeProjectValidatorPre(cmd, &workload.DescribeFlagsOptions)
		if err != nil {
			return err
		}

		clusterWorkload, err := backward.HandleUniqWorkloadBackwardCompatiblePreE(txn.GetCtx(),
			cmd, args[0], workloadType, workload.DescribeFlagsOptions.Project, consts.FailedToDescribeWorkload)
		if err != nil {
			return err
		}

		workload.WorkloadType = clusterWorkload.Type.QueryParam()
		return workload.Describe.PreRunE(cmd, args)
	},
	RunE: workload.Describe.RunE,
}

func init() {
	Cmd.AddCommand(job)
	job.Flags().StringVarP(&workload.DescribeFlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	job.Flags().StringVar(&workloadType, consts.WorkloadType, "", consts.WorkloadTypeUsage)
	_ = job.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
}
