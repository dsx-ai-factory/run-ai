/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package researcher_service

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	osv1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// Package researcher_service deletes the resources the operator created for
// researcher-service, retired in 2.26. They carry no ownerReferences, so
// nothing else removes them from upgraded clusters. The operand is always
// disabled, which routes its resources to deletion on every reconcile; objects
// that are already gone and kinds whose CRD is absent are skipped.
//
// Remove this operand once every supported cluster is past 2.26.

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups=networking.k8s.io,resources=ingresses,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups=gateway.networking.k8s.io,resources=httproutes,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups=route.openshift.io,resources=routes,verbs=get;delete

const (
	name = "researcher-service"

	deploymentName = "researcher-service"
	ingressName    = "researcher-service-ingress"
	httpRouteName  = "researcher-service-httproute"
	cliRouteName   = "researcher-cli-route"
	apiRouteName   = "runai-researcher-route"
)

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

func (o *operand) Enabled() bool { return false }

func Operand(_ kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	namespace := runaiConfig.GetNamespace()
	meta := func(objectName string) metav1.ObjectMeta {
		return metav1.ObjectMeta{Name: objectName, Namespace: namespace}
	}

	return &operand{
		resources: []operands.Resource{
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &appsv1.Deployment{ObjectMeta: meta(deploymentName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.Service{ObjectMeta: meta(deploymentName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.ServiceAccount{ObjectMeta: meta(deploymentName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &vpav1.VerticalPodAutoscaler{ObjectMeta: meta(deploymentName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &networkingv1.Ingress{ObjectMeta: meta(ingressName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &gwv1.HTTPRoute{ObjectMeta: meta(httpRouteName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &osv1.Route{ObjectMeta: meta(cliRouteName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &osv1.Route{ObjectMeta: meta(apiRouteName)}
			}),
		},
	}
}
