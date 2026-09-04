/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"fmt"
	"os"
	"strings"
)

// FileContentFlag is a flag that supports reading the
// content of a file if the flag value starts with '@'.
type FileContentFlag string

func (f *FileContentFlag) String() string {
	return string(*f)
}

func (f *FileContentFlag) Set(value string) error {
	if strings.HasPrefix(value, "@") {
		path := value[1:]
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		*f = FileContentFlag(content)
	} else {
		*f = FileContentFlag(value)
	}

	return nil
}
