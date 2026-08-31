/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/spf13/cobra"
)

// clusters represents the clusters command
// Deprecated: Use `runai project list` instead
var projects = &cobra.Command{
	Use:        "projects",
	Short:      consts.ListProjectShortMsg,
	Args:       cobra.NoArgs,
	Aliases:    []string{"project"},
	Deprecated: consts.ListProjectDepMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return project.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: project.List.PreRunE,
	RunE:    project.List.RunE,
}

func init() {
	Cmd.AddCommand(projects)
	list.SetPageFlags(Cmd, &project.PageOptions)
	listCommon.FlagOutputType(projects, &project.Output)
}
