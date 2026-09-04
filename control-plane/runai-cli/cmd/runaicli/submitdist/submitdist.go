/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submitdist

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/submitdistbc"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:        "submit-dist [TYPE]",
	Short:      consts.SubmitDistShortMsg,
	GroupID:    consts.GroupDeprecatedID,
	Deprecated: consts.DepSubmitDistMsg,
	Args:       cobra.NoArgs,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func distributedWorkloadsCmdInit(cmd *cobra.Command, framework api.Framework) {
	submitFactory := submitdistbc.NewCommandFactory(config.GetInstance(), helper.Cmd{})
	if !submitFactory.IsSupported(framework, api.Distributed) {
		return
	}

	submitCommand, err := submitFactory.GetCommand(framework, api.Distributed)
	if err != nil {
		return
	}

	err = submitCommand.RegisterFlags()
	if err != nil {
		return
	}

	submitCmd := submitCommand.GetCommand()
	submitCmd.Use = framework.String()
	submitCmd.Short = "[Deprecated] " + submitCmd.Short
	submitCmd.Example = ""

	cmd.AddCommand(submitCmd)
}

func init() {
	distributedWorkloadsCmdInit(Cmd, api.MPI)
	distributedWorkloadsCmdInit(Cmd, api.XGBoost)
	distributedWorkloadsCmdInit(Cmd, api.TF)
	distributedWorkloadsCmdInit(Cmd, api.PyTorch)
	distributedWorkloadsCmdInit(Cmd, api.Jax)
}
