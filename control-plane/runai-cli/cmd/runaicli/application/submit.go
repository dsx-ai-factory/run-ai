/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

import (
	"fmt"

	applicationservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application"
	appsubmit "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/submit"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"

	projectCmd "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

type submitOptions struct {
	Project    string
	ChartURL   string
	Values     string
	Set        []string
	Credential string
}

var submitFlags submitOptions

var Submit = &cobra.Command{
	Use:     "submit APPLICATION_NAME",
	Short:   SubmitShortMsg,
	Args:    cobra.ExactArgs(1),
	Example: SubmitExample,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if submitFlags.Project == "" {
			submitFlags.Project = instance.GetConfigValue(config.ClusterProjectName)
		}
		cmdHelper.AddValidation(validators.NewProjectValidator(submitFlags.Project))

		if msg, err := cmdHelper.Check(); err != nil {
			return cmdtxn.CmdErrorValidationMessage(cmd, msg, err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		values, err := appsubmit.BuildValues(submitFlags.Values, submitFlags.Set)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToSubmitAIApplication, err)
		}

		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		clusterId := instance.GetConfigValue(config.ClusterUUID)
		token := instance.GetToken()

		projectId, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, submitFlags.Project)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToSubmitAIApplication, err)
		}

		var credentialAssetId string
		if submitFlags.Credential != "" {
			credentialAssetId, err = applicationservice.ResolveCredentialAssetId(
				txn, cpUrl, token, submitFlags.Credential, projectId, clusterId)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToSubmitAIApplication, err)
			}
		}

		request := appsubmit.BuildCreationRequest(appsubmit.CreationOptions{
			ClusterId:         clusterId,
			ProjectId:         projectId,
			Name:              appName,
			ChartURL:          submitFlags.ChartURL,
			Values:            values,
			CredentialAssetId: credentialAssetId,
		})

		if _, err = applicationservice.CreateAIApplication(txn, cpUrl, token, request); err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToSubmitAIApplication, err)
		}

		cmd.Println(fmt.Sprintf(SubmitSuccessMsg, appName, submitFlags.Project))
		return nil
	},
}

func init() {
	Cmd.AddCommand(Submit)

	Submit.Flags().StringVar(&submitFlags.ChartURL, attributeNameChartURL, "", ChartURLFlagUsage)
	Submit.Flags().StringVarP(&submitFlags.Project, consts.ProjectStr, consts.ProjectShort, "", ProjectFlagUsage)
	Submit.Flags().StringVar(&submitFlags.Values, attributeNameValues, "", ValuesFlagUsage)
	Submit.Flags().StringArrayVar(&submitFlags.Set, attributeNameSet, nil, SetFlagUsage)
	Submit.Flags().StringVar(&submitFlags.Credential, attributeNameCredential, "", CredentialFlagUsage)

	_ = Submit.MarkFlagRequired(attributeNameChartURL)
	_ = Submit.MarkFlagFilename(attributeNameValues, "yaml", "yml")
	_ = Submit.RegisterFlagCompletionFunc(consts.ProjectStr, projectCmd.GetProjectNamesForFlag)
}
