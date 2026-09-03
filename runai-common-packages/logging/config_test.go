/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logging

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguration_FromEnv(t *testing.T) {
	os.Setenv("LOG_LEVEL", "error")
	config := readConfig()

	assert.Equal(t, config.LogLevel, "error")

	os.Clearenv()
}
