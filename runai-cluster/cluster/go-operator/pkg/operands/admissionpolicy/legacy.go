/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package admissionpolicy

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	operandsutils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewResourcesOld(ctx context.Context, operandName string, k8sReader client.Reader,
	definition Definition) ([]operands.ResourceOld, error) {
	resources := NewResources(operandName, k8sReader, definition)
	legacyResources := make([]operands.ResourceOld, 0, len(resources))
	for _, resource := range resources {
		object, err := resource.Object(ctx)
		if err != nil {
			return nil, err
		}
		legacyResource := operands.ResourceForOperand(object, operandName)
		legacyResource.Enabled = resource.Enabled()
		legacyResources = append(legacyResources, legacyResource)
	}
	return legacyResources, nil
}

func StatusOld(ctx context.Context, operandName string, k8sReader client.Reader,
	definition Definition) (operands.Status, error) {
	return operandsutils.Status(ctx, NewResources(operandName, k8sReader, definition))
}
