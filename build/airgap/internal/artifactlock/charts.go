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
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dsx-ai-factory/run-ai/build/airgap/pkg/lock"
	"gopkg.in/yaml.v3"
)

type HelmRunner interface {
	Pull(context.Context, string, string, string, string) error
}

type ExecHelmRunner struct {
	Binary string
}

type chartDependencyDefinition struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
}

type chartDefinition struct {
	Name         string                      `yaml:"name"`
	Version      string                      `yaml:"version"`
	Dependencies []chartDependencyDefinition `yaml:"dependencies"`
}

type embeddedDependency struct {
	Path    string
	Version string
	Digest  string
}

// A dependency may declare a version range rather than an exact version, in
// which case only the embedded chart knows what the build resolved it to.
var exactVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+([-+].*)?$`)

type archiveFile struct {
	Name     string
	Contents []byte
	Mode     os.FileMode
}

func (runner ExecHelmRunner) Pull(ctx context.Context, source, destination, username, password string) error {
	binary := runner.Binary
	if binary == "" {
		binary = "helm"
	}
	args := []string{"pull", source, "--destination", destination}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	command := exec.CommandContext(ctx, binary, args...)
	if err := command.Run(); err != nil {
		return fmt.Errorf("helm pull failed for %s", source)
	}
	return nil
}

func LockCharts(ctx context.Context, sources, versions []string, username, password string, runner HelmRunner) ([]lock.Chart, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("at least one Helm chart source is required")
	}
	if len(sources) != len(versions) {
		return nil, fmt.Errorf("helm chart sources and versions must have the same length")
	}
	workspace, err := os.MkdirTemp("", "airgap-charts-")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.RemoveAll(workspace)
	}()

	charts := make([]lock.Chart, 0, len(sources))
	names := map[string]struct{}{}
	for index, source := range sources {
		version := versions[index]
		if version == "" {
			return nil, fmt.Errorf("helm chart version is required for %s", source)
		}
		destination := filepath.Join(workspace, fmt.Sprintf("%d", index))
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return nil, err
		}
		if err := runner.Pull(ctx, source, destination, username, password); err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(destination)
		if err != nil {
			return nil, err
		}
		var archives []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tgz") {
				archives = append(archives, filepath.Join(destination, entry.Name()))
			}
		}
		if len(archives) != 1 {
			return nil, fmt.Errorf("helm pull produced %d archives for %s", len(archives), source)
		}
		contents, err := os.ReadFile(archives[0])
		if err != nil {
			return nil, err
		}
		definition, files, root, err := readChart(contents)
		if err != nil {
			return nil, fmt.Errorf("read chart %s: %w", source, err)
		}
		if definition.Version != version {
			return nil, fmt.Errorf("chart %s has version %s, expected %s", definition.Name, definition.Version, version)
		}
		if _, exists := names[definition.Name]; exists {
			return nil, fmt.Errorf("helm chart names must be unique: %s", definition.Name)
		}
		names[definition.Name] = struct{}{}
		dependencies, err := lockChartDependencies(definition.Dependencies, files, root)
		if err != nil {
			return nil, fmt.Errorf("chart %s: %w", definition.Name, err)
		}
		charts = append(charts, lock.Chart{
			Name:         definition.Name,
			Version:      definition.Version,
			Source:       source,
			BundlePath:   "charts/" + filepath.Base(archives[0]),
			Digest:       lock.Digest(contents),
			Dependencies: dependencies,
		})
	}
	sort.Slice(charts, func(i, j int) bool {
		return charts[i].Name < charts[j].Name
	})
	return charts, nil
}

func readChart(contents []byte) (chartDefinition, map[string][]byte, string, error) {
	entries, err := readTarGzipEntries(contents)
	if err != nil {
		return chartDefinition{}, nil, "", err
	}
	files := map[string][]byte{}
	var chartPaths []string
	for _, entry := range entries {
		files[entry.Name] = entry.Contents
		parts := strings.Split(entry.Name, "/")
		if len(parts) == 2 && parts[1] == "Chart.yaml" {
			chartPaths = append(chartPaths, entry.Name)
		}
	}
	if len(chartPaths) != 1 {
		return chartDefinition{}, nil, "", fmt.Errorf("unable to identify Chart.yaml in chart archive")
	}
	var definition chartDefinition
	if err := yaml.Unmarshal(files[chartPaths[0]], &definition); err != nil {
		return chartDefinition{}, nil, "", err
	}
	if definition.Name == "" || definition.Version == "" {
		return chartDefinition{}, nil, "", fmt.Errorf("chart name and version are required")
	}
	return definition, files, strings.Split(chartPaths[0], "/")[0], nil
}

func lockChartDependencies(definitions []chartDependencyDefinition, files map[string][]byte, root string) ([]lock.ChartDependency, error) {
	embedded := map[string]embeddedDependency{}
	byName := map[string][]string{}
	prefix := root + "/charts/"
	for name, contents := range files {
		parts := strings.Split(name, "/")
		if len(parts) == 3 && parts[0] == root && parts[1] == "charts" && strings.HasSuffix(parts[2], ".tgz") {
			definition, _, _, err := readChart(contents)
			if err != nil {
				return nil, err
			}
			key := dependencyKey(definition.Name, definition.Version)
			if _, exists := embedded[key]; exists {
				return nil, fmt.Errorf("duplicate Helm dependency %s %s", definition.Name, definition.Version)
			}
			embedded[key] = embeddedDependency{
				Path:    "charts/" + parts[2],
				Version: definition.Version,
				Digest:  lock.Digest(contents),
			}
			byName[definition.Name] = append(byName[definition.Name], key)
		}
		if len(parts) == 4 && parts[0] == root && parts[1] == "charts" && parts[3] == "Chart.yaml" {
			var definition chartDefinition
			if err := yaml.Unmarshal(contents, &definition); err != nil {
				return nil, err
			}
			key := dependencyKey(definition.Name, definition.Version)
			if _, exists := embedded[key]; exists {
				return nil, fmt.Errorf("duplicate Helm dependency %s %s", definition.Name, definition.Version)
			}
			directoryPrefix := prefix + parts[2] + "/"
			digest, err := directoryDigest(files, directoryPrefix)
			if err != nil {
				return nil, err
			}
			embedded[key] = embeddedDependency{
				Path:    "charts/" + parts[2],
				Version: definition.Version,
				Digest:  digest,
			}
			byName[definition.Name] = append(byName[definition.Name], key)
		}
	}

	dependencies := make([]lock.ChartDependency, 0, len(definitions))
	expected := map[string]struct{}{}
	for _, definition := range definitions {
		key := dependencyKey(definition.Name, definition.Version)
		// A range never equals the version it resolved to, so match on name and
		// let the embedded chart supply the version.
		if !exactVersion.MatchString(definition.Version) {
			keys := byName[definition.Name]
			if len(keys) != 1 {
				return nil, fmt.Errorf("version range %s for %s matches %d embedded packages, want exactly one",
					definition.Version, definition.Name, len(keys))
			}
			key = keys[0]
		}
		resolved, exists := embedded[key]
		if !exists {
			return nil, fmt.Errorf("missing Helm dependency package %s %s", definition.Name, definition.Version)
		}
		if _, exists := expected[key]; exists {
			return nil, fmt.Errorf("duplicate Helm dependency definition %s %s", definition.Name, definition.Version)
		}
		expected[key] = struct{}{}
		dependencies = append(dependencies, lock.ChartDependency{
			Name:         definition.Name,
			Version:      resolved.Version,
			Repository:   definition.Repository,
			EmbeddedPath: resolved.Path,
			Digest:       resolved.Digest,
		})
	}
	if len(expected) != len(embedded) {
		var unexpected []string
		for key := range embedded {
			if _, exists := expected[key]; !exists {
				unexpected = append(unexpected, strings.ReplaceAll(key, "\x00", " "))
			}
		}
		sort.Strings(unexpected)
		return nil, fmt.Errorf("unexpected Helm dependency packages: %s", strings.Join(unexpected, ", "))
	}
	sort.Slice(dependencies, func(i, j int) bool {
		return dependencies[i].Name < dependencies[j].Name
	})
	return dependencies, nil
}

func readTarGzip(contents []byte) (map[string][]byte, error) {
	entries, err := readTarGzipEntries(contents)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	for _, entry := range entries {
		files[entry.Name] = entry.Contents
	}
	return files, nil
}

func readTarGzipEntries(contents []byte) ([]archiveFile, error) {
	compressed, err := gzip.NewReader(bytes.NewReader(contents))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = compressed.Close()
	}()
	archive := tar.NewReader(compressed)
	var files []archiveFile
	for {
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if !header.FileInfo().Mode().IsRegular() {
			continue
		}
		name := strings.TrimPrefix(path.Clean(header.Name), "./")
		contents, err := io.ReadAll(archive)
		if err != nil {
			return nil, err
		}
		files = append(files, archiveFile{Name: name, Contents: contents, Mode: header.FileInfo().Mode()})
	}
	return files, nil
}

func directoryDigest(files map[string][]byte, prefix string) (string, error) {
	var names []string
	for name := range files {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "", fmt.Errorf("empty Helm dependency directory %s", prefix)
	}
	sort.Strings(names)
	digest := sha256.New()
	for _, name := range names {
		digest.Write([]byte(strings.TrimPrefix(name, prefix)))
		digest.Write([]byte{0})
		digest.Write(files[name])
		digest.Write([]byte{0})
	}
	return fmt.Sprintf("sha256:%x", digest.Sum(nil)), nil
}

func dependencyKey(name, version string) string {
	return name + "\x00" + version
}
