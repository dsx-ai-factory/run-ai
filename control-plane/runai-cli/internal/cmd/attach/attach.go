/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	commonattach "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attach"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type Command struct {
	cmd.BaseCommand
	options      commonattach.Options
	texts        cmd.CommandTexts
	workloadType api.WorkloadType
	completionFn completion.CompletionFn
}

func NewAttachCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType,
	text cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		options:      commonattach.Options{},
		workloadType: workloadType,
		texts:        text,
		completionFn: completionFn,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "attach [WORKLOAD_NAME]"
	c.Cmd.Short = c.texts.GetShort()
	c.Cmd.Example = c.texts.GetExample()
	c.Cmd.Args = cobra.ExactArgs(1)
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		c.options.WorkloadType = c.workloadType
		c.options.WorkloadFramework = c.Framework
		return commonattach.Pre(cmd, args, &c.options)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return commonattach.Run(cmd, &c.options, c.GetTxnFn())
	}
	c.Cmd.ValidArgsFunction = c.completionFn
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	commonattach.FlagOutputType(c.Cmd, &c.options)
	return nil
}

func (c *Command) GetWorkloadNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Implement the logic to get workload names for auto-completion
	// This is c placeholder and should be replaced with actual implementation
	return []string{}, cobra.ShellCompDirectiveNoFileComp
}
