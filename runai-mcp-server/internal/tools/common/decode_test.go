/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeTypedRequest(t *testing.T) {
	type target struct {
		Name   string `json:"name"`
		Nested struct {
			Enabled bool `json:"enabled"`
		} `json:"nested"`
	}

	var out target
	require.NoError(t, DecodeTypedRequest(map[string]any{"name": "training", "nested": map[string]any{"enabled": true}}, &out))
	assert.Equal(t, "training", out.Name)
	assert.True(t, out.Nested.Enabled)

	err := DecodeTypedRequest(map[string]any{"name": "training", "unexpected": true}, &out)
	assert.ErrorContains(t, err, `unknown field "unexpected"`)

	err = DecodeTypedRequest(map[string]any{"name": "training", "nested": map[string]any{"typo": true}}, &out)
	assert.ErrorContains(t, err, `unknown field "typo"`)
}

func TestDecodeTypedRequestReturnsMarshalError(t *testing.T) {
	var out map[string]any
	err := DecodeTypedRequest(map[string]any{"unsupported": make(chan struct{})}, &out)
	assert.ErrorContains(t, err, "unsupported type")
}
