/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"errors"
	"fmt"
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmd"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/spec"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	UpdateCommandAnnotationKey = "runai-update-command"
	errWorkloadNotFound        = errors.New("workload not found")
	updatingMsg                = "inference %s updated successfully\n"
	dryRunMsg                  = "%s"
	EnvFormat                  = "%s=%s"
)

type Command struct {
	cmd.BaseCommand
	generalSpec      *spec.GeneralSpec
	commandTexts     cmd.SearchableCommandTexts
	searchWorkloadFn cmd.SearchWorkloadFn
	uuid             string
	updateWorkloadFn cmd.UpdateWorkloadFn
	project          string
}

func NewCommand(specType api.Framework, workloadSpec *spec.GeneralSpec, initBaseCommandFn func() *cobra.Command,
	commandTexts cmd.SearchableCommandTexts, instance *config.Instance, cmdHelper helper.Cmd,
	searchWorkloadFn cmd.SearchWorkloadFn, updateWorkloadFn cmd.UpdateWorkloadFn) *Command {
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
		generalSpec:      workloadSpec,
		commandTexts:     commandTexts,
		searchWorkloadFn: searchWorkloadFn,
		updateWorkloadFn: updateWorkloadFn,
	}
}

func (c *Command) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *Command) buildCommand() *cobra.Command {
	c.Cmd.Use = "update [WORKLOAD_NAME]"
	c.Cmd.Short = c.commandTexts.GetShort()
	c.Cmd.Long = c.commandTexts.GetLong()
	c.Cmd.Example = c.commandTexts.GetExample()
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		token := c.Instance.GetToken()
		cpUrl := c.Instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := c.Instance.GetConfigValue(config.ClusterUUID)
		instance := config.GetInstance()

		project := c.project
		if project == "" {
			project = instance.GetConfigValue(config.ClusterProjectName)
		}

		var cmdHelper helper.Cmd
		cmdHelper.AddValidation(validators.NewProjectValidator(project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		Name, err := argshandler.ValidateArgsAndExtractName(cmd, &args)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToUpdate, err)
		}

		if Name == "" {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), errors.New(consts.ErrorWorkloadNameIsRequired))
		}

		fb := filter.NewBuilder().WithWorkloadFramework(c.Framework)
		handler, err := c.searchWorkloadFn(c.GetTxnFn(), cpUrl, token, Name, project, clusterId, fb)

		if err != nil {
			_ = cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetSearchableErrorMsg(), err)

			if handler != nil && len(handler.GetList()) > 1 {
				cmd.Println(consts.PreFoundMultipleWorkloadMessage)
				cmd.Println(handler.GenerateTable())
			}
			return err
		}
		if handler != nil && len(handler.GetList()) != 1 {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetNoResultsErrorMsg(), errWorkloadNotFound)
		}
		c.uuid = handler.GetList()[0].Id

		if execute, found := argshandler.GetExecuteString(&args); found {
			c.generalSpec.SetExecute(execute)
		}

		return nil
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		instance := config.GetInstance()
		txn := instance.GetTxn()
		token := instance.GetToken()
		cpUrl := c.Instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := c.Instance.GetConfigValue(config.ClusterUUID)

		c.setCliCommandAsAnnotation()

		workloadName, err := c.updateWorkloadFn(txn, cpUrl, token, c.uuid, clusterId, c.generalSpec)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, c.commandTexts.GetBaseErrorMsg(), err)
		}

		if c.generalSpec.IsDryRun() {
			cmd.Printf(dryRunMsg, workloadName)
		} else {
			cmd.Printf(updatingMsg, workloadName)
		}
		return nil
	}
	return c.Cmd
}

func (c *Command) RegisterFlags() error {
	c.generalSpec.SetUpdateFlags(c.Cmd)
	c.Cmd.Flags().StringVarP(&c.project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)

	return nil
}

func (c *Command) setCliCommandAsAnnotation() {
	osArgs := os.Args
	arguments, _ := argshandler.GetExecuteString(&osArgs)
	c.generalSpec.SetCliCommandAsAnnotation(fmt.Sprintf(EnvFormat, UpdateCommandAnnotationKey, arguments))
}
