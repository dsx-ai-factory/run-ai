/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// clusters represents the clusters command
// Deprecated: Use `runai workload list` instead
var jobs = &cobra.Command{
	Use:        "jobs",
	Short:      consts.ListJobShortMsg,
	Aliases:    []string{"job", "workloads", "workload"},
	Args:       cobra.NoArgs,
	Deprecated: consts.ListJobDepMsg,
	// Do not add deprecated flag, it removes the command from help

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return workload.Cmd.PersistentPreRunE(cmd, args)
	},
	PreRunE: workload.List.PreRunE,
	RunE:    workload.List.RunE,
}

func init() {
	Cmd.AddCommand(jobs)
	listCommon.FlagOutputType(jobs, &workload.FlagsOptions.Output)
	jobs.Flags().BoolVarP(&workload.FlagsOptions.All, "all-projects", "A", false, consts.AttributeAllUsage)
	jobs.Flags().StringVarP(&workload.FlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
}
