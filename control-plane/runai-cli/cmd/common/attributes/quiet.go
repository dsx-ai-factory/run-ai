/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attributes

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	FilePermission = 0755
)

func EnableQuietMode(cmd *cobra.Command) error {
	zap.L().Debug("quiet mode enabled")
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, FilePermission)
	if err != nil {
		zap.L().Error(fmt.Sprintf("failed to find %s", os.DevNull), zap.Error(err))
		return err
	}
	cmd.SetOut(devNull)
	return nil
}
