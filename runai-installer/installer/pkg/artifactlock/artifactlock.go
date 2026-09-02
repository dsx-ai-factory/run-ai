/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package artifactlock defines the artifact lock Run:ai publishes for a
// release: the exact images, charts and CLI binaries an air-gapped bundle is
// built from. It mirrors the schema produced by build/airgap/artifact-lock.
package artifactlock

import (
	"fmt"
	"path"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

const (
	APIVersion = "artifacts.run.ai/v1alpha1"
	Kind       = "ArtifactLock"
	OSLinux    = "linux"

	ArchAMD64 = "amd64"
	ArchARM64 = "arm64"
)

type ArtifactLock struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Spec struct {
	Profile    string      `yaml:"profile"`
	Platform   Platform    `yaml:"platform"`
	ImageLocks []ImageLock `yaml:"imageLocks"`
	Charts     []Chart     `yaml:"charts"`
	CLIs       []CLI       `yaml:"clis"`
}

type Platform struct {
	OS           string `yaml:"os"`
	Architecture string `yaml:"architecture"`
}

type ImageLock struct {
	Name    string  `yaml:"name"`
	Version string  `yaml:"version"`
	Digest  string  `yaml:"digest"`
	Images  []Image `yaml:"images"`
}

type Image struct {
	Name        string `yaml:"name"`
	Image       string `yaml:"image"`
	Source      string `yaml:"source"`
	IndexDigest string `yaml:"indexDigest"`
}

type ImageEntry struct {
	Group        string
	GroupVersion string
	Image        Image
}

type Chart struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Source       string            `yaml:"source"`
	BundlePath   string            `yaml:"bundlePath"`
	Digest       string            `yaml:"digest"`
	Dependencies []ChartDependency `yaml:"dependencies"`
}

type ChartDependency struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Repository   string `yaml:"repository"`
	EmbeddedPath string `yaml:"embeddedPath"`
	Digest       string `yaml:"digest"`
}

type CLI struct {
	Name         string    `yaml:"name"`
	Version      string    `yaml:"version"`
	Source       string    `yaml:"source"`
	SourceDigest string    `yaml:"sourceDigest"`
	Files        []CLIFile `yaml:"files,omitempty"`
	SourcePath   string    `yaml:"sourcePath,omitempty"`
	BundlePath   string    `yaml:"bundlePath,omitempty"`
	Digest       string    `yaml:"digest,omitempty"`
}

type CLIFile struct {
	SourcePath string `yaml:"sourcePath"`
	BundlePath string `yaml:"bundlePath"`
	Digest     string `yaml:"digest"`
	Mode       uint32 `yaml:"mode"`
}

// Images returns every locked image across all image-lock groups.
func (l *ArtifactLock) Images() []Image {
	var out []Image
	for _, group := range l.Spec.ImageLocks {
		out = append(out, group.Images...)
	}
	return out
}

func (l *ArtifactLock) ImageEntries() []ImageEntry {
	var out []ImageEntry
	for _, group := range l.Spec.ImageLocks {
		for _, image := range group.Images {
			out = append(out, ImageEntry{Group: group.Name, GroupVersion: group.Version, Image: image})
		}
	}
	return out
}

func (l *ArtifactLock) ValidateImages() error {
	groups := map[string]bool{}
	for _, group := range l.Spec.ImageLocks {
		if !validPathComponent(group.Name) {
			return fmt.Errorf("invalid image lock group %q", group.Name)
		}
		if groups[group.Name] {
			return fmt.Errorf("duplicate image lock group %q", group.Name)
		}
		groups[group.Name] = true

		images := map[string]bool{}
		for _, image := range group.Images {
			if !validPathComponent(image.Name) {
				return fmt.Errorf("invalid image name %q in group %q", image.Name, group.Name)
			}
			if images[image.Name] {
				return fmt.Errorf("duplicate image %q in group %q", image.Name, group.Name)
			}
			images[image.Name] = true
			if _, err := ImageDigest(image.Image); err != nil {
				return fmt.Errorf("image %q in group %q: %w", image.Name, group.Name, err)
			}
			if !validSHA256(image.IndexDigest) {
				return fmt.Errorf("image %q in group %q has invalid index digest %q", image.Name, group.Name, image.IndexDigest)
			}
		}
	}
	return nil
}

func (l *ArtifactLock) ValidateCLIs() error {
	for _, cli := range l.Spec.CLIs {
		if !validCLIName(cli.Name) || cli.Version == "" || cli.Version == "latest" || cli.Source == "" || !validSHA256(cli.SourceDigest) {
			return fmt.Errorf("invalid CLI metadata")
		}
		files := cli.Files
		if len(files) == 0 && cli.SourcePath != "" {
			if cli.BundlePath != "cli/"+cli.Name || !validSHA256(cli.Digest) {
				return fmt.Errorf("invalid CLI file for %q", cli.Name)
			}
			continue
		}
		if len(files) == 0 {
			return fmt.Errorf("CLI %q has no files", cli.Name)
		}
		sourcePaths := map[string]bool{}
		bundlePaths := map[string]bool{}
		for _, file := range files {
			if !validRelativeCLIPath(file.SourcePath) || !validCLIBundlePath(cli.Name, file.BundlePath) || sourcePaths[file.SourcePath] || bundlePaths[file.BundlePath] || !validSHA256(file.Digest) || file.Mode&0o111 == 0 {
				return fmt.Errorf("invalid CLI file for %q", cli.Name)
			}
			sourcePaths[file.SourcePath] = true
			bundlePaths[file.BundlePath] = true
		}
	}
	return nil
}

func validCLIName(name string) bool {
	return name != "" && name != "." && name != ".." && path.Base(name) == name
}

func validRelativeCLIPath(value string) bool {
	clean := path.Clean(value)
	return value != "" && value == clean && !path.IsAbs(clean) && clean != ".." && !strings.HasPrefix(clean, "../")
}

func validCLIBundlePath(name, value string) bool {
	prefix := "cli/" + name + "/"
	return validRelativeCLIPath(value) && strings.HasPrefix(value, prefix) && len(value) > len(prefix)
}

func ImageDigest(ref string) (string, error) {
	parsed, err := name.ParseReference(ref)
	if err != nil {
		return "", fmt.Errorf("invalid immutable image %q: %w", ref, err)
	}
	digest, ok := parsed.(name.Digest)
	if !ok || !validSHA256(digest.DigestStr()) {
		return "", fmt.Errorf("image %q must use an immutable sha256 digest", ref)
	}
	return digest.DigestStr(), nil
}

func validSHA256(value string) bool {
	hash, err := v1.NewHash(value)
	return err == nil && hash.Algorithm == "sha256"
}

func validPathComponent(value string) bool {
	if value == "" || strings.Contains(value, "/") {
		return false
	}
	_, err := name.NewRepository("registry.invalid/" + value)
	return err == nil
}

// IsValidArchitecture reports whether arch is an architecture locks are built for.
func IsValidArchitecture(arch string) bool {
	return arch == ArchAMD64 || arch == ArchARM64
}
