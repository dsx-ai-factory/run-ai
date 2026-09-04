/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	"fmt"
	"strings"

	argshandler "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"
	mycredCompletion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/completion/mycredential"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	mycreddelete "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/delete"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var deleteType string

var DeleteCmd = &cobra.Command{
	Use:     DeleteUse,
	Short:   DeleteShortMsg,
	Long:    DeleteLong,
	Example: DeleteExample,
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(argshandler.GetNameFromArgs(&args))
		if deleteType == "" {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteCredential, fmt.Errorf("required flag(s) %q not set", TypeFlag))
		}
		err := mycreddelete.Delete(instance.GetTxn(), instance.GetConfigValue(config.ControlPlaneUrl), instance.GetToken(), strings.TrimSpace(deleteType), name)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToDeleteCredential, err)
		}
		cmd.Printf("Personal credential %q deleted.\n", name)
		return nil
	},
}

func init() {
	Cmd.AddCommand(DeleteCmd)
	DeleteCmd.Flags().StringVar(&deleteType, TypeFlag, "", TypeFlagUsageCreate)
	DeleteCmd.MarkFlagRequired(TypeFlag)
	_ = DeleteCmd.RegisterFlagCompletionFunc(TypeFlag, mycredCompletion.TypeFlagCompletion)
}
