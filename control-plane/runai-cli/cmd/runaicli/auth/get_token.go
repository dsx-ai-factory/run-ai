/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	instance     = config.GetInstance()
	outputFormat string
)

var getToken = &cobra.Command{
	Use:   GetTokenUse,
	Short: GetTokenShort,
	Args:  cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		txn := instance.GetTxn()
		tokenDisplay, err := auth.GetTokenDisplay(txn.GetCtx(), outputFormat)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedGetTokenMsg, err)
		}

		cmd.Printf("%s\n", tokenDisplay)
		return nil
	},
}

func init() {
	Cmd.AddCommand(getToken)

	getToken.Flags().StringVarP(&outputFormat, FlagNameOutput, FlagShortOutput, FlagDefaultOutput, FlagHelpOutput)
	_ = getToken.RegisterFlagCompletionFunc(FlagNameOutput, formatFlagCompletion)
}

func formatFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{auth.PlaintextFormat, auth.KubeconfigFormat}, cobra.ShellCompDirectiveNoFileComp
}
