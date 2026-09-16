/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package packconfig

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/releaselock"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/pkg/artifactlock"
)

const LockFlag = "--airgap-lock-file"

type LockFetcher interface {
	Fetch(ctx context.Context, version, arch string) (*releaselock.Result, error)
}

type LockSource struct {
	Path  string
	Tag   string
	Asset string
	URL   string
}

func (c *Config) ResolveLock(ctx context.Context, lockPath string, fetcher LockFetcher) (*artifactlock.ArtifactLock, LockSource, error) {
	if lockPath != "" {
		lock, err := artifactlock.ParseFile(lockPath)
		if err != nil {
			return nil, LockSource{}, err
		}
		return lock, LockSource{Path: lockPath}, nil
	}

	version := strings.TrimSpace(c.Settings.RunaiVersion)
	if version == "" {
		return nil, LockSource{}, fmt.Errorf("settings.runaiVersion is required to look up the release's lock file; set it or provide a lock with %s", LockFlag)
	}
	arch := strings.TrimSpace(c.Settings.Architecture)
	if arch == "" {
		arch = artifactlock.ArchAMD64
	}
	if !artifactlock.IsValidArchitecture(arch) {
		return nil, LockSource{}, fmt.Errorf("unsupported architecture %q: must be %s or %s", arch, artifactlock.ArchAMD64, artifactlock.ArchARM64)
	}

	res, err := fetcher.Fetch(ctx, version, arch)
	if err != nil {
		return nil, LockSource{}, fmt.Errorf("could not fetch lock file for version %s: %s. Please provide a lock file using %s", version, lookupFailure(err), LockFlag)
	}

	c.Settings.RunaiVersion = res.Lock.Metadata.Version
	c.Settings.Architecture = arch
	return res.Lock, LockSource{Tag: res.Tag, Asset: res.Asset, URL: res.URL}, nil
}

func lookupFailure(err error) string {
	var notFound *releaselock.NotFoundError
	switch {
	case errors.Is(err, releaselock.ErrNotReleaseVersion):
		return "it is not a release version, so it has no published lock file"
	case errors.As(err, &notFound) && notFound.Tag == "":
		return "release not found"
	case errors.As(err, &notFound):
		return fmt.Sprintf("no lock files in release %s (expected %s)", notFound.Tag, notFound.Asset)
	}
	return err.Error()
}
