/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"fmt"
	"net/url"
	"sort"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/builders"
	network "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func sortIngressesByName(ingresses map[string]*network.Ingress) []*network.Ingress {
	// Create a slice to hold the names
	names := make([]string, 0, len(ingresses))

	// Extract the names from the map
	for name := range ingresses {
		names = append(names, name)
	}

	// Sort the names
	sort.Strings(names)

	// Create a new slice to hold the sorted ingresses
	sortedIngresses := make([]*network.Ingress, 0, len(ingresses))

	// Append ingresses to the new slice in sorted order
	for _, name := range names {
		sortedIngresses = append(sortedIngresses, ingresses[name])
	}

	return sortedIngresses
}

func sortRoutesByName(routes map[string]*routev1.Route) []*routev1.Route {
	names := make([]string, 0, len(routes))

	for name := range routes {
		names = append(names, name)
	}

	// Sort the names
	sort.Strings(names)

	sortedRoutes := make([]*routev1.Route, 0, len(routes))

	// Append routes to the new slice in sorted order
	for _, name := range names {
		sortedRoutes = append(sortedRoutes, routes[name])
	}

	return sortedRoutes
}

func BuildURLFromIngress(ingress *network.Ingress) (*url.URL, error) {
	if ingress == nil {
		return nil, fmt.Errorf("ingress is nil")
	}

	if len(ingress.Spec.Rules) == 0 {
		return nil, fmt.Errorf("ingress has no rules")
	}

	rule := ingress.Spec.Rules[0]
	host := rule.Host
	scheme := "https" // cluster ingress is configured to use HTTPS by default

	// Use the first path if available, otherwise use root path
	path := "/"
	if len(rule.HTTP.Paths) > 0 {
		path = rule.HTTP.Paths[0].Path
		// Strip the regex pattern used for path rewriting
		path = builders.StripPathRewritePattern(path)
	}

	return url.Parse(fmt.Sprintf("%s://%s%s", scheme, host, path))
}

func BuildURLFromRoute(route *routev1.Route) (*url.URL, error) {
	if route == nil {
		return nil, fmt.Errorf("route is nil")
	}

	host := route.Spec.Host
	if host == "" {
		return nil, fmt.Errorf("route has no host")
	}

	scheme := "https"

	path := "/"
	if route.Spec.Path != "" {
		path = route.Spec.Path
		path = builders.StripPathRewritePattern(path)
	}

	return url.Parse(fmt.Sprintf("%s://%s%s", scheme, host, path))
}

func BuildURLFromHTTPRoute(httpRoute *gatewayv1.HTTPRoute) (*url.URL, error) {
	if httpRoute == nil {
		return nil, fmt.Errorf("httpRoute is nil")
	}

	if len(httpRoute.Spec.Hostnames) == 0 {
		return nil, fmt.Errorf("httpRoute has no hostnames")
	}

	host := string(httpRoute.Spec.Hostnames[0])
	scheme := "https"

	path := "/"
	if len(httpRoute.Spec.Rules) > 0 && len(httpRoute.Spec.Rules[0].Matches) > 0 {
		match := httpRoute.Spec.Rules[0].Matches[0]
		if match.Path != nil && match.Path.Value != nil {
			path = *match.Path.Value
		}
	}

	return url.Parse(fmt.Sprintf("%s://%s%s", scheme, host, path))
}

func sortHTTPRoutesByName(httpRoutes map[string]*gatewayv1.HTTPRoute) []*gatewayv1.HTTPRoute {
	names := make([]string, 0, len(httpRoutes))
	for name := range httpRoutes {
		names = append(names, name)
	}
	sort.Strings(names)

	sorted := make([]*gatewayv1.HTTPRoute, 0, len(httpRoutes))
	for _, name := range names {
		sorted = append(sorted, httpRoutes[name])
	}
	return sorted
}
