/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package lock

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	APIVersion   = "artifacts.run.ai/v1alpha1"
	ImageKind    = "ImageLock"
	ArtifactKind = "ArtifactLock"
)

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

type Metadata struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type Platform struct {
	OS           string `yaml:"os"`
	Architecture string `yaml:"architecture"`
}

func (p Platform) String() string {
	return p.OS + "/" + p.Architecture
}

type Image struct {
	Name        string `yaml:"name"`
	Image       string `yaml:"image"`
	Source      string `yaml:"source"`
	IndexDigest string `yaml:"indexDigest"`
}

type ImageLockSpec struct {
	Profile  string   `yaml:"profile"`
	Platform Platform `yaml:"platform"`
	Images   []Image  `yaml:"images"`
}

type ImageLock struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   Metadata      `yaml:"metadata"`
	Spec       ImageLockSpec `yaml:"spec"`
}

type CollectedImageLock struct {
	Name    string  `yaml:"name"`
	Version string  `yaml:"version"`
	Digest  string  `yaml:"digest"`
	Images  []Image `yaml:"images"`
}

type ChartDependency struct {
	Name         string `yaml:"name"`
	Version      string `yaml:"version"`
	Repository   string `yaml:"repository"`
	EmbeddedPath string `yaml:"embeddedPath"`
	Digest       string `yaml:"digest"`
}

type Chart struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Source       string            `yaml:"source"`
	BundlePath   string            `yaml:"bundlePath"`
	Digest       string            `yaml:"digest"`
	Dependencies []ChartDependency `yaml:"dependencies"`
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

type ArtifactLockSpec struct {
	Profile    string               `yaml:"profile"`
	Platform   Platform             `yaml:"platform"`
	ImageLocks []CollectedImageLock `yaml:"imageLocks"`
	Charts     []Chart              `yaml:"charts"`
	CLIs       []CLI                `yaml:"clis"`
}

type ArtifactLock struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   Metadata         `yaml:"metadata"`
	Spec       ArtifactLockSpec `yaml:"spec"`
}

func ParseImageLock(contents []byte) (ImageLock, error) {
	var value ImageLock
	if err := decode(contents, &value); err != nil {
		return ImageLock{}, err
	}
	if err := value.Validate(); err != nil {
		return ImageLock{}, err
	}
	return value, nil
}

func ParseArtifactLock(contents []byte) (ArtifactLock, error) {
	var value ArtifactLock
	if err := decode(contents, &value); err != nil {
		return ArtifactLock{}, err
	}
	if err := value.Validate(); err != nil {
		return ArtifactLock{}, err
	}
	return value, nil
}

func Marshal(value any) ([]byte, error) {
	return yaml.Marshal(value)
}

func Digest(contents []byte) string {
	sum := sha256.Sum256(contents)
	return fmt.Sprintf("sha256:%x", sum)
}

func (l ImageLock) Validate() error {
	if l.APIVersion != APIVersion || l.Kind != ImageKind {
		return fmt.Errorf("expected %s %s", APIVersion, ImageKind)
	}
	if err := validateMetadata(l.Metadata); err != nil {
		return err
	}
	if err := validateProfile(l.Spec.Profile); err != nil {
		return err
	}
	if err := validatePlatform(l.Spec.Platform); err != nil {
		return err
	}
	if len(l.Spec.Images) == 0 {
		return errors.New("image lock has no images")
	}
	return validateImages(l.Spec.Images)
}

func (l ArtifactLock) Validate() error {
	if l.APIVersion != APIVersion || l.Kind != ArtifactKind {
		return fmt.Errorf("expected %s %s", APIVersion, ArtifactKind)
	}
	if err := validateMetadata(l.Metadata); err != nil {
		return err
	}
	if err := validateProfile(l.Spec.Profile); err != nil {
		return err
	}
	if err := validatePlatform(l.Spec.Platform); err != nil {
		return err
	}
	if len(l.Spec.ImageLocks) == 0 {
		return errors.New("artifact lock has no image locks")
	}
	names := map[string]struct{}{}
	for _, imageLock := range l.Spec.ImageLocks {
		if imageLock.Name == "" || imageLock.Version == "" || imageLock.Version == "latest" {
			return errors.New("invalid collected image lock metadata")
		}
		if !digestPattern.MatchString(imageLock.Digest) {
			return fmt.Errorf("invalid image lock digest %q", imageLock.Digest)
		}
		if len(imageLock.Images) == 0 {
			return fmt.Errorf("image lock %q has no images", imageLock.Name)
		}
		if _, exists := names[imageLock.Name]; exists {
			return fmt.Errorf("duplicate image lock name %q", imageLock.Name)
		}
		names[imageLock.Name] = struct{}{}
		if err := validateImages(imageLock.Images); err != nil {
			return fmt.Errorf("image lock %q: %w", imageLock.Name, err)
		}
	}
	for _, chart := range l.Spec.Charts {
		if chart.Name == "" || chart.Version == "" || chart.Source == "" {
			return errors.New("invalid chart metadata")
		}
		if !strings.HasPrefix(chart.BundlePath, "charts/") || strings.Count(chart.BundlePath, "/") != 1 || !strings.HasSuffix(chart.BundlePath, ".tgz") {
			return fmt.Errorf("invalid chart bundle path %q", chart.BundlePath)
		}
		if !digestPattern.MatchString(chart.Digest) {
			return fmt.Errorf("invalid chart digest %q", chart.Digest)
		}
		for _, dependency := range chart.Dependencies {
			if dependency.Name == "" || dependency.Version == "" || dependency.Repository == "" || dependency.EmbeddedPath == "" {
				return fmt.Errorf("invalid dependency in chart %q", chart.Name)
			}
			if !digestPattern.MatchString(dependency.Digest) {
				return fmt.Errorf("invalid chart dependency digest %q", dependency.Digest)
			}
		}
	}
	for _, cli := range l.Spec.CLIs {
		if !validCLIName(cli.Name) || cli.Version == "" || cli.Version == "latest" || cli.Source == "" || !digestPattern.MatchString(cli.SourceDigest) {
			return errors.New("invalid CLI metadata")
		}
		files := cli.Files
		if len(files) == 0 && cli.SourcePath != "" {
			if cli.BundlePath != "cli/"+cli.Name || !digestPattern.MatchString(cli.Digest) {
				return fmt.Errorf("invalid CLI file for %q", cli.Name)
			}
			continue
		}
		if len(files) == 0 {
			return fmt.Errorf("invalid CLI digest for %q", cli.Name)
		}
		sourcePaths := map[string]bool{}
		bundlePaths := map[string]bool{}
		for _, file := range files {
			if !validRelativePath(file.SourcePath) || !validCLIBundlePath(cli.Name, file.BundlePath) {
				return fmt.Errorf("invalid CLI file path for %q", cli.Name)
			}
			if sourcePaths[file.SourcePath] || bundlePaths[file.BundlePath] || !digestPattern.MatchString(file.Digest) || file.Mode&0o111 == 0 {
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

func validRelativePath(value string) bool {
	clean := path.Clean(value)
	return value != "" && value == clean && !path.IsAbs(clean) && clean != ".." && !strings.HasPrefix(clean, "../")
}

func validCLIBundlePath(name, value string) bool {
	prefix := "cli/" + name + "/"
	return validRelativePath(value) && strings.HasPrefix(value, prefix) && len(value) > len(prefix)
}

func decode(contents []byte, destination any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple YAML documents")
		}
		return err
	}
	return nil
}

func validateMetadata(metadata Metadata) error {
	if metadata.Name == "" || metadata.Version == "" {
		return errors.New("lock metadata name and version are required")
	}
	if metadata.Version == "latest" {
		return errors.New("latest is not a valid lock version")
	}
	return nil
}

func validateProfile(profile string) error {
	if profile != "standard" && profile != "fips" {
		return fmt.Errorf("invalid profile %q", profile)
	}
	return nil
}

func validatePlatform(platform Platform) error {
	if platform.OS == "" || platform.Architecture == "" {
		return errors.New("platform OS and architecture are required")
	}
	return nil
}

func validateImages(images []Image) error {
	names := map[string]struct{}{}
	for _, image := range images {
		if image.Name == "" || image.Source == "" {
			return errors.New("image name and source are required")
		}
		if _, exists := names[image.Name]; exists {
			return fmt.Errorf("duplicate image name %q", image.Name)
		}
		names[image.Name] = struct{}{}
		if !digestPattern.MatchString(image.IndexDigest) {
			return fmt.Errorf("invalid index digest %q", image.IndexDigest)
		}
		at := strings.LastIndex(image.Image, "@")
		if at < 1 || !digestPattern.MatchString(image.Image[at+1:]) {
			return fmt.Errorf("invalid immutable image %q", image.Image)
		}
	}
	return nil
}
