/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package linetail

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_KeepsLastNLines(t *testing.T) {
	var got [][]string
	w := New(3, func(l []string) { got = append(got, l) })

	// A write with no complete line yet reports nothing.
	_, _ = w.Write([]byte("partial"))
	assert.Empty(t, got, "no report until a full line is seen")

	// Completing the line plus more, split across writes, with \r\n handled.
	_, _ = w.Write([]byte("-done\r\nb\nc\nd\ne\n"))
	require.NotEmpty(t, got)
	assert.Equal(t, []string{"c", "d", "e"}, got[len(got)-1], "only the last n lines are kept")
}
