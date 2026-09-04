/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type oldResourceWrapper struct {
	ResourceOld operands.ResourceOld
}

func (resource *oldResourceWrapper) OperandName() string {
	return resource.ResourceOld.OperandName
}

func (resource *oldResourceWrapper) Object(ctx context.Context) (client.Object, error) {
	return resource.ResourceOld.Object, nil
}
func (resource *oldResourceWrapper) Status(ctx context.Context) error {
	return nil
}
func (resource *oldResourceWrapper) Enabled() bool {
	return resource.ResourceOld.Enabled
}

type oldOperandWrapper struct {
	OperandOld     operands.OperandOld
	OperandEnabled bool
	ctx            context.Context
	hybridClient   kc.HybridClient
	runaiConfig    *runaiv1.RunaiConfig
}

func (o *oldOperandWrapper) Name() string {
	return o.OperandOld.Name()
}

func (o *oldOperandWrapper) Resources() []operands.Resource {
	resourcesOld, err := o.OperandOld.ResourcesForRunaiConfig(o.ctx, o.hybridClient, o.runaiConfig)
	if err != nil {
		log.Log.Error(err, "Error getting resources form old operand", "operand", o.Name())
		return nil
	}
	resourcesNew := []operands.Resource{}
	for _, resource := range resourcesOld {
		resourcesNew = append(resourcesNew, &oldResourceWrapper{
			ResourceOld: resource,
		})
	}
	return resourcesNew
}

func (o *oldOperandWrapper) Status(ctx context.Context) (operands.Status, error) {
	return o.OperandOld.Status(ctx, o.hybridClient, o.runaiConfig)
}

func (o *oldOperandWrapper) Enabled() bool {
	return o.OperandEnabled
}
