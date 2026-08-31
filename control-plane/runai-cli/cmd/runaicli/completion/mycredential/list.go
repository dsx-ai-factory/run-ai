/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	"github.com/spf13/cobra"
)

func TypeFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"genericSecret", "dockerRegistry", "ngcApiKey"}, cobra.ShellCompDirectiveNoSpace
}
