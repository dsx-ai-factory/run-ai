/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"net/url"
	"strings"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	networkingPluginName             = "networking-plugin"
	externalAccessReadyConditionType = "ExternalAccessReady"
	kubernetesServiceKind            = "Service"
	nodePortServiceType              = "NodePort"
)

func networkingDetailsFromWorkloadRunner(wlr *dromosv1alpha1.WorkloadRunner) (*endpoints.NetworkingStatus, error) {
	if wlr == nil {
		return nil, nil
	}

	netPlugin := wlr.Status.Plugins[networkingPluginName]
	if netPlugin == nil {
		return nil, nil
	}

	networkingStatus, err := endpoints.NetworkingStatusFromRawExtension(netPlugin.AdditionalInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse networking plugin details: %w", err)
	}

	return networkingStatus, nil
}

func applyNetworkingDetailsToSyncFields(
	networkingStatus *endpoints.NetworkingStatus,
	workloadSyncFields *openapi.WorkloadSyncFields,
) {
	if workloadSyncFields == nil {
		return
	}

	if networkStatus := networkingDetailsToNetworkStatus(networkingStatus); networkStatus != nil {
		enrichNodePortServiceURLsFromExternalEndpoints(networkStatus, workloadSyncFields.Endpoints)
		workloadSyncFields.NetworkStatus = networkStatus
	}

	if urls := networkingDetailsFlatURLs(networkingStatus); len(urls) > 0 && isInternalClusterURL(workloadSyncFields.Urls) {
		workloadSyncFields.Urls = urls
	}
}

func enrichNodePortServiceURLsFromExternalEndpoints(
	networkStatus *openapi.WorkloadNetworkStatus,
	externalEndpoints []openapi.EndpointSync,
) {
	if networkStatus == nil || len(externalEndpoints) == 0 {
		return
	}

	serviceURLs := serviceURLsFromExternalEndpoints(externalEndpoints)
	if len(serviceURLs) == 0 {
		return
	}

	for i := range networkStatus.Services {
		service := &networkStatus.Services[i]
		if service.GetType() != nodePortServiceType {
			continue
		}
		urls := serviceURLs[service.GetName()]
		for _, serviceURL := range urls {
			appendUniqueServiceURL(service, serviceURL)
		}
	}
}

func serviceURLsFromExternalEndpoints(
	externalEndpoints []openapi.EndpointSync,
) map[string][]openapi.WorkloadNetworkStatusServicesInnerUrlsInner {
	result := map[string][]openapi.WorkloadNetworkStatusServicesInnerUrlsInner{}
	for _, endpoint := range externalEndpoints {
		if endpoint.SourceRef == nil || endpoint.GetUrl() == "" {
			continue
		}
		if endpoint.SourceRef.GetKind() != kubernetesServiceKind || endpoint.SourceRef.GetName() == "" {
			continue
		}
		serviceName := endpoint.SourceRef.GetName()
		result[serviceName] = append(result[serviceName], openapi.WorkloadNetworkStatusServicesInnerUrlsInner{
			Url:        strPointerIfNotEmpty(endpoint.GetUrl()),
			SourceKind: strPointerIfNotEmpty(endpoint.SourceRef.GetKind()),
			SourceName: strPointerIfNotEmpty(serviceName),
		})
	}
	return result
}

func appendUniqueServiceURL(
	service *openapi.WorkloadNetworkStatusServicesInner,
	serviceURL openapi.WorkloadNetworkStatusServicesInnerUrlsInner,
) {
	if serviceURL.GetUrl() == "" {
		return
	}
	for _, existing := range service.Urls {
		if existing.GetUrl() == serviceURL.GetUrl() {
			return
		}
	}
	service.Urls = append(service.Urls, serviceURL)
}

func networkingDetailsToNetworkStatus(networkingStatus *endpoints.NetworkingStatus) *openapi.WorkloadNetworkStatus {
	if networkingStatus == nil {
		return nil
	}

	services := []openapi.WorkloadNetworkStatusServicesInner{}
	for _, endpointStatus := range networkingStatus.Endpoints {
		for _, serviceStatus := range endpointStatus.Services {
			apiService := openapi.WorkloadNetworkStatusServicesInner{
				Name:         strPointerIfNotEmpty(serviceStatus.Name),
				Type:         strPointerIfNotEmpty(serviceStatus.Type),
				Ports:        convertNetworkingDetailsPorts(serviceStatus.Ports),
				LoadBalancer: convertNetworkingDetailsLoadBalancer(serviceStatus.LoadBalancer),
				Urls:         convertNetworkingDetailsURLs(serviceStatus.URLs),
			}
			services = append(services, apiService)
		}
	}
	if len(services) == 0 {
		return nil
	}

	return &openapi.WorkloadNetworkStatus{
		Services: services,
	}
}

func convertNetworkingDetailsPorts(ports []endpoints.ServicePortStatus) []openapi.WorkloadNetworkStatusServicesInnerPortsInner {
	if len(ports) == 0 {
		return nil
	}

	result := make([]openapi.WorkloadNetworkStatusServicesInnerPortsInner, 0, len(ports))
	for _, port := range ports {
		result = append(result, openapi.WorkloadNetworkStatusServicesInnerPortsInner{
			Name:       strPointerIfNotEmpty(port.Name),
			TargetPort: strPointerIfNotEmpty(port.TargetPort.String()),
			Port:       int32PointerIfNotEmpty(port.Port),
			Protocol:   strPointerIfNotEmpty(port.Protocol),
		})
	}
	return result
}

func convertNetworkingDetailsLoadBalancer(loadBalancerStatus *endpoints.LoadBalancerStatus) *openapi.WorkloadNetworkStatusServicesInnerLoadBalancer {
	if loadBalancerStatus == nil || len(loadBalancerStatus.Ingress) == 0 {
		return nil
	}

	ingress := make([]openapi.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner, 0, len(loadBalancerStatus.Ingress))
	for _, i := range loadBalancerStatus.Ingress {
		ingress = append(ingress, openapi.WorkloadNetworkStatusServicesInnerLoadBalancerIngressInner{
			Ip:       strPointerIfNotEmpty(i.IP),
			Hostname: strPointerIfNotEmpty(i.Hostname),
		})
	}

	return &openapi.WorkloadNetworkStatusServicesInnerLoadBalancer{
		Ingress: ingress,
	}
}

func convertNetworkingDetailsURLs(urls []endpoints.ServiceURLStatus) []openapi.WorkloadNetworkStatusServicesInnerUrlsInner {
	if len(urls) == 0 {
		return nil
	}

	result := make([]openapi.WorkloadNetworkStatusServicesInnerUrlsInner, 0, len(urls))
	for _, serviceURL := range urls {
		result = append(result, openapi.WorkloadNetworkStatusServicesInnerUrlsInner{
			Url:        strPointerIfNotEmpty(serviceURL.URL),
			SourceKind: strPointerIfNotEmpty(string(serviceURL.SourceKind)),
			SourceName: strPointerIfNotEmpty(serviceURL.SourceName),
		})
	}
	return result
}

func networkingDetailsFlatURLs(networkingStatus *endpoints.NetworkingStatus) []string {
	if networkingStatus == nil {
		return nil
	}

	seen := map[string]struct{}{}
	result := []string{}
	for _, endpointStatus := range networkingStatus.Endpoints {
		for _, serviceStatus := range endpointStatus.Services {
			for _, serviceURL := range serviceStatus.URLs {
				if serviceURL.URL == "" {
					continue
				}
				if _, exists := seen[serviceURL.URL]; exists {
					continue
				}
				seen[serviceURL.URL] = struct{}{}
				result = append(result, serviceURL.URL)
			}
		}
	}
	return result
}

func externalUrlFromWorkloadRunner(wlr *dromosv1alpha1.WorkloadRunner) string {
	if wlr == nil {
		return ""
	}

	netPlugin := wlr.Status.Plugins[networkingPluginName]
	if netPlugin == nil {
		return ""
	}
	for _, c := range netPlugin.Conditions {
		if c.Type == externalAccessReadyConditionType && c.Status == metav1.ConditionTrue {
			if i := strings.Index(c.Message, ", url: "); i >= 0 {
				raw := strings.TrimSpace(c.Message[i+len(", url: "):])
				if parsed, err := url.ParseRequestURI(raw); err == nil &&
					(parsed.Scheme == "http" || parsed.Scheme == "https") &&
					parsed.Host != "" {
					return raw
				}
			}
		}
	}
	return ""
}

func isInternalClusterURL(urls []string) bool {
	if len(urls) == 0 {
		return true
	}
	for _, u := range urls {
		if !strings.Contains(u, ".cluster.local") {
			return false
		}
	}
	return true
}
