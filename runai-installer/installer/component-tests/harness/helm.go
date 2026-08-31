/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package harness exposes shared fakes and utilities for orchestrator-level
// component tests. Helpers live in a regular (non-_test) package so they can
// be imported from any test package under component-tests/.
package harness

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/step"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/helmclient"
	"helm.sh/helm/v3/pkg/release"
)

type HelmCall struct {
	RepoUrl     string
	Version     string
	ReleaseName string
	ChartName   string
	Namespace   string
	IsOci       bool
	Values      []string
	ValueFiles  []string
	KubeVersion string
	Auth        step.ChartAuth
}

// FakeHelmInstaller records every helm invocation made by the orchestrator
// and lets tests seed pre-existing releases or simulate failures. It
// implements both step.HelmInstaller (install/upgrade path) and
// uninstall.HelmWrapper (uninstall path).
type FakeHelmInstaller struct {
	mu                 sync.Mutex
	Calls              []HelmCall
	Releases           map[string]string
	PendingReleases    map[string]bool
	UninstallErrors    map[string]error
	UninstallResponses map[string]*release.UninstallReleaseResponse
	// InstallErrors maps a release name to an error returned by
	// InstallOrUpgradeRelease, letting tests simulate a failing install. The
	// call is still recorded so tests can assert the attempt was made.
	InstallErrors map[string]error
	// Block, when set for a release name, makes InstallOrUpgradeRelease block
	// until the context is cancelled. Tests use it to exercise step timeouts
	// through a real install plan.
	Block map[string]bool
	// BlockAttempts blocks only the first N install attempts for a release
	// (each attempt decrements the counter), letting tests exercise a step that
	// times out and then succeeds on retry.
	BlockAttempts map[string]int
	// BrokenReleases, when set for a release name, makes ReleaseStatus return an
	// error, simulating a release stuck in a pending/uninstalling state.
	BrokenReleases map[string]bool
}

type ReleaseEntry struct {
	Name    string
	Version string
}

func NewFakeHelmWithReleases(entries ...ReleaseEntry) *FakeHelmInstaller {
	releases := make(map[string]string, len(entries))
	for _, e := range entries {
		releases[e.Name] = e.Version
	}
	return &FakeHelmInstaller{Releases: releases}
}

func (f *FakeHelmInstaller) InstallOrUpgradeRelease(ctx context.Context, repoUrl, version, releaseName, chartName, namespace string, isOci bool, values []string) error {
	return f.InstallOrUpgradeReleaseWithAuth(ctx, repoUrl, version, releaseName, chartName, namespace, isOci, values, step.ChartAuth{})
}

func (f *FakeHelmInstaller) InstallOrUpgradeReleaseWithAuth(ctx context.Context, repoUrl, version, releaseName, chartName, namespace string, isOci bool, values []string, auth step.ChartAuth) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, HelmCall{
		RepoUrl:     repoUrl,
		Version:     version,
		ReleaseName: releaseName,
		ChartName:   chartName,
		Namespace:   namespace,
		IsOci:       isOci,
		Values:      append([]string{}, values...),
		Auth:        auth,
	})
	block := f.Block[releaseName]
	if f.BlockAttempts[releaseName] > 0 {
		f.BlockAttempts[releaseName]--
		block = true
	}
	installErr, hasErr := f.InstallErrors[releaseName]
	f.mu.Unlock()

	if block {
		<-ctx.Done()
		return ctx.Err()
	}
	if hasErr {
		return installErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Releases == nil {
		f.Releases = make(map[string]string)
	}
	f.Releases[releaseName] = version
	return nil
}

func (f *FakeHelmInstaller) InstallOrUpgradeReleaseWithKubeVersion(ctx context.Context, repoUrl, version, releaseName, chartName, namespace string, isOci bool, values []string, kubeVersion string) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, HelmCall{
		RepoUrl:     repoUrl,
		Version:     version,
		ReleaseName: releaseName,
		ChartName:   chartName,
		Namespace:   namespace,
		IsOci:       isOci,
		Values:      append([]string{}, values...),
		KubeVersion: kubeVersion,
	})
	installErr, hasErr := f.InstallErrors[releaseName]
	f.mu.Unlock()

	if hasErr {
		return installErr
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Releases == nil {
		f.Releases = make(map[string]string)
	}
	f.Releases[releaseName] = version
	return nil
}

func (f *FakeHelmInstaller) InstallOrUpgradeFromLocalChart(_ context.Context, chartPath, releaseName, namespace string, values []string, valueFiles []string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Calls = append(f.Calls, HelmCall{
		ReleaseName: releaseName,
		ChartName:   chartPath,
		Namespace:   namespace,
		Values:      append([]string{}, values...),
		ValueFiles:  append([]string{}, valueFiles...),
	})
	if f.Releases == nil {
		f.Releases = make(map[string]string)
	}
	if _, exists := f.Releases[releaseName]; !exists {
		f.Releases[releaseName] = ""
	}
	return nil
}

func (f *FakeHelmInstaller) FixPendingRelease(releaseName, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.PendingReleases, releaseName)
	return nil
}

func (f *FakeHelmInstaller) UninstallRelease(releaseName, _ string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.PendingReleases[releaseName] {
		return fmt.Errorf("release '%s' is in pending state", releaseName)
	}
	delete(f.Releases, releaseName)
	if f.UninstallErrors != nil {
		if err, ok := f.UninstallErrors[releaseName]; ok {
			return helmclient.NormalizeUninstallError(f.UninstallResponses[releaseName], err)
		}
	}
	return nil
}

func (f *FakeHelmInstaller) GetRelease(releaseName string) (version string, exists bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	version, exists = f.Releases[releaseName]
	return
}

// ReleaseStatus reports whether a release exists and its recorded version,
// mirroring the production HelmInstaller used by step detection.
func (f *FakeHelmInstaller) ReleaseStatus(releaseName, _ string) (bool, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.BrokenReleases[releaseName] {
		return false, "", fmt.Errorf("release %q is in a broken state", releaseName)
	}
	version, exists := f.Releases[releaseName]
	return exists, version, nil
}

// CallsByRelease returns calls whose ReleaseName matches name, in order.
func (f *FakeHelmInstaller) CallsByRelease(name string) []HelmCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []HelmCall
	for _, c := range f.Calls {
		if c.ReleaseName == name {
			out = append(out, c)
		}
	}
	return out
}
