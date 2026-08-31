/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
)

// NoOpPlugin is the default plugin that does nothing
type NoOpPlugin struct{}

func (n *NoOpPlugin) ProcessManifest(_ map[string]any, _ *orgUnitApi.Project) error {
	return nil
}
