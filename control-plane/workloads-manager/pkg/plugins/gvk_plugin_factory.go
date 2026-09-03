/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

const (
	DynamoGraphDeploymentKind  = "DynamoGraphDeployment"
	DynamoGraphDeploymentGroup = "nvidia.com"
)

// Factory creates GVK-specific plugins based on GVK
type Factory struct{}

// NewFactory creates a new instance of the factory
func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) CreatePlugin(gvk api.Gvk) GVKPlugin {
	if gvk.Kind == DynamoGraphDeploymentKind &&
		convutils.Deref(gvk.Group, "") == DynamoGraphDeploymentGroup {
		return &DynamoGraphDeploymentPlugin{}
	}

	return &NoOpPlugin{}
}
