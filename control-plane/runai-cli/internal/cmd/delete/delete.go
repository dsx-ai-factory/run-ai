/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package delete

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/spf13/cobra"
)

type Command struct {
	cmd.SearchBaseCommand
	completionFn completion.CompletionFn
}

func NewDeleteCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, messages cmd.SearchableCommandTexts,
	instance *config.Instance, cmdHelper helper.Cmd, operationFn cmd.RunOperationFn, searchFn cmd.SearchWorkloadFn, workloadType api.WorkloadType, completionFn completion.CompletionFn) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		SearchBaseCommand: cmd.SearchBaseCommand{
			BaseCommand: cmd.BaseCommand{
				Cmd:       initBaseCommandFn(),
				Framework: framework,
				GetTxnFn:  instance.GetTxn,
				Instance:  instance,
				CmdHelper: cmdHelper,
			},
			OperationType: "delete",
			OperationFunc: operationFn,
			GetWorkloadFn: searchFn,
			Texts:         messages,
			WorkloadType:  workloadType,
		},
		completionFn: completionFn,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	baseCmd := c.SearchBaseCommand.GetCommand()
	baseCmd.ValidArgsFunction = c.completionFn
	return baseCmd
}
