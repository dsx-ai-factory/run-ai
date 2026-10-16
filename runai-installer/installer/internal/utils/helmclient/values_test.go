/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helmclient

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHandleValues_NativeYAML(t *testing.T) {
	file := filepath.Join(t.TempDir(), "airgap-values.yaml")
	require.NoError(t, os.WriteFile(file, []byte("global:\n  domain: file.example.com\n  registry: offline\ncommand: [default]\n"), 0o600))
	custom := map[string]any{
		"global":  map[string]any{"domain": "custom.example.com"},
		"command": []any{"/bin/sh", "-c", "echo a,b\nexec app"},
		"quoted":  "true",
		"empty":   []any{},
		"unset":   nil,
	}
	for _, files := range [][]string{nil, {file}} {
		c := &helmClient{data: &helmChartData{
			valuesOverride: []string{"global.domain=default.example.com", "global.enabled=true"},
			customValues:   custom,
			valueFiles:     files,
		}}
		got, err := c.handleValues()
		require.NoError(t, err)
		global := map[string]any{"domain": "custom.example.com", "enabled": true}
		if len(files) > 0 {
			global["registry"] = "offline"
		}
		require.Equal(t, map[string]any{
			"global": global, "command": custom["command"], "quoted": "true", "empty": []any{}, "unset": nil,
		}, got)
		require.Equal(t, map[string]any{"domain": "custom.example.com"}, custom["global"], "do not mutate input values")
	}
}
