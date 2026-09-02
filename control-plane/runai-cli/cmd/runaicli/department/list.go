/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package department

import (
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/department/service"

	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"

	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	departmentservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/department"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/department/model"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	Output listCommon.Output
)

type ListOptions struct {
	limit  int32
	offset int32
}

var (
	options ListOptions
)

const (
	DefaultItems = 50
)

// List represents the list command
// set as a global variable to be used by alias commands for backward compatibility
// when the alias commands are removed, this variable should be return to be a local variable
var List = &cobra.Command{
	Use:   "list",
	Short: "list available departments",
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if options.offset < 0 {
			return fmt.Errorf(apilist.ErrNegativeOffset, options.offset)
		}

		err := commonconfig.SetOutput(&Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		handler, err := getList(options)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
		}

		if handler == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, errors.New("failed to parse list"))
		}

		var printOut string
		printOut, err = Output.Generate(handler)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListDepartments, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func getList(flagOption ListOptions) (*service.BriefDepartmentHandler, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	filters := []model.Filter{*model.NewClusterIDFilter(clusterId)}
	return departmentservice.GetDepartmentList(txn, cpUrl, token, filters, flagOption.limit, flagOption.offset)
}

func init() {
	Cmd.AddCommand(List)
	SetFlags(List)
	listCommon.FlagOutputType(List, &Output)
}

func SetFlags(cmd *cobra.Command) {
	cmd.Flags().Int32Var(&options.limit, "limit", DefaultItems, consts.AttributeLimitUsage)
	cmd.Flags().Int32Var(&options.offset, "offset", 0, consts.AttributeOffsetUsage)
}
