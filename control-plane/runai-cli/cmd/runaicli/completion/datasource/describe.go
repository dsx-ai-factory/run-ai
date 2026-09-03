/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource

import (
	"github.com/spf13/cobra"
)

// OutputFlagCompletion provides autocompletion for the output flag
func OutputFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"table", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
}
