/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"2.18.0", "2.18.0", 0},
		{"2.18", "2.18.0", 0},
		{"v2.18", "2.18.0", 0},
		{"2.18.0", "2.19.0", -1},
		{"2.19.0", "2.18.0", 1},
		{"2.18.1", "2.18.0", 1},
		{"2.18.0-rc1", "2.18.0", 0},
		{"2.0.0", "1.99.99", 1},
	}
	for _, tc := range cases {
		got, err := compareVersions(tc.a, tc.b)
		assert.NoError(t, err, "compareVersions(%q, %q)", tc.a, tc.b)
		assert.Equal(t, tc.want, got, "compareVersions(%q, %q)", tc.a, tc.b)
	}
}

func TestCompareVersions_PropagatesParseError(t *testing.T) {
	_, err := compareVersions("abc", "1.0.0")
	assert.Error(t, err, "expected parse error for left operand")
	_, err = compareVersions("1.0.0", "abc")
	assert.Error(t, err, "expected parse error for right operand")
}
