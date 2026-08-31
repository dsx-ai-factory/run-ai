/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertToStringMap_DifferentScenarios(t *testing.T) {
	tests := []struct {
		name   string
		source map[string]interface{}
		want   map[string]string
	}{
		{
			name:   "All string values",
			source: map[string]interface{}{"key1": "value1", "key2": "value2"},
			want:   map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:   "Mixed types",
			source: map[string]interface{}{"key1": "value1", "key2": 123, "key3": true},
			want:   map[string]string{"key1": "value1"},
		},
		{
			name:   "Empty map",
			source: map[string]interface{}{},
			want:   map[string]string{},
		},
		{
			name:   "Nil map",
			source: nil,
			want:   map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToStringMap(tt.source)
			require.Equal(t, len(got), len(tt.want))
			for key, value := range tt.want {
				require.Equal(t, got[key], value)
			}
		})
	}
}
