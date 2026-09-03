/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Deprecated: Use `runai node list ` instead
var nodeCmd = &cobra.Command{
	Use:        "node",
	Short:      consts.DescribeNodeShortMsgDep,
	Aliases:    []string{"nodes"},
	Deprecated: consts.DepDescribeNodeMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_ = node.Output.SetDefault("yaml")
		return node.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: node.List.PreRunE,
	RunE:    node.List.RunE,
}

func init() {
	Cmd.AddCommand(nodeCmd)
}
