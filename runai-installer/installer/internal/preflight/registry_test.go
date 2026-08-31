/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryTargets_NGCWithToken(t *testing.T) {
	targets, configErr := registryTargets(&spec.Spec{
		Registry: &spec.Registry{Type: spec.RegistryTypeNGC, Token: "nvapi-secret"},
	})

	assert.Empty(t, configErr, "an NGC API key must not be rejected as invalid base64")
	require.Len(t, targets, 1)
	assert.Equal(t, "nvcr.io", targets[0].Host)
	assert.Equal(t, "$oauthtoken", targets[0].Username)
	assert.Equal(t, "nvapi-secret", targets[0].Password)
}

func TestRegistryTargets_NGCTokenFree(t *testing.T) {
	targets, configErr := registryTargets(&spec.Spec{
		Registry: &spec.Registry{Type: spec.RegistryTypeNGC},
	})

	assert.Empty(t, configErr, "a token-free NGC registry has nothing to probe")
	assert.Empty(t, targets)
}

func TestRegistryTargets_JFrogDockerConfig(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"auths": map[string]any{
			"runai.jfrog.io": map[string]any{"username": "u", "password": "p"},
		},
	})
	require.NoError(t, err)
	token := base64.StdEncoding.EncodeToString(raw)

	targets, configErr := registryTargets(&spec.Spec{
		Registry: &spec.Registry{Type: spec.RegistryTypeJFrog, Token: token},
	})

	assert.Empty(t, configErr)
	require.Len(t, targets, 1)
	assert.Equal(t, "runai.jfrog.io", targets[0].Host)
	assert.Equal(t, "u", targets[0].Username)
	assert.Equal(t, "p", targets[0].Password)
}
