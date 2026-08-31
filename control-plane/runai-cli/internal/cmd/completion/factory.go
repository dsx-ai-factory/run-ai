/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package completion

import (
	"errors"

	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"

	workloadlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/spf13/cobra"
)

type ListCommand interface {
	cmd.ActionCommand
	GetList(project, workloadState string, pagingOptions *apilist.PageOptions) (*workloadlist.Output, error)
}

type CompletionFactory struct {
	listCmdFactory *list.Factory
	initInstanceFn func(command *cobra.Command) error
}

func NewCompletionFactory(listCmdFactory *list.Factory, initInstanceFn func(command *cobra.Command) error) *CompletionFactory {
	return &CompletionFactory{
		listCmdFactory: listCmdFactory,
		initInstanceFn: initInstanceFn,
	}
}

func (f *CompletionFactory) GetCompletionFn(framework api.Framework, workloadType api.WorkloadType) (completion.CompletionFn, error) {
	ac, err := f.listCmdFactory.GetCommand(framework, workloadType)
	if err != nil {
		return nil, err
	}
	listCmd, ok := ac.(ListCommand)
	if !ok {
		// should not happen
		return nil, errors.New("completion factory got wrong list command")
	}

	optionsFn := func(cmd *cobra.Command) ([]string, error) {
		projectOption, err := workload.GetProjectOption(cmd)
		if err != nil {
			return nil, err
		}

		workloadBriefs, err := listCmd.GetList(projectOption, "", apilist.NewFilterPageOption())
		if err != nil {
			return nil, err
		}

		names := make([]string, 0)
		for _, brief := range workloadBriefs.GetList() {
			names = append(names, brief.Name)
		}
		return names, nil
	}

	return completion.GetArgCompletionFn(completion.Workload, f.initInstanceFn, optionsFn), nil
}
