/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

func helmSetArgs(t *testing.T, customHelmValues string) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := "controlPlane:\n  enabled: true\n  domain: runai.example.com\n  customHelmValues:\n" + customHelmValues
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	return actions.MergeValues(nil, cfg.ToSpec(RunOptions{}).ControlPlane.CustomHelmValues)
}

func TestCustomHelmValues_NestedAndDottedAreEquivalent(t *testing.T) {
	nested := helmSetArgs(t, `    supportEntitled: true
    telemetryService:
      enabled: true
      telemetryAuth:
        existingSecret: my-ngc-secret
        secretKey: apiKey
    global:
      fipsMode: only
`)
	dotted := helmSetArgs(t, `    supportEntitled: true
    telemetryService.enabled: true
    telemetryService.telemetryAuth.existingSecret: my-ngc-secret
    telemetryService.telemetryAuth.secretKey: apiKey
    global.fipsMode: only
`)
	want := []string{
		"global.fipsMode=only",
		"supportEntitled=true",
		"telemetryService.enabled=true",
		"telemetryService.telemetryAuth.existingSecret=my-ngc-secret",
		"telemetryService.telemetryAuth.secretKey=apiKey",
	}
	assert.Equal(t, want, dotted, "dotted form")
	assert.Equal(t, want, nested, "nested form")
}

func TestCustomHelmValues_MixedFormAndNoGoMapStrings(t *testing.T) {
	args := helmSetArgs(t, `    supportEntitled: true
    global.fipsMode: only
    telemetryService:
      enabled: true
`)
	assert.Equal(t, []string{
		"global.fipsMode=only",
		"supportEntitled=true",
		"telemetryService.enabled=true",
	}, args)
	for _, a := range args {
		assert.NotContains(t, a, "map[", "no Go map string must reach Helm")
	}
}

func TestCustomHelmValues_ParsedByHelm(t *testing.T) {
	cases := map[string]struct {
		in   map[string]any
		want map[string]any
	}{
		"nested": {
			in:   map[string]any{"telemetryService": map[string]any{"enabled": true}},
			want: map[string]any{"telemetryService": map[string]any{"enabled": true}},
		},
		"dot in nested key": {
			in:   map[string]any{"ann": map[string]any{"foo.bar/baz": "v"}},
			want: map[string]any{"ann": map[string]any{"foo.bar/baz": "v"}},
		},
		"brackets and equals in nested key": {
			in:   map[string]any{"a": map[string]any{"route[canary]": "x", "k=1": "y"}},
			want: map[string]any{"a": map[string]any{"route[canary]": "x", "k=1": "y"}},
		},
		"comma in nested value": {
			in:   map[string]any{"env": map[string]any{"NO_PROXY": "localhost,127.0.0.1"}},
			want: map[string]any{"env": map[string]any{"NO_PROXY": "localhost,127.0.0.1"}},
		},
		"list with comma, brace and backslash": {
			in:   map[string]any{"hosts": []any{"a,b", "c}d", `e\f`}},
			want: map[string]any{"hosts": []any{"a,b", "c}d", `e\f`}},
		},
		"list of maps": {
			in: map[string]any{"hosts": []any{map[string]any{"name": "a"}, map[string]any{"name": "b"}}},
			want: map[string]any{"hosts": []any{
				map[string]any{"name": "a"}, map[string]any{"name": "b"},
			}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var args []string
			for k, v := range toStringMap(tc.in) {
				args = append(args, k+"="+v)
			}
			got, err := (&values.Options{Values: args}).MergeValues(getter.Providers{})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
