/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logout

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/token"
	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/io"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/spf13/cobra"
)

var (
	instance = config.GetInstance()
)

// Cmd represents the logs command
var Cmd = &cobra.Command{
	Use:     "logout",
	Short:   ShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupUserManagementID,

	RunE: func(cmd *cobra.Command, args []string) error {

		logger := instance.GetLogger()
		defer func() {
			deleteErr := io.CheckExistAndDeleteFile(instance.GetAuthenticationPath())
			if deleteErr != nil {
				logger.Debug("failed to delete authentication file", zap.Error(deleteErr))
			}
		}()
		tokenStr, err := auth.GetToken(instance.GetTxn().GetCtx())
		if err != nil {
			// lie to the user, we don't want to print the error
			logger.Debug("failed to get token assume logout", zap.Error(err))
			cmd.Println("No active session found. You are already logged out.")
			return nil
		}

		loginClient, err := auth.NewLoginClient(instance.GetConfigValue(config.AuthenticationUrl))
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedError, err)
		}
		tokenRevoker := token.NewControlPlaneTokenRevoker(loginClient)
		err = tokenRevoker.Revoke(instance.GetTxn().GetCtx(), tokenStr)
		if err != nil && !errors.Is(err, token.ErrTokenNotFound) {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedError, err)
		}

		// Token was revoked, clear the credentials
		errClearCred := instance.ClearCredentials()
		errDeleteAuth := io.CheckExistAndDeleteFile(instance.GetAuthenticationPath())

		if errClearCred != nil {
			logger.Debug("failed to clear credentials from config file", zap.Error(errClearCred))
		}
		if errDeleteAuth != nil {
			logger.Debug("failed to delete authentication file", zap.Error(errDeleteAuth))
		}

		if err == token.ErrTokenNotFound {
			cmd.Println(SuccessMsgTokenNotFound)
			return nil
		}
		cmd.Println(SuccessMsg)
		return nil
	},
}

func init() {
}
