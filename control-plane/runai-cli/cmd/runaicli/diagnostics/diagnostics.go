/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package diagnostics

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/diagnostics"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     diagnostics.DiagnosticsUse,
	Short:   diagnostics.DiagnosticsShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupConfigID,
	Aliases: []string{diagnostics.DiagnosticsAlias},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmd.HelpFunc()(cmd, args)
		return cmdtxn.CmdErrorSetMissingParams(cmd)
	},
}
