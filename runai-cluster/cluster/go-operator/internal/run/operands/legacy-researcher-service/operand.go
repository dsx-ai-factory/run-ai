/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacy_researcher_service

import (
	"context"
	"encoding/base64"
	"fmt"

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

// Package legacy_researcher_service deletes the resources the operator created
// for researcher-service, retired in 2.26. They carry no ownerReferences, so
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
	name = "legacy-researcher-service-cleanup"

	legacyNameB64      = "cmVzZWFyY2hlci1zZXJ2aWNl"
	legacyIngressB64   = "cmVzZWFyY2hlci1zZXJ2aWNlLWluZ3Jlc3M="
	legacyHTTPRouteB64 = "cmVzZWFyY2hlci1zZXJ2aWNlLWh0dHByb3V0ZQ=="
	legacyCLIRouteB64  = "cmVzZWFyY2hlci1jbGktcm91dGU="
	legacyAPIRouteB64  = "cnVuYWktcmVzZWFyY2hlci1yb3V0ZQ=="
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
	legacyName := mustDecodeBase64(legacyNameB64)
	ingressName := mustDecodeBase64(legacyIngressB64)
	httpRouteName := mustDecodeBase64(legacyHTTPRouteB64)
	cliRouteName := mustDecodeBase64(legacyCLIRouteB64)
	apiRouteName := mustDecodeBase64(legacyAPIRouteB64)

	return &operand{
		resources: []operands.Resource{
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &appsv1.Deployment{ObjectMeta: meta(legacyName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.Service{ObjectMeta: meta(legacyName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &corev1.ServiceAccount{ObjectMeta: meta(legacyName)}
			}),
			operands.NewDeleteOnlyResource(name, func() client.Object {
				return &vpav1.VerticalPodAutoscaler{ObjectMeta: meta(legacyName)}
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

func mustDecodeBase64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// The inputs are compile-time constants; a decode failure is a bug.
		panic(fmt.Sprintf("legacy_researcher_service: invalid base64 constant %q: %v", encoded, err))
	}
	return string(decoded)
}
