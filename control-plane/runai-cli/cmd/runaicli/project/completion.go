/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/spf13/cobra"
)

func GetProjectNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetArgCompletionFn(completion.Project, initProjectInstance, getProjectNames)
	return fn(cmd, args, toComplete)
}

func GetProjectNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Project, initProjectInstance, getProjectNames)
	return fn(cmd, args, toComplete)
}

func getProjectNames(_ *cobra.Command) ([]string, error) {
	projectBriefs, err := getList(list.NewFilterPageOption())
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, brief := range projectBriefs.GetList() {
		names = append(names, brief.Name)
	}
	return names, nil
}
