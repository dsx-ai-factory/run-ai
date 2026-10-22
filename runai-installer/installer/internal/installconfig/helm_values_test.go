/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	"github.com/stretchr/testify/require"
)

func TestCustomHelmValues_NativeYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`controlPlane:
  enabled: true
  customHelmValues: &custom
    global:
      fipsMode: only
    telemetryService:
      enabled: true
    replicas: 3
    ratio: 1.5
    quotedBoolean: "true"
    quotedNumber: "123"
    affinity: null
    emptyList: []
    tolerations:
      - key: nvidia.com/gpu
        operator: Exists
    assetsService:
      podAnnotations:
        runai.example/nested: "true"
        'a,b=c[d\e': literal
    keycloakx:
      command:
        - /bin/sh
        - -c
        - |
          KC_BOOTSTRAP_ADMIN_USERNAME="$(cat /run/secrets/keycloak/admin-username)" \
          KC_BOOTSTRAP_ADMIN_PASSWORD="$(cat /run/secrets/keycloak/admin-password)" \
          exec /opt/keycloak/bin/kc.sh start \
            --features=token-exchange,admin-fine-grained-authz:v1,fips
cluster:
  enabled: true
  customHelmValues: *custom
features:
  nimOperator:
    mode: install
    customHelmValues: *custom
`), 0o600))

	cfg, err := installconfig.Load(path)
	require.NoError(t, err)
	spec := cfg.ToSpec(installconfig.RunOptions{})
	script := `KC_BOOTSTRAP_ADMIN_USERNAME="$(cat /run/secrets/keycloak/admin-username)" \
KC_BOOTSTRAP_ADMIN_PASSWORD="$(cat /run/secrets/keycloak/admin-password)" \
exec /opt/keycloak/bin/kc.sh start \
  --features=token-exchange,admin-fine-grained-authz:v1,fips
`
	for name, custom := range map[string]map[string]any{
		"control plane": spec.ControlPlane.CustomHelmValues,
		"cluster":       spec.Cluster.CustomHelmValues,
		"feature":       spec.Features.NIMOperator.CustomHelmValues,
	} {
		t.Run(name, func(t *testing.T) {
			input := helmclient.Values{Overrides: []string{"global.domain=runai.example.com", "global.fipsMode=disabled", "emptyList[0]=default"}, Custom: custom}
			parsed, err := input.Merge(nil)
			require.NoError(t, err)
			require.Equal(t, map[string]any{
				"global":           map[string]any{"domain": "runai.example.com", "fipsMode": "only"},
				"telemetryService": map[string]any{"enabled": true},
				"replicas":         3,
				"ratio":            1.5,
				"quotedBoolean":    "true",
				"quotedNumber":     "123",
				"affinity":         nil,
				"emptyList":        []any{},
				"tolerations":      []any{map[string]any{"key": "nvidia.com/gpu", "operator": "Exists"}},
				"assetsService":    map[string]any{"podAnnotations": map[string]any{"runai.example/nested": "true", `a,b=c[d\e`: "literal"}},
				"keycloakx":        map[string]any{"command": []any{"/bin/sh", "-c", script}},
			}, parsed)
			require.NotContains(t, custom["global"], "domain", "merging defaults must not mutate the config")
		})
	}
}

func TestCustomHelmValues_RejectsOverrideSyntax(t *testing.T) {
	for _, key := range []string{"global.domain", "keycloakx.command[0]", `global\.domain`} {
		t.Run(key, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(path, []byte("controlPlane:\n  customHelmValues:\n    "+key+": value\n"), 0o600))
			_, err := installconfig.Load(path)
			require.ErrorContains(t, err, "use nested YAML maps and lists instead")
		})
	}
}
