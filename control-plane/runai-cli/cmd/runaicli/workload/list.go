/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"

	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list"

	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	commWorkload "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/workload"

	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Output       listCommon.Output
	All          bool
	Project      string
	workloadType string
	framework    string
	status       string
	page         apilist.PageOptions
	deleted      bool
}

var (
	FlagsOptions ListOptions
)

// List represents the list command
// set as a global variable to be used by alias commands for backward compatibility
// when the alias commands are removed, this variable should be return to be a local variable
var List = &cobra.Command{
	Use:     "list",
	Short:   "List workloads",
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if FlagsOptions.page.Offset < 0 {
			return fmt.Errorf(apilist.ErrNegativeOffset, FlagsOptions.page.Offset)
		}

		err := commonconfig.SetOutput(&FlagsOptions.Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = FlagsOptions.Output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		if !FlagsOptions.All {
			if FlagsOptions.Project == "" {
				FlagsOptions.Project = instance.GetConfigValue(config.ClusterProjectName)
			}
			cmdHelper.AddValidation(validators.NewProjectValidator(FlagsOptions.Project))
		}
		err = HandleColumnsFlags(FlagsOptions.Output, cmd)
		if err != nil {
			return err
		}

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := getList(FlagsOptions)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListWorkload, err)
		}

		var printOut string
		printOut, err = FlagsOptions.Output.Generate(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListWorkload, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)

	listCommon.FlagOutputType(List, &FlagsOptions.Output)
	listCommon.FlagOutputTypeWithColumns(List, &FlagsOptions.Output)
	List.Flags().BoolVarP(&FlagsOptions.All, consts.ProjectAll, "A", false, consts.AttributeAllUsage)
	List.Flags().BoolVar(&FlagsOptions.deleted, consts.DeletedWorkload, false, consts.AttributeDeletedWorkloads)
	List.Flags().StringVar(&FlagsOptions.workloadType, consts.WorkloadType, "", consts.AttributeTypeUsage)
	List.Flags().StringVar(&FlagsOptions.framework, consts.WorkloadFramework, "", consts.AttributeFrameworkUsage)
	List.Flags().StringVar(&FlagsOptions.status, "status", "", consts.AttributeStateUsage)
	List.Flags().StringVarP(&FlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)

	apilist.SetPageFlags(List, &FlagsOptions.page)

	// Mark Flag Section
	List.MarkFlagsMutuallyExclusive(consts.WorkloadType, consts.WorkloadFramework)
	List.MarkFlagsMutuallyExclusive(consts.ProjectAll, consts.ProjectStr)

	_ = List.RegisterFlagCompletionFunc(statusAttribute, commWorkload.StatusFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(consts.WorkloadType, commWorkload.TypeFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(consts.WorkloadFramework, commWorkload.FrameworkFlagCompletion)
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commWorkload.ColumnsFlagCompletion)
}

func getList(flagOptions ListOptions) (*list.Output, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()
	fb, err := GetOptionFilter(flagOptions, cpUrl, token, clusterId)
	if err != nil {
		return nil, err
	}
	return workload.GetWorkloadPagination(txn, cpUrl, token, &flagOptions.page, fb, flagOptions.deleted)
}

func GetOptionFilter(flagOptions ListOptions, cpUrl string, token string, clusterId string) (*filter.Builder, error) {
	fb := filter.NewBuilder().WithClusterID(clusterId)

	if flagOptions.Project != "" {
		_, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, flagOptions.Project)
		if err != nil {
			return nil, err
		}
		fb.WithProject(flagOptions.Project)
	}

	if flagOptions.status != "" {
		fb.WithWorkloadStatus(flagOptions.status)
	}

	if flagOptions.workloadType != "" {
		workloadType, err := api.ParseWorkloadType(strings.ToLower(flagOptions.workloadType))
		if err != nil {
			return nil, err
		}
		fb.WithWorkloadType(workloadType.QueryParam())
	}

	if flagOptions.framework != "" {
		workloadFramework, err := api.ParseWorkloadFramework(flagOptions.framework)
		if err != nil {
			return nil, err
		}
		fb.WithWorkloadFramework(workloadFramework)
	}

	if flagOptions.page.NextToken.IsSet() {
		fb.WithCreatedAt(flagOptions.page.NextToken.GetTimestamp())
	}
	return fb, nil
}

func HandleColumnsFlags(output listCommon.Output, cmd *cobra.Command) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseWorkloadColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetWorkloadColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(consts.DefaultWorkloadColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(consts.DefaultWorkloadColumnsSet, userColumns)
		}
	}
	return nil
}
