/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	templateservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Project string
	All     bool
	Page    apilist.PageOptions
}

var (
	Output      listCommon.Output
	listOptions ListOptions
)

// List represents the template list command
var List = &cobra.Command{
	Use:     "list",
	Short:   ListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: ListExample,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if listOptions.Page.Offset < 0 {
			return fmt.Errorf(apilist.ErrNegativeOffset, listOptions.Page.Offset)
		}

		if err := validateAndSetOutput(cmd); err != nil {
			return err
		}
		if err := validateAndSetColumns(cmd); err != nil {
			return err
		}

		// Validate project requirement unless -A flag is used
		if !listOptions.All {
			commonconfig.SetProject(&listOptions.Project)
			cmdHelper.AddValidation(validators.NewProjectValidator(listOptions.Project))
		}
		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl, token, clusterId := getClusterConfig()

		projectId, err := resolveProjectId(cmd, cpUrl, token, clusterId)
		if err != nil {
			return err
		}

		params, err := buildListParams(cmd, projectId)
		if err != nil {
			return err
		}

		return fetchAndPrintTemplates(cmd, cpUrl, token, clusterId, params)
	},
}

// validateAndSetOutput configures and validates the output format
func validateAndSetOutput(cmd *cobra.Command) error {
	if err := commonconfig.SetOutput(&Output); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
	}

	if err := Output.RawFlagValidator(); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
	}

	return nil
}

// validateAndSetColumns validates and configures column settings
func validateAndSetColumns(cmd *cobra.Command) error {
	return list.HandleColumnsFlags(Output, cmd)
}

// getClusterConfig retrieves cluster configuration values
func getClusterConfig() (cpUrl, token, clusterId string) {
	cpUrl = instance.GetConfigValue(config.ControlPlaneUrl)
	token = instance.GetToken()
	clusterId = instance.GetConfigValue(config.ClusterUUID)
	return
}

// resolveProjectId gets the project ID if a project is specified
func resolveProjectId(cmd *cobra.Command, cpUrl, token, clusterId string) (string, error) {
	if listOptions.All || listOptions.Project == "" {
		return "", nil
	}

	projectId, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, listOptions.Project)
	if err != nil {
		return "", cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListTemplates, err)
	}

	return projectId, nil
}

// buildListParams constructs the parameters for the template list request
func buildListParams(cmd *cobra.Command, projectId string) (list.GetListParams, error) {
	params := list.GetListParams{
		ProjectId: projectId,
		All:       listOptions.All,
	}

	if err := setPaginationParams(&params); err != nil {
		return params, cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListTemplates, err)
	}

	return params, nil
}

// setPaginationParams adds pagination parameters to the request params
func setPaginationParams(params *list.GetListParams) error {
	if listOptions.Page.MaxItems > 0 {
		params.Limit = &listOptions.Page.MaxItems
	} else if listOptions.Page.PageSize > 0 && !listOptions.Page.NoPagination {
		params.Limit = &listOptions.Page.PageSize
	}

	if listOptions.Page.Offset > 0 {
		params.Offset = &listOptions.Page.Offset
	}

	// Handle next-token for pagination continuation
	if listOptions.Page.NextToken.IsSet() {
		offset, err := structs.NextTokenToOffset(&listOptions.Page.NextToken)
		if err != nil {
			return fmt.Errorf("invalid next token: %w", err)
		}
		params.Offset = offset
	}

	return nil
}

// fetchAndPrintTemplates retrieves templates and prints the output
func fetchAndPrintTemplates(cmd *cobra.Command, cpUrl, token, clusterId string, params list.GetListParams) error {
	handler, err := templateservice.GetTemplateList(txn, cpUrl, token, clusterId, params)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListTemplates, err)
	}

	printOut, err := Output.Generate(handler)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListTemplates, err)
	}

	cmd.Println(printOut)
	return nil
}

func init() {
	// Add flags
	List.Flags().StringVarP(&listOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)
	List.Flags().BoolVarP(&listOptions.All, "all", "A", false, AllFlagUsage)

	// Pagination flags (using standard page flags like training list)
	apilist.SetPageFlags(List, &listOptions.Page)

	listCommon.FlagOutputType(List, &Output)
	listCommon.FlagOutputTypeWithColumns(List, &Output)

	// Register autocomplete function for columns flag
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, template.ColumnsFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	List.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)
}
