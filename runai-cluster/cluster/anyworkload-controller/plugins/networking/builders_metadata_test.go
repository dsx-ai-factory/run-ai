/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"reflect"
	"testing"
	"time"

	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestNetworkingBuildersDiscardCallerLifecycleMetadata(t *testing.T) {
	t.Setenv("INGRESS_ANNOTATIONS", "")
	plugin := NewPlugin(nil)
	runner := newOwnershipRunner()
	metadata := metav1.ObjectMeta{
		Name: "requested-name", Namespace: "caller-namespace", GenerateName: "caller-prefix-",
		UID: "caller-uid", ResourceVersion: "123", Generation: 7,
		CreationTimestamp: metav1.NewTime(time.Unix(1, 0)),
		OwnerReferences:   []metav1.OwnerReference{{APIVersion: "v1", Kind: "Pod", Name: "caller", UID: "owner-uid"}},
		Finalizers:        []string{"example.com/caller"},
		ManagedFields:     []metav1.ManagedFieldsEntry{{Manager: "caller"}},
		Labels:            map[string]string{"app": "retained"},
		Annotations:       map[string]string{"example.com/setting": "retained"},
	}
	service := &corev1.Service{
		ObjectMeta: metadata,
		Status: corev1.ServiceStatus{LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{{IP: "192.0.2.1"}},
		}},
	}
	tests := []struct {
		name  string
		input client.Object
		build func(client.Object) client.Object
	}{
		{
			name: "legacy Service", input: service,
			build: func(input client.Object) client.Object {
				return plugin.buildWorkloadService(runner, input.(*corev1.Service))
			},
		},
		{
			name: "endpoint Service", input: service,
			build: func(input client.Object) client.Object {
				return plugin.buildEndpointWorkloadService(runner, input.(*corev1.Service))
			},
		},
		{
			name: "Ingress",
			input: &networkingv1.Ingress{ObjectMeta: metadata, Status: networkingv1.IngressStatus{
				LoadBalancer: networkingv1.IngressLoadBalancerStatus{Ingress: []networkingv1.IngressLoadBalancerIngress{{IP: "192.0.2.1"}}},
			}},
			build: func(input client.Object) client.Object {
				return plugin.buildIngress(runner, input.(*networkingv1.Ingress), "", 0)
			},
		},
		{
			name: "Route",
			input: &openshiftv1.Route{ObjectMeta: metadata, Status: openshiftv1.RouteStatus{
				Ingress: []openshiftv1.RouteIngress{{Host: "caller.example"}},
			}},
			build: func(input client.Object) client.Object {
				return plugin.buildRoute(runner, input.(*openshiftv1.Route), "")
			},
		},
		{
			name: "HTTPRoute",
			input: &gatewayv1.HTTPRoute{ObjectMeta: metadata, Status: gatewayv1.HTTPRouteStatus{
				RouteStatus: gatewayv1.RouteStatus{Parents: []gatewayv1.RouteParentStatus{{ControllerName: "example.com/controller"}}},
			}},
			build: func(input client.Object) client.Object {
				return plugin.buildHTTPRoute(runner, input.(*gatewayv1.HTTPRoute), "", 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.input.DeepCopyObject().(client.Object)
			before := input.DeepCopyObject()
			built := tt.build(input)
			if built.GetName() != metadata.Name || built.GetNamespace() != runner.Namespace {
				t.Fatalf("resource identity = %s/%s, want %s/%s", built.GetNamespace(), built.GetName(), runner.Namespace, metadata.Name)
			}
			if built.GetUID() != "" || built.GetResourceVersion() != "" || built.GetGeneration() != 0 ||
				built.GetGenerateName() != "" || !built.GetCreationTimestamp().Time.IsZero() ||
				len(built.GetOwnerReferences()) != 0 || len(built.GetFinalizers()) != 0 || len(built.GetManagedFields()) != 0 {
				t.Fatalf("builder retained caller lifecycle metadata: %#v", built)
			}
			if !reflect.ValueOf(built).Elem().FieldByName("Status").IsZero() {
				t.Fatal("builder retained caller status")
			}
			if built.GetLabels()["app"] != "retained" || built.GetAnnotations()["example.com/setting"] != "retained" {
				t.Fatal("builder lost caller labels or annotations")
			}
			if !reflect.DeepEqual(input, before) {
				t.Fatal("builder modified its input")
			}
			built.GetLabels()["app"] = "changed"
			built.GetAnnotations()["example.com/setting"] = "changed"
			if !reflect.DeepEqual(input, before) {
				t.Fatal("built metadata aliases the input maps")
			}
		})
	}
}
