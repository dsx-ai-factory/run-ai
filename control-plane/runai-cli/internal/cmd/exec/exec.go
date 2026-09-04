/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec

import (
	commonexec "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/exec"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	MinimalArgs  = 2
	MainUseName  = "exec"
	BashUseAlias = "bash"
	bashCmd      = "/bin/bash"
)

type Command struct {
	cmd.BaseCommand
	options      commonexec.Options
	texts        cmd.CommandTexts
	workloadType api.WorkloadType
	completionFn completion.CompletionFn
	preFunc      func(cmd *cobra.Command, args []string, execOptions *commonexec.Options) error
	ArgsFunc     cobra.PositionalArgs
	useName      string
}

func NewExecCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType,
	text cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn,
	preFunc func(cmd *cobra.Command, args []string, execOptions *commonexec.Options) error, argFunc cobra.PositionalArgs) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	if preFunc == nil {
		preFunc = commonexec.Pre
	}
	if argFunc == nil {
		argFunc = cobra.MinimumNArgs(MinimalArgs)
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		options:      commonexec.Options{},
		workloadType: workloadType,
		texts:        text,
		completionFn: completionFn,
		preFunc:      preFunc,
		ArgsFunc:     argFunc,
		useName:      MainUseName,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = c.useName + " [WORKLOAD_NAME]"
	c.Cmd.Short = c.texts.GetShort()
	c.Cmd.Long = c.texts.GetLong()
	c.Cmd.Example = c.texts.GetExample()
	c.Cmd.Args = c.ArgsFunc
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		c.options.WorkloadType = c.workloadType
		c.options.WorkloadFramework = c.Framework
		return c.preFunc(cmd, args, &c.options)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return commonexec.Run(cmd, &c.options, c.GetTxnFn())
	}
	c.Cmd.ValidArgsFunction = c.completionFn
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	commonexec.FlagOutputType(c.Cmd, &c.options)
	return nil
}

func NewBashCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType,
	text cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn) *Command {
	bashAlias := NewExecCommand(framework, initBaseCommandFn, workloadType, text, instance, cmdHelper, completionFn, commonexec.BashPreRun, cobra.ExactArgs(1))
	bashAlias.useName = BashUseAlias
	return bashAlias
}
