/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource

import (
	"fmt"
	"strings"

	dsCompletion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/datasource"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	datasourcedescribe "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeOptions struct {
	Output  string
	Project string
	Type    string
}

var (
	describeOptions DescribeOptions
)

var Describe = &cobra.Command{
	Use:     "describe [DATASOURCE_NAME] --type TYPE",
	Aliases: []string{"desc"},
	Short:   DescribeShortMsg,
	Args:    cobra.ExactArgs(1),
	Example: DescribeExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		return DescribeProjectValidatorPre(cmd, &describeOptions)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return RunDatasourceDescribeCMD(cmd, args, describeOptions, txn)
	},
}

func DescribeProjectValidatorPre(cmd *cobra.Command, options *DescribeOptions) error {
	if options.Project == "" {
		options.Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	validator := validators.NewProjectValidator(options.Project)
	if _, err := validator.Execute(); err != nil {
		wrappedErr := fmt.Errorf("project is missing. %w", err)
		if cmd != nil { // Production
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeDatasource, wrappedErr)
		}
		// Testing: return plain error when cmd is nil
		return wrappedErr
	}
	return nil
}

func RunDatasourceDescribeCMD(cmd *cobra.Command, args []string, options DescribeOptions, txn *transaction.Txn) error {
	datasourceName, cpUrl, token, clusterId, projectName, err := DescribeInitializerCMD(args, options)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeDatasource, err)
	}

	datasource, err := datasourcedescribe.Describe(txn, cpUrl, token, datasourceName, projectName, clusterId, options.Type)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeDatasource, err)
	}

	if err := datasourcedescribe.RenderOutput(datasource, options.Output); err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeDatasource, err)
	}

	return nil
}

func DescribeInitializerCMD(args []string, options DescribeOptions) (string, string, string, string, string, error) {
	datasourceName := strings.TrimSpace(args[0])
	if datasourceName == "" {
		return "", "", "", "", "", fmt.Errorf("datasource name cannot be empty")
	}

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()
	clusterId := instance.GetConfigValue(config.ClusterUUID)

	projectName := options.Project

	return datasourceName, cpUrl, token, clusterId, projectName, nil
}

func init() {
	Cmd.AddCommand(Describe)

	// Register flags
	Describe.Flags().StringVarP(&describeOptions.Output, "output", "o", "table",
		consts.DescribeOutputUsage)
	Describe.Flags().StringVarP(&describeOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)
	Describe.Flags().StringVar(&describeOptions.Type, TypeFlag, "", TypeFlagUsage)

	// Mark required flags
	_ = Describe.MarkFlagRequired(TypeFlag)

	// Register auto-completion for flags
	_ = Describe.RegisterFlagCompletionFunc("output", dsCompletion.OutputFlagCompletion)
	_ = Describe.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
	_ = Describe.RegisterFlagCompletionFunc(TypeFlag, dsCompletion.TypeFlagCompletion)
}
