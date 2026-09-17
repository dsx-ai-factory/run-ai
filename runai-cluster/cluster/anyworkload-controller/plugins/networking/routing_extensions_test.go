/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"errors"
	"strings"
	"testing"

	openshiftv1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestRoutingAnnotationContract(t *testing.T) {
	t.Parallel()
	constraints := routingConstraints{}
	constructors := map[string]func() client.Object{
		"Ingress":   func() client.Object { return ingressWithServices("", "owned") },
		"Route":     func() client.Object { return routeWithServices("owned") },
		"HTTPRoute": func() client.Object { return httpRouteWithServices("owned", "", "") },
		"Service":   func() client.Object { return &corev1.Service{} },
	}
	for kind, construct := range constructors {
		for _, key := range []string{
			"nginx.ingress.kubernetes.io/mirror-target", "nginx.ingress.kubernetes.io/default-backend",
			"nginx.ingress.kubernetes.io/configuration-snippet", "traefik.ingress.kubernetes.io/router.middlewares",
			"haproxy.org/backend-config-snippet", "example.com/unknown-extension",
		} {
			t.Run(kind+"/"+key, func(t *testing.T) {
				object := construct()
				object.SetAnnotations(map[string]string{key: "private-value-not-for-errors"})
				err := validateRoutingAnnotations(object, constraints)
				if !errors.Is(err, errOwnershipViolation) || !strings.Contains(err.Error(), key) {
					t.Fatalf("expected field-specific violation, got %v", err)
				}
				if strings.Contains(err.Error(), "private-value-not-for-errors") {
					t.Fatal("annotation value leaked into error")
				}
			})
		}
	}
	tests := []struct {
		kind, key, value string
		allowed          bool
	}{
		{"Ingress", "nginx.ingress.kubernetes.io/proxy-body-size", "8g", true},
		{"Ingress", "nginx.ingress.kubernetes.io/proxy-body-size", "8g; proxy_pass http://foreign", false},
		{"Ingress", "nginx.ingress.kubernetes.io/rewrite-target", "/$2", true},
		{"Ingress", "nginx.ingress.kubernetes.io/rewrite-target", "http://foreign/$2", false},
		{"Ingress", "nginx.ingress.kubernetes.io/backend-protocol", "GRPC", true},
		{"Ingress", "haproxy.ingress.kubernetes.io/backend-protocol", "grpc", true},
		{"Ingress", "haproxy.ingress.kubernetes.io/rewrite-target", `/\1`, true},
		{"Ingress", "haproxy.ingress.kubernetes.io/request-set-header", "X-Forwarded-Uri %[req.uri]", true},
		{"Ingress", "haproxy.ingress.kubernetes.io/request-set-header", "Host foreign", false},
		{"Route", "haproxy.router.openshift.io/timeout", "60s", true},
		{"Route", "openshift.io/host.generated", "true", true},
		{"Route", "openshift.io/host.generated", "unrecognized", false},
		{"Route", "haproxy.router.openshift.io/timeout", "60s\nhttp-request redirect", false},
		{"Service", "openshift.io/scc", "nonroot", true},
		{"Service", "openshift.io/required-scc", "hostmount-anyuid", true},
		{"Service", "nvidia.com/parent-spec-hash", "6bb75d4567", true},
		{"Service", "gpu-memory", "1024", true},
		{"Service", "gpu-fraction", "0.5", true},
		{"Service", "gpu-fraction-num-devices", "2", true},
		{"Service", "deployment.kubernetes.io/revision", "3", true},
		{"Service", "kubectl.kubernetes.io/last-applied-configuration", "", true},
		{"HTTPRoute", "kubectl.kubernetes.io/last-applied-configuration", `{}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.kind+"/"+tt.key+"/"+tt.value, func(t *testing.T) {
			object := constructors[tt.kind]()
			object.SetAnnotations(map[string]string{tt.key: tt.value})
			err := validateRoutingAnnotations(object, constraints)
			if (err == nil) != tt.allowed {
				t.Fatalf("allowed=%t, got %v", tt.allowed, err)
			}
		})
	}
}

func TestNIMHAProxyPathRewrite(t *testing.T) {
	t.Parallel()
	ingress := ingressWithServices("", "owned")
	ingress.Spec.Rules[0].HTTP.Paths[0].Path = "/api/team/nim(/|$)(.*)"
	for _, tt := range []struct {
		value   string
		allowed bool
	}{
		{`/api/team/nim(/|$)(.*) /\2`, true},
		{`/other-path(/|$)(.*) /\2`, false},
		{`/api/team/nim(/|$)(.*) http://foreign/\2`, false},
		{"/api/team/nim(/|$)(.*) /\\2\nhttp-request redirect", false},
	} {
		if allowed := allowedRoutingAnnotation(ingress, "haproxy.org/path-rewrite", tt.value, routingConstraints{}); allowed != tt.allowed {
			t.Errorf("value %q: allowed=%t, want %t", tt.value, allowed, tt.allowed)
		}
	}
}

func TestConfiguredIngressClass(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, approved, class, legacy string
		allowed                       bool
	}{
		{"configured", "nginx", "nginx", "", true},
		{"legacy", "haproxy", "", "haproxy", true},
		{"matching representations", "nginx", "nginx", "nginx", true},
		{"conflict", "nginx", "nginx", "foreign", false},
		{"arbitrary class", "nginx", "traefik", "", false},
		{"missing configured class", "nginx", "", "", false},
		{"Kubernetes default", "", "", "", true},
		{"unset configuration is not wildcard", "", "foreign", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ingress := &networkingv1.Ingress{Spec: networkingv1.IngressSpec{IngressClassName: ptr.To(tt.class)}}
			if tt.legacy != "" {
				ingress.Annotations = map[string]string{ingressClassAnnotation: tt.legacy}
			}
			if err := validateIngressClass(ingress, tt.approved); (err == nil) != tt.allowed {
				t.Fatalf("allowed=%t, got %v", tt.allowed, err)
			}
		})
	}
}

func TestConfiguredGatewayParents(t *testing.T) {
	t.Parallel()
	constraints := routingConstraints{gatewayName: "external", gatewayNamespace: "infra"}
	approved := gatewayv1.ParentReference{Name: "external", Namespace: ptr.To(gatewayv1.Namespace("infra"))}
	for _, tt := range []struct {
		name    string
		parent  gatewayv1.ParentReference
		allowed bool
	}{
		{"configured cross-namespace", approved, true},
		{"explicit defaults and listener", gatewayv1.ParentReference{
			Name: "external", Namespace: approved.Namespace, Group: ptr.To(gatewayv1.Group(gatewayv1.GroupName)),
			Kind: ptr.To(gatewayv1.Kind("Gateway")), SectionName: ptr.To(gatewayv1.SectionName("https")), Port: ptr.To(gatewayv1.PortNumber(443)),
		}, true},
		{"wrong name", gatewayv1.ParentReference{Name: "foreign", Namespace: approved.Namespace}, false},
		{"wrong namespace", gatewayv1.ParentReference{Name: "external"}, false},
		{"mesh Service", gatewayv1.ParentReference{Name: "external", Namespace: approved.Namespace, Group: ptr.To(gatewayv1.Group("")), Kind: ptr.To(gatewayv1.Kind("Service"))}, false},
		{"custom kind", gatewayv1.ParentReference{Name: "external", Namespace: approved.Namespace, Kind: ptr.To(gatewayv1.Kind("CustomGateway"))}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			route := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "project"}}
			route.Spec.ParentRefs = []gatewayv1.ParentReference{tt.parent}
			if err := validateGatewayParents(route, constraints); (err == nil) != tt.allowed {
				t.Fatalf("allowed=%t, got %v", tt.allowed, err)
			}
		})
	}
	// No parent routes nowhere. Omission of namespace means the route namespace,
	// not the configured namespace; accept it when those genuinely match.
	route := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "infra"}}
	if err := validateGatewayParents(route, constraints); err != nil {
		t.Fatal(err)
	}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{{Name: "external"}}
	if err := validateGatewayParents(route, constraints); err != nil {
		t.Fatal(err)
	}
	if err := validateGatewayParents(route, routingConstraints{}); !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("unconfigured Gateway must not allow arbitrary bindings: %v", err)
	}
}

func TestRoutingExtensionsAggregateChecks(t *testing.T) {
	t.Parallel()
	route := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"unrecognized": "secret"}}}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{{Name: "foreign"}}
	route.Spec.Rules = []gatewayv1.HTTPRouteRule{{Filters: []gatewayv1.HTTPRouteFilter{{Type: "Unknown"}}}}
	err := validateExposureExtensions(route, routingConstraints{})
	for _, part := range []string{"parentRefs[0]", "filter[0]", "unrecognized"} {
		if err == nil || !strings.Contains(err.Error(), part) {
			t.Fatalf("missing %q from aggregated error: %v", part, err)
		}
	}
	// Route is also part of the shared gate, without an Ingress/Gateway binding.
	if err := validateExposureExtensions(&openshiftv1.Route{}, routingConstraints{}); err != nil {
		t.Fatal(err)
	}
}

func TestConfiguredIngressAnnotationsDoNotAuthorizeExtensions(t *testing.T) {
	t.Setenv("INGRESS_ANNOTATIONS", `{"platform.example/default":"approved"}`)
	constraints := configuredRoutingConstraints()
	for _, object := range []client.Object{
		&networkingv1.Ingress{},
		&openshiftv1.Route{},
	} {
		object.SetAnnotations(map[string]string{"platform.example/default": "approved"})
		if err := validateRoutingAnnotations(object, constraints); !errors.Is(err, errOwnershipViolation) {
			t.Fatalf("configured default must not authorize an unsupported annotation on %T: %v", object, err)
		}
	}
}
