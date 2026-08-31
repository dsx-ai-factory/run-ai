/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pf

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/portforward"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	pfoptions "github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/portforward/options"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type PortForwardCommand struct {
	cmd.BaseCommand
	options      pfoptions.Options
	commandMeta  cmd.CommandTexts
	workloadType api.WorkloadType
	completionFn completion.CompletionFn
}

func NewPortForwardCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, workloadType api.WorkloadType,
	texts cmd.CommandTexts, instance *config.Instance, cmdHelper helper.Cmd, completionFn completion.CompletionFn) *PortForwardCommand {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &PortForwardCommand{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		options:      pfoptions.Options{},
		commandMeta:  texts,
		workloadType: workloadType,
		completionFn: completionFn,
	}
}

func (pf *PortForwardCommand) GetCommand() *cobra.Command {
	return pf.buildCommand()
}

func (pf *PortForwardCommand) buildCommand() *cobra.Command {
	pf.Cmd.Use = "port-forward [WORKLOAD_NAME]"
	pf.Cmd.Short = pf.commandMeta.GetShort()
	pf.Cmd.Long = pf.commandMeta.GetLong()
	pf.Cmd.Aliases = pf.commandMeta.GetAliases()
	pf.Cmd.Example = pf.commandMeta.GetExample()
	pf.Cmd.Args = cobra.ExactArgs(1)
	pf.Cmd.PreRunE = pf.preRunE
	pf.Cmd.RunE = pf.runE
	pf.Cmd.ValidArgsFunction = pf.completionFn
	return pf.Cmd
}

func (pf *PortForwardCommand) preRunE(cmd *cobra.Command, args []string) error {
	pf.options.WorkloadType = pf.workloadType
	pf.options.WorkloadFramework = pf.Framework
	return portforward.Pre(cmd, args, &pf.options)
}

func (pf *PortForwardCommand) runE(cmd *cobra.Command, args []string) error {
	return portforward.Run(cmd, &pf.options)
}

func (pf *PortForwardCommand) RegisterFlags() error {
	portforward.FlagOutputType(pf.Cmd, &pf.options, pf.Framework)
	return nil
}
