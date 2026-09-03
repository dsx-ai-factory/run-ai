/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pack_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPack_DryRun_PlansEveryStepAndWritesNothing(t *testing.T) {
	res := run(t, packConfig(t), harness.TestLock(), true)
	require.NoError(t, res.err)

	for _, step := range []string{"resolve-artifacts", "download-images", "download-charts", "assemble-bundle"} {
		ev := findEvent(res.events, "step_skipped", step)
		require.NotNil(t, ev, "step %q must be planned", step)
		assert.Equal(t, "dry_run", ev["skipReason"])
	}

	assert.Contains(t, res.stdout, "resolve 2 image(s) and 2 chart(s) with 0 override(s)")
	assert.Contains(t, res.stdout, "NVIDIA Run:ai air-gapped bundle completed (dry run)")
	assert.NotContains(t, res.stdout, "installation completed", "a pack does not install anything")
	assert.Contains(t, res.stdout, "download the locked container images")
	assert.Contains(t, res.stdout, "write the bundle to")

	_, err := os.Stat(res.output)
	assert.True(t, os.IsNotExist(err), "dry-run must not write a bundle")

	runEv := findEventByKind(res.events, "run_completed")
	require.NotNil(t, runEv)
	run := runEv["run"].(map[string]any)
	assert.Equal(t, "succeeded", run["outcome"])
	assert.Equal(t, "airgap-pack", run["command"])
	assert.Equal(t, true, run["dryRun"])
}

func TestPack_TextOutput_UsesHumanTitles(t *testing.T) {
	res := run(t, packConfig(t), harness.TestLock(), true)
	require.NoError(t, res.err)

	assert.Contains(t, res.stdout, "# Air-gapped bundle")
	assert.Contains(t, res.stdout, "Air-gapped bundle creation")
	assert.Contains(t, res.stdout, "Resolving artifact inventory")
	assert.Contains(t, res.stdout, "Downloading container images")

	assert.NotContains(t, res.stdout, "resolve-artifacts")
	assert.NotContains(t, res.stdout, "download-images")
}

func TestPack_OverridesAreCounted(t *testing.T) {
	cfg := packConfig(t) + `
imageOverrides:
  - name: postgresql
    source: harbor.corp.example.com/runai/postgresql:16
  - name: extra-tool
    source: harbor.corp.example.com/runai/extra-tool:1.0
`
	res := run(t, cfg, harness.TestLock(), true)
	require.NoError(t, res.err)
	assert.Contains(t, res.stdout, "with 2 override(s)")
}

func TestPack_DefaultOutputPathFollowsTheLock(t *testing.T) {
	res := run(t, "settings:\n  architecture: amd64\n", harness.TestLock(), true)
	require.NoError(t, res.err)
	assert.Equal(t, "runai-2.26.13-airgap.tar", res.output)
}

// The fixture is a lock produced by build/airgap/artifact-lock, so packing it
// is what guards the Go schema against the generator's.
func TestPack_GeneratedLockFixture(t *testing.T) {
	cfg := packConfig(t) + `
registryAuth:
  - registry: nvcr.io
    username: $oauthtoken
    password: token
`
	res := runWithLockFile(t, cfg, filepath.Join("testdata", "runai-2.26.13-standard-linux-amd64.yaml"), true)
	require.NoError(t, res.err)
	assert.Contains(t, res.stdout, "resolve 51 image(s) and 2 chart(s)")
}

func TestPack_RejectsAForeignLockFile(t *testing.T) {
	lock := harness.TestLock()
	lock.Kind = "ImageLock"

	res := run(t, packConfig(t), lock, true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "is not a artifacts.run.ai/v1alpha1 ArtifactLock")
}
