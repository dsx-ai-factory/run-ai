/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"errors"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/workloadtype"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	workloadtypeservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype"
	workloadtypelist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/list"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	Output       listCommon.Output
	externalOnly bool
)

// List represents the list command
var List = &cobra.Command{
	Use:   "list",
	Short: "list available workload types",
	Args:  cobra.NoArgs,
	Example: `  # List all workload types
  runai workload-type list

  # List only external workload types
  runai workload-type list --external-only

  # List with specific columns
  runai workload-type list --columns name,category,priority

  # Output as JSON
  runai workload-type list --json`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
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
		output, err := getList()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListWorkloadTypes, err)
		}

		// Sanity check: this code path should never be reached.
		if output == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListWorkloadTypes, errors.New("failed to parse list"))
		}

		var printOut string
		printOut, err = Output.Generate(output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListWorkloadTypes, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func getList() (*workloadtypelist.Output, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()

	fb := filter.NewBuilder()
	return workloadtypeservice.List(txn, cpUrl, token, fb, &externalOnly)
}

func init() {
	Cmd.AddCommand(List)
	listCommon.FlagOutputType(List, &Output)
	listCommon.FlagOutputTypeWithColumns(List, &Output)
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, workloadtype.ColumnsFlagCompletion)

	List.Flags().BoolVar(&externalOnly, consts.ExternalOnly, false, consts.ExternalOnlyUsage)
}

func HandleColumnsFlags(cmd *cobra.Command, output listCommon.Output) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseWorkloadTypeColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetWorkloadTypeColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(consts.DefaultWorkloadTypeColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(consts.DefaultWorkloadTypeColumnsSet, userColumns)
		}
	}
	return nil
}
