/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"

	commWorkload "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	statusAttribute = "status"
)

type Options struct {
	output        listCommon.Output
	all           bool
	deleted       bool
	project       string
	workloadState string
	page          apilist.PageOptions
}

type Command struct {
	cmd.BaseCommand
	options      Options
	texts        cmd.CommandTexts
	listFn       cmd.ListWorkloadsFn
	workloadType api.WorkloadType
}

func NewListCommand(framework api.Framework, initBaseCommandFn func() *cobra.Command,
	instance *config.Instance, texts cmd.CommandTexts, cmdHelper helper.Cmd, listFn cmd.ListWorkloadsFn, workloadType api.WorkloadType) *Command {
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
		options:      Options{},
		texts:        texts,
		listFn:       listFn,
		workloadType: workloadType,
	}
}

func (l *Command) GetCommand() *cobra.Command {
	return l.buildCommand()
}

func (l *Command) buildCommand() *cobra.Command {
	l.Cmd.Use = "list"
	l.Cmd.Short = l.texts.GetShort()
	l.Cmd.Aliases = []string{"ls"}
	l.Cmd.Args = cobra.NoArgs
	l.Cmd.Example = l.texts.GetExample()
	l.Cmd.PreRunE = l.preRunE
	l.Cmd.RunE = l.runE
	return l.Cmd
}

func (l *Command) preRunE(cmd *cobra.Command, _ []string) error {
	err := commonconfig.SetOutput(&l.options.output)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
	}

	err = l.options.output.RawFlagValidator()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
	}

	if !l.options.all {
		if l.options.project == "" {
			l.options.project = l.Instance.GetConfigValue(config.ClusterProjectName)
		}
		l.CmdHelper.AddValidation(validators.NewProjectValidator(l.options.project))
	}

	err = workload.HandleColumnsFlags(l.options.output, cmd)
	if err != nil {
		return err
	}

	if msg, err := l.CmdHelper.Check(); err != nil {
		return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
	}
	return nil
}

func (l *Command) runE(cmd *cobra.Command, _ []string) error {
	handler, err := l.GetList(l.options.project, l.options.workloadState, &l.options.page)
	if err != nil || handler == nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, l.texts.GetBaseErrorMsg(), err)
	}

	var printOut string
	printOut, err = l.options.output.Generate(handler)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, l.texts.GetBaseErrorMsg(), err)
	}
	cmd.Println(printOut)
	return nil
}

func (l *Command) GetList(project, workloadState string, pagingOptions *apilist.PageOptions) (*list.Output, error) {
	cpUrl := l.Instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := l.Instance.GetConfigValue(config.ClusterUUID)

	token := l.Instance.GetToken()
	fb := filter.NewBuilder().
		WithClusterID(clusterId).
		WithWorkloadFramework(l.Framework)

	// NIM sets its own workload type ("NIMService") in nim.List; skip here to avoid filtering by both "Inference" and "NIMService"
	if l.Framework != api.NimService {
		fb.WithWorkloadType(l.workloadType.QueryParam())
	}

	if project != "" {
		_, err := projectservice.GetProjectId(l.GetTxnFn(), cpUrl, token, clusterId, project)
		if err != nil {
			return nil, err
		}

		fb.WithProject(project)
	}

	if workloadState != "" {
		fb.WithWorkloadStatus(workloadState)
	}

	return l.listFn(l.GetTxnFn(), cpUrl, token, pagingOptions, fb, l.options.deleted)
}

func (l *Command) RegisterFlags() error {
	listCommon.FlagOutputType(l.Cmd, &l.options.output)
	listCommon.FlagOutputTypeWithColumns(l.Cmd, &l.options.output)
	l.Cmd.Flags().BoolVarP(&l.options.all, consts.ProjectAll, "A", false, consts.AttributeAllUsage)
	l.Cmd.Flags().BoolVar(&l.options.deleted, consts.DeletedWorkload, false, consts.AttributeDeletedWorkloads)
	l.Cmd.Flags().StringVar(&l.options.workloadState, statusAttribute, "", consts.AttributeStateUsage)
	l.Cmd.Flags().StringVarP(&l.options.project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = l.Cmd.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	apilist.SetPageFlags(l.Cmd, &l.options.page)
	l.Cmd.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)

	_ = l.Cmd.RegisterFlagCompletionFunc(statusAttribute, commWorkload.StatusFlagCompletion)

	_ = l.Cmd.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commWorkload.ColumnsFlagCompletion)
	return nil
}
