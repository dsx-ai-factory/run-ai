/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/inputs"
	"github.com/spf13/cobra"
)

const (
	userName      = "user"
	userNameShort = "u"
	passwordArg   = "password"
	passwordShort = "p"
	userTimerName = "user_login"
	passwordKey   = "Password"
	passwordTitle = "Please enter your password"
)

var (
	username string
	password string
)

var passwordInput = []textinput.Model{
	{
		Placeholder:   passwordKey,
		EchoMode:      textinput.EchoPassword,
		EchoCharacter: '*',
		CharLimit:     charLimitation,
	},
}

// application represents the base command when called without any subcommands
var user = &cobra.Command{
	Use:     "user",
	Short:   consts.LoginUserShortMsg,
	Long:    consts.LoginUserLongMsg,
	Args:    cobra.NoArgs,
	Example: consts.ExampleLoginUser,

	RunE: func(cmd *cobra.Command, args []string) error {
		loginClient, err := auth.NewLoginClient(instance.GetConfigValue(config.AuthenticationUrl))
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
		}
		if password == "" {
			response, err := inputs.MultiTexts(passwordInput, passwordTitle)
			if err != nil {
				return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
			}
			password = (*response)[passwordKey]
		}
		if password == "" {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, errors.New("password is empty"))
		}
		authenticator := auth.NewLocalUser(username, password, loginClient.TokensAPI)

		err = auth.Login(txn, userTimerName, authenticator, authFilePath)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
		}
		cmd.Println(consts.SuccessfullyLoggedIn)
		return nil
	},
}

func init() {
	Cmd.AddCommand(user)
	user.Flags().StringVarP(&username, userName, userNameShort, "", consts.LoginUserUsernameUsage)
	user.Flags().StringVarP(&password, passwordArg, passwordShort, "", consts.LoginUserPasswordUsage)
	_ = user.MarkFlagRequired(userName)
}
