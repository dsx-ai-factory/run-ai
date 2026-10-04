/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestHTTPRouteFilterSchema(t *testing.T) {
	t.Parallel()
	// New union members must be reviewed even when paired with an allowed Type.
	const expectedFields = 9 // Type and eight configuration members.
	if fields := reflect.TypeFor[gatewayv1.HTTPRouteFilter]().NumField(); fields != expectedFields {
		t.Fatalf("HTTPRouteFilter has %d fields, expected %d; update validation for every union member", fields, expectedFields)
	}
}

func TestValidateHTTPRouteFilters(t *testing.T) {
	t.Parallel()

	allowed := []gatewayv1.HTTPRouteFilter{
		{Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier, RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{}},
		{Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier, ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{}},
		{Type: gatewayv1.HTTPRouteFilterURLRewrite, URLRewrite: &gatewayv1.HTTPURLRewriteFilter{}},
		{Type: gatewayv1.HTTPRouteFilterRequestRedirect, RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{}},
		{Type: gatewayv1.HTTPRouteFilterCORS, CORS: &gatewayv1.HTTPCORSFilter{}},
		{Type: gatewayv1.HTTPRouteFilterRequestMirror, RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
			BackendRef: gatewayBackend("mirror"),
		}},
	}
	type testCase struct {
		name   string
		filter gatewayv1.HTTPRouteFilter
		denied bool
	}
	tests := []testCase{
		{
			name: "extension reference",
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: "example.com", Kind: "CustomFilter", Name: "custom",
				},
			},
			denied: true,
		},
		{
			name: "external authorization",
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExternalAuth,
				ExternalAuth: &gatewayv1.HTTPExternalAuthFilter{
					BackendRef: gatewayBackend("auth"),
				},
			},
			denied: true,
		},
		{name: "unknown type", filter: gatewayv1.HTTPRouteFilter{Type: "FutureFilter"}, denied: true},
		{name: "empty type", filter: gatewayv1.HTTPRouteFilter{}, denied: true},
		{
			name: "mismatched configuration",
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite, RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{},
			},
			denied: true,
		},
		{
			name: "allowed type hiding external authorization",
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterCORS, CORS: &gatewayv1.HTTPCORSFilter{},
				ExternalAuth: &gatewayv1.HTTPExternalAuthFilter{BackendRef: gatewayBackend("foreign")},
			},
			denied: true,
		},
		{
			name: "allowed type hiding extension reference",
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterCORS, CORS: &gatewayv1.HTTPCORSFilter{},
				ExtensionRef: &gatewayv1.LocalObjectReference{Name: "custom"},
			},
			denied: true,
		},
	}
	for _, filter := range allowed {
		tests = append(tests,
			testCase{name: string(filter.Type), filter: filter},
			testCase{name: string(filter.Type) + " missing configuration", filter: gatewayv1.HTTPRouteFilter{Type: filter.Type}, denied: true},
		)
	}
	for _, scope := range []string{"rule", "backend"} {
		for _, tt := range tests {
			t.Run(scope+"/"+tt.name, func(t *testing.T) {
				route := httpRouteWithFilter(scope, tt.filter)
				err := validateHTTPRouteFilters(route)
				if tt.denied {
					if !errors.Is(err, errOwnershipViolation) {
						t.Fatalf("expected ownership violation, got %v", err)
					}
					if !strings.Contains(err.Error(), "filter[0]") {
						t.Fatalf("error does not identify the filter: %v", err)
					}
				} else if err != nil {
					t.Fatalf("expected supported filter, got %v", err)
				}
			})
		}
	}
}

func TestValidateHTTPRouteFiltersAggregatesScopes(t *testing.T) {
	t.Parallel()
	route := httpRouteWithFilter("rule", gatewayv1.HTTPRouteFilter{Type: "UnknownRuleFilter"})
	backendRoute := httpRouteWithFilter("backend", gatewayv1.HTTPRouteFilter{Type: "UnknownBackendFilter"})
	route.Spec.Rules = append(route.Spec.Rules, backendRoute.Spec.Rules...)
	err := validateHTTPRouteFilters(route)
	if !errors.Is(err, errOwnershipViolation) {
		t.Fatalf("expected ownership violation, got %v", err)
	}
	for _, location := range []string{"rule[0] filter[0]", "rule[1].backend[0] filter[0]"} {
		if !strings.Contains(err.Error(), location) {
			t.Errorf("error is missing %s: %v", location, err)
		}
	}
}

func httpRouteWithFilter(scope string, filter gatewayv1.HTTPRouteFilter) *gatewayv1.HTTPRoute {
	rule := gatewayv1.HTTPRouteRule{}
	if scope == "rule" {
		rule.Filters = []gatewayv1.HTTPRouteFilter{filter}
	} else {
		rule.BackendRefs = []gatewayv1.HTTPBackendRef{{Filters: []gatewayv1.HTTPRouteFilter{filter}}}
	}
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: ownershipNamespace, Name: "http-route"},
		Spec:       gatewayv1.HTTPRouteSpec{Rules: []gatewayv1.HTTPRouteRule{rule}},
	}
}
