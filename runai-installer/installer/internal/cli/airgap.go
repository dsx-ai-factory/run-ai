/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import "github.com/spf13/cobra"

func newAirgapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airgap",
		Short: "Air-gapped bundle tooling",
	}
	cmd.AddCommand(newAirgapPackCmd())
	cmd.AddCommand(newAirgapLoadCmd())
	return cmd
}
