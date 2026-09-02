/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jax

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

const (
	ShortMsg = "jax management"
)

// Cmd represents the cluster command
var Cmd = &cobra.Command{
	Use:        "jax",
	Short:      ShortMsg,
	Args:       cobra.NoArgs,
	GroupID:    consts.GroupTrainingID,
	Deprecated: consts.DepFrameworkMsg,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

func Init(cmd *cobra.Command) {
	cmdfactory.WorkloadsCmdInit(cmd, api.Jax, api.Distributed)
}

func init() {
	Init(Cmd)
}
