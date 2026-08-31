/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	"fmt"

	nodecp "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node"

	commonNode "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/node"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/settings"

	commonconfig "github.com/run-ai/runai/control-plane/runai-cli/internal/config"
	nodeservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/node/list"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	Output listCommon.Output
)

// List represents the list command
// set as a global variable to be used by alias commands for backward compatibility
// when the alias commands are removed, this variable should be return to be a local variable
var List = &cobra.Command{
	Use:   "list",
	Short: "List node",
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonconfig.SetOutput(&Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		err = HandleColumnsFlags(cmd, Output)
		if err != nil {
			return err
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var printOut string
		output, err := executeNodeList()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListNodes, err)
		}

		// Sanity check: this code path should never be reached.
		if output == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListNodes, fmt.Errorf("failed to parse list"))
		}

		printOut, err = Output.Generate(output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListNodes, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)

	listCommon.FlagOutputType(List, &Output)
	listCommon.FlagOutputTypeWithColumns(List, &Output)

	// Register autocomplete function for columns flag
	_ = List.RegisterFlagCompletionFunc(listCommon.ColumnsStr, commonNode.ColumnsFlagCompletion)
}

func executeNodeList() (*nodeservice.Output, error) {
	txn := instance.GetTxn()
	ctx := txn.GetCtx()
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()
	clusterUuid := instance.GetConfigValue(config.ClusterUUID)

	return nodecp.GetNodeList(ctx, cpUrl, token, clusterUuid)
}

func HandleColumnsFlags(cmd *cobra.Command, output listCommon.Output) error {
	var err error
	userColumns := output.GetColumns()

	if len(userColumns) > 0 {
		_, err = manipulate.ParseNodeColumns(userColumns)
		if err != nil {
			return err
		}
	}
	if output.IsSetDefaultColumns() {
		err = settings.GetInstance().SetNodeColumns(userColumns)
		if err != nil {
			return err
		}
		if len(userColumns) == 0 {
			cmd.Printf(consts.DefaultNodeColumnsSet, manipulate.GetDefaultColumns())
		} else {
			cmd.Printf(consts.DefaultNodeColumnsSet, userColumns)
		}
	}
	return nil
}
