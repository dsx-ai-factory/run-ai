/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"strings"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// firstService returns endpoint.Services[0] or nil; admission caps Services at one per Endpoint.
func firstService(endpoint *endpoints.Endpoint) *endpoints.Service {
	if endpoint == nil || len(endpoint.Services) == 0 {
		return nil
	}
	return &endpoint.Services[0]
}

const (
	AuthProxySuffix     = "-auth-proxy"
	LabelApp            = "app"
	LabelWorkloadRunner = "run.ai/workload-runner"
)

func isMetricsService(svc *corev1.Service) bool {
	return svc != nil && svc.Labels[constants.LabelWorkloadMetricsService] == "true"
}

// WorkloadServiceAction indicates whether the workload service is create or ref.
type WorkloadServiceAction string

const (
	WorkloadServiceActionCreate WorkloadServiceAction = "create"
	WorkloadServiceActionRef    WorkloadServiceAction = "ref"
	WorkloadServiceActionNone   WorkloadServiceAction = ""
)

// WorkloadServiceInfo holds name, port, and action (create/ref) for the workload service.
type WorkloadServiceInfo struct {
	Name   string
	Port   int32
	Action WorkloadServiceAction
}

func extractNetworkingConfig(runner *dromosv1alpha1.WorkloadRunner) *dromosv1alpha1.NetworkingConfig {
	if runner.Spec.Instructions == nil {
		return nil
	}
	return runner.Spec.Instructions.Networking
}

func hasService(config *dromosv1alpha1.NetworkingConfig) bool {
	return config != nil && len(config.Services) > 0
}

func shouldReportNetworkingStatus(config *dromosv1alpha1.NetworkingConfig) bool {
	return config != nil && (hasService(config) || config.ExternalAccess != nil)
}

func isAuthEnabled(config *dromosv1alpha1.NetworkingConfig) bool {
	if !hasService(config) {
		return false
	}
	return config.Services[0].Auth != nil
}

func getWorkloadServiceInfo(config *dromosv1alpha1.NetworkingConfig, defaultName string) WorkloadServiceInfo {
	if !hasService(config) {
		return WorkloadServiceInfo{Name: defaultName, Port: 0, Action: WorkloadServiceActionNone}
	}

	svc := config.Services[0]

	if svc.Create != nil {
		name := svc.Create.Name
		if name == "" {
			name = defaultName
		}
		port := int32(0)
		if len(svc.Create.Spec.Ports) > 0 {
			port = svc.Create.Spec.Ports[0].Port
		}
		return WorkloadServiceInfo{Name: name, Port: port, Action: WorkloadServiceActionCreate}
	}

	if svc.Ref != nil {
		return WorkloadServiceInfo{Name: svc.Ref.Name, Port: svc.Ref.Port, Action: WorkloadServiceActionRef}
	}

	return WorkloadServiceInfo{Name: defaultName, Port: 0, Action: WorkloadServiceActionNone}
}

func getResourceName(specName string, runner *dromosv1alpha1.WorkloadRunner) string {
	if specName != "" {
		return specName
	}
	return runner.Name
}

// Legacy auth-proxy Service and Deployment share this name by design (Service selects Deployment by label).
func getAuthProxyName(runner *dromosv1alpha1.WorkloadRunner) string {
	return runner.Name + AuthProxySuffix
}

func getExternalHostAndPath(config *dromosv1alpha1.NetworkingConfig) (host, path string) {
	if config.ExternalAccess == nil {
		return "", "/"
	}
	ea := config.ExternalAccess
	return resolveExternalHostAndPath(ea.Ingress, ea.Route, ea.HTTPRoute)
}

// Last non-nil kind wins; migration payloads may carry HTTPRoute alongside Ingress or Route.
func resolveExternalHostAndPath(
	ingress *networkingv1.Ingress,
	route *openshiftv1.Route,
	httpRoute *gatewayv1.HTTPRoute,
) (host, path string) {
	if ingress != nil {
		host, path = ingressHostAndPath(ingress)
	}
	if route != nil {
		host, path = routeHostAndPath(route)
	}
	if httpRoute != nil {
		host, path = httpRouteHostAndPath(httpRoute)
	}
	if path == "" {
		path = "/"
	}
	return host, path
}

func ingressHostAndPath(ingress *networkingv1.Ingress) (host, path string) {
	if len(ingress.Spec.Rules) == 0 {
		return "", ""
	}
	host = ingress.Spec.Rules[0].Host
	if ingress.Spec.Rules[0].HTTP != nil && len(ingress.Spec.Rules[0].HTTP.Paths) > 0 {
		// Strip nginx-style rewrite capture-group suffix so oauth2-proxy sees the user-facing prefix
		// (e.g. "/api/v1(/|$)(.*)" -> "/api/v1").
		rawPath := ingress.Spec.Rules[0].HTTP.Paths[0].Path
		if idx := strings.Index(rawPath, "(/|$)"); idx >= 0 {
			path = rawPath[:idx]
		} else {
			path = rawPath
		}
	}
	return host, path
}

func routeHostAndPath(route *openshiftv1.Route) (host, path string) {
	return route.Spec.Host, route.Spec.Path
}

func httpRouteHostAndPath(hr *gatewayv1.HTTPRoute) (host, path string) {
	if len(hr.Spec.Hostnames) > 0 {
		host = string(hr.Spec.Hostnames[0])
	}
	if len(hr.Spec.Rules) > 0 && len(hr.Spec.Rules[0].Matches) > 0 {
		match := hr.Spec.Rules[0].Matches[0]
		if match.Path != nil && match.Path.Value != nil {
			path = *match.Path.Value
		}
	}
	return host, path
}
