/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/nodepool-controller/resources"
	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	name = "nodepool-controller"
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=roles,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:groups=run.ai,resources=nodepools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.ai,resources=nodepools/status,verbs=get;update;patch;watch
//+kubebuilder:rbac:groups=run.ai,resources=nodepools/finalizers,verbs=update;patch
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;watch;update;patch;list
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;watch;list
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;watch;list
//+kubebuilder:rbac:groups="",resources=services,verbs=get;watch;list
//+kubebuilder:rbac:groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;watch;list
//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=roles,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=configmaps,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="autoscaling.k8s.io",resources=verticalpodautoscalers,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	runaiConfig *runv1.RunaiConfig
	resources   []operands.Resource
}

func (o *operand) Name() string {
	return name
}

func (o *operand) Resources() []operands.Resource {
	return o.resources
}

func (o *operand) Status(ctx context.Context) (operands.Status, error) {
	return utils.Status(ctx, o.resources)
}

// Enabled is pinned off: KRM supplies this controller now, and the operand is
// kept registered only to delete what KRM does not take over.
func (o *operand) Enabled() bool {
	return false
}

func Operand(k8sReader kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	return &operand{
		runaiConfig: runaiConfig,
		// A disabled operand deletes exactly what it lists, and KRM names its
		// Deployment, ServiceAccount, Service, ServiceMonitors and VPA the same,
		// so listing those would delete objects KRM has already adopted and
		// recreates seconds later. Only the Role and RoleBinding are left: KRM
		// prefixes its own with kai-, so nothing else would ever remove ours.
		resources: []operands.Resource{
			resources.NewRoleResource(name, k8sReader, runaiConfig),
			resources.NewRoleBindingResource(name, k8sReader, runaiConfig),
		},
	}
}
