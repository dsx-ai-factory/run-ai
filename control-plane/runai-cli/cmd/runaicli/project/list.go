/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"errors"
	"fmt"

	commonProject "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	projectlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project/list"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	Output      listCommon.Output
	PageOptions list.PageOptions

	ErrParseList = errors.New("failed to parse list")
)

// List represents the list command
// set as a global variable to be used by alias commands for backward compatibility
// when the alias commands are removed, this variable should be return to be a local variable
var List = &cobra.Command{
	Use:   "list",
	Short: "list available project",
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if PageOptions.Offset < 0 {
			return fmt.Errorf(list.ErrNegativeOffset, PageOptions.Offset)
		}

		err := commonconfig.SetOutput(&Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}
		err = Output.RawFlagValidator()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.InvalidFlags, err)
		}

		err = HandleColumnsFlags(cmd, Output)
		if err != nil {
			return err
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getList(&PageOptions)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToProjectList, err)
		}

		// Sanity check: this code path should never be reached.
		if output == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToProjectList, ErrParseList)
		}

		var printOut string
		printOut, err = Output.Generate(output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToProjectList, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func getList(pageOptions *list.PageOptions) (*projectlist.Output, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	fb := filter.NewBuilder().WithClusterID(clusterId)
	return projectservice.GetProjectList(txn, cpUrl, token, pageOptions, fb)
}

func init() {
	Cmd.AddCommand(List)
	list.SetPageFlags(List, &PageOptions)
	listCommon.FlagOutputType(List, &Output)
	listCommon.FlagOutputTypeWithColumns(List, &Output)
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commonProject.ColumnsFlagCompletion)
}

func HandleColumnsFlags(cmd *cobra.Command, output listCommon.Output) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseProjectColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetProjectColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(consts.DefaultWProjectColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(consts.DefaultWProjectColumnsSet, userColumns)
		}
	}
	return nil
}
