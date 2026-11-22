/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestKnativeServingOverrideAutoscaler(t *testing.T) {
	for _, key := range []string{"autoscaler", "config-autoscaler"} {
		for _, tc := range []struct {
			name          string
			allowZero     any
			wantAllowZero string
		}{
			{name: "default", wantAllowZero: "true"},
			{name: "explicit opt out", allowZero: "false", wantAllowZero: "false"},
			{name: "explicit opt in", allowZero: "true", wantAllowZero: "true"},
		} {
			t.Run(key+"/"+tc.name, func(t *testing.T) {
				autoscaler := map[string]any{"initial-scale": "2", "stable-window": "90s"}
				if tc.allowZero != nil {
					autoscaler["allow-zero-initial-scale"] = tc.allowZero
				}
				override := map[string]any{"config": map[string]any{key: autoscaler}}
				result := knativeServingOverride(nil, override)
				got, found, err := unstructured.NestedStringMap(result, "config", key)
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, tc.wantAllowZero, got["allow-zero-initial-scale"])
				assert.Equal(t, "2", got["initial-scale"])
				assert.Equal(t, "90s", got["stable-window"])
				assert.Equal(t, tc.allowZero, autoscaler["allow-zero-initial-scale"], "the input override is not mutated")
			})
		}
	}
}
