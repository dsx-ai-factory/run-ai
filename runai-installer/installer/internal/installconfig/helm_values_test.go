/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/cli/values"
)

func TestCustomHelmValues_ParsedByHelm(t *testing.T) {
	cases := []struct {
		name string
		yaml string
	}{
		{
			name: "nested",
			yaml: `global: {fipsMode: only}
telemetryService: {enabled: true}
replicas: 3`,
		},
		{
			name: "dotted",
			yaml: `global.fipsMode: only
telemetryService.enabled: true
replicas: 3`,
		},
		{
			name: "mixed with explicit override",
			yaml: `global: {fipsMode: disabled}
global.fipsMode: only
telemetryService: {enabled: true}
replicas: 3
empty: {}`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			// YAML aliases exercise the same user input for all three consumers.
			configYAML := "controlPlane:\n  enabled: true\n  customHelmValues: &custom\n"
			for _, line := range strings.Split(tc.yaml, "\n") {
				configYAML += "    " + line + "\n"
			}
			configYAML += "cluster:\n  enabled: true\n  customHelmValues: *custom\nfeatures:\n  nimOperator:\n    mode: install\n    customHelmValues: *custom\n"
			require.NoError(t, os.WriteFile(path, []byte(configYAML), 0o600))

			cfg, err := installconfig.Load(path)
			require.NoError(t, err)
			spec := cfg.ToSpec(installconfig.RunOptions{})

			for _, custom := range []map[string]string{
				spec.ControlPlane.CustomHelmValues,
				spec.Cluster.CustomHelmValues,
				spec.Features.NIMOperator.CustomHelmValues,
			} {
				opts := values.Options{Values: actions.MergeValues([]string{"global.domain=runai.example.com"}, custom)}
				parsed, err := opts.MergeValues(nil)
				require.NoError(t, err)
				require.Equal(t, map[string]any{
					"global":           map[string]any{"domain": "runai.example.com", "fipsMode": "only"},
					"telemetryService": map[string]any{"enabled": true},
					"replicas":         int64(3),
				}, parsed)
			}
		})
	}
}

func TestCustomHelmValues_LiteralNestedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`controlPlane:
  enabled: true
  customHelmValues:
    assetsService:
      podAnnotations:
        runai.example/nested: nested
        "a,b=c[d\\e": literal
    assetsService.podAnnotations.runai\.example/dotted: dotted
`), 0o600))

	cfg, err := installconfig.Load(path)
	require.NoError(t, err)
	opts := values.Options{Values: actions.MapToValues(cfg.ToSpec(installconfig.RunOptions{}).ControlPlane.CustomHelmValues)}
	parsed, err := opts.MergeValues(nil)
	require.NoError(t, err)
	require.Equal(t, map[string]any{
		"assetsService": map[string]any{
			"podAnnotations": map[string]any{
				"runai.example/nested": "nested",
				"runai.example/dotted": "dotted",
				`a,b=c[d\e`:            "literal",
			},
		},
	}, parsed)
}
