/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

import (
	"errors"
	"fmt"

	commWorkload "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	apilist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/training"
	workloadlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	statusAttribute = "status"
	NoTrainingFound = "No training found"
)

type ListOptions struct {
	output    listCommon.Output
	all       bool
	project   string
	status    string
	framework string
	page      apilist.PageOptions
	deleted   bool
}

var (
	options ListOptions
)

// list represents the list command
var list = &cobra.Command{
	Use:     "list",
	Short:   consts.TrainingListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	GroupID: consts.GroupTrainingBaseID,
	Example: consts.TrainingListExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if options.page.Offset < 0 {
			return fmt.Errorf(apilist.ErrNegativeOffset, options.page.Offset)
		}

		err := commonconfig.SetOutput(&options.output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = options.output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		if !options.all {
			if options.project == "" {
				options.project = instance.GetConfigValue(config.ClusterProjectName)
			}
			cmdHelper.AddValidation(validators.NewProjectValidator(options.project))
		}

		err = workload.HandleColumnsFlags(options.output, cmd)
		if err != nil {
			return err
		}

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := getList(options)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListTraining, err)
		}

		if handler == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListTraining, errors.New("failed to parse list"))
		}

		var printOut string
		printOut, err = options.output.Generate(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListTraining, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	listCommon.FlagOutputType(list, &options.output)
	listCommon.FlagOutputTypeWithColumns(list, &options.output)
	list.Flags().BoolVarP(&options.all, consts.ProjectAll, "A", false, consts.AttributeAllUsage)
	list.Flags().BoolVar(&options.deleted, consts.DeletedWorkload, false, consts.AttributeDeletedWorkloads)
	list.Flags().StringVar(&options.status, "status", "", consts.AttributeStateUsage)
	list.Flags().StringVar(&options.framework, consts.WorkloadFramework, "", consts.AttributeFrameworkUsage)
	list.Flags().StringVarP(&options.project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = list.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	apilist.SetPageFlags(list, &options.page)

	_ = list.RegisterFlagCompletionFunc(statusAttribute, commWorkload.StatusFlagCompletion)
	_ = list.RegisterFlagCompletionFunc(consts.WorkloadFramework, commWorkload.FrameworkFlagCompletion)
	_ = list.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commWorkload.ColumnsFlagCompletion)
	list.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}

func getList(flagOption ListOptions) (*workloadlist.Output, error) {
	txn := instance.GetTxn()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()
	workloadFramework, err := api.ParseWorkloadFramework(flagOption.framework)
	if err != nil {
		return nil, err
	}

	fb := filter.NewBuilder().WithClusterID(clusterId).WithWorkloadFramework(workloadFramework)

	if flagOption.project != "" {
		_, err = projectservice.GetProjectId(txn, cpUrl, token, clusterId, flagOption.project)
		if err != nil {
			return nil, err
		}

		fb.WithProject(flagOption.project)
	}
	if flagOption.status != "" {
		fb.WithWorkloadStatus(flagOption.status)
	}
	return training.List(txn, cpUrl, token, &flagOption.page, fb, flagOption.deleted)
}
