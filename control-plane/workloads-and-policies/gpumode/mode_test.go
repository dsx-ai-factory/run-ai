/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpumode

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasGPU(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       bool
	}{
		{"portion", `{"gpuMemoryPortionRequest":0.25}`, true},
		{"legacy whole GPU", `{"gpuPortionRequest":2}`, true},
		{"devices", `{"gpuDevicesRequest":2}`, true},
		{"absolute memory", `{"gpuMemoryRequest":"1Gi"}`, true},
		{"CPU only", `{"gpuDevicesRequest":0}`, false},
		{"no allocation", `{}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var compute map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(tc.body), &compute))
			require.Equal(t, tc.want, HasGPU(compute))
		})
	}
}
