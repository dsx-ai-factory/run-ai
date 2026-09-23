/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

// GVKPlugin defines the interface for GVK-specific processing
type GVKPlugin interface {
	ProcessManifest(manifest map[string]any, project *orgUnitApi.Project) error
}
