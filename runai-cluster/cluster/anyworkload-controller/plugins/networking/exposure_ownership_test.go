/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"reflect"
	"testing"

	openshiftv1 "github.com/openshift/api/route/v1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestExposureServiceBackends(t *testing.T) {
	t.Parallel()

	crossNamespace := gatewayv1.Namespace("foreign")
	nonServiceGroup := gatewayv1.Group("example.com")
	nonServiceKind := gatewayv1.Kind("StorageBucket")
	crossNamespaceBackend := gatewayBackend("foreign")
	crossNamespaceBackend.Namespace = &crossNamespace
	nonServiceBackend := gatewayBackend("foreign")
	nonServiceBackend.Group, nonServiceBackend.Kind = &nonServiceGroup, &nonServiceKind

	tests := []struct {
		name          string
		exposure      client.Object
		wantNames     []string
		wantViolation bool
	}{
		{
			name:      "Ingress default and path backends",
			exposure:  ingressWithServices("default", "main", "secondary"),
			wantNames: []string{"default", "main", "secondary"},
		},
		{
			name: "Ingress resource backend",
			exposure: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "ingress"},
				Spec: networkingv1.IngressSpec{DefaultBackend: &networkingv1.IngressBackend{
					Resource: &corev1.TypedLocalObjectReference{Kind: "StorageBucket", Name: "foreign"},
				}},
			},
			wantViolation: true,
		},
		{
			name: "Ingress backend contains both Service and resource",
			exposure: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "ingress"},
				Spec: networkingv1.IngressSpec{DefaultBackend: &networkingv1.IngressBackend{
					Service:  &networkingv1.IngressServiceBackend{Name: "service"},
					Resource: &corev1.TypedLocalObjectReference{Kind: "StorageBucket", Name: "foreign"},
				}},
			},
			wantViolation: true,
		},
		{
			name:      "Route primary and alternate backends",
			exposure:  routeWithServices("main", "canary"),
			wantNames: []string{"main", "canary"},
		},
		{
			name: "Route non-Service backend",
			exposure: &openshiftv1.Route{
				ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "route"},
				Spec:       openshiftv1.RouteSpec{To: openshiftv1.RouteTargetReference{Kind: "Pod", Name: "foreign"}},
			},
			wantViolation: true,
		},
		{
			name:      "HTTPRoute rule backend and request mirrors",
			exposure:  httpRouteWithServices("main", "backend-mirror", "rule-mirror"),
			wantNames: []string{"main", "backend-mirror", "rule-mirror"},
		},
		{
			name:          "HTTPRoute cross-namespace backend",
			exposure:      httpRouteWithBackend(crossNamespaceBackend),
			wantViolation: true,
		},
		{
			name:          "HTTPRoute non-Service backend",
			exposure:      httpRouteWithBackend(nonServiceBackend),
			wantViolation: true,
		},
		{
			name: "HTTPRoute malformed request mirror",
			exposure: &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "http-route"},
				Spec: gatewayv1.HTTPRouteSpec{Rules: []gatewayv1.HTTPRouteRule{{
					Filters: []gatewayv1.HTTPRouteFilter{{Type: gatewayv1.HTTPRouteFilterRequestMirror}},
				}}},
			},
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			backends, err := exposureServiceBackends(tt.exposure)
			if tt.wantViolation {
				if !errors.Is(err, errOwnershipViolation) {
					t.Fatalf("exposureServiceBackends() error = %v, want ownership violation", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("exposureServiceBackends() error = %v", err)
			}
			if got := backendNames(backends); !reflect.DeepEqual(got, tt.wantNames) {
				t.Fatalf("exposureServiceBackends() names = %v, want %v", got, tt.wantNames)
			}
		})
	}
}

func TestValidateExposureServiceOwnershipChecksSecondaryBackends(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	service, pod, slice := newNamedOwnershipBackend(runner, "main", "10.0.1.1")
	reader := newOwnershipFakeClient(t, service, pod, slice)

	tests := []struct {
		name     string
		exposure client.Object
	}{
		{
			name:     "Ingress default backend",
			exposure: ingressWithServices("missing", "main"),
		},
		{
			name:     "Route alternate backend",
			exposure: routeWithServices("main", "missing"),
		},
		{
			name:     "HTTPRoute request-mirror backend",
			exposure: httpRouteWithServices("main", "", "missing"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateExposureServiceOwnership(context.Background(), reader, runner, tt.exposure)
			if !apierrors.IsNotFound(err) {
				t.Fatalf("validateExposureServiceOwnership() error = %v, want missing secondary Service", err)
			}
			if errors.Is(err, errOwnershipViolation) {
				t.Fatalf("missing Service must remain retryable, got violation: %v", err)
			}
		})
	}
}

func TestValidateExposureServiceOwnershipAcceptsOwnedDestinations(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	service, pod, slice := newNamedOwnershipBackend(runner, "main", "10.0.1.2")
	ingress := ingressWithServices("", service.Name)

	err := validateExposureServiceOwnership(
		context.Background(), newOwnershipFakeClient(t, service, pod, slice), runner, ingress,
	)
	if err != nil {
		t.Fatalf("validateExposureServiceOwnership() error = %v", err)
	}
}

func TestValidateExposureServiceOwnershipPrioritizesViolationAcrossObjects(t *testing.T) {
	t.Parallel()

	runner := newOwnershipRunner()
	foreignService, foreignPod, foreignSlice := newNamedOwnershipBackend(runner, "foreign", "10.0.1.3")
	foreignPod.OwnerReferences[0].UID = "foreign-runner-uid"
	missingIngress := ingressWithServices("missing")
	foreignRoute := routeWithServices(foreignService.Name)

	err := validateExposureServiceOwnership(
		context.Background(),
		newOwnershipFakeClient(t, foreignService, foreignPod, foreignSlice),
		runner,
		missingIngress,
		foreignRoute,
	)
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("validation error = %v, want ownership violation", err)
	}
	if apierrors.IsNotFound(err) {
		t.Fatalf("missing Service masked an ownership violation: %v", err)
	}
}

func TestValidateExposureServiceOwnershipHandlesTypedNil(t *testing.T) {
	t.Parallel()

	var ingress *networkingv1.Ingress
	err := validateExposureServiceOwnership(
		context.Background(), newOwnershipFakeClient(t), newOwnershipRunner(), ingress,
	)
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("validation error = %v, want typed-nil ownership violation", err)
	}
}

func ingressBackend(name string) *networkingv1.IngressBackend {
	return &networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{Name: name}}
}

func ingressWithServices(defaultService string, pathServices ...string) *networkingv1.Ingress {
	spec := networkingv1.IngressSpec{}
	if defaultService != "" {
		spec.DefaultBackend = ingressBackend(defaultService)
	}
	if len(pathServices) > 0 {
		paths := make([]networkingv1.HTTPIngressPath, 0, len(pathServices))
		for _, name := range pathServices {
			paths = append(paths, networkingv1.HTTPIngressPath{Backend: *ingressBackend(name)})
		}
		spec.Rules = []networkingv1.IngressRule{{IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{Paths: paths},
		}}}
	}
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "ingress"},
		Spec:       spec,
	}
}

func routeWithServices(primary string, alternates ...string) *openshiftv1.Route {
	alternateBackends := make([]openshiftv1.RouteTargetReference, 0, len(alternates))
	for _, name := range alternates {
		alternateBackends = append(alternateBackends, openshiftv1.RouteTargetReference{Name: name})
	}
	return &openshiftv1.Route{
		ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "route"},
		Spec: openshiftv1.RouteSpec{
			To:                openshiftv1.RouteTargetReference{Kind: serviceKind, Name: primary},
			AlternateBackends: alternateBackends,
		},
	}
}

func gatewayBackend(name string) gatewayv1.BackendObjectReference {
	return gatewayv1.BackendObjectReference{Name: gatewayv1.ObjectName(name)}
}

func httpRouteWithServices(backend, backendMirror, ruleMirror string) *gatewayv1.HTTPRoute {
	backendRef := gatewayv1.HTTPBackendRef{
		BackendRef: gatewayv1.BackendRef{BackendObjectReference: gatewayBackend(backend)},
	}
	if backendMirror != "" {
		backendRef.Filters = []gatewayv1.HTTPRouteFilter{requestMirrorFilter(backendMirror)}
	}
	rule := gatewayv1.HTTPRouteRule{BackendRefs: []gatewayv1.HTTPBackendRef{backendRef}}
	if ruleMirror != "" {
		rule.Filters = []gatewayv1.HTTPRouteFilter{requestMirrorFilter(ruleMirror)}
	}
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "http-route"},
		Spec:       gatewayv1.HTTPRouteSpec{Rules: []gatewayv1.HTTPRouteRule{rule}},
	}
}

func requestMirrorFilter(serviceName string) gatewayv1.HTTPRouteFilter {
	return gatewayv1.HTTPRouteFilter{
		Type: gatewayv1.HTTPRouteFilterRequestMirror,
		RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
			BackendRef: gatewayBackend(serviceName),
		},
	}
}

func httpRouteWithBackend(backend gatewayv1.BackendObjectReference) *gatewayv1.HTTPRoute {
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "http-route"},
		Spec: gatewayv1.HTTPRouteSpec{Rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{BackendObjectReference: backend},
			}},
		}}},
	}
}

func newNamedOwnershipBackend(
	runner *dromosv1alpha1.WorkloadRunner,
	name, ip string,
) (*corev1.Service, *corev1.Pod, *discoveryv1.EndpointSlice) {
	selector := map[string]string{"app": name}
	service := newOwnershipService(runner, false, selector)
	service.Name = name
	service.UID = types.UID(name + "-service-uid")
	pod := newOwnershipPod(runner, name, selector, ip)
	slice := newOwnershipEndpointSlice(service, pod, ip)
	return service, pod, slice
}

func backendNames(backends []serviceBackend) []string {
	names := make([]string, 0, len(backends))
	for _, backend := range backends {
		names = append(names, backend.key.Name)
	}
	return names
}
