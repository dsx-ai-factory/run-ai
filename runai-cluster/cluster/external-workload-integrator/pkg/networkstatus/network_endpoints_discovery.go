/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networkstatus

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/ptr"
	"net/url"
	"sort"

	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconf "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	infv1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"strconv"
	"strings"
)

const (
	kindService       = "Service"
	kindIngress       = "Ingress"
	kindHTTPRoute     = "HTTPRoute"
	kindGRPCRoute     = "GRPCRoute"
	kindTLSRoute      = "TLSRoute"
	kindInferencePool = "InferencePool"
	kindRoute         = "Route"

	protocolHTTP  = "http"
	protocolHTTPS = "https"
	port80        = "80"
	port443       = "443"

	controlPlaneNodeLabel    = "node-role.kubernetes.io/control-plane"
	externalNodeIPAnnotation = "runai/external-node-ip"
)

// appProtocolInfo holds the canonical scheme and TLS flag for a Kubernetes AppProtocol value.
type appProtocolInfo struct {
	scheme string
	tls    bool
}

// knownAppProtocols maps Kubernetes reserved AppProtocol strings to their
// canonical HTTP scheme and TLS flag.
// See https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol
var knownAppProtocols = map[string]appProtocolInfo{
	protocolHTTP:        {scheme: protocolHTTP, tls: false},
	protocolHTTPS:       {scheme: protocolHTTPS, tls: true},
	"kubernetes.io/h2c": {scheme: protocolHTTP, tls: false},
	"kubernetes.io/ws":  {scheme: protocolHTTP, tls: false},
	"kubernetes.io/wss": {scheme: protocolHTTPS, tls: true},
	"h2":                {scheme: protocolHTTPS, tls: true},
	"h2c":               {scheme: protocolHTTP, tls: false},
}

// normalizeAppProtocol returns the canonical scheme and TLS flag for the given
// AppProtocol string. Unknown protocols default to "http" / no TLS.
func normalizeAppProtocol(appProtocol string) (scheme string, tlsEnabled bool) {
	if info, ok := knownAppProtocols[strings.ToLower(appProtocol)]; ok {
		return info.scheme, info.tls
	}
	return protocolHTTP, false
}

// gatewayRouteRule represents a single rule extracted from a Gateway API route,
// containing the backend refs that may target our service and the paths to expose.
type gatewayRouteRule struct {
	backendRefs []gwv1.BackendObjectReference
	paths       []string // for HTTPRoute these come from matches; for GRPC/TLS always ["/"]
}

// gatewayRoute holds the common fields extracted from any Gateway API route type.
type gatewayRoute struct {
	name         string
	namespace    string
	kind         string
	hostnames    []gwv1.Hostname
	protocols    []gwv1.ProtocolType
	rules        []gatewayRouteRule
	parentStatus []gwv1.RouteParentStatus
}

func (n *NetworkStatusUpdater) DiscoverEndpointsForEwAndPods(ctx context.Context, externalWorkload *v1alpha1.ExternalWorkload, pods []corev1.Pod) ([]v1alpha1.ExternalEndpoint, error) {

	var endpoints []v1alpha1.ExternalEndpoint
	if externalWorkload == nil {
		return endpoints, nil
	}

	namespace := externalWorkload.Namespace
	services, err := n.findServicesForPods(ctx, pods, namespace)
	if err != nil {
		return nil, err
	}
	inferencePools := n.findInferencePoolsForPods(ctx, pods, namespace)

	var httpRouteList *gwv1.HTTPRouteList
	if len(inferencePools) > 0 {
		endpoints, httpRouteList, err = n.discoverEndpointsFromInferencePools(ctx, namespace, inferencePools)
		if err != nil {
			return nil, err
		}
	}
	if len(services) > 0 {
		nodeAddress := n.findNodeAddress(ctx, pods)
		serviceEndpoints, err := n.discoverEndpointsFromServices(ctx, namespace, nodeAddress, services, httpRouteList)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, serviceEndpoints...)
	}

	if len(externalWorkload.Spec.Urls) > 0 {
		// for specific types which adds urls, parse them to the endpoints as well
		urlsToEndpoints := n.parseUrlsToEndpoints(externalWorkload.Spec.Urls, endpoints)
		endpoints = append(endpoints, urlsToEndpoints...)
	}

	return endpoints, nil
}

func (n *NetworkStatusUpdater) findNodeAddress(ctx context.Context, pods []corev1.Pod) string {
	nodeNames := []string{}
	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			nodeNames = append(nodeNames, pod.Spec.NodeName)
		}
	}
	// single node address is needed for nodeport ip
	nodeAddress := ""
	if len(nodeNames) > 0 {
		sort.Slice(nodeNames, func(i, j int) bool {
			return nodeNames[i] < nodeNames[j]
		})
		nodeAddress = getNodeAddress(ctx, n.cachedClient, nodeNames[0])
	}

	if nodeAddress != "" {
		return nodeAddress
	}

	// Fallback: pick a worker node (skip control-plane nodes).
	return n.findNodeAddressFromCluster(ctx)
}

func (n *NetworkStatusUpdater) findNodeAddressFromCluster(ctx context.Context) string {
	nodeList := &corev1.NodeList{}
	if err := n.cachedClient.List(ctx, nodeList); err != nil {
		log.Debug().Msgf("NetworkStatusUpdater: failed listing nodes for fallback address: %s", err.Error())
		return ""
	}
	if len(nodeList.Items) == 0 {
		return ""
	}
	sort.Slice(nodeList.Items, func(i, j int) bool {
		return nodeList.Items[i].Name < nodeList.Items[j].Name
	})
	for i := range nodeList.Items {
		if _, exists := (&nodeList.Items[i]).Labels[controlPlaneNodeLabel]; exists {
			continue
		}
		if addr := getNodeAddressFromNode(&nodeList.Items[i]); addr != "" {
			return addr
		}
	}
	return ""
}

func (n *NetworkStatusUpdater) parseUrlsToEndpoints(urls []string, existingEndpoints []v1alpha1.ExternalEndpoint) []v1alpha1.ExternalEndpoint {
	var endpoints []v1alpha1.ExternalEndpoint
	existingEndpointsUrls := map[string]bool{}
	for _, endpoint := range existingEndpoints {
		urlStr := strings.TrimSuffix(endpoint.Url, "/")
		existingEndpointsUrls[urlStr] = true
	}
	for _, rawUrl := range urls {
		exists := existingEndpointsUrls[strings.TrimSuffix(rawUrl, "/")]
		if exists {
			continue
		}
		parsed, err := url.Parse(rawUrl)
		if err != nil || parsed.Host == "" {
			log.Warn().Msgf("parseUrlsToEndpoints: url invalid %q: %v", rawUrl, err)
			continue
		}

		scheme := strings.ToLower(parsed.Scheme)
		host := parsed.Hostname()
		port := parsed.Port()
		path := parsed.Path
		if path == "" {
			path = "/"
		}

		if port == "" {
			if scheme == protocolHTTPS {
				port = port443
			} else {
				port = port80
			}
		}

		tls := scheme == protocolHTTPS

		endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
			SourceType: v1alpha1.ExternalEndpointOtherSource,
			Host:       host,
			Path:       path,
			Port:       port,
			Protocol:   scheme,
			TlsEnabled: ptr.To(tls),
			Url:        rawUrl,
		})
	}
	return endpoints
}

// findInferencePoolsForPods returns all InferencePools whose selector matches at least one of the given pods.
func (n *NetworkStatusUpdater) findInferencePoolsForPods(ctx context.Context, pods []corev1.Pod, namespace string) []infv1.InferencePool {
	if len(pods) == 0 {
		return []infv1.InferencePool{}
	}

	poolList := &infv1.InferencePoolList{}
	if err := n.cachedClient.List(ctx, poolList, client.InNamespace(namespace)); err != nil {
		logListFail(err, "inferencePools", namespace)
		return nil
	}

	seen := map[types.UID]bool{}
	var matched []infv1.InferencePool
	for _, pool := range poolList.Items {
		if pool.Spec.Selector.MatchLabels == nil {
			continue
		}
		selector := make(map[string]string, len(pool.Spec.Selector.MatchLabels))
		for k, v := range pool.Spec.Selector.MatchLabels {
			selector[string(k)] = string(v)
		}
		for i := range pods {
			if labelsMatchSelector(pods[i].Labels, selector) && !seen[pool.UID] {
				seen[pool.UID] = true
				matched = append(matched, pool)
				break
			}
		}
	}
	return matched
}

func (n *NetworkStatusUpdater) discoverEndpointsFromInferencePools(ctx context.Context, namespace string, inferencePools []infv1.InferencePool) ([]v1alpha1.ExternalEndpoint, *gwv1.HTTPRouteList, error) {
	httpRouteList := &gwv1.HTTPRouteList{}
	if err := n.cachedClient.List(ctx, httpRouteList, client.InNamespace(namespace)); err != nil {
		logListFail(err, "httpRoutes", namespace)
		return nil, nil, err
	}
	if len(inferencePools) == 0 || len(httpRouteList.Items) == 0 {
		return nil, httpRouteList, nil
	}

	// Convert HTTPRoutes once and reuse across all pools.
	routes := httpRoutesToGatewayRoutes(httpRouteList)
	var endpoints []v1alpha1.ExternalEndpoint
	for i := range inferencePools {
		pool := &inferencePools[i]
		log.Debug().Msgf("looking for inferencepool %s endpoints", pool.Name)
		poolEndpoints := n.findGatewayRouteEndpoints(ctx, pool.Name, func(ref gwv1.BackendObjectReference) bool {
			return isInferencePoolRef(ref, pool.Name, pool.Namespace)
		}, routes)
		endpoints = append(endpoints, poolEndpoints...)
	}
	return endpoints, httpRouteList, nil
}

func logListFail(err error, listType string, namespace string) {
	errMsg := fmt.Sprintf("Failed to list %s in namespace %s", listType, namespace)
	if meta.IsNoMatchError(err) {
		log.Debug().Msgf("%s: type doesn't exist in the k8s api server", errMsg)
	} else {
		log.Warn().Msgf("%s: %s", errMsg, err.Error())
	}
}

func (n *NetworkStatusUpdater) discoverEndpointsFromServices(ctx context.Context, namespace string, nodeAddress string, services []corev1.Service, httpRouteList *gwv1.HTTPRouteList) ([]v1alpha1.ExternalEndpoint, error) {
	var endpoints []v1alpha1.ExternalEndpoint

	if len(services) == 0 {
		return endpoints, nil
	}

	// pre-load ingresses, httproutes and tlsroutes for the namespace (used for ClusterIP services)
	if httpRouteList == nil || len(httpRouteList.Items) == 0 {
		httpRouteList = &gwv1.HTTPRouteList{}
		if err := n.cachedClient.List(ctx, httpRouteList, client.InNamespace(namespace)); err != nil {
			logListFail(err, "httpRoutes", namespace)
			httpRouteList = nil
		}
	}

	grpcRouteList := &gwv1.GRPCRouteList{}
	if err := n.cachedClient.List(ctx, grpcRouteList, client.InNamespace(namespace)); err != nil {
		logListFail(err, "grpcRoutes", namespace)
		grpcRouteList = nil
	}

	tlsRouteList := &gwv1alpha2.TLSRouteList{}
	if err := n.cachedClient.List(ctx, tlsRouteList, client.InNamespace(namespace)); err != nil {
		logListFail(err, "tlsRoutes", namespace)
		tlsRouteList = nil
	}

	ingressList := &networkingv1.IngressList{}
	if err := n.cachedClient.List(ctx, ingressList, client.InNamespace(namespace)); err != nil {
		log.Warn().Msgf("Failed to list ingresses in namespace %s: %s", namespace, err.Error())
		ingressList = nil
	}

	routeList := &openshiftV1.RouteList{}
	isOpenshift := commonconf.IsOpenshift()
	if isOpenshift {
		if err := n.cachedClient.List(ctx, routeList, client.InNamespace(namespace)); err != nil {
			log.Warn().Msgf("Failed to list routes in namespace %s: %s", namespace, err.Error())
			routeList = nil
		}
	}

	for i := range services {
		svcEndpoints := n.extractEndpointsFromService(ctx, &services[i], ingressList, httpRouteList, grpcRouteList, tlsRouteList, routeList, nodeAddress, isOpenshift)
		endpoints = append(endpoints, svcEndpoints...)
	}

	return endpoints, nil
}

func (n *NetworkStatusUpdater) extractEndpointsFromService(
	ctx context.Context,
	svc *corev1.Service,
	ingressList *networkingv1.IngressList,
	httpRouteList *gwv1.HTTPRouteList,
	grpcRouteList *gwv1.GRPCRouteList,
	tlsRouteList *gwv1alpha2.TLSRouteList,
	routeList *openshiftV1.RouteList,
	nodeAddress string,
	isOpenShift bool,
) []v1alpha1.ExternalEndpoint {
	var endpoints []v1alpha1.ExternalEndpoint

	switch svc.Spec.Type {
	case corev1.ServiceTypeNodePort:
		endpoints = append(endpoints, n.findNodePortEndpoints(svc, nodeAddress)...)
	case corev1.ServiceTypeLoadBalancer:
		endpoints = append(endpoints, n.findLoadBalancerEndpoints(svc)...)
	case corev1.ServiceTypeClusterIP:
		endpoints = append(endpoints, n.findHTTPRouteEndpointsForService(ctx, svc, httpRouteList)...)
		endpoints = append(endpoints, n.findGRPCRouteEndpointsForService(ctx, svc, grpcRouteList)...)
		endpoints = append(endpoints, n.findTLSRouteEndpointsForService(ctx, svc, tlsRouteList)...)
		if isOpenShift {
			endpoints = append(endpoints, n.findRouteEndpointsForService(svc, routeList)...)
		} else {
			endpoints = append(endpoints, n.findIngressEndpointsForService(svc, ingressList)...)
		}
	}

	return endpoints
}

func (n *NetworkStatusUpdater) findNodePortEndpoints(svc *corev1.Service, nodeAddress string) []v1alpha1.ExternalEndpoint {
	log.Debug().Msgf("found nodeport service %s, extracting endpoints", svc.Name)
	var endpoints []v1alpha1.ExternalEndpoint
	if nodeAddress == "" {
		return endpoints
	}
	for _, port := range svc.Spec.Ports {
		if port.NodePort == 0 {
			continue
		}
		nodePortStr := strconv.FormatInt(int64(port.NodePort), 10)
		var tls *bool
		endpointUrl := fmt.Sprintf("%s:%s", nodeAddress, nodePortStr)
		if port.AppProtocol != nil {
			scheme, tlsEnabled := normalizeAppProtocol(*port.AppProtocol)
			tls = ptr.To(tlsEnabled)
			endpointUrl = getUrl(scheme, nodeAddress, nodePortStr, "/")
		}
		endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
			SourceType: v1alpha1.ExternalEndpointNodePortSource,
			Host:       nodeAddress,
			Path:       "/",
			Port:       nodePortStr,
			Protocol:   string(port.Protocol),
			TlsEnabled: tls,
			Url:        endpointUrl,
			SourceRef: v1alpha1.ExternalEndpointSourceRef{
				Kind: kindService,
				Name: svc.Name,
			},
		})
	}
	log.Debug().Msgf("found %d nodeport enpoints for service %s", len(endpoints), svc.Name)
	return endpoints
}

func (n *NetworkStatusUpdater) findLoadBalancerEndpoints(svc *corev1.Service) []v1alpha1.ExternalEndpoint {
	log.Debug().Msgf("found loadbalancer service %s, extracting endpoints", svc.Name)
	var endpoints []v1alpha1.ExternalEndpoint
	for _, ingress := range svc.Status.LoadBalancer.Ingress {
		host := ingress.IP
		if host == "" {
			host = ingress.Hostname
		}
		if host == "" {
			continue
		}
		for _, port := range svc.Spec.Ports {
			portStr := strconv.FormatInt(int64(port.Port), 10)
			var tls *bool
			endpointUrl := fmt.Sprintf("%s:%s", host, portStr)
			if port.AppProtocol != nil {
				scheme, tlsEnabled := normalizeAppProtocol(*port.AppProtocol)
				tls = ptr.To(tlsEnabled)
				endpointUrl = getUrl(scheme, host, portStr, "/")
			} else {
				tls = ptr.To(!config.Config().LbServiceEndpointsTlsByExplicitAnnotation || hasTlsAnnotation(svc.Annotations))
			}
			endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
				SourceType: v1alpha1.ExternalEndpointLoadBalanceSource,
				Host:       host,
				Path:       "/",
				Port:       portStr,
				Protocol:   string(port.Protocol),
				TlsEnabled: tls,
				Url:        endpointUrl,
				SourceRef: v1alpha1.ExternalEndpointSourceRef{
					Kind: kindService,
					Name: svc.Name,
				},
			})
		}
	}
	log.Debug().Msgf("found %d loadbalancer enpoints for service %s", len(endpoints), svc.Name)
	return endpoints
}

func (n *NetworkStatusUpdater) findIngressEndpointsForService(svc *corev1.Service, ingressList *networkingv1.IngressList) []v1alpha1.ExternalEndpoint {
	if ingressList == nil {
		return nil
	}
	var endpoints []v1alpha1.ExternalEndpoint
	log.Debug().Msgf("found ipcluster service %s, looking for ingress endpoints", svc.Name)

	clusterDomain := commonconf.GetClusterDomain()
	parsedUrlPattern, err := apis.ParseURL(clusterDomain)
	if err != nil || parsedUrlPattern == nil {
		return endpoints
	}
	port := parsedUrlPattern.URL().Port()

	for _, ing := range ingressList.Items {
		for _, rule := range ing.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service == nil || path.Backend.Service.Name != svc.Name {
					continue
				}
				log.Debug().Msgf("found ingress service %s, extracting endpoints", svc.Name)
				tlsEnabled := hasTLSForHost(&ing, rule.Host)
				protocol := protocolHTTP
				if tlsEnabled {
					protocol = protocolHTTPS
				}
				endpointPort := port
				if endpointPort == "" {
					if protocol == protocolHTTPS {
						endpointPort = port443
					} else {
						endpointPort = port80
					}
				}
				ingressPath := path.Path
				if ingressPath == "" {
					ingressPath = "/"
				}
				endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
					SourceType: v1alpha1.ExternalEndpointIngressSource,
					Host:       rule.Host,
					Port:       endpointPort,
					Path:       ingressPath,
					TlsEnabled: ptr.To(tlsEnabled),
					Protocol:   protocol,
					Url:        getUrl(protocol, rule.Host, endpointPort, ingressPath),
					SourceRef: v1alpha1.ExternalEndpointSourceRef{
						Kind: kindIngress,
						Name: ing.Name,
					},
				})
			}
		}
	}
	log.Debug().Msgf("found %d ingress enpoints for service %s", len(endpoints), svc.Name)
	return endpoints
}

func getRouteAdmitted(route *openshiftV1.Route) (bool, string) {
	for _, ing := range route.Status.Ingress {
		for _, cond := range ing.Conditions {
			if cond.Type == openshiftV1.RouteAdmitted && cond.Status == corev1.ConditionTrue {
				return true, ing.Host
			}
		}
	}
	return false, ""
}

func (n *NetworkStatusUpdater) findRouteEndpointsForService(svc *corev1.Service, routeList *openshiftV1.RouteList) []v1alpha1.ExternalEndpoint {
	if routeList == nil {
		return nil
	}
	endpoints := []v1alpha1.ExternalEndpoint{}
	for _, route := range routeList.Items {
		if !openShiftRouteReferencesService(&route, svc.Name) {
			continue
		}

		// Ensure the route is actually "Admitted" by a router
		isAdmitted, admittedHost := getRouteAdmitted(&route)

		if !isAdmitted {
			log.Debug().Msgf("Route %s found but not yet admitted by any router", route.Name)
			continue
		}

		// Use the host from status (admittedHost) as primary, fallback to spec
		host := admittedHost
		if host == "" {
			host = route.Spec.Host
		}

		routePath := route.Spec.Path
		if routePath == "" {
			routePath = "/"
		}
		tlsEnabled := route.Spec.TLS != nil
		protocol := protocolHTTP
		routePort := port80
		if tlsEnabled {
			protocol = protocolHTTPS
			routePort = port443
		}

		endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
			SourceType: v1alpha1.ExternalEndpointIngressSource,
			Host:       host,
			Port:       routePort,
			Path:       routePath,
			TlsEnabled: ptr.To(tlsEnabled),
			Protocol:   protocol,
			Url:        getUrl(protocol, host, routePort, routePath),
			SourceRef: v1alpha1.ExternalEndpointSourceRef{
				Kind: kindRoute,
				Name: route.Name,
			},
		})
	}
	return endpoints
}

// openShiftRouteReferencesService checks if a Route targets the given service name
// via its primary backend or any alternate backends.
func openShiftRouteReferencesService(route *openshiftV1.Route, svcName string) bool {
	if route.Spec.To.Name == svcName {
		return true
	}
	for _, backend := range route.Spec.AlternateBackends {
		if backend.Name == svcName {
			return true
		}
	}
	return false
}

func (n *NetworkStatusUpdater) findHTTPRouteEndpointsForService(ctx context.Context, svc *corev1.Service, httpRouteList *gwv1.HTTPRouteList) []v1alpha1.ExternalEndpoint {
	if httpRouteList == nil || len(httpRouteList.Items) == 0 {
		return nil
	}
	log.Debug().Msgf("found ipcluster service %s, looking for httproutes endpoints", svc.Name)
	routes := httpRoutesToGatewayRoutes(httpRouteList)
	return n.findGatewayRouteEndpoints(ctx, svc.Name, func(ref gwv1.BackendObjectReference) bool {
		return isServiceRoute(ref, svc)
	}, routes)
}

func (n *NetworkStatusUpdater) findGRPCRouteEndpointsForService(ctx context.Context, svc *corev1.Service, grpcRouteList *gwv1.GRPCRouteList) []v1alpha1.ExternalEndpoint {
	if grpcRouteList == nil || len(grpcRouteList.Items) == 0 {
		return nil
	}
	log.Debug().Msgf("found ipcluster service %s, looking for grpcroutes endpoints", svc.Name)
	routes := make([]gatewayRoute, 0, len(grpcRouteList.Items))
	for _, route := range grpcRouteList.Items {
		rules := make([]gatewayRouteRule, 0, len(route.Spec.Rules))
		for _, rule := range route.Spec.Rules {
			refs := make([]gwv1.BackendObjectReference, 0, len(rule.BackendRefs))
			for _, ref := range rule.BackendRefs {
				refs = append(refs, ref.BackendObjectReference)
			}
			rules = append(rules, gatewayRouteRule{backendRefs: refs, paths: []string{"/"}})
		}
		routes = append(routes, gatewayRoute{
			name:         route.Name,
			namespace:    route.Namespace,
			kind:         kindGRPCRoute,
			hostnames:    route.Spec.Hostnames,
			protocols:    []gwv1.ProtocolType{gwv1.HTTPProtocolType, gwv1.HTTPSProtocolType},
			rules:        rules,
			parentStatus: route.Status.Parents,
		})
	}
	return n.findGatewayRouteEndpoints(ctx, svc.Name, func(ref gwv1.BackendObjectReference) bool {
		return isServiceRoute(ref, svc)
	}, routes)
}

func (n *NetworkStatusUpdater) findTLSRouteEndpointsForService(ctx context.Context, svc *corev1.Service, tlsRouteList *gwv1alpha2.TLSRouteList) []v1alpha1.ExternalEndpoint {
	if tlsRouteList == nil || len(tlsRouteList.Items) == 0 {
		return nil
	}
	log.Debug().Msgf("found ipcluster service %s, looking for tlsroutes endpoints", svc.Name)
	routes := make([]gatewayRoute, 0, len(tlsRouteList.Items))
	for _, route := range tlsRouteList.Items {
		rules := make([]gatewayRouteRule, 0, len(route.Spec.Rules))
		for _, rule := range route.Spec.Rules {
			refs := make([]gwv1.BackendObjectReference, 0, len(rule.BackendRefs))
			for _, ref := range rule.BackendRefs {
				refs = append(refs, ref.BackendObjectReference)
			}
			rules = append(rules, gatewayRouteRule{backendRefs: refs, paths: []string{"/"}})
		}
		routes = append(routes, gatewayRoute{
			name:         route.Name,
			namespace:    route.Namespace,
			kind:         kindTLSRoute,
			hostnames:    route.Spec.Hostnames,
			protocols:    []gwv1.ProtocolType{gwv1.TLSProtocolType},
			rules:        rules,
			parentStatus: route.Status.Parents,
		})
	}
	return n.findGatewayRouteEndpoints(ctx, svc.Name, func(ref gwv1.BackendObjectReference) bool {
		return isServiceRoute(ref, svc)
	}, routes)
}

// httpRoutesToGatewayRoutes converts an HTTPRouteList into the generic []gatewayRoute representation
// used by findGatewayRouteEndpoints. This allows the conversion to be done once and reused across
// multiple backends (e.g. when iterating over InferencePools).
func httpRoutesToGatewayRoutes(httpRouteList *gwv1.HTTPRouteList) []gatewayRoute {
	routes := make([]gatewayRoute, 0, len(httpRouteList.Items))
	for _, route := range httpRouteList.Items {
		rules := make([]gatewayRouteRule, 0, len(route.Spec.Rules))
		for _, rule := range route.Spec.Rules {
			refs := make([]gwv1.BackendObjectReference, 0, len(rule.BackendRefs))
			for _, ref := range rule.BackendRefs {
				refs = append(refs, ref.BackendObjectReference)
			}
			rules = append(rules, gatewayRouteRule{backendRefs: refs, paths: routeMatchPaths(rule.Matches)})
		}
		routes = append(routes, gatewayRoute{
			name:         route.Name,
			namespace:    route.Namespace,
			kind:         kindHTTPRoute,
			hostnames:    route.Spec.Hostnames,
			protocols:    []gwv1.ProtocolType{gwv1.HTTPProtocolType, gwv1.HTTPSProtocolType},
			rules:        rules,
			parentStatus: route.Status.Parents,
		})
	}
	return routes
}

// findGatewayRouteEndpoints is the shared core for all Gateway API route types.
// backendName is used only for log messages; matchRef decides whether a BackendObjectReference targets our backend.
func (n *NetworkStatusUpdater) findGatewayRouteEndpoints(ctx context.Context, backendName string, matchRef func(gwv1.BackendObjectReference) bool, routes []gatewayRoute) []v1alpha1.ExternalEndpoint {
	var endpoints []v1alpha1.ExternalEndpoint
	for _, route := range routes {
		log.Debug().Msgf("found backend %s, looking for %s endpoints", backendName, route.kind)
		for _, rule := range route.rules {
			for _, ref := range rule.backendRefs {
				if !matchRef(ref) {
					continue
				}
				for _, parentStatus := range route.parentStatus {
					if !isRouteAccepted(parentStatus) {
						log.Debug().Msgf("found parent of route %s with status which is not Accepted - ignoring it.", route.name)
						continue
					}
					gw := n.getGatewayForRef(ctx, parentStatus, route.namespace)
					if gw == nil {
						log.Debug().Msgf("no gateways found for route %s.", route.name)
						continue
					}
					log.Debug().Msgf("found %s backend %s, extracting endpoints", route.kind, backendName)
					for _, listener := range findGwListeners(gw, parentStatus, route.protocols) {
						protocol := strings.ToLower(string(listener.Protocol))
						forceTls := route.kind == kindTLSRoute
						tlsEnabled := forceTls || listener.TLS != nil || protocol == protocolHTTPS
						urlProtocol := protocol
						if forceTls {
							urlProtocol = protocolHTTPS
						}
						port := strconv.FormatInt(int64(listener.Port), 10)
						hostnames := routeOrListenerHostnames(route.hostnames, listener.Hostname)

						for _, host := range hostnames {
							for _, path := range rule.paths {
								endpoints = append(endpoints, v1alpha1.ExternalEndpoint{
									SourceType: v1alpha1.ExternalEndpointGatewayAPISource,
									Host:       host,
									Path:       path,
									TlsEnabled: ptr.To(tlsEnabled),
									Protocol:   protocol,
									Port:       port,
									Url:        getUrl(urlProtocol, host, port, path),
									SourceRef: v1alpha1.ExternalEndpointSourceRef{
										Kind: route.kind,
										Name: route.name,
									},
								})
							}
						}
					}
				}
			}
		}
		log.Debug().Msgf("found %d %s endpoints for backend %s", len(endpoints), route.kind, backendName)
	}
	return endpoints
}

func (n *NetworkStatusUpdater) getGatewayForRef(ctx context.Context, parentStatus gwv1.RouteParentStatus, routeNamespace string) *gwv1.Gateway {
	ref := parentStatus.ParentRef
	if ref.Group != nil && *ref.Group != gwv1.GroupName {
		return nil
	}
	if ref.Kind != nil && *ref.Kind != "Gateway" {
		return nil
	}
	ns := routeNamespace
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}
	gw := &gwv1.Gateway{}
	if err := n.cachedClient.Get(ctx, types.NamespacedName{Name: string(ref.Name), Namespace: ns}, gw); err != nil {
		log.Warn().Msgf("Failed to get gateway %s/%s: %s", ns, ref.Name, err.Error())
		return nil
	}
	return gw
}

// findGwListeners filters a Gateway's listeners to only those that match the route's ParentReference.
// A listener matches if:
//  1. SectionName: if the ParentRef specifies a SectionName, the listener's name must match it
//  2. Port: if the ParentRef specifies a Port, the listener's port must match it
//  3. Protocol: the listener's protocol must be one of the compatibleProtocols
//     (e.g. HTTP/HTTPS for HTTPRoutes, TLS for TLSRoutes)
func findGwListeners(gw *gwv1.Gateway, parentStatus gwv1.RouteParentStatus, compatibleProtocols []gwv1.ProtocolType) []gwv1.Listener {
	ref := parentStatus.ParentRef
	var matched []gwv1.Listener
	for _, listener := range gw.Spec.Listeners {
		// Skip if ParentRef targets a specific listener by name and this isn't it
		if ref.SectionName != nil && listener.Name != *ref.SectionName {
			continue
		}
		// Skip if ParentRef targets a specific port and this listener uses a different one
		if ref.Port != nil && listener.Port != *ref.Port {
			continue
		}
		// Skip if the listener's protocol isn't compatible with the route type
		protocolMatch := false
		for _, p := range compatibleProtocols {
			if listener.Protocol == p {
				protocolMatch = true
				break
			}
		}
		if !protocolMatch {
			continue
		}
		matched = append(matched, listener)
	}
	return matched
}

// routeOrListenerHostnames returns the route's hostnames if specified,
// otherwise falls back to the listener's hostname. Per the Gateway API spec,
// an empty route hostname list means the route matches all listener-allowed hostnames.
func routeOrListenerHostnames(routeHostnames []gwv1.Hostname, listenerHostname *gwv1.Hostname) []string {
	if len(routeHostnames) > 0 {
		hosts := make([]string, len(routeHostnames))
		for i, h := range routeHostnames {
			hosts[i] = string(h)
		}
		return hosts
	}
	if listenerHostname != nil {
		return []string{string(*listenerHostname)}
	}
	return nil
}

// routeMatchPaths extracts paths from HTTPRoute matches.
// Per the Gateway API spec, an empty matches list means "match everything" (path "/").
func routeMatchPaths(matches []gwv1.HTTPRouteMatch) []string {
	if len(matches) == 0 {
		return []string{"/"}
	}
	paths := make([]string, 0, len(matches))
	for _, m := range matches {
		path := "/"
		if m.Path != nil && m.Path.Value != nil && *m.Path.Value != "" {
			path = *m.Path.Value
		}
		paths = append(paths, path)
	}
	return paths
}

func isServiceRoute(ref gwv1.BackendObjectReference, svc *corev1.Service) bool {
	return ref.Name == gwv1.ObjectName(svc.Name) && (ref.Kind == nil || *ref.Kind == "Service")
}

func isInferencePoolRef(ref gwv1.BackendObjectReference, poolName string, namespace string) bool {
	return ref.Kind != nil && string(*ref.Kind) == kindInferencePool &&
		ref.Group != nil && string(*ref.Group) == infv1.GroupName &&
		(ref.Namespace == nil || string(*ref.Namespace) == namespace) &&
		ref.Name == gwv1.ObjectName(poolName)
}

func isRouteAccepted(status gwv1.RouteParentStatus) bool {
	for _, cond := range status.Conditions {
		if cond.Type == string(gwv1.RouteConditionAccepted) {
			return cond.Status == metav1.ConditionTrue
		}
	}
	return false
}

// getUrl builds a display URL, omitting the port when it's the default for the protocol
// (443 for https, 80 for http).
func getUrl(protocol string, host string, port string, path string) string {
	if port == "" || (protocol == protocolHTTPS && port == port443) || (protocol == protocolHTTP && port == port80) {
		return fmt.Sprintf("%s://%s%s", protocol, host, path)
	}
	return fmt.Sprintf("%s://%s:%s%s", protocol, host, port, path)
}

func hasTlsAnnotation(annotations map[string]string) bool {
	// List of common keys that indicate TLS is configured
	tlsMarkers := []string{
		// AWS
		"service.beta.kubernetes.io/aws-load-balancer-ssl-cert",
		// DigitalOcean
		"service.beta.kubernetes.io/do-loadbalancer-certificate-id",
		// Azure
		"service.beta.kubernetes.io/azure-load-balancer-ssl-cert-managed",
		// Alibaba Cloud
		"service.beta.kubernetes.io/alibaba-cloud-loadbalancer-cert-id",

		// Specific port-based TLS markers
		"service.beta.kubernetes.io/aws-load-balancer-ssl-ports",
		"service.beta.kubernetes.io/do-loadbalancer-tls-ports",
	}

	for _, key := range tlsMarkers {
		if val, exists := annotations[key]; exists && val != "" {
			return true
		}
	}

	return false
}

func hasTLSForHost(ing *networkingv1.Ingress, host string) bool {
	for _, tls := range ing.Spec.TLS {
		for _, tlsHost := range tls.Hosts {
			if tlsHost == host {
				return true
			}
		}
	}
	return false
}

func getNodeAddress(ctx context.Context, cachedClient client.Client, nodeName string) string {
	if nodeName == "" {
		return ""
	}
	node := &corev1.Node{}
	if err := cachedClient.Get(ctx, types.NamespacedName{Name: nodeName}, node); err != nil {
		log.Debug().Msgf("Failed to get node %s: %s", nodeName, err.Error())
		return ""
	}
	return getNodeAddressFromNode(node)
}

func getNodeAddressFromNode(node *corev1.Node) string {
	var internalIP string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			return addr.Address
		}
		if addr.Type == corev1.NodeInternalIP && internalIP == "" {
			internalIP = addr.Address
		}
	}
	if annotatedExternalIP := node.Annotations[externalNodeIPAnnotation]; annotatedExternalIP != "" {
		return annotatedExternalIP
	}
	return internalIP
}

func (n *NetworkStatusUpdater) findServicesForPod(ctx context.Context,
	pod *corev1.Pod) ([]corev1.Service, error) {
	serviceList := &corev1.ServiceList{}
	err := n.cachedClient.List(ctx, serviceList, client.InNamespace(pod.Namespace))
	if err != nil {
		return nil, err
	}

	var matchingServices []corev1.Service
	for _, svc := range serviceList.Items {
		if svc.Spec.Selector == nil || len(svc.Spec.Selector) == 0 {
			continue
		}
		if labelsMatchSelector(pod.Labels, svc.Spec.Selector) {
			matchingServices = append(matchingServices, svc)
		}
	}
	return matchingServices, nil
}

func (n *NetworkStatusUpdater) findServicesForPods(ctx context.Context,
	pods []corev1.Pod, namespace string) ([]corev1.Service, error) {
	if len(pods) == 0 {
		return []corev1.Service{}, nil
	}

	serviceList := &corev1.ServiceList{}
	err := n.cachedClient.List(ctx, serviceList, client.InNamespace(namespace))
	if err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	var matchingServices []corev1.Service
	for _, svc := range serviceList.Items {
		if len(svc.Spec.Selector) == 0 {
			continue
		}
		for i := range pods {
			if labelsMatchSelector(pods[i].Labels, svc.Spec.Selector) {
				if !seen[svc.Name] {
					seen[svc.Name] = true
					matchingServices = append(matchingServices, svc)
				}
				break
			}
		}
	}
	return matchingServices, nil
}

func labelsMatchSelector(labels, selector map[string]string) bool {
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

// FindPodsForService lists Pods in the same namespace whose labels match the service's selector.
func (n *NetworkStatusUpdater) FindPodsForService(ctx context.Context, svc *corev1.Service) ([]corev1.Pod, error) {
	if svc.Spec.Selector == nil || len(svc.Spec.Selector) == 0 {
		return nil, nil
	}
	podList := &corev1.PodList{}
	err := n.cachedClient.List(ctx, podList,
		client.InNamespace(svc.Namespace),
		client.MatchingLabels(svc.Spec.Selector))
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func extractServiceNamesFromIngress(ing *networkingv1.Ingress) []string {
	seen := map[string]bool{}
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil && path.Backend.Service.Name != "" {
				seen[path.Backend.Service.Name] = true
			}
		}
	}
	if ing.Spec.DefaultBackend != nil && ing.Spec.DefaultBackend.Service != nil {
		seen[ing.Spec.DefaultBackend.Service.Name] = true
	}
	return mapKeys(seen)
}

func extractServiceNamesFromHTTPRoute(route *gwv1.HTTPRoute) []string {
	seen := map[string]bool{}
	for _, rule := range route.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if (ref.Kind == nil || *ref.Kind == "Service") && string(ref.Name) != "" {
				seen[string(ref.Name)] = true
			}
		}
	}
	return mapKeys(seen)
}

// extractInferencePoolNamesFromHTTPRoute returns the names of InferencePools referenced
// by the HTTPRoute's backendRefs (Kind == "InferencePool" with the inference extension group).
func extractInferencePoolNamesFromHTTPRoute(route *gwv1.HTTPRoute) []string {
	seen := map[string]bool{}
	for _, rule := range route.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if ref.Kind != nil && string(*ref.Kind) == kindInferencePool &&
				ref.Group != nil && string(*ref.Group) == infv1.GroupName &&
				string(ref.Name) != "" {
				seen[string(ref.Name)] = true
			}
		}
	}
	return mapKeys(seen)
}

func extractServiceNamesFromGRPCRoute(route *gwv1.GRPCRoute) []string {
	seen := map[string]bool{}
	for _, rule := range route.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if (ref.Kind == nil || *ref.Kind == "Service") && string(ref.Name) != "" {
				seen[string(ref.Name)] = true
			}
		}
	}
	return mapKeys(seen)
}

func extractServiceNamesFromOpenshiftRoute(route *openshiftV1.Route) []string {
	seen := map[string]bool{}
	if route.Spec.To.Name != "" {
		seen[route.Spec.To.Name] = true
	}
	for _, backend := range route.Spec.AlternateBackends {
		if backend.Name != "" {
			seen[backend.Name] = true
		}
	}
	return mapKeys(seen)
}

func extractServiceNamesFromTLSRoute(route *gwv1alpha2.TLSRoute) []string {
	seen := map[string]bool{}
	for _, rule := range route.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if (ref.Kind == nil || *ref.Kind == "Service") && string(ref.Name) != "" {
				seen[string(ref.Name)] = true
			}
		}
	}
	return mapKeys(seen)
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
