/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runai_scheduler

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/common"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands/runai-scheduler/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/operands/utils"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// resourceFunctions returns the functions that build this operand's resources. The
// ServiceMonitor is included only when its CRD is present (i.e. Prometheus is installed);
func (o *operand) resourceFunctions() []operands.ResourceFunction {
	funcs := []operands.ResourceFunction{
		resources.SchedulingShardForNodePool,
	}
	if o.includeServiceMonitor {
		funcs = append(funcs, resources.ServiceMonitorForNodePool)
	}
	return funcs
}

// resourceStatusFunctions returns the functions that compute this operand's resource
// statuses, gating the ServiceMonitor status the same way as resourceFunctions.
func (o *operand) resourceStatusFunctions() []operands.ResourceStatusFunction {
	funcs := []operands.ResourceStatusFunction{
		resources.SchedulingShardStatus,
	}
	if o.includeServiceMonitor {
		funcs = append(funcs, resources.ServiceMonitorStatus)
	}
	return funcs
}

// RBAC permissions the operand requires for resource management
//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=roles,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="",resources=configmaps,verbs=get;create;update;watch;patch;delete;list
//+kubebuilder:rbac:namespace=runai,groups="monitoring.coreos.com",resources=servicemonitors,verbs=get;create;update;watch;patch;delete;list

type operand struct {
	name                  string
	includeServiceMonitor bool
}

func (o *operand) Name() string {
	return o.name
}

// ResourcesForNodePool returns a slice of Kubernetes resources
// that should be Created/Updated/Deleted by the controller to satisfy
// the needs of this operand.
func (o *operand) ResourcesForNodePool(ctx context.Context, k8sReader client.Reader,
	nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams) ([]operands.ResourceOld, error) {
	return utils.ResourcesForNodePool(ctx, k8sReader, nodePool, params, o.resourceFunctions(), o.Name())
}

func (o *operand) Status(ctx context.Context, k8sReader client.Reader,
	nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams) (operands.Status, error) {
	return utils.Status(ctx, k8sReader, nodePool, params, o.resourceStatusFunctions(), o.Name())
}

func Operand(ownerNodePoolName string, includeServiceMonitor bool) operands.NodePoolOperand {
	return &operand{name: ownerNodePoolName, includeServiceMonitor: includeServiceMonitor}
}
