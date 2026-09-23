/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pack_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/releaselock"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func releaseLock(version, arch string) *artifactlock.ArtifactLock {
	lock := harness.TestLock()
	lock.Metadata.Version = version
	lock.Spec.Platform.Architecture = arch
	return lock
}

func releases() []harness.GitHubRelease {
	return []harness.GitHubRelease{
		harness.GitHubRelease{Tag: "3.1.0"}.WithLock("amd64", releaseLock("3.1.0", "amd64")),
		harness.GitHubRelease{Tag: "3.0.7", Prerelease: true}.WithLock("amd64", releaseLock("3.0.7", "amd64")),
		harness.GitHubRelease{Tag: "3.0.6", Draft: true}.WithLock("amd64", releaseLock("3.0.6", "amd64")),
		harness.GitHubRelease{Tag: "3.0.5"}.
			WithLock("amd64", releaseLock("3.0.5", "amd64")).
			WithLock("arm64", releaseLock("3.0.5", "arm64")),
		harness.GitHubRelease{Tag: "3.0.4"}.WithLock("amd64", releaseLock("3.0.4", "amd64")),
		{Tag: "3.0.3"},
		harness.GitHubRelease{Tag: "2.26.13"}.WithLock("amd64", releaseLock("2.26.13", "amd64")),
	}
}

func releaseFetcher(server *harness.GitHubReleaseServer) *releaselock.Fetcher {
	return &releaselock.Fetcher{
		APIBaseURL: server.URL,
		Repository: server.Repository,
		HTTPClient: server.Client(),
	}
}

func releaseConfig(t *testing.T, version, arch string) string {
	t.Helper()
	cfg := "settings:\n  runaiVersion: " + version + "\n"
	if arch != "" {
		cfg += "  architecture: " + arch + "\n"
	}
	return cfg + "  output: " + filepath.Join(t.TempDir(), "runai-airgap.tar") + "\n"
}

func TestPack_ReleaseLock_ProvidedLockFileSkipsTheLookup(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)
	lockPath := harness.WriteLockFile(t, harness.TestLock())

	res := runResolving(t, packConfig(t), lockPath, releaseFetcher(server), true, newFakes())
	require.NoError(t, res.err)

	assert.Equal(t, lockPath, res.source.Path)
	assert.Empty(t, res.source.Tag)
	assert.Empty(t, server.Requests(), "a provided lock file must not trigger a release lookup")
}

func TestPack_ReleaseLock_MajorMinorResolvesToNewestPublishedPatch(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)

	res := runRelease(t, "settings:\n  runaiVersion: \"3.0\"\n", releaseFetcher(server), true)
	require.NoError(t, res.err)

	assert.Equal(t, "3.0.5", res.source.Tag, "the prerelease 3.0.7 and the draft 3.0.6 must be skipped")
	assert.Equal(t, "runai-3.0.5-standard-linux-amd64.yaml", res.source.Asset, "amd64 is the default architecture")
	assert.Equal(t, server.DownloadURL("3.0.5", "runai-3.0.5-standard-linux-amd64.yaml"), res.source.URL, "the source carries the asset's download link")
	assert.Equal(t, "runai-3.0.5-airgap.tar", res.output, "the fetched lock decides the exact version")
	assert.Nil(t, findEventByKind(res.events, "input_validation_failed"), "3.0 resolving to 3.0.5 is not a version mismatch")
	assert.Contains(t, res.stdout, "NVIDIA Run:ai air-gapped bundle completed (dry run)")

	var downloads int
	for _, r := range server.Requests() {
		if r.Header.Get("Accept") == "application/octet-stream" {
			downloads++
		}
	}
	assert.Equal(t, 1, downloads, "exactly one asset must be downloaded")
}

func TestPack_ReleaseLock_ExactVersionSelectsThatRelease(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)

	res := runRelease(t, releaseConfig(t, "3.0.4", ""), releaseFetcher(server), true)
	require.NoError(t, res.err)

	assert.Equal(t, "3.0.4", res.source.Tag)
	assert.Equal(t, "runai-3.0.4-standard-linux-amd64.yaml", res.source.Asset)
}

func TestPack_ReleaseLock_ArchitectureSelectsTheMatchingAsset(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)

	res := runRelease(t, releaseConfig(t, "3.0.5", "arm64"), releaseFetcher(server), true)
	require.NoError(t, res.err)

	assert.Equal(t, "runai-3.0.5-standard-linux-arm64.yaml", res.source.Asset)
	assert.Nil(t, findEventByKind(res.events, "input_validation_failed"))
}

func TestPack_ReleaseLock_FetchedLockIsPacked(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)

	res := runRelease(t, releaseConfig(t, "3.0.5", ""), releaseFetcher(server), false)
	require.NoError(t, res.err)

	assert.ElementsMatch(t, []string{"control-plane", "runai-cluster"}, res.fetcher.Fetches)
	assert.Len(t, res.puller.Pulls, 2)
	assert.FileExists(t, res.output)
}

func TestPack_ReleaseLock_WalksReleasePages(t *testing.T) {
	var many []harness.GitHubRelease
	for i := 0; i < 100; i++ {
		many = append(many, harness.GitHubRelease{Tag: fmt.Sprintf("4.%d.0", 100-i)})
	}
	many = append(many, harness.GitHubRelease{Tag: "3.0.5"}.WithLock("amd64", releaseLock("3.0.5", "amd64")))
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, many...)

	res := runRelease(t, releaseConfig(t, "3.0", ""), releaseFetcher(server), true)
	require.NoError(t, res.err)
	assert.Equal(t, "3.0.5", res.source.Tag)

	var pages []string
	for _, r := range server.Requests() {
		if page := r.URL.Query().Get("page"); page != "" {
			pages = append(pages, page)
		}
	}
	assert.Equal(t, []string{"1", "2"}, pages)
}

func TestPack_ReleaseLock_Failures(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		contains []string
		noLookup bool
	}{
		{
			name:     "release not found",
			config:   "settings:\n  runaiVersion: \"2.9\"\n",
			contains: []string{"could not fetch lock file for version 2.9: release not found. Please provide a lock file using --airgap-lock-file"},
		},
		{
			name:     "exact version not released",
			config:   "settings:\n  runaiVersion: 3.0.9\n",
			contains: []string{"could not fetch lock file for version 3.0.9: release not found"},
		},
		{
			name:     "release without lock files",
			config:   "settings:\n  runaiVersion: 3.0.3\n",
			contains: []string{"could not fetch lock file for version 3.0.3: no lock files in release 3.0.3 (expected runai-3.0.3-standard-linux-amd64.yaml)"},
		},
		{
			name:     "release without a lock for the architecture",
			config:   "settings:\n  runaiVersion: 3.0.4\n  architecture: arm64\n",
			contains: []string{"no lock files in release 3.0.4 (expected runai-3.0.4-standard-linux-arm64.yaml)"},
		},
		{
			name:     "development version",
			config:   "settings:\n  runaiVersion: 0.0.0-30002779.main-bef491b\n",
			contains: []string{"could not fetch lock file for version 0.0.0-30002779.main-bef491b: it is not a release version", "--airgap-lock-file"},
			noLookup: true,
		},
		{
			name:     "branch name",
			config:   "settings:\n  runaiVersion: main\n",
			contains: []string{"could not fetch lock file for version main: it is not a release version"},
			noLookup: true,
		},
		{
			name:     "version missing",
			config:   "settings:\n  architecture: amd64\n",
			contains: []string{"settings.runaiVersion is required", "--airgap-lock-file"},
			noLookup: true,
		},
		{
			name:     "unsupported architecture",
			config:   "settings:\n  runaiVersion: 3.0.5\n  architecture: s390x\n",
			contains: []string{`unsupported architecture "s390x"`},
			noLookup: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)

			res := runRelease(t, tc.config, releaseFetcher(server), true)
			require.Error(t, res.err)
			for _, want := range tc.contains {
				assert.Contains(t, res.err.Error(), want)
			}
			if tc.noLookup {
				assert.Empty(t, server.Requests(), "the failure must be decided without contacting GitHub")
			}
		})
	}
}

func TestPack_ReleaseLock_GitHubUnreachable(t *testing.T) {
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository, releases()...)
	fetcher := releaseFetcher(server)
	server.Close()

	res := runRelease(t, releaseConfig(t, "3.0", ""), fetcher, true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), "could not fetch lock file for version 3.0: list releases of dsx-ai-factory/run-ai")
	assert.Contains(t, res.err.Error(), "--airgap-lock-file")
}

func TestPack_ReleaseLock_ForeignAssetIsRejected(t *testing.T) {
	foreign := harness.TestLock()
	foreign.Kind = "ImageLock"
	foreign.Metadata.Version = "3.0.5"
	server := harness.NewGitHubReleaseServer(t, releaselock.DefaultRepository,
		harness.GitHubRelease{Tag: "3.0.5"}.WithLock("amd64", foreign))

	res := runRelease(t, releaseConfig(t, "3.0.5", ""), releaseFetcher(server), true)
	require.Error(t, res.err)
	assert.Contains(t, res.err.Error(), `artifact lock "runai-3.0.5-standard-linux-amd64.yaml" is not a artifacts.run.ai/v1alpha1 ArtifactLock`)
}
