/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	nimv1alpha1 "github.com/NVIDIA/k8s-nim-operator/api/apps/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	openshiftv1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ingressClassHAProxy = "haproxy"

	annotationNginxBackendProtocol    = "nginx.ingress.kubernetes.io/backend-protocol"
	annotationNginxRewriteTarget      = "nginx.ingress.kubernetes.io/rewrite-target"
	annotationHAProxyBackendProtocol  = "haproxy.ingress.kubernetes.io/backend-protocol"
	annotationHAProxyRewriteTarget    = "haproxy.ingress.kubernetes.io/rewrite-target"
	annotationHAProxyRequestSetHeader = "haproxy.ingress.kubernetes.io/request-set-header"
	backendProtocolGRPCNginx          = "GRPC"
	backendProtocolGRPCHAProxy        = "grpc"
	haproxyForwardedUriHeader         = "X-Forwarded-Uri %[req.uri]"
)

type nimServiceNetworkingInfo struct {
	AutoExposureUrl         *string
	IngressClass            string
	IngressEnabled          *bool
	UseOpenShiftRoute       bool
	UseWorkloadV2Networking bool
	UseGatewayAPI           bool
	GatewayName             string
	GatewayNamespace        string
}

func nimServingPortExposureWanted(sp *api.NimServiceServingPort) bool {
	if sp == nil {
		return false
	}
	if sp.ExposeExternally == nil {
		return true // nil means expose by default
	}
	return *sp.ExposeExternally
}

func validateNimServingPortUrlRequiresExposure(sp *api.NimServiceServingPort) *api.Error {
	if sp == nil || sp.ExposedUrl == nil || *sp.ExposedUrl == "" {
		return nil
	}
	if nimServingPortExposureWanted(sp) {
		return nil
	}
	return api.NewError(http.StatusBadRequest, "servingPort: exposedUrl requires exposeExternally to be true", nil)
}

func validateNimServingPortAuthFieldsConsistency(sp *api.NimServiceServingPort) *api.Error {
	if sp == nil {
		return nil
	}

	hasUsersField := sp.AuthorizedUsers != nil
	hasGroupsField := sp.AuthorizedGroups != nil
	hasUsers := hasUsersField && len(*sp.AuthorizedUsers) > 0
	hasGroups := hasGroupsField && len(*sp.AuthorizedGroups) > 0

	if hasUsers && hasGroups {
		return api.NewError(http.StatusBadRequest, "authorizedUsers and authorizedGroups are mutually exclusive", nil)
	}

	if sp.AuthorizationType == nil {
		if hasUsersField || hasGroupsField {
			return api.NewError(http.StatusBadRequest, "authorizationType is required when authorizedUsers or authorizedGroups is specified", nil)
		}
		return nil
	}
	t := *sp.AuthorizationType

	switch t {
	case api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC,
		api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS:
		if sp.AuthorizedUsers != nil || sp.AuthorizedGroups != nil {
			return api.NewError(http.StatusBadRequest,
				fmt.Sprintf("authorizedUsers and authorizedGroups cannot be set for %s authorization type", t), nil)
		}

	case api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS:
		if sp.AuthorizedGroups != nil {
			return api.NewError(http.StatusBadRequest, "authorizedGroups cannot be set for authorizedUsers authorization type", nil)
		}
		if !hasUsers {
			return api.NewError(http.StatusBadRequest, "authorizedUsers must be provided and non-empty for authorizedUsers authorization type", nil)
		}

	case api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS:
		if sp.AuthorizedUsers != nil {
			return api.NewError(http.StatusBadRequest, "authorizedUsers cannot be set for authorizedGroups authorization type", nil)
		}
		if !hasGroups {
			return api.NewError(http.StatusBadRequest, "authorizedGroups must be provided and non-empty for authorizedGroups authorization type", nil)
		}
	}

	return nil
}

func validateNimServingPortAuthRequiresExposure(sp *api.NimServiceServingPort) *api.Error {
	if sp == nil || !hasNimServingPortAuth(sp) {
		return nil
	}
	if nimServingPortExposureWanted(sp) {
		return nil
	}
	return api.NewError(http.StatusBadRequest, "servingPort: authorization requires external exposure (set exposeExternally to true)", nil)
}

func validateNimServingPortAuthSupportedForClusterVersion(ctx context.Context, clusterVersion string, sp *api.NimServiceServingPort) *api.Error {
	if sp == nil || !hasNimServingPortAuth(sp) {
		return nil
	}
	if clusterVersion == "" {
		return api.NewError(http.StatusBadRequest, "servingPort authorization fields are supported for cluster version 2.25 and above", nil)
	}
	cap, err := checker.NewClusterCapabilitiesCheckerForVersion(clusterVersion)
	if err != nil {
		return api.NewError(http.StatusBadRequest, fmt.Sprintf("cluster version: %v", err), nil)
	}
	supported, err := cap.ValidateClusterCapability(ctx, checker.MinVersionNIMIngressViaWorkloadV2)
	if err != nil {
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("cluster capability check: %v", err), nil)
	}
	if supported {
		return nil
	}
	return api.NewError(http.StatusBadRequest, "servingPort authorization fields are supported for cluster version 2.25 and above", nil)
}

func nimServingPortToNetworkingConfig(workloadName string, sp *api.NimServiceServingPort, info nimServiceNetworkingInfo) (*[]api.NetworkingEndpoint, *api.Error) {
	if sp == nil || sp.Port == nil || *sp.Port < 1 {
		return nil, nil
	}

	hasAuth := hasNimServingPortAuth(sp)
	hasExpose := nimServingPortExposureWanted(sp)
	if !hasAuth && !hasExpose {
		return nil, nil
	}

	port := int32(*sp.Port)
	endpoint := api.NetworkingEndpoint{}

	if hasAuth {
		// ServiceRef is set to workloadName because the NIM operator creates a K8s Service named after the workload.
		svc := api.NetworkingService{
			Ref: &api.ServiceRef{
				Name: workloadName,
				Port: port,
			},
			Authorization: nimServingPortToAuthorization(sp),
		}
		services := []api.NetworkingService{svc}
		endpoint.Services = &services
	}

	if hasExpose {
		externalAccess, err := nimServingPortToExternalAccess(sp, workloadName, port, info, hasAuth)
		if err != nil {
			return nil, err
		}
		endpoint.ExternalAccess = externalAccess
	}

	return &[]api.NetworkingEndpoint{endpoint}, nil
}

func hasNimServingPortAuth(sp *api.NimServiceServingPort) bool {
	if sp.AuthorizationType == nil {
		return false
	}
	t := *sp.AuthorizationType
	return t == api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS ||
		t == api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS ||
		t == api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS
}

func nimServingPortToAuthorization(sp *api.NimServiceServingPort) *api.Workloadv2Authorization {
	if sp == nil || sp.AuthorizationType == nil {
		return nil
	}
	t := *sp.AuthorizationType
	if t == api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC {
		return nil
	}
	authType := api.AuthorizationType(t)
	out := &api.Workloadv2Authorization{
		AuthorizationType: &authType,
	}
	if sp.AuthorizedUsers != nil {
		out.AuthorizedUsers = sp.AuthorizedUsers
	}
	if sp.AuthorizedGroups != nil {
		out.AuthorizedGroups = sp.AuthorizedGroups
	}
	return out
}

func nimServingPortToExternalAccess(sp *api.NimServiceServingPort, serviceName string, servicePort int32, info nimServiceNetworkingInfo, hasAuth bool) (*api.ExternalAccess, *api.Error) {
	urlStr := resolveExposureUrl(sp, info.AutoExposureUrl)
	if urlStr == "" {
		return nil, nil
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("invalid URL %q: %v", urlStr, err), nil)
	}

	host := u.Hostname()
	if host == "" {
		return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("invalid URL %q: missing host", urlStr), nil)
	}

	path := u.Path
	if path == "" {
		path = "/"
	}

	ea := &api.ExternalAccess{}
	ingressRouteEnabled := nimIngressRouteEnabled(info.IngressEnabled)

	if ingressRouteEnabled {
		if info.UseOpenShiftRoute {
			ea.Route, err = buildOpenShiftRouteMap(host, path, serviceName, hasAuth)
			if err != nil {
				return nil, api.NewError(http.StatusInternalServerError, "failed to build NIM OpenShift Route", err)
			}
		} else {
			ingressEA := buildIngressExternalAccess(host, path, resolveProtocol(sp), info.IngressClass, serviceName, servicePort, hasAuth)
			if ingressEA != nil {
				ea.Ingress = ingressEA.Ingress
			}
		}
	}
	if info.UseGatewayAPI {
		if info.GatewayName == "" {
			return nil, api.NewError(http.StatusServiceUnavailable, "cluster Gateway API is enabled but gateway name is not reported", nil)
		}
		if httpRouteEA := buildHTTPRouteExternalAccess(host, path, resolveProtocol(sp), serviceName, servicePort, info.GatewayName, info.GatewayNamespace); httpRouteEA != nil {
			ea.HttpRoute = httpRouteEA.HttpRoute
		}
	}

	if ea.Ingress == nil && ea.Route == nil && ea.HttpRoute == nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "cluster has no enabled external URL routing mechanism", nil)
	}

	return ea, nil
}

func nimIngressRouteEnabled(enabled *bool) bool {
	return enabled == nil || *enabled
}

func effectiveNimGatewayAPIEnabled(reported *bool, dependencyAvailable bool, gatewayName, gatewayNamespace string) bool {
	if reported != nil {
		return *reported
	}
	if gatewayName == "" {
		return false
	}
	return dependencyAvailable && gatewayNamespace != ""
}

func resolveExposureUrl(sp *api.NimServiceServingPort, autoGeneratedUrl *string) string {
	if sp.ExposedUrl != nil && *sp.ExposedUrl != "" {
		return *sp.ExposedUrl
	}
	if autoGeneratedUrl != nil && *autoGeneratedUrl != "" {
		return *autoGeneratedUrl
	}
	return ""
}

func resolveProtocol(sp *api.NimServiceServingPort) api.ServingPortExposedProtocol {
	if sp.ExposedProtocol != nil {
		return *sp.ExposedProtocol
	}
	if sp.GrpcPort != nil && *sp.GrpcPort > 0 {
		return api.SERVINGPORTEXPOSEDPROTOCOL_GRPC
	}
	return ""
}

func buildIngressExternalAccess(host, path string, protocol api.ServingPortExposedProtocol, ingressClass, serviceName string, servicePort int32, hasAuth bool) *api.ExternalAccess {
	pathType := networkingv1.PathTypePrefix
	annotations := make(map[string]string)

	if path != "" && path != "/" {
		path = strings.TrimSuffix(path, "/") + "(/|$)(.*)"
		pathType = networkingv1.PathTypeImplementationSpecific
		annotations = pathRewriteAnnotations(ingressClass)
		if hasAuth {
			for k, v := range authForwardedUriAnnotations(ingressClass) {
				annotations[k] = v
			}
		}
	}

	if protocol == api.SERVINGPORTEXPOSEDPROTOCOL_GRPC {
		if strings.ToLower(ingressClass) == ingressClassHAProxy {
			annotations[annotationHAProxyBackendProtocol] = backendProtocolGRPCHAProxy
		} else {
			annotations[annotationNginxBackendProtocol] = backendProtocolGRPCNginx
		}
	}

	ing := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.NetworkingV1,
			Kind:       constants.KindIngress,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     path,
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: serviceName,
									Port: networkingv1.ServiceBackendPort{Number: servicePort},
								},
							},
						}},
					},
				},
			}},
		},
		ObjectMeta: metav1.ObjectMeta{Annotations: annotations},
	}
	var m map[string]interface{}
	if b, err := json.Marshal(ing); err == nil && json.Unmarshal(b, &m) == nil {
		ea := &api.ExternalAccess{Ingress: &m}
		return ea
	}
	return nil
}

func buildHTTPRouteExternalAccess(host, path string, _ api.ServingPortExposedProtocol, serviceName string, servicePort int32, gatewayName, gatewayNamespace string) *api.ExternalAccess {
	if gatewayName == "" {
		return nil
	}
	if path == "/" {
		path = ""
	}

	pathValue := "/"
	if path != "" {
		pathValue = path
	}

	rule := map[string]interface{}{
		"matches": []map[string]interface{}{
			{
				"path": map[string]interface{}{
					"type":  "PathPrefix",
					"value": pathValue,
				},
			},
		},
		"backendRefs": []map[string]interface{}{
			{
				"name": serviceName,
				"port": servicePort,
			},
		},
	}

	if path != "" {
		rule["filters"] = []map[string]interface{}{
			{
				"type": "URLRewrite",
				"urlRewrite": map[string]interface{}{
					"path": map[string]interface{}{
						"type":               "ReplacePrefixMatch",
						"replacePrefixMatch": "/",
					},
				},
			},
		}
	}

	httpRoute := map[string]interface{}{
		"apiVersion": constants.GatewayV1,
		"kind":       constants.KindHTTPRoute,
		"spec": map[string]interface{}{
			"parentRefs": []map[string]interface{}{
				{
					"name": gatewayName,
				},
			},
			"hostnames": []string{host},
			"rules":     []map[string]interface{}{rule},
		},
	}

	if gatewayNamespace != "" {
		parentRefs := httpRoute["spec"].(map[string]interface{})["parentRefs"].([]map[string]interface{})
		parentRefs[0]["namespace"] = gatewayNamespace
	}

	return &api.ExternalAccess{HttpRoute: &httpRoute}
}

func buildOpenShiftRouteMap(host, path, serviceName string, hasAuth bool) (*map[string]interface{}, error) {
	route := &openshiftv1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.RouteV1,
			Kind:       constants.KindRoute,
		},
		Spec: openshiftv1.RouteSpec{
			Host: host,
			Path: path,
			To: openshiftv1.RouteTargetReference{
				Kind: constants.KindService,
				Name: serviceName,
			},
		},
	}
	// Authenticated Routes are redirected to a single-port auth proxy by the plugin,
	// which rewrites the backend Service but leaves the Route's port unchanged.
	if !hasAuth {
		route.Spec.Port = &openshiftv1.RoutePort{TargetPort: intstr.FromString(nimv1alpha1.DefaultNamedPortAPI)}
	}

	var routeMap map[string]interface{}
	if err := convutils.CopySpecs(&routeMap, route); err != nil {
		return nil, fmt.Errorf("convert OpenShift Route to networking configuration: %w", err)
	}
	// Omit fields that the typed API serializes even when unset.
	delete(routeMap, "metadata")
	delete(routeMap, "status")
	unstructured.RemoveNestedField(routeMap, "spec", "to", "weight")
	return &routeMap, nil
}

func pathRewriteAnnotations(ingressClass string) map[string]string {
	if strings.EqualFold(ingressClass, ingressClassHAProxy) {
		return map[string]string{annotationHAProxyRewriteTarget: "/\\1"}
	}
	return map[string]string{annotationNginxRewriteTarget: "/$2"}
}

func authForwardedUriAnnotations(ingressClass string) map[string]string {
	if strings.EqualFold(ingressClass, ingressClassHAProxy) {
		return map[string]string{annotationHAProxyRequestSetHeader: haproxyForwardedUriHeader}
	}
	return map[string]string{}
}
