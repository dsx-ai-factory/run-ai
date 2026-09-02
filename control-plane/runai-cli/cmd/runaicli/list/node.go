/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// clusters represents the clusters command
// Deprecated: Use `runai node list` instead
var nodes = &cobra.Command{
	Use:        "nodes",
	Short:      consts.ListNodeShortMsg,
	Args:       cobra.NoArgs,
	Aliases:    []string{"node"},
	Deprecated: consts.ListNodeDepMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return node.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: node.List.PreRunE,
	RunE:    node.List.RunE,
}

func init() {
	Cmd.AddCommand(nodes)

	listCommon.FlagOutputType(nodes, &node.Output)
}
