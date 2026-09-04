/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/credentials"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	configAtt "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config/attributes"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/choises"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/file"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/inputs"
	"github.com/spf13/cobra"
)

type CredentialArgs struct {
	ClientId   string
	Secret     string
	SecretFile string
	SaveCreds  bool
}

type accessKeyArgs = CredentialArgs

const (
	attributeAccessKeyClientId = "client-id"
	attributeAccessKeySecret   = "secret"
	attributeAccessKeyFile     = "secret-file"
	attributeAccessKeySave     = "save"

	accessKeyClientIDUsage       = "client ID for service account or user"
	accessKeySecretUsage         = "access key secret" //nolint:gosec
	accessKeySecretFromFileUsage = "use access key secret from file"
	accessKeySecretSaveCreds     = "save access key credentials in config file"

	exampleLoginAccessKey = `  
  # Login interactive using access key
  runai login access-key

  # Login using access key credentials
  runai login access-key --client-id=<client_id> --secret=<secret> --interactive=disabled

  # Login and save access key credentials
  runai login access-key --client-id=<client_id> --secret=<secret> --interactive=disabled --save
`
)

var SharedFieldMap = []textinput.Model{
	{
		Placeholder: credentials.InteractiveAppName,
		CharLimit:   charLimitation,
	},
	{
		Placeholder:   credentials.InteractiveAppSecret,
		EchoMode:      textinput.EchoPassword,
		EchoCharacter: '*',
		CharLimit:     charLimitation,
	},
}

var (
	accessKeyInteractive configAtt.EnableFlag
	accessKeyCredential  credentials.Type
	accessKeyArgsCmd     accessKeyArgs
)

var accessKey = &cobra.Command{
	Use:     "access-key",
	Args:    cobra.NoArgs,
	Short:   "login as a service account or a user using an access key",
	Example: exampleLoginAccessKey,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		return AccessKeyPreRunE(cmd, args, &accessKeyInteractive, &accessKeyCredential, &accessKeyArgsCmd)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		return AccessKeyRunE(cmd, args, accessKeyInteractive, accessKeyCredential, accessKeyArgsCmd, useConfiguredCredsAccessKey)
	},
}

func useConfiguredCredsAccessKey(cmd *cobra.Command) error {
	return UseConfiguredAccessKeyCredentials(cmd, &accessKeyCredential)
}

func AccessKeyPreRunE(
	cmd *cobra.Command,
	args []string,
	interactive *configAtt.EnableFlag,
	credential *credentials.Type,
	argsCommand *CredentialArgs,
) error {
	var err error

	options := instance.GetOptions()
	if cmdHelper.CheckInteractive(txn, *interactive, options.Interactive) {
		*interactive = configAtt.EnableTrue
		return nil
	} else {
		*interactive = configAtt.EnableFalse
	}

	// no args, and no config will fail the access key login
	if cap(args) == 0 {
		*credential, err = instance.LoadCredentials()
		if err != nil {
			cmd.PrintErrf(consts.LoginAccessKeyMissingAttributes)
			return cmdtxn.CmdErrorSetMissingParams(cmd)
		}
		return nil
	}

	if argsCommand.ClientId == "" {
		cmd.PrintErrf(consts.LoginAccessKeyMissingAttributes)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	}

	err = attributes.MutuallyExclusiveAtLeastOne(argsCommand.SecretFile, argsCommand.Secret)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorLoginAccessKeySecretRequired, err)
	}

	if argsCommand.Secret != "" {
		*credential = credentials.NewAccessKeyAuth(argsCommand.ClientId, argsCommand.Secret)
	} else if argsCommand.SecretFile != "" {
		secret, err := auth.ReadApplicationSecret(argsCommand.SecretFile)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, auth.AuthenticationSecretFileReadError, err)
		}
		*credential = credentials.NewAccessKeyAuth(argsCommand.ClientId, secret)
	}
	return nil
}

func AccessKeyRunE(
	cmd *cobra.Command,
	args []string,
	interactive configAtt.EnableFlag,
	credential credentials.Type,
	argsCommand CredentialArgs,
	useConfiguredCredsFunc func(*cobra.Command) error,
) error {
	if interactive.IsEnabled() {
		err := useConfiguredCredsFunc(cmd)
		if err != nil {
			return err
		}
		return nil
	}

	msg, err := auth.ApplicationLogin(txn, credential.GetID(), credential.GetSecret(), instance.GetConfigValue(config.AuthenticationUrl), authFilePath)
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, msg, err)
	}
	cmd.Println(msg)

	if argsCommand.SaveCreds {
		err := instance.SetCredentials(credential)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.LoginAccessKeyCredentialSetError, err)
		}
	}
	return nil
}

func InteractiveAccessKeyLogin() (string, error) {
	var (
		msg      string
		err      error
		option   string
		response *map[string]string
	)

	option, err = choises.ExecuteTea(consts.AccessKeyMethodSelection, "", auth.AppOption)
	if err != nil {
		return msg, err
	}

	switch option {
	case auth.FilePath:
		err = file.Picker()
		if err != nil {
			return msg, err
		}
	case auth.ApplicationKey:
		response, err = inputs.MultiTexts(SharedFieldMap, "")
		if err != nil {
			return msg, err
		}
		appKey := credentials.NewAccessKeyAuth((*response)[credentials.InteractiveAppName], (*response)[credentials.InteractiveAppSecret])
		msg, err = auth.ApplicationLogin(txn, appKey.GetID(), appKey.GetSecret(), instance.GetConfigValue(config.AuthenticationUrl), authFilePath)
		if err != nil {
			return msg, err
		}
		err = instance.SetCredentials(appKey)
		if err != nil {
			return msg, err
		}
	}
	return fmt.Sprintf("%s\n%s", consts.AccessKeySuccessfullySet, msg), nil
}

func UseConfiguredAccessKeyCredentials(cmd *cobra.Command, credential *credentials.Type) error {
	var loginOpt []string
	var err error
	var msg string

	authType := instance.GetConfigValue(config.CredentialsType)
	if authType != "" {
		displayName := credentials.GetDisplayNameFromString(authType)
		loginOpt = append(loginOpt, fmt.Sprintf(consts.ChooseLoginViaByExistingLoginValues, displayName))
		loginOpt = append(loginOpt, consts.ChooseReplaceExistingLoginValues)
		authType, err = choises.ExecuteTea(consts.ChooseUsingExistingLoginSubject, "", loginOpt)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ChooseErrorExitingWithoutSelecting, err)
		}

		if strings.HasPrefix(authType, "Use configured credentials:") {
			*credential, err = instance.LoadCredentials()
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ChooseErrorExitingCredLoadFailed, err)
			}

			msg, err = auth.ApplicationLogin(txn, (*credential).GetID(), (*credential).GetSecret(), instance.GetConfigValue(config.AuthenticationUrl), authFilePath)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, msg, err)
			}
			cmd.Println(msg)
			return nil
		}
	}

	msg, err = InteractiveAccessKeyLogin()
	if err != nil {
		return cmdtxn.CmdErrorPrintAndSet(cmd, msg, err)
	}

	cmd.Println(msg)
	return nil
}

func init() {
	Cmd.AddCommand(accessKey)

	accessKey.Flags().Var(&accessKeyInteractive, consts.Interactive, consts.InteractiveArgumentUsage)
	_ = accessKey.RegisterFlagCompletionFunc(consts.Interactive, configAtt.EnableFlagCompletion)

	accessKey.Flags().StringVar(&accessKeyArgsCmd.ClientId, attributeAccessKeyClientId, "", accessKeyClientIDUsage)
	accessKey.Flags().StringVar(&accessKeyArgsCmd.Secret, attributeAccessKeySecret, "", accessKeySecretUsage)
	accessKey.Flags().StringVar(&accessKeyArgsCmd.SecretFile, attributeAccessKeyFile, "", accessKeySecretFromFileUsage)
	accessKey.Flags().BoolVar(&accessKeyArgsCmd.SaveCreds, attributeAccessKeySave, false, accessKeySecretSaveCreds)
	accessKey.MarkFlagsMutuallyExclusive(attributeAccessKeySecret, attributeAccessKeyFile)
}
