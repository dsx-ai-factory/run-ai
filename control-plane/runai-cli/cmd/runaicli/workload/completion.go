/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func GetWorkloadNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetArgCompletionFn(completion.Workload, initWorkloadInstance, getWorkloadNames)
	return fn(cmd, args, toComplete)
}

func getWorkloadNames(cmd *cobra.Command) ([]string, error) {
	typeOption := getFlagValue(cmd, consts.WorkloadType)
	frameworkOption := getFlagValue(cmd, consts.WorkloadFramework)

	if len(typeOption) == 0 && len(frameworkOption) == 0 {
		// we don't complete if the mandatory type flag is not set
		// return empty instead of error so the `--type` flag completion will still occur
		return []string{}, nil
	}

	projectOption, err := GetProjectOption(cmd)
	if err != nil {
		return nil, err
	}

	flagOptions := ListOptions{
		workloadType: typeOption,
		Project:      projectOption,
		page:         *list.NewFilterPageOption(),
		framework:    frameworkOption,
	}

	workloadBriefs, err := getList(flagOptions)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, brief := range workloadBriefs.GetList() {
		names = append(names, brief.Name)
	}
	return names, nil
}

func GetProjectOption(cmd *cobra.Command) (string, error) {
	fromFlag := getFlagValue(cmd, consts.ProjectStr)
	if len(fromFlag) != 0 {
		return fromFlag, nil
	}

	return getProjectFromContext()
}

func getFlagValue(cmd *cobra.Command, flagName string) string {
	value, err := cmd.Flags().GetString(flagName)
	if err != nil {
		value = ""
	}
	return value
}

func getProjectFromContext() (string, error) {
	ctxProject := instance.GetConfigValue(config.ClusterProjectName)
	cmdHelper.AddValidation(validators.NewProjectValidator(ctxProject))
	_, err := cmdHelper.Check()
	if err != nil {
		return "", err
	}
	return ctxProject, nil
}
