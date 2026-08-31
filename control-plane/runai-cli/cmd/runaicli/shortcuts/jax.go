/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package shortcuts

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/training/jax"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

var Jax = &cobra.Command{
	Use:        jax.Cmd.Use,
	Short:      fmt.Sprintf("alias for %s", jax.Cmd.Short),
	Aliases:    jax.Cmd.Aliases,
	Args:       jax.Cmd.Args,
	GroupID:    consts.GroupWorkloadID,
	Deprecated: jax.Cmd.Deprecated,

	PersistentPreRunE: jax.Cmd.PersistentPreRunE,
	PreRunE:           jax.Cmd.PersistentPreRunE,
}

func init() {
	jax.Init(Jax)
}
