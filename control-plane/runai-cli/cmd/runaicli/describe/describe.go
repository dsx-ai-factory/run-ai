/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/cobra"
)

// Deprecated: Use `runai workload describe` instead
var Cmd = &cobra.Command{
	Use:        "describe",
	Short:      consts.DescribeShortMsgDep,
	GroupID:    consts.GroupDeprecatedID,
	Deprecated: consts.DepDescribeMsg,
	// Do not add deprecated flag, it removes the command from help
}
