/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package whoami

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	instance = config.GetInstance()
)

const (
	claimError = "failed to get email and ID from token"
)

// Cmd represents the version command
var Cmd = &cobra.Command{
	Use:     "whoami",
	Short:   consts.WhoamiShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupUserManagementID,

	PreRunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		tokenStr, err := auth.GetToken(instance.GetTxn().GetCtx())
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, claimError, err)
		}

		claims, err := auth.GetClaims(tokenStr)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, claimError, err)
		}
		if claims == nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, claimError, fmt.Errorf("claims are empty"))
		}
		cmd.Println(fmt.Sprintf("User %s logged in with ID %s", claims.UserEmail, claims.UserID))
		return nil
	},
}
