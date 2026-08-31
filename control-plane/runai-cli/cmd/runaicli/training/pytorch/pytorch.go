/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pytorch

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/run-ai/runai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:     "pytorch",
	Short:   consts.PytorchShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupTrainingID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.PyTorch, api.Distributed)
}

func init() {
	Init(Cmd)
}
