/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package krm_operator

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/krm-operator/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	// Name specifies the name of this operand for logging purposes.
	Name = "krm-operator"
)

var (
	resourceFunctions = []operands.ResourceFunction{
		resources.KRMConfigForRunaiConfig,
	}

	resourceStatusFunctions = []operands.ResourceStatusFunction{
		resources.KRMConfigStatus,
	}
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace="",groups="kai.resources",resources=krmconfigs,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace="",groups="kai.resources",resources=krmconfigs/status,verbs=get

type operand struct{}

func (o *operand) Name() string {
	return Name
}

func (o *operand) ResourcesForRunaiConfig(ctx context.Context, hybridClient kc.HybridClient,
	runaiConfig *runv1.RunaiConfig) ([]operands.ResourceOld, error) {
	return utils.ResourcesForRunaiConfig(ctx, hybridClient, runaiConfig, resourceFunctions, o.Name())
}

func (o *operand) Status(ctx context.Context, hybridClient kc.HybridClient,
	runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	return utils.StatusOld(ctx, hybridClient, runaiConfig, resourceStatusFunctions)
}

func Operand() operands.OperandOld {
	return &operand{}
}
