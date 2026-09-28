/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project_controller

import (
	"context"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name = "project-controller"

	roleBindingsConfigMapName = "runai-project-controller-rolebindings-plugin"
)

//+kubebuilder:rbac:namespace=runai,groups=apps,resources=deployments,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=serviceaccounts,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=configmaps,verbs=get;delete
//+kubebuilder:rbac:namespace=runai,groups="",resources=services,verbs=get;delete

type operand struct {
	resources []operands.Resource
}

func (o *operand) Name() string { return name }

func (o *operand) Resources() []operands.Resource { return o.resources }

func (o *operand) Status(context.Context) (operands.Status, error) {
	return operands.ReadyStatus(), nil
}

// Enabled is always false: KRM supplies the project-controller now, under a
// different name, so this operand only deletes what the Run:ai-built one left.
func (o *operand) Enabled() bool { return false }

func Operand(_ kc.HybridClient, runaiConfig *runv1.RunaiConfig) operands.Operand {
	namespace := runaiConfig.GetNamespace()
	meta := func(objectName string) metav1.ObjectMeta {
		return metav1.ObjectMeta{Name: objectName, Namespace: namespace}
	}

	return &operand{
		resources: []operands.Resource{
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &appsv1.Deployment{ObjectMeta: meta(operands.RunaiProjectControllerName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.Service{ObjectMeta: meta(operands.RunaiProjectControllerName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.ServiceAccount{ObjectMeta: meta(operands.RunaiProjectControllerName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &monitoringv1.ServiceMonitor{ObjectMeta: meta(operands.RunaiProjectControllerName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.ConfigMap{ObjectMeta: meta(roleBindingsConfigMapName)}
			}),
		},
	}
}
