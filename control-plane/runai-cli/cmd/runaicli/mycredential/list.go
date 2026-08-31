/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	mycredCompletion "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/completion/mycredential"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cmdtxn"
	mycredlist "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/list"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var listType string

var List = &cobra.Command{
	Use:     "list",
	Short:   ListShortMsg,
	Aliases: []string{"ls"},
	Args:    cobra.NoArgs,
	Example: ListExample,

	RunE: func(cmd *cobra.Command, args []string) error {
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()

		handler, err := mycredlist.List(instance.GetTxn(), cpUrl, token, listType)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListCredentials, err)
		}

		printOut, err := handler.GenerateTable()
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, FailedToListCredentials, err)
		}

		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(List)
	List.Flags().StringVar(&listType, TypeFlag, "", TypeFlagUsage)
	_ = List.RegisterFlagCompletionFunc(TypeFlag, mycredCompletion.TypeFlagCompletion)
}
