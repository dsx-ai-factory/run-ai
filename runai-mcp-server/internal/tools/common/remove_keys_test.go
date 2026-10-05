/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

func TestRemoveKeysReturnsIndependentRecursiveCopy(t *testing.T) {
	original := map[string]any{
		"name":            "template",
		"AWSSessionToken": "top-secret",
		"nested": map[string]any{
			"password": "nested-secret",
			"visible":  "original",
		},
		"items": []any{map[string]any{
			"awsSessionToken": "slice-secret",
			"visible":         "original",
		}},
		"typedItems": []map[string]any{{
			"AwsSessionToken": nil,
			"visible":         "original",
		}},
	}

	clean := common.RemoveKeys(original, []string{"awsSessionToken", "password"}).(map[string]any)

	assert.NotContains(t, clean, "AWSSessionToken")
	cleanNested := clean["nested"].(map[string]any)
	assert.NotContains(t, cleanNested, "password")
	cleanItems := clean["items"].([]any)
	assert.NotContains(t, cleanItems[0].(map[string]any), "awsSessionToken")
	cleanTypedItems := clean["typedItems"].([]map[string]any)
	assert.NotContains(t, cleanTypedItems[0], "AwsSessionToken")

	assert.Contains(t, original, "AWSSessionToken", "the source must not be modified")
	assert.Contains(t, original["nested"].(map[string]any), "password")
	assert.Contains(t, original["items"].([]any)[0].(map[string]any), "awsSessionToken")
	assert.Contains(t, original["typedItems"].([]map[string]any)[0], "AwsSessionToken")

	cleanNested["visible"] = "changed"
	cleanItems[0].(map[string]any)["visible"] = "changed"
	cleanTypedItems[0]["visible"] = "changed"
	require.Equal(t, "original", original["nested"].(map[string]any)["visible"])
	require.Equal(t, "original", original["items"].([]any)[0].(map[string]any)["visible"])
	require.Equal(t, "original", original["typedItems"].([]map[string]any)[0]["visible"])
}
