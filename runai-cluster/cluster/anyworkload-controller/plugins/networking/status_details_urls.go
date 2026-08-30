package networking

import (
	"fmt"
	"strings"

	openshiftv1 "github.com/openshift/api/route/v1"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func addURLsToServices(endpointStatus *endpoints.EndpointStatus, urls []endpoints.ServiceURLStatus) {
	if endpointStatus == nil || len(endpointStatus.Services) == 0 || len(urls) == 0 {
		return
	}
	for i := range endpointStatus.Services {
		endpointStatus.Services[i].URLs = append(endpointStatus.Services[i].URLs, urls...)
	}
}

func ingressHosts(ingress *networkingv1.Ingress) []string {
	if ingress == nil {
		return nil
	}
	hosts := make([]string, 0, len(ingress.Spec.Rules))
	for _, rule := range ingress.Spec.Rules {
		if rule.Host != "" {
			hosts = append(hosts, rule.Host)
		}
	}
	return hosts
}

func ingressURLs(ingress *networkingv1.Ingress) []string {
	if ingress == nil {
		return nil
	}
	var urls []string
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			continue
		}
		if rule.HTTP == nil || len(rule.HTTP.Paths) == 0 {
			urls = append(urls, fmt.Sprintf("https://%s/", rule.Host))
			continue
		}
		for _, path := range rule.HTTP.Paths {
			urls = append(urls, fmt.Sprintf("https://%s%s", rule.Host, normalizeURLPath(path.Path)))
		}
	}
	return urls
}

func routeHosts(route *openshiftv1.Route) []string {
	if route == nil || route.Spec.Host == "" {
		return nil
	}
	return []string{route.Spec.Host}
}

func routeURLs(route *openshiftv1.Route) []string {
	if route == nil || route.Spec.Host == "" {
		return nil
	}
	scheme := "http"
	if route.Spec.TLS != nil {
		scheme = "https"
	}
	return []string{fmt.Sprintf("%s://%s%s", scheme, route.Spec.Host, normalizeURLPath(route.Spec.Path))}
}

func httpRouteHosts(httpRoute *gatewayv1.HTTPRoute) []string {
	if httpRoute == nil || len(httpRoute.Spec.Hostnames) == 0 {
		return nil
	}
	hosts := make([]string, 0, len(httpRoute.Spec.Hostnames))
	for _, hostname := range httpRoute.Spec.Hostnames {
		hosts = append(hosts, string(hostname))
	}
	return hosts
}

func httpRouteURLs(httpRoute *gatewayv1.HTTPRoute) []string {
	if httpRoute == nil || len(httpRoute.Spec.Hostnames) == 0 {
		return nil
	}
	paths := httpRoutePaths(httpRoute)
	urls := make([]string, 0, len(httpRoute.Spec.Hostnames)*len(paths))
	for _, hostname := range httpRoute.Spec.Hostnames {
		for _, path := range paths {
			urls = append(urls, fmt.Sprintf("https://%s%s", hostname, path))
		}
	}
	return urls
}

func httpRoutePaths(httpRoute *gatewayv1.HTTPRoute) []string {
	paths := []string{"/"}
	if httpRoute == nil {
		return paths
	}
	var configured []string
	for _, rule := range httpRoute.Spec.Rules {
		for _, match := range rule.Matches {
			if match.Path != nil && match.Path.Value != nil {
				configured = append(configured, normalizeURLPath(*match.Path.Value))
			}
		}
	}
	if len(configured) == 0 {
		return paths
	}
	return configured
}

func serviceURLs(kind endpoints.ExternalAccessKind, sourceName string, urls []string) []endpoints.ServiceURLStatus {
	statuses := make([]endpoints.ServiceURLStatus, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			continue
		}
		statuses = append(statuses, endpoints.ServiceURLStatus{
			URL:        url,
			SourceKind: kind,
			SourceName: sourceName,
		})
	}
	return statuses
}

func normalizeURLPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}
