/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLegacyWorkloadCleanupFlag(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() { os.Args = originalArgs })
	os.Args = []string{"go-operator", "--legacy-workload-cleanup"}

	opts := &Options{}
	opts.parseCommandLineArgs(flag.NewFlagSet("test", flag.ContinueOnError))

	require.True(t, opts.LegacyWorkloadCleanup)
}
