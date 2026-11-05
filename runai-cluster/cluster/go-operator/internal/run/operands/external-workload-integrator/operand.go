/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package external_workload_integrator

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/external-workload-integrator/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	Name = operands.RunaiExternalWorkloadIntegratorName
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return Name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

func (o *operand) Enabled() bool {
	return *o.runaiConfig.Spec.ExternalWorkloadIntegrator.Enabled
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	operandResources := []operands.Resource{
		resources.NewDeploymentResource(Name, k8sReader, runaiConfig),
		resources.NewServiceAccountResource(Name, k8sReader, runaiConfig),
		resources.NewServiceMonitorResource(Name, k8sReader, runaiConfig),
		resources.NewServiceResource(Name, k8sReader, runaiConfig),
		resources.NewVPAResource(Name, k8sReader, runaiConfig),
	}
	operandResources = append(operandResources, admissionpolicy.NewResources(
		Name, k8sReader, resources.PodOwnershipPolicyDefinition(runaiConfig),
	)...)
	return &operand{runaiConfig: runaiConfig, resources: operandResources}
}
