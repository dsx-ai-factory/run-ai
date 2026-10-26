/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
)

// applyGVKSpecificProcessing applies GVK-specific processing to a manifest
func (w *WorkloadsV2Impl) applyGVKSpecificProcessing(manifest map[string]any, project *orgUnitApi.Project) error {
	gvk, err := w.k8sUtil.GetGVK(manifest)
	if err != nil {
		return fmt.Errorf("failed to extract GVK: %w", err)
	}

	plugin := w.gvkPluginFactory.CreatePlugin(gvk)

	if err := plugin.ProcessManifest(manifest, project); err != nil {
		return fmt.Errorf("failed to process manifest with GVK plugin: %w", err)
	}
	return nil
}
