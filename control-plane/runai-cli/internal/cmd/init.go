/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cmd

import "github.com/spf13/cobra"

func NewEmptyCobraCommand() *cobra.Command {
	return &cobra.Command{}
}
