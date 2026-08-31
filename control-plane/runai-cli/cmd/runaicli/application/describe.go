/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

import (
	"fmt"

	applicationservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application"
	appdescribe "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/describe"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/helper/validators"

	projectCmd "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	DescribeShortMsg              = "Describe an AI application"
	FailedToDescribeAIApplication = "application describe failed"
	DescribeExample               = `# Describe an AI application
runai application describe my-grafana -p team-alpha

# Describe in JSON format
runai application describe my-grafana -p team-alpha -o json`
)

type describeOptions struct {
	Project string
	Output  string
}

var describeFlags describeOptions

var Describe = &cobra.Command{
	Use:               "describe APPLICATION_NAME",
	Aliases:           []string{"desc"},
	Short:             DescribeShortMsg,
	Args:              cobra.ExactArgs(1),
	Example:           DescribeExample,
	ValidArgsFunction: GetAIApplicationNamesForArg,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if describeFlags.Project == "" {
			describeFlags.Project = instance.GetConfigValue(config.ClusterProjectName)
		}
		cmdHelper.AddValidation(validators.NewProjectValidator(describeFlags.Project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		token := instance.GetToken()

		if _, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, describeFlags.Project); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeAIApplication, err)
		}

		app, err := applicationservice.GetAIApplicationByName(txn, cpUrl, token, clusterId, describeFlags.Project, appName)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeAIApplication, err)
		}
		if app == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeAIApplication,
				fmt.Errorf(ErrAppNotFound, appName, describeFlags.Project))
		}

		if err = appdescribe.RenderOutput(cmd.OutOrStdout(), app, describeFlags.Output); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeAIApplication, err)
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(Describe)

	Describe.Flags().StringVarP(&describeFlags.Project, consts.ProjectStr, consts.ProjectShort, "", ProjectFlagUsage)
	Describe.Flags().StringVarP(&describeFlags.Output, "output", "o", "table", consts.DescribeOutputUsage)
	_ = Describe.RegisterFlagCompletionFunc(consts.ProjectStr, projectCmd.GetProjectNamesForFlag)
	_ = Describe.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
}
