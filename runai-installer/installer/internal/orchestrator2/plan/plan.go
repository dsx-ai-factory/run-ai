/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package plan builds executable orchestrator plans for every command
// (install, upgrade, uninstall). It lives outside the orchestrator2 core
// because it depends on the actions package (which itself depends on
// orchestrator2); keeping it here breaks the import cycle while letting the
// Runner stay action-agnostic.
//
// Resolve is the command dispatcher and satisfies orchestrator2.PlanBuilder;
// per-command construction lives in the matching file (install.go, ...).
package plan

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
)

// Resolve is an orchestrator2.PlanBuilder that dispatches on the command. It
// returns the preflight plan and the install plan separately so the runner can
// stage them (preflight → confirm → install). Only install is populated today;
// upgrade and uninstall have no preflight and an empty install plan.
func Resolve(s *spec.Spec, deps orchestrator2.Deps) (*orchestrator2.Plan, *orchestrator2.Plan, error) {
	switch s.Command {
	case spec.CommandInstall:
		preflight, install := buildInstall(s, deps)
		return preflight, install, nil
	case spec.CommandAirgapPack:
		preflight, pack := buildAirgapPack(s)
		return preflight, pack, nil
	case spec.CommandAirgapLoad:
		preflight, load := buildAirgapLoad(s)
		return preflight, load, nil
	case spec.CommandUpgrade, spec.CommandUninstall:
		return nil, &orchestrator2.Plan{}, nil
	}
	return nil, nil, fmt.Errorf("unsupported command: %q", s.Command)
}
