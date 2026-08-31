/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pack_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPack_BuildsTheBundle(t *testing.T) {
	res := run(t, packConfig(t), harness.TestLock(), false)
	require.NoError(t, res.err)

	assert.ElementsMatch(t, []string{
		"nvcr.io/nvidia/runai/assets-service@sha256:9741abc964a10bc41271a1bf2cffd56b4d05664cd78782c71b55469950b572d3",
		"docker.io/bitnamilegacy/postgresql@sha256:8f50ea374acc3e8da24b1c2ab0d2340cb98b7045ca9b25eff95d634e4874991d",
	}, res.puller.Pulls, "every locked image is pulled by digest")
	assert.Equal(t, []string{"control-plane"}, res.fetcher.Fetches)

	entries := tarEntries(t, res.output)
	assert.Contains(t, entries, "artifact-lock.yaml")
	assert.Contains(t, entries, "charts/control-plane-2.26.13.tgz")
	assert.Contains(t, entries, "images/runai-control-plane/assets-service/index.json")
	assert.Contains(t, entries, "images/runai-control-plane/postgresql/index.json")

	lock := string(tarFile(t, res.output, "artifact-lock.yaml"))
	assert.Contains(t, lock, "name: assets-service")
	assert.Contains(t, lock, "kind: ArtifactLock")

	_, err := os.Stat(res.workDir)
	assert.True(t, os.IsNotExist(err), "the staging directory is removed once the bundle exists")
}

func TestPack_IncludesAllCLIFiles(t *testing.T) {
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

	res := run(t, packConfig(t), lock, false)
	require.NoError(t, res.err)
	assert.Equal(t, []string{"runai-installer"}, res.cli.Fetches)
	entries := tarEntries(t, res.output)
	assert.Contains(t, entries, "cli/runai-installer/darwin/arm64/runai-installer")
	assert.Contains(t, entries, "cli/runai-installer/linux/amd64/runai-installer")
}

func TestPack_PreservesLegacyCLIFile(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       "https://api.example.com/runai-installer.tgz",
		SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		SourcePath:   "linux/amd64/runai-installer",
		BundlePath:   "cli/runai-installer",
		Digest:       harness.CLIDigest("runai-installer", "linux/amd64/runai-installer"),
	}}

	res := run(t, packConfig(t), lock, false)
	require.NoError(t, res.err)
	assert.Contains(t, tarEntries(t, res.output), "cli/runai-installer")
}

func TestPack_DownloadsCLIArchiveWithBearerAuthentication(t *testing.T) {
	archive := cliArchive(t, map[string][]byte{
		"./linux/amd64/runai-installer": []byte("linux"),
		"darwin/arm64/runai-installer":  []byte("darwin"),
	})
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		require.Equal(t, "Bearer secret", request.Header.Get("Authorization"))
		_, _ = response.Write(archive)
	}))
	t.Cleanup(server.Close)
	lock := harness.TestLock()
	lock.Spec.CLIs = []artifactlock.CLI{{
		Name:         "runai-installer",
		Version:      "2.26.13",
		Source:       server.URL,
		SourceDigest: digest(archive),
		Files: []artifactlock.CLIFile{
			{SourcePath: "darwin/arm64/runai-installer", BundlePath: "cli/runai-installer/darwin/arm64/runai-installer", Digest: digest([]byte("darwin")), Mode: 0o755},
			{SourcePath: "linux/amd64/runai-installer", BundlePath: "cli/runai-installer/linux/amd64/runai-installer", Digest: digest([]byte("linux")), Mode: 0o755},
		},
	}}
	host := strings.TrimPrefix(server.URL, "http://")
	f := newFakes()
	f.realCLI = artifacts.NewCLIFetcher(artifacts.NewKeychain(artifacts.Credentials{{
		Prefix: host, BearerToken: "secret", Insecure: true,
	}}))
	config := packConfig(t) + fmt.Sprintf("registryAuth:\n  - registry: %s\n    bearerToken: secret\n    insecure: true\n", host)

	res := runFakes(t, config, lock, f)
	require.NoError(t, res.err)
	entries := tarEntries(t, res.output)
	assert.Contains(t, entries, "cli/runai-installer/darwin/arm64/runai-installer")
	assert.Contains(t, entries, "cli/runai-installer/linux/amd64/runai-installer")
}

func cliArchive(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	compressed := gzip.NewWriter(&buffer)
	writer := tar.NewWriter(compressed)
	for name, content := range files {
		require.NoError(t, writer.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}))
		_, err := io.Copy(writer, bytes.NewReader(content))
		require.NoError(t, err)
	}
	require.NoError(t, writer.Close())
	require.NoError(t, compressed.Close())
	return buffer.Bytes()
}

func digest(content []byte) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(content))
}

func TestPack_PreservesImageGroups(t *testing.T) {
	lock := harness.TestLock()
	lock.Spec.ImageLocks = []artifactlock.ImageLock{
		{
			Name:    "gpu-sharing",
			Version: "v0.1.0",
			Images: []artifactlock.Image{{
				Name: "operator", Image: "ghcr.io/kai-scheduler/gpu-sharing/operator@sha256:1111111111111111111111111111111111111111111111111111111111111111",
				Source: "ghcr.io/kai-scheduler/gpu-sharing/operator:v0.1.0", IndexDigest: "sha256:1111111111111111111111111111111111111111111111111111111111111111",
			}},
		},
		{
			Name:    "kai-scheduler",
			Version: "v0.16.8",
			Images: []artifactlock.Image{{
				Name: "operator", Image: "ghcr.io/kai-scheduler/kai-scheduler/operator@sha256:2222222222222222222222222222222222222222222222222222222222222222",
				Source: "ghcr.io/kai-scheduler/kai-scheduler/operator:v0.16.8", IndexDigest: "sha256:2222222222222222222222222222222222222222222222222222222222222222",
			}},
		},
		{
			Name:    "karta",
			Version: "0.2.3",
			Images: []artifactlock.Image{{
				Name: "operator", Image: "ghcr.io/run-ai/karta/karta-operator@sha256:3333333333333333333333333333333333333333333333333333333333333333",
				Source: "ghcr.io/run-ai/karta/karta-operator:0.2.3", IndexDigest: "sha256:3333333333333333333333333333333333333333333333333333333333333333",
			}},
		},
	}
	lock.Spec.Charts = nil

	res := run(t, packConfig(t), lock, false)
	require.NoError(t, res.err)
	entries := tarEntries(t, res.output)
	assert.Contains(t, entries, "images/gpu-sharing/operator/index.json")
	assert.Contains(t, entries, "images/kai-scheduler/operator/index.json")
	assert.Contains(t, entries, "images/karta/operator/index.json")
}

func TestPack_WritesAVerifiableChecksum(t *testing.T) {
	res := run(t, packConfig(t), harness.TestLock(), false)
	require.NoError(t, res.err)

	content, err := os.ReadFile(res.output + ".sha256")
	require.NoError(t, err)

	fields := strings.Fields(string(content))
	require.Len(t, fields, 2, "the sidecar is a sha256sum line")
	assert.Equal(t, filepathBase(res.output), fields[1])

	bundle, err := os.ReadFile(res.output)
	require.NoError(t, err)
	assert.Equal(t, sha256Hex(bundle), fields[0], "the checksum is of the bundle actually written")
}

func TestPack_ImageOverrideReplacesAndAdds(t *testing.T) {
	cfg := packConfig(t) + `
imageOverrides:
  - name: postgresql
    source: harbor.corp.example.com/runai/postgresql:16
  - name: extra-tool
    source: harbor.corp.example.com/runai/extra-tool:1.0
`
	res := run(t, cfg, harness.TestLock(), false)
	require.NoError(t, res.err)

	assert.Contains(t, res.stdout, "1 replaced, 1 added")

	lock := string(tarFile(t, res.output, "artifact-lock.yaml"))
	assert.Contains(t, lock, "source: harbor.corp.example.com/runai/postgresql:16")
	assert.Contains(t, lock, "name: extra-tool")
	assert.NotContains(t, lock, "source: docker.io/bitnamilegacy/postgresql:16", "the replaced source is gone")

	entries := tarEntries(t, res.output)
	assert.Contains(t, entries, "images/runai-control-plane/postgresql/index.json")
	assert.Contains(t, entries, "images/custom/extra-tool/index.json")
}

// The pack attempts every image before reporting, so one run surfaces the whole
// list rather than one failure at a time.
func TestPack_ImageFailuresAreAllReported(t *testing.T) {
	f := newFakes()
	f.puller.PullErrors = map[string]error{
		"nvcr.io/nvidia/runai/assets-service@sha256:9741abc964a10bc41271a1bf2cffd56b4d05664cd78782c71b55469950b572d3": errors.New("unauthorized"),
		"docker.io/bitnamilegacy/postgresql@sha256:8f50ea374acc3e8da24b1c2ab0d2340cb98b7045ca9b25eff95d634e4874991d":  errors.New("not found"),
	}

	res := runFakes(t, packConfig(t), harness.TestLock(), f)
	require.Error(t, res.err)

	assert.Len(t, res.puller.Pulls, 2, "every image is attempted despite the failures")
	assert.Empty(t, res.fetcher.Fetches, "charts are not fetched after the images fail")

	ev := findEvent(res.events, "step_failed", "download-images")
	require.NotNil(t, ev)
	failure := ev["error"].(string)
	assert.Contains(t, failure, "2 of 2 image(s) failed")
	assert.Contains(t, failure, "assets-service: unauthorized")
	assert.Contains(t, failure, "postgresql: not found")

	_, err := os.Stat(res.output)
	assert.True(t, os.IsNotExist(err), "a failed pack must not leave a bundle")
}

func TestPack_ChartDigestMismatchIsRejected(t *testing.T) {
	f := newFakes()
	f.fetcher.Content = map[string][]byte{"control-plane": []byte("tampered")}

	res := runFakes(t, packConfig(t), harness.TestLock(), f)
	require.Error(t, res.err)

	ev := findEvent(res.events, "step_failed", "download-charts")
	require.NotNil(t, ev)
	assert.Contains(t, ev["error"], "chart control-plane: the lock pins")

	_, err := os.Stat(res.output)
	assert.True(t, os.IsNotExist(err))
}

func TestPack_KeepsTheStagingDirectoryOnFailure(t *testing.T) {
	f := newFakes()
	f.fetcher.FetchErrors = map[string]error{"control-plane": errors.New("connection refused")}

	res := runFakes(t, packConfig(t), harness.TestLock(), f)
	require.Error(t, res.err)

	_, err := os.Stat(res.workDir)
	assert.NoError(t, err, "the staging directory is left for diagnosis")
}
