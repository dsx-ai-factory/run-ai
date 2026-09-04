/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pack_test

import (
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPack_UnknownConfigKeyIsRejected(t *testing.T) {
	res := run(t, "settings:\n  architecture: amd64\n  parallelism: 8\n", harness.TestLock(), true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "field parallelism not found")
}

func TestPack_InputValidationFindings(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		field   string
		message string
	}{
		{
			name:    "version does not match the lock",
			config:  "settings:\n  runaiVersion: 2.25.0\n  architecture: amd64\n",
			field:   "settings.runaiVersion",
			message: "the artifact lock is for 2.26.13, not 2.25.0",
		},
		{
			name:    "architecture does not match the lock",
			config:  "settings:\n  architecture: arm64\n",
			field:   "settings.architecture",
			message: "the artifact lock is for amd64, not arm64",
		},
		{
			name:    "unsupported architecture",
			config:  "settings:\n  architecture: s390x\n",
			field:   "settings.architecture",
			message: `unsupported architecture "s390x": must be amd64 or arm64`,
		},
		{
			name:    "registry is required",
			config:  "settings:\n  architecture: amd64\nregistryAuth:\n  - registry: \"\"\n    username: u\n    password: p\n",
			field:   "registryAuth[0].registry",
			message: "registry is required",
		},
		{
			name:    "duplicate registry",
			config:  "settings:\n  architecture: amd64\nregistryAuth:\n  - registry: nvcr.io\n    token: a\n  - registry: nvcr.io\n    token: b\n",
			field:   "registryAuth[1].registry",
			message: `duplicate registry "nvcr.io"`,
		},
		{
			name:    "override without a tag or digest",
			config:  "settings:\n  architecture: amd64\nimageOverrides:\n  - name: postgresql\n    source: harbor.corp.example.com/runai/postgresql\n",
			field:   "imageOverrides[0].source",
			message: `must carry an explicit tag or digest`,
		},
		{
			name:    "override without a name",
			config:  "settings:\n  architecture: amd64\nimageOverrides:\n  - name: \"\"\n    source: harbor.corp.example.com/runai/postgresql:16\n",
			field:   "imageOverrides[0].name",
			message: "name is required",
		},
		{
			name:    "duplicate override",
			config:  "settings:\n  architecture: amd64\nimageOverrides:\n  - name: postgresql\n    source: a.example.com/postgresql:16\n  - name: postgresql\n    source: b.example.com/postgresql:16\n",
			field:   "imageOverrides[1].name",
			message: `duplicate override "postgresql"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := run(t, tc.config, harness.TestLock(), true)
			require.Error(t, res.err)
			assert.Contains(t, res.err.Error(), "input validation failed")

			got := findings(t, res.events)
			require.Contains(t, got, tc.field)
			assert.Contains(t, got[tc.field], tc.message)

			assert.Nil(t, findEventByKind(res.events, "pipeline_started"), "no work may start after a validation failure")
		})
	}
}

func TestPack_SecretSources(t *testing.T) {
	t.Run("token from the environment", func(t *testing.T) {
		t.Setenv("PACK_TEST_TOKEN", "ngc-token")
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    tokenEnv: PACK_TEST_TOKEN\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.NoError(t, res.err)
	})

	t.Run("password from a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "password")
		require.NoError(t, writeFile(path, "s3cr3t\n"))
		cfg := packConfig(t) + "registryAuth:\n  - registry: harbor.corp.example.com/runai\n    username: runai\n    passwordPath: " + path + "\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.NoError(t, res.err)
	})

	t.Run("missing environment variable", func(t *testing.T) {
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    tokenEnv: PACK_TEST_ABSENT\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "environment variable PACK_TEST_ABSENT is not set")
	})

	t.Run("missing password file", func(t *testing.T) {
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    username: runai\n    passwordPath: /nonexistent/password\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "read passwordPath")
	})

	t.Run("two sources for one secret", func(t *testing.T) {
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    token: inline\n    tokenEnv: PACK_TEST_TOKEN\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "set only one of token, tokenPath or tokenEnv")
	})

	t.Run("token and basic auth together", func(t *testing.T) {
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    token: inline\n    username: runai\n    password: s3cr3t\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "set either a token or a username and password, not both")
	})

	t.Run("password without a username", func(t *testing.T) {
		cfg := packConfig(t) + "registryAuth:\n  - registry: nvcr.io\n    password: s3cr3t\n"

		res := run(t, cfg, harness.TestLock(), true)
		require.Error(t, res.err)
		assert.Contains(t, res.err.Error(), "a password requires a username")
	})
}

func TestPack_AmbiguousOverrideRequiresGroup(t *testing.T) {
	lock := harness.TestLock()
	image := lock.Spec.ImageLocks[0].Images[0]
	image.Name = "operator"
	lock.Spec.ImageLocks = []artifactlock.ImageLock{
		{Name: "gpu-sharing", Version: "v0.1.0", Images: []artifactlock.Image{image}},
		{Name: "kai-scheduler", Version: "v0.16.8", Images: []artifactlock.Image{image}},
	}

	res := run(t, packConfig(t)+`
imageOverrides:
  - name: operator
    source: registry.example.com/operator:v1
`, lock, true)
	require.Error(t, res.err)
	assert.Contains(t, findings(t, res.events)["imageOverrides[0].name"], "set group")

	res = run(t, packConfig(t)+`
imageOverrides:
  - group: kai-scheduler
    name: operator
    source: registry.example.com/operator:v1
`, lock, true)
	require.NoError(t, res.err)
}

func TestPack_AcceptsLockWithCLIArtifacts(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       "https://api.example.com/runai-installer.tgz",
		SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Files: []artifactlock.CLIFile{{
			SourcePath: "linux/amd64/runai-installer",
			BundlePath: "cli/runai-installer/linux/amd64/runai-installer",
			Digest:     harness.CLIDigest("runai-installer", "linux/amd64/runai-installer"),
			Mode:       0o755,
		}},
	}}

	res := run(t, packConfig(t), lock, true)
	require.NoError(t, res.err)
}

func TestPack_RejectsCLITraversal(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       "https://api.example.com/runai-installer.tgz",
		SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Files: []artifactlock.CLIFile{{
			SourcePath: "linux/amd64/runai-installer",
			BundlePath: "cli/runai-installer/../../../outside",
			Digest:     harness.CLIDigest("runai-installer", "linux/amd64/runai-installer"),
			Mode:       0o755,
		}},
	}}

	res := run(t, packConfig(t), lock, true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "invalid CLI file")
}

func TestPack_RejectsDuplicateCLIBundlePaths(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       "https://api.example.com/runai-installer.tgz",
		SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Files: []artifactlock.CLIFile{
			{
				SourcePath: "linux/amd64/runai-installer",
				BundlePath: "cli/runai-installer/bin/runai-installer",
				Digest:     harness.CLIDigest("runai-installer", "linux/amd64/runai-installer"),
				Mode:       0o755,
			},
			{
				SourcePath: "darwin/arm64/runai-installer",
				BundlePath: "cli/runai-installer/bin/runai-installer",
				Digest:     harness.CLIDigest("runai-installer", "darwin/arm64/runai-installer"),
				Mode:       0o755,
			},
		},
	}}

	res := run(t, packConfig(t), lock, true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "invalid CLI file")
}
