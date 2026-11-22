/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package knativescaling

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type plugin struct{}

func New(context.Context, framework.PluginDeps) (framework.Plugin, error) { return plugin{}, nil }

func (plugin) Name() string { return "knative-scaling" }

func (plugin) BeforeApply(ctx context.Context, c client.Client, resource *unstructured.Unstructured, preview framework.PreviewApply) error {
	labels := resource.GetLabels()
	if labels[constants.ConvertedFromNativeLabel] != "true" ||
		labels[constants.ControlPlaneWorkloadKindLabelForWorkload] != constants.V2InferenceWorkloadKind {
		return nil
	}
	return Prepare(ctx, c, resource, preview)
}

var _ framework.BeforeApplyPlugin = plugin{}
