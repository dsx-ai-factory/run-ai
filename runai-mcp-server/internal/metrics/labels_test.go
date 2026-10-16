/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolLabel_KnownTool(t *testing.T) {
	known := map[string]bool{"list_workloads": true, "get_project": true}
	assert.Equal(t, "list_workloads", toolLabel(&mcp.CallToolParamsRaw{Name: "list_workloads"}, known))
	assert.Equal(t, "get_project", toolLabel(&mcp.CallToolParamsRaw{Name: "get_project"}, known))
}

func TestToolLabel_UnknownTool(t *testing.T) {
	known := map[string]bool{"list_workloads": true}
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParamsRaw{Name: "injected_tool"}, known))
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParamsRaw{Name: ""}, known))
}

func TestToolLabel_NilParams(t *testing.T) {
	known := map[string]bool{"list_workloads": true}
	assert.Equal(t, "unknown", toolLabel(nil, known))
}

func TestToolLabel_CallToolParams(t *testing.T) {
	known := map[string]bool{"submit_job": true}
	assert.Equal(t, "submit_job", toolLabel(&mcp.CallToolParams{Name: "submit_job"}, known))
	assert.Equal(t, "unknown", toolLabel(&mcp.CallToolParams{Name: "other"}, known))
}

func TestNormalizeLabel_ShortString(t *testing.T) {
	s := "initialize"
	assert.Equal(t, s, normalizeLabel(s))
}

func TestNormalizeLabel_ExactlyMaxLen(t *testing.T) {
	s := strings.Repeat("a", maxLabelLen)
	assert.Equal(t, s, normalizeLabel(s))
}

func TestNormalizeLabel_LongString(t *testing.T) {
	s := strings.Repeat("x", maxLabelLen+1)
	got := normalizeLabel(s)
	assert.Len(t, got, 57) // 48 prefix + 1 tilde + 8 hex chars
	assert.Contains(t, got, "~")
}

func TestNormalizeLabel_TwoLongStringsAreDistinct(t *testing.T) {
	a := strings.Repeat("a", maxLabelLen+1)
	b := strings.Repeat("b", maxLabelLen+1)
	assert.NotEqual(t, normalizeLabel(a), normalizeLabel(b))
}

// This method name hashes to 0x0eb6abe8, so its hexadecimal form is only seven
// characters and slicing it to eight used to panic.
func TestNormalizeLabel_HashWithLeadingZero(t *testing.T) {
	s := "tools/call/" + strings.Repeat("a", 54) + "620"
	require.Greater(t, len(s), maxLabelLen)
	require.Less(t, fnv1a(s), uint32(0x10000000), "the test input must hash below 0x10000000")

	got := normalizeLabel(s)
	assert.Len(t, got, 57) // 48 prefix + 1 tilde + 8 hex chars
	assert.Equal(t, "0eb6abe8", strings.TrimPrefix(got[48:], "~"))
}

func TestNormalizeLabel_AlwaysProducesAFixedWidthHash(t *testing.T) {
	for i := 0; i < 1000; i++ {
		s := strings.Repeat("a", 54) + fmt.Sprint(i) + strings.Repeat("b", 20)
		got := normalizeLabel(s)
		assert.Len(t, got, 57, "input %q", s)
	}
}
