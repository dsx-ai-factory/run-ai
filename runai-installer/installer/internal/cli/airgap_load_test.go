/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAirgapLoadExecutionTimeoutDefault(t *testing.T) {
	flag := newAirgapLoadCmd().Flags().Lookup("execution-timeout")

	require.NotNil(t, flag)
	assert.Equal(t, "3h0m0s", flag.DefValue)
}
