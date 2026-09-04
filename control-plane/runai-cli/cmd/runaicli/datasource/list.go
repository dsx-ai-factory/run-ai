/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	dsCompletion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/datasource"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	dslist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	All     bool
	Project string
	Output  list.Output
	Type    string
}

var (
	listOptions ListOptions
)

var List = &cobra.Command{
	Use:     "list",
	Short:   ListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: ListExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonconfig.SetOutput(&listOptions.Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = listOptions.Output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		// Handle columns flags
		err = dslist.HandleColumnsFlags(listOptions.Output, cmd, DefaultDatasourceColumnsSet)
		if err != nil {
			return err
		}

		if !listOptions.All {
			if listOptions.Project == "" {
				listOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
			}
			cmdHelper.AddValidation(validators.NewProjectValidator(listOptions.Project))
		}

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		clusterId := instance.GetConfigValue(config.ClusterUUID)

		projectName := ""
		if !listOptions.All {
			projectName = listOptions.Project
		}

		handler, err := dslist.List(instance.GetTxn(), cpUrl, token, projectName, clusterId, listOptions.Type)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListDatasource, err)
		}

		if err := renderList(cmd, listOptions.Output, handler); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListDatasource, err)
		}
		return nil
	},
}

func renderList(cmd *cobra.Command, output list.Output, handler *dslist.Output) error {
	printOut, err := output.Generate(handler)
	if err != nil {
		return err
	}

	cmd.Println(printOut)
	return nil
}

func init() {
	Cmd.AddCommand(List)

	// Register output flags
	list.FlagOutputType(List, &listOptions.Output)
	list.FlagOutputTypeWithColumns(List, &listOptions.Output)

	// Register flags
	List.Flags().BoolVarP(&listOptions.All, consts.ProjectAll, "A", false,
		ListAllFlagUsage)
	List.Flags().StringVarP(&listOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)
	List.Flags().StringVar(&listOptions.Type, TypeFlag, "", TypeFlagUsage)

	// Register auto-completion for flags
	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	_ = List.RegisterFlagCompletionFunc(list.ColumnsStr, dsCompletion.ColumnsFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(TypeFlag, dsCompletion.TypeFlagCompletion)

	// Mark flags as mutually exclusive
	List.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}
