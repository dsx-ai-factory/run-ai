/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_HappyPath(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)

	assert.Equal(t, []string{
		destination + "/runai-control-plane/assets-service:2.26.13",
		destination + "/runai-control-plane/postgresql:16",
	}, sorted(res.pusher.Pushes), "every locked image is pushed under its own name and tag")

	assert.Equal(t, []string{"preflight-load-registry", "extract-bundle", "push-images", "generate-values"},
		stepNames(res.events, "step_succeeded"))

	completed := findEventByKind(res.events, "run_completed")
	require.NotNil(t, completed)
	summary := completed["run"].(map[string]any)
	assert.Equal(t, "succeeded", summary["outcome"])
	assert.Equal(t, "airgap-load", summary["command"])

	for _, name := range []string{"control-plane-airgap-values.yaml", "cluster-airgap-values.yaml"} {
		assert.FileExists(t, filepath.Join(res.valuesDir, name))
	}
	assert.FileExists(t, filepath.Join(res.workingDir, "bundle", "artifact-lock.yaml"))
}

func TestLoad_PushesFromTheBundledLayout(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)

	assert.Equal(t,
		destination+"/runai-control-plane/assets-service:2.26.13",
		res.pusher.Layouts[filepath.Join(res.workingDir, "bundle", "images", "runai-control-plane", "assets-service")],
		"an image is pushed from its own directory in the extracted bundle")
}

func TestLoad_VerifiesAllCLIFiles(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       "https://api.example.com/runai-installer.tgz",
		SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Files: []artifactlock.CLIFile{
			{
				SourcePath: "darwin/arm64/runai-installer",
				BundlePath: "cli/runai-installer/darwin/arm64/runai-installer",
				Digest:     harness.CLIDigest("runai-installer", "darwin/arm64/runai-installer"),
				Mode:       0o755,
			},
			{
				SourcePath: "linux/amd64/runai-installer",
				BundlePath: "cli/runai-installer/linux/amd64/runai-installer",
				Digest:     harness.CLIDigest("runai-installer", "linux/amd64/runai-installer"),
				Mode:       0o755,
			},
		},
	}}
	bundle := harness.CreateLoadBundle(t, lock)

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)
	assert.FileExists(t, filepath.Join(res.workingDir, "bundle", "cli", "runai-installer", "darwin", "arm64", "runai-installer"))
	assert.FileExists(t, filepath.Join(res.workingDir, "bundle", "cli", "runai-installer", "linux", "amd64", "runai-installer"))
}

func TestLoad_GzipBundleStillExtracts(t *testing.T) {
	bundle := harness.GzipFile(t, harness.CreateLoadBundle(t, harness.TestLock()))

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)
	assert.Len(t, res.pusher.Pushes, 2)
}

func TestLoad_ValuesDirDefaultsToWorkingDir(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)
	assert.Equal(t, res.workingDir, res.valuesDir)
}

func TestLoad_ValuesDirIsHonoured(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())
	valuesDir := filepath.Join(t.TempDir(), "generated", "values")

	res := run(t, fmt.Sprintf(`settings:
  bundlePath: %s
  workingDir: %s
  valuesDir: %s
registry:
  url: %s
`, bundle, filepath.Join(t.TempDir(), "airgap"), valuesDir, destination), false)

	require.NoError(t, res.err)
	assert.Equal(t, valuesDir, res.valuesDir)
	assert.FileExists(t, filepath.Join(valuesDir, "control-plane-airgap-values.yaml"))
}

func TestLoad_OldFormatBundleFails(t *testing.T) {
	res := run(t, loadConfig(t, harness.CreateTestTarball(t)), false)

	require.Error(t, res.err)
	failed := findEvent(res.events, "step_failed", "extract-bundle")
	require.NotNil(t, failed, "a package that is not an air-gapped bundle fails at extraction")
	assert.Contains(t, failed["error"], "not an air-gapped bundle")
	assert.Empty(t, res.pusher.Pushes)
}

func TestLoad_BadArchiveFails(t *testing.T) {
	bundle := filepath.Join(t.TempDir(), "broken.tar")
	require.NoError(t, os.WriteFile(bundle, []byte("not a tar"), 0o600))

	res := run(t, loadConfig(t, bundle), false)

	require.Error(t, res.err)
	assert.NotNil(t, findEvent(res.events, "step_failed", "extract-bundle"))
	assert.Empty(t, res.pusher.Pushes)
}

func TestLoad_PushFailuresAreAllReported(t *testing.T) {
	lock := harness.TestLock()
	bundle := harness.CreateLoadBundle(t, lock)

	f := newFakes()
	f.pusher.PushErrors = map[string]error{
		destination + "/runai-control-plane/assets-service:2.26.13": errors.New("quota exceeded"),
		destination + "/runai-control-plane/postgresql:16":          errors.New("connection reset"),
	}

	res := runFakes(t, loadConfig(t, bundle), f)

	require.Error(t, res.err)
	assert.Len(t, res.pusher.Pushes, 2, "every image is attempted even after one fails")

	failed := findEvent(res.events, "step_failed", "push-images")
	require.NotNil(t, failed)
	assert.Contains(t, failed["error"], "quota exceeded")
	assert.Contains(t, failed["error"], "connection reset")
	assert.Contains(t, failed["error"], "2 of 2 image(s) failed")

	assert.Nil(t, findEvent(res.events, "step_succeeded", "generate-values"))
	assert.NoFileExists(t, filepath.Join(res.valuesDir, "control-plane-airgap-values.yaml"))
}

func TestLoad_DigestMismatchFails(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	f := newFakes()
	f.pusher.Digests = map[string]string{
		destination + "/runai-control-plane/postgresql:16": "sha256:0000000000000000000000000000000000000000000000000000000000000000",
	}

	res := runFakes(t, loadConfig(t, bundle), f)

	require.Error(t, res.err)
	failed := findEvent(res.events, "step_failed", "push-images")
	require.NotNil(t, failed)
	assert.Contains(t, failed["error"], "postgresql")
	assert.Contains(t, failed["error"], "the lock pins")
	assert.NotContains(t, res.pusher.Pushes, destination+"/runai-control-plane/postgresql:16")
}

func TestLoad_SourceWithoutTagFails(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.ImageLocks[0].Images[1].Source = "docker.io/bitnamilegacy/postgresql"
	bundle := harness.CreateLoadBundle(t, lock)

	res := run(t, loadConfig(t, bundle), false)

	require.Error(t, res.err)
	failed := findEvent(res.events, "step_failed", "push-images")
	require.NotNil(t, failed)
	assert.Contains(t, failed["error"], "no tag to mirror")
}

func TestLoad_UnpinnedImageFailsBeforeTransfer(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.ImageLocks[0].Images[0].Image = "nvcr.io/nvidia/runai/assets-service:2.26.13"
	bundle := harness.CreateLoadBundle(t, lock)

	res := run(t, loadConfig(t, bundle), false)

	require.Error(t, res.err)
	failed := findEvent(res.events, "step_failed", "extract-bundle")
	require.NotNil(t, failed)
	assert.Contains(t, failed["error"], "must use an immutable sha256 digest")
	assert.Equal(t, []string{destination + "/runai-airgap-probe:probe"}, res.pusher.Checks)
	assert.Empty(t, res.pusher.Pushes)
}

func TestLoad_DryRunTouchesNothing(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), true)
	require.NoError(t, res.err)

	assert.Empty(t, res.pusher.Pushes)
	assert.NoDirExists(t, filepath.Join(res.workingDir, "bundle"))
	assert.Equal(t, []string{"extract-bundle", "push-images", "generate-values"},
		stepNames(res.events, "step_skipped"))
	assert.Contains(t, res.stdout, "push the bundled container images to "+destination)
}

func TestLoad_TextOutputReportsALoad(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)

	assert.Contains(t, res.stdout, "Pushing container images")
	assert.Contains(t, res.stdout, "Air-gapped load complete")
	assert.NotContains(t, res.stdout, "push-images", "machine step names stay out of the terminal")
}

func TestLoad_SkipPreflightStillChecksTargetRepositories(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	f := newFakes()
	res := runStaged(t, loadConfig(t, bundle), f, func(o *specrun.ExecOptions) { o.SkipPreflight = true })

	require.NoError(t, res.err)
	assert.NotContains(t, f.pusher.Checks, destination+"/runai-airgap-probe:probe")
	assert.ElementsMatch(t, []string{
		destination + "/runai-control-plane/assets-service:2.26.13",
		destination + "/runai-control-plane/postgresql:16",
	}, f.pusher.Checks)
	assert.Len(t, res.pusher.Pushes, 2)
}

func sorted(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}
