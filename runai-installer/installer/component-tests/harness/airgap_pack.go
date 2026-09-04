/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
)

// FakeImagePuller records pulls and digest resolutions, writing a marker file
// into each layout so bundle contents can be asserted on.
type FakeImagePuller struct {
	// PullErrors and ResolveErrors fail one reference, keyed by image reference.
	PullErrors    map[string]error
	ResolveErrors map[string]error
	// Digests overrides the digest a reference resolves to, keyed by reference.
	Digests map[string]string

	mu       sync.Mutex
	Pulls    []string
	Resolves []string
}

func (f *FakeImagePuller) Pull(_ context.Context, ref, dir string, _ artifactlock.Platform) error {
	f.mu.Lock()
	f.Pulls = append(f.Pulls, ref)
	f.mu.Unlock()

	if err := f.PullErrors[ref]; err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "blobs", "sha256"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`), 0o644)
}

func (f *FakeImagePuller) ResolveDigests(_ context.Context, ref string, _ artifactlock.Platform) (string, string, error) {
	f.mu.Lock()
	f.Resolves = append(f.Resolves, ref)
	f.mu.Unlock()

	if err := f.ResolveErrors[ref]; err != nil {
		return "", "", err
	}
	if digest, ok := f.Digests[ref]; ok {
		return digest, digest, nil
	}
	digest := fmt.Sprintf("sha256:%064x", len(ref))
	return digest, digest, nil
}

// FakeChartFetcher writes canned archive bytes, so a digest mismatch can be
// staged without a registry.
type FakeChartFetcher struct {
	// Content is the archive bytes served for a chart, keyed by chart name.
	// A chart with no entry gets bytes matching its locked digest.
	Content map[string][]byte
	// FetchErrors fails one chart, keyed by chart name.
	FetchErrors map[string]error
	// ProbeErrors fails one chart's preflight probe, keyed by chart name.
	ProbeErrors map[string]error

	mu      sync.Mutex
	Fetches []string
	Probes  []string
}

type FakeCLIFetcher struct {
	FetchErrors map[string]error
	ProbeErrors map[string]error
	Fetches     []string
	Probes      []string
}

func (f *FakeCLIFetcher) Probe(_ context.Context, cli artifactlock.CLI) error {
	f.Probes = append(f.Probes, cli.Name)
	return f.ProbeErrors[cli.Name]
}

func (f *FakeCLIFetcher) Fetch(_ context.Context, cli artifactlock.CLI, workDir string) error {
	f.Fetches = append(f.Fetches, cli.Name)
	if err := f.FetchErrors[cli.Name]; err != nil {
		return err
	}
	files := cli.Files
	if len(files) == 0 && cli.SourcePath != "" {
		files = []artifactlock.CLIFile{{SourcePath: cli.SourcePath, BundlePath: cli.BundlePath, Digest: cli.Digest, Mode: 0o755}}
	}
	for _, file := range files {
		path := filepath.Join(workDir, "bundle", filepath.FromSlash(file.BundlePath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, CLIContent(cli.Name, file.SourcePath), os.FileMode(file.Mode)); err != nil {
			return err
		}
	}
	return nil
}

func (f *FakeChartFetcher) Probe(_ context.Context, chart artifactlock.Chart) error {
	f.mu.Lock()
	f.Probes = append(f.Probes, chart.Name)
	f.mu.Unlock()

	return f.ProbeErrors[chart.Name]
}

func (f *FakeChartFetcher) Fetch(_ context.Context, chart artifactlock.Chart, dest string) error {
	f.mu.Lock()
	f.Fetches = append(f.Fetches, chart.Name)
	f.mu.Unlock()

	if err := f.FetchErrors[chart.Name]; err != nil {
		return err
	}
	content, ok := f.Content[chart.Name]
	if !ok {
		content = ChartContent(chart.Name)
	}
	return os.WriteFile(dest, content, 0o644)
}

// ChartContent is the archive a lock's chart digest is computed over, so a
// fetched chart verifies against the lock.
func ChartContent(name string) []byte {
	var buffer bytes.Buffer
	compressor := gzip.NewWriter(&buffer)
	archive := tar.NewWriter(compressor)
	files := []struct {
		name    string
		content string
	}{
		{name + "/Chart.yaml", fmt.Sprintf("apiVersion: v2\nname: %s\nversion: 0.0.0-test\n", name)},
		{name + "/values.yaml", chartValues(name)},
	}
	for _, file := range files {
		if err := archive.WriteHeader(&tar.Header{Name: file.name, Mode: 0o644, Size: int64(len(file.content))}); err != nil {
			panic(err)
		}
		if _, err := archive.Write([]byte(file.content)); err != nil {
			panic(err)
		}
	}
	if err := archive.Close(); err != nil {
		panic(err)
	}
	if err := compressor.Close(); err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

func chartValues(name string) string {
	if name == "control-plane" {
		return `thirdPartyImages:
  keycloak: keycloakx.image
  nats: nats.image
  postgres: global.postgresImage
  postgresql: postgresql.image
  thanos: thanos.image
  traefik: traefik.image
  wait-ubi-minimal: global.waitImage
global:
  postgresImage: {registry: docker.io, repository: postgres, tag: 16-alpine}
  waitImage: {registry: registry.access.redhat.com, repository: ubi9/ubi-minimal, tag: "9.8"}
keycloakx:
  image: {repository: quay.io/keycloak/keycloak, tag: "26.6"}
nats:
  image: {registry: docker.io, repository: bitnamilegacy/nats, tag: 2.11.8-debian-12-r0}
postgresql:
  image: {registry: docker.io, repository: bitnamilegacy/postgresql, tag: "16"}
thanos:
  image: {registry: docker.io, repository: bitnamilegacy/thanos, tag: 0.38.0-debian-12-r3}
traefik:
  image: {registry: docker.io, repository: traefik, tag: v3.7}
`
	}
	return `thirdPartyImages:
  git-sync: {registry: registry.k8s.io, repository: git-sync/git-sync, tag: v4.7.0}
  valkey: {registry: docker.io, repository: valkey/valkey, tag: 9.0.4}
  wait-ubi-minimal: {registry: registry.access.redhat.com, repository: ubi9/ubi-minimal, tag: "9.8"}
`
}

// ChartDigest is the digest of a chart's canned content, for building a lock
// whose charts verify.
func ChartDigest(name string) string {
	return fmt.Sprintf("sha256:%x", sha256.Sum256(ChartContent(name)))
}
