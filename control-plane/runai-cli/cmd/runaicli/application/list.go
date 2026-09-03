/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	applicationservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application"
	applist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/list/manipulate"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	applicationCompletion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/application"
	projectCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Output  listCommon.Output
	All     bool
	Project string
	page    apilist.PageOptions
}

var flagsOptions listOptions

var List = &cobra.Command{
	Use:     "list",
	Short:   ListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: ListExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if flagsOptions.page.Offset < 0 {
			return fmt.Errorf(apilist.ErrNegativeOffset, flagsOptions.page.Offset)
		}

		if err := commonconfig.SetOutput(&flagsOptions.Output); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}
		if err := flagsOptions.Output.RawFlagValidator(); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		if !flagsOptions.All {
			if flagsOptions.Project == "" {
				flagsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
			}
			cmdHelper.AddValidation(validators.NewProjectValidator(flagsOptions.Project))
		}

		if err := HandleColumnsFlags(flagsOptions.Output, cmd); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getList(flagsOptions)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListAIApplication, err)
		}

		printOut, err := flagsOptions.Output.Generate(output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListAIApplication, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)
	listCommon.FlagOutputType(List, &flagsOptions.Output)
	listCommon.FlagOutputTypeWithColumns(List, &flagsOptions.Output)
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, applicationCompletion.ColumnsFlagCompletion)
	List.Flags().BoolVarP(&flagsOptions.All, consts.ProjectAll, "A", false, ListAllFlagUsage)
	List.Flags().StringVarP(&flagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", ProjectFlagUsage)
	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, projectCmd.GetProjectNamesForFlag)
	apilist.SetPageFlags(List, &flagsOptions.page)
	List.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}

func getList(opts listOptions) (*applist.Output, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	fb := filter.NewBuilder().WithClusterID(clusterId)

	if opts.Project != "" {
		_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, opts.Project)
		if err != nil {
			return nil, err
		}
		fb.WithProject(opts.Project)
	}

	if opts.page.NextToken.IsSet() {
		fb.WithCreatedAt(opts.page.NextToken.GetTimestamp())
	}

	return applicationservice.GetAIApplicationList(txn, cpUrl, token, &opts.page, fb)
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command) error {
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		if _, err := manipulate.ParseAIApplicationColumns(userColumns); err != nil {
			return err
		}
	}

	if output.IsSetDefaultColumns() {
		if err := settings.GetInstance().SetAIApplicationColumns(userColumns); err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(DefaultAIApplicationColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(DefaultAIApplicationColumnsSet, userColumns)
		}
	}
	return nil
}
