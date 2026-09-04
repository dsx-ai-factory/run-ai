/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
)

// PushAirgapImages mirrors the offline bundle's container images into the
// customer's registry.
//
// It does not push the images itself: it shells out to the setup.sh script that
// ships inside the extracted bundle (run from WorkingDir, with REGISTRY_URL and
// optional REGISTRY_USERNAME/REGISTRY_PASSWORD passed via the environment). The
// installer relies on that script being present and doing the actual push and
// tagging; we only locate the working dir and invoke it. The script is produced
// by the bundle build, so a missing or non-executable setup.sh means the bundle
// is malformed.
type PushAirgapImages struct {
	WorkingDir  string
	RegistryURL string
	Username    string
	Password    string
}

func (a PushAirgapImages) Kind() string { return "airgap.push_images" }

func (a PushAirgapImages) Args() map[string]any {
	return map[string]any{
		"workingDir": a.WorkingDir,
		"registry":   a.RegistryURL,
	}
}

func (a PushAirgapImages) Describe() string {
	return fmt.Sprintf("push airgap images from %s to %s", a.WorkingDir, a.RegistryURL)
}

func (a PushAirgapImages) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	return orchestrator2.ActionResult{}, airgap.RunSetupScript(ctx, a.WorkingDir, a.RegistryURL, a.Username, a.Password, deps.Progress)
}
