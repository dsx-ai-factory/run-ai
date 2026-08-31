/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeOptions struct {
	Output string
}

var describeOptions DescribeOptions

var Describe = &cobra.Command{
	Use:   "describe WORKLOAD_TYPE_NAME",
	Short: "Describe a workload type",
	Args:  cobra.ExactArgs(1),
	Example: `  # Describe a workload type by name
  runai workload-type describe MyWorkloadType

  # Output as JSON
  runai workload-type describe MyWorkloadType --output json

  # Output as YAML
  runai workload-type describe MyWorkloadType -o yaml`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()

		wtClient, err := client.NewClient(cpUrl, token)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkloadType, err)
		}

		service := describe.NewService(wtClient, wtClient)

		wt, err := service.Describe(txn.GetCtx(), args[0])
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkloadType, err)
		}

		if err := describe.RenderOutput(wt, describeOptions.Output); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToDescribeWorkloadType, err)
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(Describe)

	Describe.Flags().StringVarP(&describeOptions.Output, "output", "o", "table",
		consts.DescribeOutputUsage)

	_ = Describe.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
}
