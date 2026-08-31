/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	ssoTimerName = "sso_login"
)

// set represents the set credentials command
var sso = &cobra.Command{
	Use:     "sso",
	Aliases: []string{"remote-browser"},
	Args:    cobra.NoArgs,
	Short:   consts.LoginSSOShortMsg,

	RunE: func(cmd *cobra.Command, args []string) error {
		authenticator, err := auth.InitRemoteBrowserLogin(instance.GetConfigValue(config.AuthenticationUrl))

		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
		}

		err = auth.Login(txn, ssoTimerName, authenticator, authFilePath)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToAuthenticate, err)
		}
		cmd.Println(consts.SuccessfullyLoggedIn)
		return nil
	},
}

func init() {
	Cmd.AddCommand(sso)
}
