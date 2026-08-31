/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/run-ai/runai/runai-mcp-server/internal/tools/common"
)

func TestDiffMaps(t *testing.T) {
	before := map[string]any{"name": "a", "gpu": 4, "keep": true}
	after := map[string]any{"name": "b", "gpu": 4, "added": "x"}

	diff := common.DiffMaps(before, after)

	// changed
	assert.Equal(t, map[string]any{"before": "a", "after": "b"}, diff["name"])
	// removed (present only in before)
	assert.Equal(t, map[string]any{"before": true, "after": nil}, diff["keep"])
	// added (present only in after)
	assert.Equal(t, map[string]any{"before": nil, "after": "x"}, diff["added"])
	// unchanged → absent from diff
	assert.NotContains(t, diff, "gpu")
}

func TestDiffMaps_Empty(t *testing.T) {
	assert.Empty(t, common.DiffMaps(map[string]any{}, map[string]any{}))
	assert.Empty(t, common.DiffMaps(nil, nil))
}
