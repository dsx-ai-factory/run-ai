/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type K8sCreate struct {
	Object client.Object
}

func (a K8sCreate) Kind() string { return "k8s.create" }

func (a K8sCreate) Args() map[string]any {
	return map[string]any{
		"name":      a.Object.GetName(),
		"namespace": a.Object.GetNamespace(),
		"kind":      a.Object.GetObjectKind().GroupVersionKind().Kind,
	}
}

func (a K8sCreate) Describe() string {
	gvk := a.Object.GetObjectKind().GroupVersionKind()
	return fmt.Sprintf("kubectl create %s %s", gvk.Kind, a.Object.GetName())
}

func (a K8sCreate) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	return orchestrator2.ActionResult{}, deps.K8sExecutor.Create(ctx, a.Object)
}
