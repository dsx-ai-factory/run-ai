/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

import (
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	templateclient "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/client"
	templatedescribe "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type DescribeOptions struct {
	Output  string
	Project string
}

var (
	describeOptions DescribeOptions
)

// Describe represents the template describe command
var Describe = &cobra.Command{
	Use:     "describe [TEMPLATE_NAME]",
	Short:   DescribeShortMsg,
	Args:    cobra.ExactArgs(1),
	Example: DescribeExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		return DescribeProjectValidatorPre(cmd, &describeOptions)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return RunTemplateDescribeCMD(cmd, args, describeOptions, txn)
	},
}

// DescribeProjectValidatorPre validates project requirement before running describe command
func DescribeProjectValidatorPre(cmd *cobra.Command, options *DescribeOptions) error {
	// Get project from flag or config
	if options.Project == "" {
		options.Project = instance.GetConfigValue(config.ClusterProjectName)
	}

	// Validate that project is set
	validator := validators.NewProjectValidator(options.Project)
	if _, err := validator.Execute(); err != nil {
		wrappedErr := fmt.Errorf("project is missing. %w", err)
		if cmd != nil { // Production
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeTemplate, wrappedErr)
		}
		// Testing: return plain error when cmd is nil
		return wrappedErr
	}
	return nil
}

// RunTemplateDescribeCMD executes the template describe command
func RunTemplateDescribeCMD(cmd *cobra.Command, args []string, options DescribeOptions, txn *transaction.Txn) error {
	templateName, cpUrl, token, clusterId, projectId, err := DescribeInitializerCMD(args, options, txn)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeTemplate, err)
	}

	service, err := getTemplateDescribeService(cpUrl, token, clusterId)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeTemplate, fmt.Errorf("client error: %w", err))
	}

	// Get template (returns highest priority template based on scope hierarchy)
	tmpl, err := service.GetByName(txn.GetCtx(), templateName, projectId)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeTemplate, err)
	}

	displayer := templatedescribe.NewTemplateDisplayer(tmpl)
	err = RenderDescription(options, displayer)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDescribeTemplate, err)
	}

	return nil
}

func getTemplateDescribeService(domain, token, clusterId string) (*templatedescribe.Service, error) {
	templateClient, err := templateclient.NewClient(domain, token)
	if err != nil {
		return nil, err
	}

	return templatedescribe.NewService(templateClient, clusterId), nil
}

// DescribeInitializerCMD initializes parameters for template describe command
// Note: Project validation is done in PreRunE, so options.Project is guaranteed to be set
func DescribeInitializerCMD(args []string, options DescribeOptions, txn *transaction.Txn) (string, string, string, string, string, error) {
	templateName := strings.TrimSpace(args[0])
	if templateName == "" {
		return "", "", "", "", "", fmt.Errorf("template name cannot be empty")
	}

	cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
	token := instance.GetToken()
	clusterId := instance.GetConfigValue(config.ClusterUUID)

	// Project is guaranteed to be set by PreRunE validation
	// Get project ID for the filter (IncludeAncestors requires projectId, not projectName)
	projectId, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, options.Project)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get project ID for '%s': %w", options.Project, err)
	}

	return templateName, cpUrl, token, clusterId, projectId, nil
}

// RenderDescription renders template description based on output format
func RenderDescription(options DescribeOptions, displayer templatedescribe.Displayer) error {
	var err error

	switch options.Output {
	case "table":
		err = displayer.AsTable()
	case "json":
		err = displayer.AsJSON()
	case "yaml":
		err = displayer.AsYAML()
	default:
		err = fmt.Errorf("unsupported output format: %s", options.Output)
	}

	return err
}

func init() {
	// Register flags
	Describe.Flags().StringVarP(&describeOptions.Output, "output", "o", "table",
		consts.DescribeOutputUsage)
	Describe.Flags().StringVarP(&describeOptions.Project, consts.ProjectStr, consts.ProjectShort, "",
		ProjectFlagUsage)

	// Register auto-completion for flags
	_ = Describe.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
	_ = Describe.RegisterFlagCompletionFunc(consts.ProjectStr, project.GetProjectNamesForFlag)
}
