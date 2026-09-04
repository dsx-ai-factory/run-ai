/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/pvc"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	FailedToListPVC = "PVC list failed"
	ListUse         = "list"
	ListShort       = "List PVC"

	Examples = `# List of PVC by project with table output
runai pvc list -p <project-name>

# List of PVC by project with table JSON output
runai pvc list -p <project-name> --json

# List of PVC by project with YAML  output
runai pvc list -p <project-name> --yaml`
)

type ListOptions struct {
	Output  listCommon.Output
	Project string
}

var (
	FlagsOptions ListOptions
)

// List represents the pvc list command
var List = &cobra.Command{
	Use:     ListUse,
	Short:   ListShort,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: Examples,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := commonconfig.SetOutput(&FlagsOptions.Output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		commonconfig.SetProject(&FlagsOptions.Project)
		cmdHelper.AddValidation(validators.NewProjectValidator(FlagsOptions.Project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		output, err := getList(FlagsOptions)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListPVC, err)
		}

		var printOut string
		printOut, err = FlagsOptions.Output.Generate(output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListPVC, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)

	listCommon.FlagOutputType(List, &FlagsOptions.Output)
	List.Flags().StringVarP(&FlagsOptions.Project, consts.ProjectStr, consts.ProjectShort, "", consts.ProjectAttributeExplanation)
	_ = List.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
}

func getList(flagOptions ListOptions) (*pvc.Output, error) {
	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	clusterId := instance.GetConfigValue(config.ClusterUUID)
	token := instance.GetToken()

	return pvc.List(txn, cpUrl, token, clusterId, flagOptions.Project)
}
