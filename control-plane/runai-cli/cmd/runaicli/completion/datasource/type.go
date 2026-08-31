/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource

import (
	"github.com/spf13/cobra"
)

// TypeFlagCompletion provides autocompletion for the type flag
func TypeFlagCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"nfs", "pvc", "s3", "git", "hostPath", "configMap", "secretVolume", "dataVolume"}, cobra.ShellCompDirectiveNoFileComp
}
