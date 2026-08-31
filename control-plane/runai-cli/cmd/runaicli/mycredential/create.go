/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	"fmt"
	"strings"

	argshandler "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/args"
	mycredCompletion "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/mycredential"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	mycredcreate "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/create"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var createType string
var createItems []string
var createDescription string

var CreateCmd = &cobra.Command{
	Use:     CreateUse,
	Short:   CreateShortMsg,
	Long:    CreateLong,
	Example: CreateExample,
	Args:    cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(argshandler.GetNameFromArgs(&args))
		if createType == "" {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToCreateCredentials, fmt.Errorf("required flag(s) %q not set", TypeFlag))
		}
		err := mycredcreate.Create(instance.GetTxn(), instance.GetConfigValue(config.ControlPlaneUrl), instance.GetToken(), strings.TrimSpace(createType), name, strings.TrimSpace(createDescription), createItems)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToCreateCredentials, err)
		}
		cmd.Printf("Credential %q created.\n", name)
		return nil
	},
}

func init() {
	Cmd.AddCommand(CreateCmd)
	CreateCmd.Flags().StringVar(&createType, TypeFlag, "", TypeFlagUsageCreate)
	CreateCmd.MarkFlagRequired(TypeFlag)
	CreateCmd.Flags().StringArrayVar(&createItems, ItemFlag, nil, ItemFlagUsage)
	CreateCmd.Flags().StringVar(&createDescription, DescriptionFlag, "", DescriptionFlagUsage)
	_ = CreateCmd.RegisterFlagCompletionFunc(TypeFlag, mycredCompletion.TypeFlagCompletion)
}
