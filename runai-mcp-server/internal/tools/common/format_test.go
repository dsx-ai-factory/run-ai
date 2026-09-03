/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func TestRoundTo(t *testing.T) {
	cases := []struct {
		name     string
		v        float64
		decimals int
		want     float64
	}{
		{"3dp_round_up", 1.2345, 3, 1.235},
		{"3dp_round_down", 1.2344, 3, 1.234},
		{"3dp_negative_half_away_from_zero", -1.2345, 3, -1.235},
		{"3dp_zero", 0, 3, 0},
		{"3dp_integer_unchanged", 42, 3, 42},
		{"3dp_tiny_half_up", 0.0005, 3, 0.001},
		{"3dp_tiny_below_half_down", 0.0004, 3, 0},
		{"0dp_round_to_int", 1.6, 0, 2},
		{"0dp_round_half_away", 2.5, 0, 3},
		// 1.005 would be a classic test, but 1.005 isn't exactly representable in
		// binary (it's actually 1.00499…), so RoundTo(1.005, 2) = 1, not 1.01. Use
		// values without that float quirk.
		{"2dp_exact_quarter", 0.125, 2, 0.13},
		{"2dp_negative", -1.27, 1, -1.3},
		{"negative_decimals_clamped_to_zero", 1.6, -1, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, common.RoundTo(tc.v, tc.decimals))
		})
	}
	// NaN/Inf pass through (callers gate on finite checks before rounding).
	assert.True(t, math.IsNaN(common.RoundTo(math.NaN(), 3)))
	assert.True(t, math.IsInf(common.RoundTo(math.Inf(1), 3), 1))
}
