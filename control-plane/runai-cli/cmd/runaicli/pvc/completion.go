/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
)

func GetExistingPVCFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.ExistingPvc, initPVCInstance, getExistingPVCNames)
	return fn(cmd, args, toComplete)
}

func getExistingPVCNames(cmd *cobra.Command) ([]string, error) {
	projectOption, err := workload.GetProjectOption(cmd)
	if err != nil {
		return nil, err
	}

	flagOptions := ListOptions{
		Project: projectOption,
	}

	existingPvcBriefs, err := getList(flagOptions)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, brief := range existingPvcBriefs.GetList() {
		names = append(names, fmt.Sprintf("claimname=%s,", brief.Name))
	}
	return names, nil
}
