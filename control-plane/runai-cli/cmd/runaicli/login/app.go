/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/credentials"
	configAtt "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/spf13/cobra"
)

type appArgs = CredentialArgs

const (
	attributeAppName     = "name"
	attributeAppClientId = "client-id"
	attributeAppSecret   = "secret"
	attributeSecretFile  = "secret-file"
	attributeSaveCreds   = "save"

	applicationClientIDUsage       = "application client ID"
	applicationSecretUsage         = "application secret"
	applicationSecretFromFileUsage = "use application secret from file"
	applicationSecretSaveCreds     = "save application credentials in config file"

	nameDeprecationMsg = "use --client-id instead."

	exampleLoginApplication = `  
  # Login interactive using application credentials
  runai login app

  # Login using application credentials
  login app --client-id=<app_client_id> --secret=<app_secret> --interactive=disabled

  # Login and Save application credentials
  login app --client-id=<app_client_id> --secret=<app_secret> --interactive=disabled --save
`
)

var (
	interactive configAtt.EnableFlag
	credential  credentials.Type
	argsCommand appArgs
)

// application represents the base command when called without any subcommands
// Deprecated: This command is deprecated and hidden. Use 'access-key' instead.
var application = &cobra.Command{
	Use:        "application",
	Aliases:    []string{"app"},
	Args:       cobra.NoArgs,
	Short:      "login as an application",
	Example:    exampleLoginApplication,
	Deprecated: "use 'runai login access-key' instead",
	Hidden:     true,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		return AccessKeyPreRunE(cmd, args, &interactive, &credential, &argsCommand)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return AccessKeyRunE(cmd, args, interactive, credential, argsCommand, useConfiguredCreds)
	},
}

func useConfiguredCreds(cmd *cobra.Command) error {
	return UseConfiguredAccessKeyCredentials(cmd, &credential)
}

func init() {
	Cmd.AddCommand(application)

	application.Flags().Var(&interactive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = application.RegisterFlagCompletionFunc(consts.Interactive, configAtt.EnableFlagCompletion)

	application.Flags().StringVar(&argsCommand.ClientId, attributeAppName, "", "")
	application.Flags().StringVar(&argsCommand.ClientId, attributeAppClientId, "", applicationClientIDUsage)
	application.Flags().StringVar(&argsCommand.Secret, attributeAppSecret, "", applicationSecretUsage)
	application.Flags().StringVar(&argsCommand.SecretFile, attributeSecretFile, "", applicationSecretFromFileUsage)
	application.Flags().BoolVar(&argsCommand.SaveCreds, attributeSaveCreds, false, applicationSecretSaveCreds)
	application.MarkFlagsMutuallyExclusive(attributeAppSecret, attributeSecretFile)

	_ = application.Flags().MarkDeprecated(attributeAppName, nameDeprecationMsg)
}
