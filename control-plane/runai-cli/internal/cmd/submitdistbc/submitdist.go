/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdistbc

import (
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/submit"
	workloadcommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"
)

type Command struct {
	cmd.BaseCommand
	distBcSpec   *spec.DistBCSpec
	commandTexts cmd.CommandTexts
}

func NewCommand(specType api.Framework, workloadSpec *spec.DistBCSpec, initBaseCommandFn func() *cobra.Command,
	commandTexts cmd.CommandTexts,
	instance *config.Instance, cmdHelper helper.Cmd) *Command {
	if initBaseCommandFn == nil {
		initBaseCommandFn = cmd.NewEmptyCobraCommand
	}
	return &Command{
		BaseCommand: cmd.BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: specType,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
			CmdHelper: cmdHelper,
		},
		distBcSpec:   workloadSpec,
		commandTexts: commandTexts,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "submit-dist"
	c.Cmd.Short = c.commandTexts.GetShort()
	c.Cmd.Example = c.commandTexts.GetExample()
	c.Cmd.Deprecated = consts.DepSubmitDistMsg
	c.Cmd.PostRunE = func(cmd *cobra.Command, args []string) error {
		c.distBcSpec.Clear()
		return nil
	}
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		err := c.distBcSpec.Mutate()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
		}
		submit.Spec = &c.distBcSpec.GeneralSpec
		return submit.PreDistBc(cmd, args, c.distBcSpec.BcFlags)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := c.distBcSpec.CheckDistributeFlags()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
		}
		if c.Framework == api.MPI {
			err = c.distBcSpec.CheckDistributeMpiFlags()
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
			}
		}
		return submit.Run(cmd)
	}
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	c.distBcSpec.SetFlagsBc(c.Cmd)
	c.distBcSpec.SetFlags(c.Cmd, &submit.Project)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	_ = c.Cmd.MarkFlagRequired(consts.AttributeNameImage)
	c.Cmd.MarkFlagsMutuallyExclusive(consts.AttributeNameRequiredPodAffinityKey, consts.AttributeNamePreferredPodAffinityKey)

	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameCleanPodPolicy, workloadcommon.CleanPodPolicyFlagCompletion)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameMasterRestartPolicy, workloadcommon.RestartPolicyFlagCompletion)
	_ = c.Cmd.RegisterFlagCompletionFunc(consts.AttributeNameRestartPolicy, workloadcommon.RestartPolicyFlagCompletion)

	return nil
}
