/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package steps holds the concrete orchestrator2 Step implementations used to
// build install/upgrade plans. Each step type lives in its own file and owns
// its full lifecycle: PreValidate, Detect (read-only "is it already there?"),
// BuildActions, and PostValidate. Steps capture the orchestrator2.Deps they
// need at construction time, mirroring how actions receive Deps in Execute.
//
// To add a new step: create a file here with a typed step and a New... factory
// that the plan package can call. Implement Detect when the step has a
// meaningful "already installed" check; otherwise embed BaseStep (or use
// ActionStep) and the default no-op Detect applies.
package steps

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

const readinessPollInterval = 5 * time.Second

// pollUntil blocks until ready returns true or ctx is cancelled.
func pollUntil(ctx context.Context, ready func(context.Context) bool) error {
	for {
		if ready(ctx) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readinessPollInterval):
		}
	}
}

// helmReleaseDetect reports a Helm release as installed (with its chart version)
// when the release exists in the namespace.
func helmReleaseDetect(deps orchestrator2.Deps, releaseName, namespace string) (orchestrator2.DetectResult, error) {
	exists, version, err := deps.HelmInstaller.ReleaseStatus(releaseName, namespace)
	if err != nil {
		return orchestrator2.DetectResult{}, err
	}
	if exists {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled, Version: version}, nil
	}
	return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
}
