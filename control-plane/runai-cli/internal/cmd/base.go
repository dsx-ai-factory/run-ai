/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/inference"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workspace"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/distributed"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/nim"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/training"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/common/attributes"
	workloadCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	UuidAttribute = "uuid"
)

type SearchBaseCommand struct {
	BaseCommand
	Options       workloadCommon.ActionsOptions
	Texts         SearchableCommandTexts
	WorkloadType  api.WorkloadType
	OperationType string
	OperationFunc RunOperationFn
	GetWorkloadFn SearchWorkloadFn
	cmdHelper     helper.Cmd
}

func NewSearchBaseCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command, operationType string,
	operationFunc RunOperationFn, getWorkloadFn SearchWorkloadFn, workloadType api.WorkloadType,
	instance *config.Instance, commandTexts SearchableCommandTexts, cmdHelper helper.Cmd) *SearchBaseCommand {
	if initBaseCommandFn == nil {
		initBaseCommandFn = NewEmptyCobraCommand
	}
	return &SearchBaseCommand{
		BaseCommand: BaseCommand{
			Cmd:       initBaseCommandFn(),
			Framework: framework,
			GetTxnFn:  instance.GetTxn,
			Instance:  instance,
		},
		Options:       workloadCommon.ActionsOptions{},
		cmdHelper:     cmdHelper,
		Texts:         commandTexts,
		OperationType: operationType,
		OperationFunc: operationFunc,
		GetWorkloadFn: getWorkloadFn,
		WorkloadType:  workloadType,
	}
}
func (c *SearchBaseCommand) GetCommand() *cobra.Command {
	return c.buildCommand()
}

func (c *SearchBaseCommand) buildCommand() *cobra.Command {
	c.Cmd.Use = fmt.Sprintf("%s [WORKLOAD_NAME]", c.OperationType)
	c.Cmd.Short = c.Texts.GetShort()
	c.Cmd.Args = cobra.MaximumNArgs(1)
	c.Cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		return c.preRunE(cmd, args)
	}
	c.Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return c.runE(cmd, args)
	}
	c.Cmd.Aliases = c.Texts.GetAliases()
	c.Cmd.Example = c.Texts.GetExample()
	return c.Cmd
}

func (c *SearchBaseCommand) preRunE(cmd *cobra.Command, args []string) error {
	c.Options.Name = argshandler.GetNameFromArgs(&args)
	err := attributes.MutuallyExclusiveAtLeastOne(c.Options.Name, c.Options.Uuid)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, c.Texts.GetBaseErrorMsg(), fmt.Errorf(consts.ErrorNameOrUuidIsNeeded))
	}

	if c.Options.Project == "" {
		c.Options.Project = c.Instance.GetConfigValue(config.ClusterProjectName)
	}

	c.cmdHelper.AddValidation(validators.NewProjectValidator(c.Options.Project))

	if msg, err := c.cmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	return nil
}

func (c *SearchBaseCommand) runE(cmd *cobra.Command, _ []string) error {
	token := c.Instance.GetToken()
	cpUrl := c.Instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := c.Instance.GetConfigValue(config.ClusterUUID)

	if c.Options.Name != "" {
		fb := filter.NewBuilder().WithWorkloadFramework(c.Framework).WithWorkloadType(c.WorkloadType.QueryParam())
		handler, err := c.GetWorkloadFn(c.GetTxnFn(), cpUrl, token, c.Options.Name, c.Options.Project, clusterId, fb)

		if err != nil {
			_ = cmdtxn.CmdErrorPrintAndSet(cmd, c.Texts.GetSearchableErrorMsg(), err)

			if handler != nil && len(handler.GetList()) > 1 {
				cmd.Println(consts.PreFoundMultipleWorkloadMessage)
				cmd.Println(handler.GenerateTable())
			}
			return err
		}
		c.Options.Uuid = handler.GetList()[0].Id
	}

	err := c.OperationFunc(c.GetTxnFn(), cpUrl, token, c.Options.Uuid)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, c.Texts.GetBaseErrorMsg(), err)
	}

	target := c.Options.Uuid
	if c.Options.Name != "" {
		target = c.Options.Name
	}
	cmd.Print(c.Texts.GetSuccessMsg(target))
	return nil
}

func CommonFlags(cmd *cobra.Command, actions *workloadCommon.ActionsOptions) {
	cmd.Flags().StringVarP(&actions.Uuid, UuidAttribute, "u", "", consts.AttributeUuidUsage)
	cmd.Flags().StringVarP(&actions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
}

func (c *SearchBaseCommand) RegisterFlags() error {
	CommonFlags(c.Cmd, &c.Options)
	return nil
}

type BaseCommand struct {
	Cmd       *cobra.Command
	Framework api.Framework
	GetTxnFn  func() *transaction.Txn
	Instance  *config.Instance
	CmdHelper helper.Cmd
}

func GetSearchFunc(framework api.Framework, workloadType api.WorkloadType) (SearchWorkloadFn, error) {
	if framework <= api.UndefinedFramework || (framework >= api.AnyDistributed && framework != api.NimService) {
		return nil, NewUnsupportedError(framework, workloadType)
	}

	if framework == api.NimService {
		return nim.GetByNameAndProj, nil
	}

	switch workloadType {
	case api.Distributed, api.DistributedInference:
		return distributed.GetByNameAndProj, nil
	case api.Inference:
		return inference.GetByNameAndProj, nil
	case api.Training:
		return training.GetByNameAndProj, nil
	case api.Workspace:
		return workspace.GetByNameAndProj, nil
	default:
		return nil, NewUnsupportedError(framework, workloadType)
	}
}
