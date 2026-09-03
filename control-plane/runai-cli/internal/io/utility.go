/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package io

import (
	"fmt"
	"os"
)

func CheckExistAndDeleteFile(filePath string) error {
	stats, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if stats.IsDir() {
		return fmt.Errorf("path is a directory (%s)", filePath)
	}

	return os.Remove(filePath)
}
