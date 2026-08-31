/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package artifacts

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/pkg/artifactlock"
	"gopkg.in/yaml.v3"
)

const (
	bundleDirName = "bundle"
	lockFileName  = "artifact-lock.yaml"
	imagesDirName = "images"
	chartsDirName = "charts"
	cliDirName    = "cli"
)

// BundleDir is the staged bundle inside the working directory: its contents are
// what the bundle tar holds.
func BundleDir(workDir string) string { return filepath.Join(workDir, bundleDirName) }

// ImageDir is the OCI layout directory of one image inside the bundle.
func ImageDir(workDir, group, name string) string {
	return filepath.Join(BundleDir(workDir), imagesDirName, group, name)
}

// ChartPath resolves a chart's bundlePath inside the bundle.
func ChartPath(workDir, bundlePath string) string {
	return filepath.Join(BundleDir(workDir), filepath.FromSlash(bundlePath))
}

func CLIPath(workDir, bundlePath string) (string, error) {
	root := BundleDir(workDir)
	resolved := filepath.Join(root, filepath.FromSlash(bundlePath))
	relative, err := filepath.Rel(root, resolved)
	if err != nil {
		return "", err
	}
	if relative == ".." || filepath.IsAbs(relative) || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("CLI bundle path %q escapes the bundle", bundlePath)
	}
	return resolved, nil
}

// LockedChartPath resolves and verifies a chart from the extracted artifact lock.
func LockedChartPath(workDir, name string) (string, error) {
	lock, err := ReadLock(workDir)
	if err != nil {
		return "", err
	}
	for _, chart := range lock.Spec.Charts {
		if chart.Name != name {
			continue
		}
		path := ChartPath(workDir, chart.BundlePath)
		digest, err := FileDigest(path)
		if err != nil {
			return "", fmt.Errorf("read chart %s: %w", name, err)
		}
		if chart.Digest != "" && digest != chart.Digest {
			return "", fmt.Errorf("chart %s has digest %s but the lock pins %s", name, digest, chart.Digest)
		}
		return path, nil
	}
	return "", fmt.Errorf("artifact lock does not contain chart %q", name)
}

// LockedChartPreviewPath returns the dry-run path for a locked chart.
func LockedChartPreviewPath(workDir, name string) string {
	return filepath.Join(BundleDir(workDir), chartsDirName, name+".tgz")
}

// InitBundleDir creates an empty bundle, discarding anything a previous run
// staged so a bundle never mixes two runs' artifacts.
func InitBundleDir(workDir string) error {
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("clear working directory %s: %w", workDir, err)
	}
	for _, dir := range []string{imagesDirName, chartsDirName, cliDirName} {
		if err := os.MkdirAll(filepath.Join(BundleDir(workDir), dir), 0o755); err != nil {
			return fmt.Errorf("create working directory %s: %w", workDir, err)
		}
	}
	return nil
}

// WriteLock stores the effective lock in the bundle, where it is both the
// later steps' input and the record of what the bundle holds.
func WriteLock(workDir string, lock *artifactlock.ArtifactLock) error {
	encoded, err := yaml.Marshal(lock)
	if err != nil {
		return fmt.Errorf("encode artifact lock: %w", err)
	}
	return os.WriteFile(filepath.Join(BundleDir(workDir), lockFileName), encoded, 0o644)
}

// ReadLock loads the effective lock a previous step staged.
func ReadLock(workDir string) (*artifactlock.ArtifactLock, error) {
	path := filepath.Join(BundleDir(workDir), lockFileName)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read staged artifact lock: %w", err)
	}
	var lock artifactlock.ArtifactLock
	if err := yaml.Unmarshal(content, &lock); err != nil {
		return nil, fmt.Errorf("parse staged artifact lock: %w", err)
	}
	if err := lock.ValidateImages(); err != nil {
		return nil, fmt.Errorf("validate staged artifact lock: %w", err)
	}
	if err := lock.ValidateCLIs(); err != nil {
		return nil, fmt.Errorf("validate staged artifact lock: %w", err)
	}
	return &lock, nil
}

// TarDir writes the directory's contents to a tar archive, returning the
// archive's sha256.
func TarDir(ctx context.Context, dir, output string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	out, err := os.Create(output)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", output, err)
	}
	defer out.Close()

	sum := sha256.New()
	tw := tar.NewWriter(io.MultiWriter(out, sum))

	walkErr := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil || rel == "." {
			return err
		}
		return writeTarEntry(tw, path, filepath.ToSlash(rel), entry)
	})
	if walkErr != nil {
		return "", fmt.Errorf("write %s: %w", output, walkErr)
	}
	if err := tw.Close(); err != nil {
		return "", fmt.Errorf("write %s: %w", output, err)
	}
	return "sha256:" + hex.EncodeToString(sum.Sum(nil)), nil
}

func writeTarEntry(tw *tar.Writer, path, name string, entry fs.DirEntry) error {
	info, err := entry.Info()
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = name
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if entry.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(tw, file)
	return err
}

// WriteChecksum writes the bundle's checksum sidecar, in the format sha256sum
// verifies.
func WriteChecksum(output, digest string) error {
	line := fmt.Sprintf("%s  %s\n", trimAlgorithm(digest), filepath.Base(output))
	return os.WriteFile(output+".sha256", []byte(line), 0o644)
}

// FileDigest returns the sha256 of a file, in the lock's digest format.
func FileDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	sum := sha256.New()
	if _, err := io.Copy(sum, file); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(sum.Sum(nil)), nil
}

func trimAlgorithm(digest string) string {
	if len(digest) > 7 && digest[:7] == "sha256:" {
		return digest[7:]
	}
	return digest
}
