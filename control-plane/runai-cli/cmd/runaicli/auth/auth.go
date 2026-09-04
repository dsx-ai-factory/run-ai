/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"github.com/spf13/cobra"
)

// Cmd represents the auth command group
var Cmd = &cobra.Command{
	Use:   AuthUse,
	Short: AuthShortMsg,
	Args:  cobra.NoArgs,
}

func init() {
}
