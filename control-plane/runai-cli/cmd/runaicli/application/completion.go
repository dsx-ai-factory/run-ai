/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

import (
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func GetAIApplicationNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetArgCompletionFn(completion.AIApplication, cmdfactory.InstanceInit, getAIApplicationNames)
	return fn(cmd, args, toComplete)
}

func getAIApplicationNames(cmd *cobra.Command) ([]string, error) {
	project := getFlagValue(cmd, "project")
	if project == "" {
		project = instance.GetConfigValue(config.ClusterProjectName)
	}
	output, err := getList(listOptions{Project: project, page: *apilist.NewFilterPageOption()})
	if err != nil {
		return nil, err
	}
	briefs := output.GetList()
	names := make([]string, 0, len(briefs))
	for _, app := range briefs {
		names = append(names, app.Name)
	}
	return names, nil
}

func getFlagValue(cmd *cobra.Command, flagName string) string {
	value, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return ""
	}
	return value
}
