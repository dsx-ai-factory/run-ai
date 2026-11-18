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
)

// helmSetArgs loads a config YAML, converts it to a spec and returns the
// control plane --set arguments the installer would pass to Helm.
func helmSetArgs(t *testing.T, customHelmValues string) []string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	yaml := "controlPlane:\n  enabled: true\n  domain: runai.example.com\n  customHelmValues:\n" + customHelmValues
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o600))

	cfg, err := Load(path)
	require.NoError(t, err)
	return actions.MergeValues(nil, cfg.ToSpec(RunOptions{}).ControlPlane.CustomHelmValues)
}

// Regression test for #2464: the nested form used in the guided-installation
// docs and the dotted-key form must produce identical Helm --set arguments.
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
