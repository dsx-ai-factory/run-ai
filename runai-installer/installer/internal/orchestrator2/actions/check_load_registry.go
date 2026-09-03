/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// probeRepository is the repository push access is probed against. The probe
// starts and cancels an upload session without writing image content.
const probeRepository = "runai-airgap-probe"

// CheckLoadRegistry verifies the destination registry accepts the bundle before
// the upload begins: it must be reachable, accept the configured credentials,
// and grant push access. Unlike the install-time registry check this fails
// rather than warns, since a load without push access cannot succeed.
type CheckLoadRegistry struct {
	Destination string
}

func (a CheckLoadRegistry) Kind() string { return "preflight.load_registry" }

func (a CheckLoadRegistry) Args() map[string]any {
	return map[string]any{"registry": a.Destination}
}

func (a CheckLoadRegistry) Describe() string {
	return fmt.Sprintf("verify %s is reachable and grants push access", a.Destination)
}

func (a CheckLoadRegistry) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if deps.ImagePusher == nil {
		return orchestrator2.ActionResult{}, errors.New("no image pusher configured")
	}

	probeRef := fmt.Sprintf("%s/%s:probe", a.Destination, probeRepository)
	if err := deps.ImagePusher.CheckPushPermission(ctx, probeRef); err != nil {
		problem := fmt.Sprintf("%s: %v", a.Destination, err)
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      problem,
			Remediation: "provide credentials with push access under registry, and create the project first if the registry requires one",
			Data:        map[string]any{"registry": a.Destination, "problem": problem},
		}, nil
	}

	return orchestrator2.ActionResult{
		Reason: fmt.Sprintf("push access to %s verified", a.Destination),
		Data:   map[string]any{"registry": a.Destination},
	}, nil
}
