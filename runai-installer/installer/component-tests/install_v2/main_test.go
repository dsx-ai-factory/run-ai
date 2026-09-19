/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

// TestMain gives every connected install in this package a chart repository
// base, the way CI does for the installer. Tests that exercise the missing
// base override it with t.Setenv.
func TestMain(m *testing.M) {
	if os.Getenv(step.ChartRepoBaseEnv) == "" {
		os.Setenv(step.ChartRepoBaseEnv, "https://charts.example.com/api/helm")
	}
	os.Exit(m.Run())
}
