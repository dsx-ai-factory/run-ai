/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference/distributed"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference/nim"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/inference/standard"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	cmdfactory "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmd/factory"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Cmd represents the inference command
var Cmd = &cobra.Command{
	Use:     "inference",
	Short:   consts.InferenceShortMsg,
	Args:    cobra.NoArgs,
	GroupID: consts.GroupWorkloadID,

	PersistentPreRunE: cmdfactory.WorkloadsPersistentPreRunE,
	PreRunE:           cmdfactory.WorkloadsPreRunE,
}

// InitInferenceCmd initializes the inference command structure with all subcommands.
// This is called both at package init time and by tests that need to reset the command.
func InitInferenceCmd() {
	// Add groups (only if not already added)
	hasInferenceGroup := false
	for _, g := range Cmd.Groups() {
		if g.ID == consts.GroupInferenceID {
			hasInferenceGroup = true
			break
		}
	}

	if !hasInferenceGroup {
		Cmd.AddGroup(
			&cobra.Group{
				ID:    consts.GroupInferenceID,
				Title: consts.GroupInferenceTitle,
			},
			&cobra.Group{
				ID:    consts.GroupInferenceBaseID,
				Title: consts.GroupInferenceBaseTitle,
			},
			&cobra.Group{
				ID:    consts.GroupInferenceStdID,
				Title: consts.GroupInferenceStdTitle,
			},
		)
	}

	// Initialize standard inference commands (shortcuts at inference level)
	cmdfactory.WorkloadsCmdInit(Cmd, api.Knative, api.Inference)
	for _, c := range Cmd.Commands() {
		c.GroupID = consts.GroupInferenceStdID
	}

	// Remove standard list command from inference (we'll add custom one)
	for _, c := range Cmd.Commands() {
		if c.Use == "list" {
			Cmd.RemoveCommand(c)
		}
	}

	// Add custom list command that shows all inference architectures
	Cmd.AddCommand(list)

	Cmd.AddCommand(standard.Cmd)
	Cmd.AddCommand(distributed.Cmd)
	Cmd.AddCommand(nim.Cmd)
	Cmd.AddCommand(newRecommendCommand())
}

func init() {
	InitInferenceCmd()
}
