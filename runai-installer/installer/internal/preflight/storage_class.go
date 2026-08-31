/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package preflight

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/steps"
)

func init() { builders = append(builders, storageClassBuilder) }

// storageClassBuilder assembles the default-StorageClass check. It needs a
// ClusterProbe, so it returns nil when one is not wired (e.g. component tests
// not exercising preflight).
func storageClassBuilder(_ *spec.Spec, deps orchestrator2.Deps) orchestrator2.Step {
	if deps.ClusterProbe == nil {
		return nil
	}
	return steps.NewActionStepTitled("preflight-storage-class", "Default storage class", actions.CheckDefaultStorageClass{})
}
