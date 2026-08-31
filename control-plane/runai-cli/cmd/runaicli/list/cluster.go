/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/cluster"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// clusters represents the clusters command
// Deprecated: Use `runai cluster list` instead
var clusters = &cobra.Command{
	Use:        "clusters",
	Short:      consts.ListClusterShortMsg,
	Aliases:    []string{"cluster"},
	Args:       cobra.NoArgs,
	Deprecated: consts.ListClusterDepMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cluster.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: cluster.List.PreRunE,
	RunE:    cluster.List.RunE,
}

func init() {
	Cmd.AddCommand(clusters)

	listCommon.FlagOutputType(clusters, &cluster.Output)
}
