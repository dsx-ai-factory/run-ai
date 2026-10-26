/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool_controller

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name = "nodepool-controller"
)

//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=roles,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;delete

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

// Enabled is always false: KRM supplies this controller now, and the operand is
// kept registered only to delete what KRM does not take over.
func (o *operand) Enabled() bool { return false }

func Operand(_ kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	namespace := runaiConfig.GetNamespace()

	// KRM names its Deployment, ServiceAccount, Service, ServiceMonitors and VPA
	// the same, so listing those would delete objects KRM has already adopted and
	// recreates seconds later. Only the Role and RoleBinding are left: KRM
	// prefixes its own with kai-, so nothing else would ever remove ours.
	return &operand{
		resources: []operands.Resource{
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
			}),
		},
	}
}
