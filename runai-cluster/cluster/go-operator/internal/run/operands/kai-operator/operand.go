/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kai_operator

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/kai-operator/resources"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	name = "kai-operator"
)

var (
	resourceFunctions = []operands.ResourceFunction{
		resources.KAIConfigForRunaiConfig,
		resources.PrometheusRuleForRunaiConfig,
	}

	resourceStatusFunctions = []operands.ResourceStatusFunction{
		resources.KAIConfigStatus,
		resources.PrometheusRuleStatus,
	}
)

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:groups=kai.scheduler,resources=configs,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=kai.scheduler,resources=configs/status,verbs=get
//+kubebuilder:rbac:groups=monitoring.coreos.com,resources=prometheusrules,verbs=get;create;update;watch;patch;delete;list

type operand struct{}

func (o *operand) Name() string {
	return name
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
