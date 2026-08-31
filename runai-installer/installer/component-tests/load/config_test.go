/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_RejectsUnknownConfigKeys(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle)+"  registryUrl: typo.example.com\n", false)

	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "registryUrl")
}

func TestLoad_InputValidation(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())
	workingDir := filepath.Join(t.TempDir(), "airgap")

	tests := []struct {
		name    string
		config  string
		field   string
		message string
	}{
		{
			name:    "missing bundle path",
			config:  fmt.Sprintf("settings:\n  workingDir: %s\nregistry:\n  url: %s\n", workingDir, destination),
			field:   "settings.bundlePath",
			message: "bundlePath is required",
		},
		{
			name:    "bundle does not exist",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n", filepath.Join(t.TempDir(), "missing.tar"), workingDir, destination),
			field:   "settings.bundlePath",
			message: "does not exist",
		},
		{
			name:    "bundle is a directory",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n", t.TempDir(), workingDir, destination),
			field:   "settings.bundlePath",
			message: "is a directory",
		},
		{
			name:    "missing working directory",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\nregistry:\n  url: %s\n", bundle, destination),
			field:   "settings.workingDir",
			message: "workingDir is required",
		},
		{
			name:    "missing registry",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\n", bundle, workingDir),
			field:   "registry.url",
			message: "url is required",
		},
		{
			name:    "registry with a scheme",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: https://registry.local/runai\n", bundle, workingDir),
			field:   "registry.url",
			message: "must not carry a scheme",
		},
		{
			name:    "registry with a tag",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: registry.local/runai:2.26.13\n", bundle, workingDir),
			field:   "registry.url",
			message: "without a tag or digest",
		},
		{
			name:    "ambiguous short registry host",
			config:  fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: harbor/runai\n", bundle, workingDir),
			field:   "registry.url",
			message: "must use a fully qualified registry host",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := run(t, tc.config, false)

			require.Error(t, res.err)
			got := findings(t, res.events)
			require.Contains(t, got, tc.field)
			assert.Contains(t, got[tc.field], tc.message)
			assert.Empty(t, res.pusher.Pushes, "nothing runs when the input is invalid")
		})
	}
}

func TestLoad_RegistryPortIsNotATag(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: registry.local:5000\n",
		bundle, filepath.Join(t.TempDir(), "airgap")), false)

	require.NoError(t, res.err)
	assert.Contains(t, res.pusher.Pushes, "registry.local:5000/runai-control-plane/postgresql:16")
}

func TestLoad_DockerHubAliasIsNormalized(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: docker.io/runai\n",
		bundle, filepath.Join(t.TempDir(), "airgap")), false)

	require.NoError(t, res.err)
	assert.Contains(t, res.pusher.Checks, "index.docker.io/runai/runai-airgap-probe:probe")
	assert.Contains(t, res.pusher.Pushes, "index.docker.io/runai/runai-control-plane/postgresql:16")
}

func TestLoad_AuthSchemes(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())
	workingDir := filepath.Join(t.TempDir(), "airgap")

	passwordPath := filepath.Join(t.TempDir(), "password")
	require.NoError(t, os.WriteFile(passwordPath, []byte("from-file\n"), 0o600))

	t.Run("token authenticates as the token user", func(t *testing.T) {
		res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n  token: ngc-token\n",
			bundle, workingDir, destination), false)

		require.NoError(t, res.err)
		assert.Equal(t, "$oauthtoken", res.resolved.AirgapLoad.Credential.Username)
		assert.Equal(t, "ngc-token", res.resolved.AirgapLoad.Credential.Password)
	})

	t.Run("password path is read", func(t *testing.T) {
		res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n  username: admin\n  passwordPath: %s\n",
			bundle, workingDir, destination, passwordPath), false)

		require.NoError(t, res.err)
		assert.Equal(t, "from-file", res.resolved.AirgapLoad.Credential.Password)
	})

	t.Run("password env is read", func(t *testing.T) {
		t.Setenv("LOAD_TEST_REGISTRY_PASSWORD", "from-env")

		res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n  username: admin\n  passwordEnv: LOAD_TEST_REGISTRY_PASSWORD\n",
			bundle, workingDir, destination), false)

		require.NoError(t, res.err)
		assert.Equal(t, "from-env", res.resolved.AirgapLoad.Credential.Password)
	})

	t.Run("two schemes are rejected", func(t *testing.T) {
		res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n  username: admin\n  password: secret\n  token: ngc-token\n",
			bundle, workingDir, destination), false)

		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "not both")
		assert.Nil(t, res.events, "the run never starts")
	})

	t.Run("an unset environment variable is rejected", func(t *testing.T) {
		res := run(t, fmt.Sprintf("settings:\n  bundlePath: %s\n  workingDir: %s\nregistry:\n  url: %s\n  username: admin\n  passwordEnv: LOAD_TEST_UNSET_PASSWORD\n",
			bundle, workingDir, destination), false)

		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "LOAD_TEST_UNSET_PASSWORD")
	})
}
