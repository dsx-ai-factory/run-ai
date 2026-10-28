/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifactlock

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type fakeHelmRunner struct {
	archive  []byte
	filename string
	calls    int
}

type chartFixture struct {
	archive  []byte
	filename string
}

type multiHelmRunner map[string]chartFixture

func (runner *fakeHelmRunner) Pull(_ context.Context, _, destination, _, _ string) error {
	runner.calls++
	return os.WriteFile(filepath.Join(destination, runner.filename), runner.archive, 0o644)
}

func (runner multiHelmRunner) Pull(_ context.Context, source, destination, _, _ string) error {
	fixture := runner[source]
	return os.WriteFile(filepath.Join(destination, fixture.filename), fixture.archive, 0o644)
}

func TestLockChartsLocksPackagedDependencies(t *testing.T) {
	dependency := chartArchive(t, "dependency", "1.2.3", nil, false)
	root := chartArchive(t, "published", "3.0.0", []chartDependencyDefinition{{
		Name:       "dependency",
		Version:    "1.2.3",
		Repository: "https://charts.example.com",
	}}, false)
	files, err := readTarGzip(root)
	if err != nil {
		t.Fatal(err)
	}
	files["published/charts/dependency-1.2.3.tgz"] = dependency
	root = archiveFiles(t, files)
	runner := &fakeHelmRunner{archive: root, filename: "published-3.0.0.tgz"}

	charts, err := LockCharts(
		context.Background(),
		[]string{"https://charts.example.com/published-3.0.0.tgz"},
		[]string{"3.0.0"},
		"$oauthtoken",
		"secret",
		runner,
	)
	if err != nil {
		t.Fatal(err)
	}
	if charts[0].Name != "published" || charts[0].BundlePath != "charts/published-3.0.0.tgz" {
		t.Fatalf("chart = %#v", charts[0])
	}
	if charts[0].Dependencies[0].EmbeddedPath != "charts/dependency-1.2.3.tgz" {
		t.Fatalf("dependency = %#v", charts[0].Dependencies[0])
	}
	if runner.calls != 1 {
		t.Fatalf("calls = %d", runner.calls)
	}
}

func TestLockChartsLocksUnpackedDependencies(t *testing.T) {
	root := chartArchive(t, "published", "3.0.0", []chartDependencyDefinition{{
		Name:       "dependency",
		Version:    "1.2.3",
		Repository: "https://charts.example.com",
	}}, true)
	runner := &fakeHelmRunner{archive: root, filename: "published-3.0.0.tgz"}

	charts, err := LockCharts(
		context.Background(),
		[]string{"https://charts.example.com/published-3.0.0.tgz"},
		[]string{"3.0.0"},
		"",
		"",
		runner,
	)
	if err != nil {
		t.Fatal(err)
	}
	if charts[0].Dependencies[0].EmbeddedPath != "charts/dependency" {
		t.Fatalf("dependency = %#v", charts[0].Dependencies[0])
	}
	if !strings.HasPrefix(charts[0].Dependencies[0].Digest, "sha256:") {
		t.Fatalf("digest = %s", charts[0].Dependencies[0].Digest)
	}
}

func TestLockChartsAllowsDuplicatePayloadPaths(t *testing.T) {
	definition, err := yaml.Marshal(chartDefinition{Name: "runai-cluster", Version: "3.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	archive := archiveEntries(t, []archiveFile{
		{Name: "runai-cluster/Chart.yaml", Contents: definition},
		{Name: "runai-cluster/charts/crds/values.yaml", Contents: []byte("first: true\n")},
		{Name: "runai-cluster/charts/crds/values.yaml", Contents: []byte("second: true\n")},
	})
	runner := &fakeHelmRunner{archive: archive, filename: "runai-cluster-3.0.0.tgz"}
	if _, err := LockCharts(
		context.Background(),
		[]string{"runai-cluster"},
		[]string{"3.0.0"},
		"",
		"",
		runner,
	); err != nil {
		t.Fatal(err)
	}
}

func TestLockChartsRejectsVersionMismatch(t *testing.T) {
	runner := &fakeHelmRunner{
		archive:  chartArchive(t, "published", "3.0.0", nil, false),
		filename: "published-3.0.0.tgz",
	}
	_, err := LockCharts(context.Background(), []string{"source"}, []string{"9.9.9"}, "", "", runner)
	if err == nil || !strings.Contains(err.Error(), "expected 9.9.9") {
		t.Fatalf("error = %v", err)
	}
}

func TestLockChartsRequiresVersionForEverySource(t *testing.T) {
	_, err := LockCharts(
		context.Background(),
		[]string{"control-plane", "runai-cluster"},
		[]string{"3.0.0"},
		"",
		"",
		&fakeHelmRunner{},
	)
	if err == nil || !strings.Contains(err.Error(), "same length") {
		t.Fatalf("error = %v", err)
	}
}

func TestLockChartsAcceptsIndependentVersions(t *testing.T) {
	runner := multiHelmRunner{
		"control-plane": {
			archive:  chartArchive(t, "control-plane", "3.0.1", nil, false),
			filename: "control-plane-3.0.1.tgz",
		},
		"runai-cluster": {
			archive:  chartArchive(t, "runai-cluster", "3.0.2", nil, false),
			filename: "runai-cluster-3.0.2.tgz",
		},
	}
	charts, err := LockCharts(
		context.Background(),
		[]string{"control-plane", "runai-cluster"},
		[]string{"3.0.1", "3.0.2"},
		"",
		"",
		runner,
	)
	if err != nil {
		t.Fatal(err)
	}
	if charts[0].Version != "3.0.1" || charts[1].Version != "3.0.2" {
		t.Fatalf("charts = %#v", charts)
	}
}

func TestExecHelmRunnerDoesNotExposePassword(t *testing.T) {
	runner := ExecHelmRunner{Binary: filepath.Join(t.TempDir(), "missing")}
	err := runner.Pull(context.Background(), "source", t.TempDir(), "$oauthtoken", "secret")
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatalf("error = %v", err)
	}
}

func chartArchive(t *testing.T, name, version string, dependencies []chartDependencyDefinition, unpacked bool) []byte {
	t.Helper()
	definition, err := yaml.Marshal(chartDefinition{
		Name:         name,
		Version:      version,
		Dependencies: dependencies,
	})
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{name + "/Chart.yaml": definition}
	if unpacked {
		for _, dependency := range dependencies {
			dependencyChart, err := yaml.Marshal(chartDefinition{
				Name:    dependency.Name,
				Version: dependency.Version,
			})
			if err != nil {
				t.Fatal(err)
			}
			files[name+"/charts/"+dependency.Name+"/Chart.yaml"] = dependencyChart
			files[name+"/charts/"+dependency.Name+"/values.yaml"] = []byte("enabled: true\n")
		}
	}
	return archiveFiles(t, files)
}

func archiveFiles(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	entries := make([]archiveFile, 0, len(files))
	for name, contents := range files {
		entries = append(entries, archiveFile{Name: name, Contents: contents})
	}
	return archiveEntries(t, entries)
}

func archiveEntries(t *testing.T, files []archiveFile) []byte {
	t.Helper()
	var buffer bytes.Buffer
	compressed := gzip.NewWriter(&buffer)
	archive := tar.NewWriter(compressed)
	for _, file := range files {
		mode := file.Mode
		if mode == 0 {
			mode = 0o644
		}
		header := &tar.Header{Name: file.Name, Mode: int64(mode.Perm()), Size: int64(len(file.Contents))}
		if err := archive.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := archive.Write(file.Contents); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func TestLockChartsResolvesRangedDependencies(t *testing.T) {
	for _, tc := range []struct {
		name     string
		declared string
		want     string
		wantErr  string
	}{
		{name: "tilde range", declared: "~1.2.0", want: "1.2.3"},
		{name: "caret range", declared: "^1.0.0", want: "1.2.3"},
		{name: "comparator range", declared: ">=1.2.0 <2.0.0", want: "1.2.3"},
		{name: "exact version still matches itself", declared: "1.2.3", want: "1.2.3"},
		{name: "exact version that did not ship", declared: "1.2.4", wantErr: "missing Helm dependency package"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dependency := chartArchive(t, "dependency", "1.2.3", nil, false)
			root := chartArchive(t, "published", "3.0.0", []chartDependencyDefinition{{
				Name:       "dependency",
				Version:    tc.declared,
				Repository: "https://charts.example.com",
			}}, false)
			files, err := readTarGzip(root)
			if err != nil {
				t.Fatal(err)
			}
			files["published/charts/dependency-1.2.3.tgz"] = dependency
			runner := &fakeHelmRunner{archive: archiveFiles(t, files), filename: "published-3.0.0.tgz"}

			charts, err := LockCharts(
				context.Background(),
				[]string{"https://charts.example.com/published-3.0.0.tgz"},
				[]string{"3.0.0"},
				"",
				"",
				runner,
			)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			// The lock records what shipped, never the range that selected it.
			if got := charts[0].Dependencies[0].Version; got != tc.want {
				t.Fatalf("version = %q, want %q", got, tc.want)
			}
		})
	}
}
