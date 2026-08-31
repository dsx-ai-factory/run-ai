/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *WorkloadRunnerController) getAttachedResource(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner) (*unstructured.Unstructured, error) {
	template, err := c.templateParser.Parse(runner.Spec.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	existing := &unstructured.Unstructured{}
	existing.SetNamespace(runner.GetNamespace())
	existing.SetGroupVersionKind(template.GroupVersionKind())

	if err = c.client.Get(ctx, client.ObjectKey{
		Namespace: runner.Namespace,
		Name:      template.GetName(),
	}, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
