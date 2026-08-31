/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"

	openshiftv1 "github.com/openshift/api/route/v1"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func convertWorkloadV2NetworkingConfig(configuration *workloadsManagerApi.WorkloadV2UpdateConfiguration, useOpenShiftRoute bool) (*dromosv1alpha1.NetworkingConfig, error) {
	if configuration == nil {
		return nil, nil
	}
	return convertAgentNetworkingConfig(configuration.Networking.Get(), useOpenShiftRoute)
}

func convertNetworkingEndpoints(apiConfig []workloadsManagerApi.NetworkingEndpoint, useOpenShiftRoute bool) (*endpoints.Config, error) {
	if len(apiConfig) == 0 {
		return nil, nil
	}

	config := &endpoints.Config{
		Endpoints: make([]endpoints.Endpoint, 0, len(apiConfig)),
	}

	for endpointIndex := range apiConfig {
		endpoint, err := convertNetworkingEndpoint(&apiConfig[endpointIndex], useOpenShiftRoute)
		if err != nil {
			return nil, fmt.Errorf("endpoints[%d]: %w", endpointIndex, err)
		}
		config.Endpoints = append(config.Endpoints, endpoint)
	}

	return config, nil
}

func convertNetworkingEndpoint(apiEndpoint *workloadsManagerApi.NetworkingEndpoint, useOpenShiftRoute bool) (endpoints.Endpoint, error) {
	endpoint := endpoints.Endpoint{}

	for serviceIndex := range apiEndpoint.Services {
		item, err := convertEndpointService(&apiEndpoint.Services[serviceIndex])
		if err != nil {
			return endpoint, fmt.Errorf("services[%d]: %w", serviceIndex, err)
		}
		endpoint.Services = append(endpoint.Services, item)
	}

	if ea := apiEndpoint.ExternalAccess.Get(); ea != nil {
		endpoint.ExternalAccess = convertEndpointExternalAccess(ea, useOpenShiftRoute)
	}

	return endpoint, nil
}

func convertEndpointService(apiSvc *workloadsManagerApi.NetworkingService) (endpoints.Service, error) {
	item := endpoints.Service{}

	if err := validateNetworkingService(apiSvc); err != nil {
		return item, err
	}

	if apiSvc.Spec.IsSet() {
		svc, err := convertServiceSpec(apiSvc.Spec.Get())
		if err != nil {
			return item, err
		}
		item.Create = svc
	}

	if apiSvc.Ref.IsSet() {
		item.Ref = &endpoints.ServiceRef{
			Name: apiSvc.Ref.Get().Name,
			Port: apiSvc.Ref.Get().Port,
		}
	}

	if auth := apiSvc.Authorization.Get(); auth != nil {
		item.Auth = convertEndpointAuthorization(auth)
	}

	return item, nil
}

func convertAgentNetworkingConfig(apiConfig *workloadsManagerApi.AgentNetworkingConfiguration, useOpenShiftRoute bool) (*dromosv1alpha1.NetworkingConfig, error) {
	if apiConfig == nil {
		return nil, nil
	}

	runnerConfig := &dromosv1alpha1.NetworkingConfig{}

	for i := range apiConfig.Services {
		item, err := convertAgentService(&apiConfig.Services[i])
		if err != nil {
			return nil, err
		}
		runnerConfig.Services = append(runnerConfig.Services, item)
	}

	if ea := apiConfig.ExternalAccess.Get(); ea != nil {
		runnerConfig.ExternalAccess = convertExternalAccess(ea, useOpenShiftRoute)
	}

	return runnerConfig, nil
}

func validateNetworkingService(apiSvc *workloadsManagerApi.NetworkingService) error {
	hasSpec := apiSvc.Spec.IsSet()
	hasRef := apiSvc.Ref.IsSet()

	if hasSpec && hasRef {
		return fmt.Errorf("networking service: spec and ref are mutually exclusive")
	}

	if !hasSpec && !hasRef {
		return fmt.Errorf("networking service: either spec or ref must be set")
	}

	if hasSpec && apiSvc.Spec.Get() == nil {
		return fmt.Errorf("networking service: spec value is required when spec is set")
	}

	if hasRef && apiSvc.Ref.Get() == nil {
		return fmt.Errorf("networking service: ref value is required when ref is set")
	}

	return nil
}

func validateAgentNetworkingService(apiSvc *workloadsManagerApi.AgentNetworkingService) error {
	hasCreate := apiSvc.Create.IsSet()
	hasRef := apiSvc.Ref.IsSet()

	if hasCreate && hasRef {
		return fmt.Errorf("networking service: create and ref are mutually exclusive")
	}

	if !hasCreate && !hasRef {
		return fmt.Errorf("networking service: either create or ref must be set")
	}

	if hasCreate && apiSvc.Create.Get() == nil {
		return fmt.Errorf("networking service: create value is required when create is set")
	}

	if hasRef && apiSvc.Ref.Get() == nil {
		return fmt.Errorf("networking service: ref value is required when ref is set")
	}

	return nil
}

func convertAgentService(apiSvc *workloadsManagerApi.AgentNetworkingService) (dromosv1alpha1.NetworkingServiceItem, error) {
	item := dromosv1alpha1.NetworkingServiceItem{}

	if err := validateAgentNetworkingService(apiSvc); err != nil {
		return item, err
	}

	if apiSvc.Create.IsSet() {
		svc, err := convertAgentServiceCreate(apiSvc.Create.Get())
		if err != nil {
			return item, err
		}
		item.Create = svc
	}

	if apiSvc.Ref.IsSet() {
		item.Ref = &dromosv1alpha1.NetworkingServiceRef{
			Name: apiSvc.Ref.Get().Name,
			Port: apiSvc.Ref.Get().Port,
		}
	}

	if auth := apiSvc.Authorization.Get(); auth != nil {
		item.Auth = convertAuthorization(auth)
	}

	return item, nil
}

func convertServiceSpec(apiSpec *workloadsManagerApi.ServiceSpec) (*corev1.Service, error) {
	if apiSpec == nil {
		return nil, fmt.Errorf("service spec: value is required")
	}
	if apiSpec.Name == "" {
		return nil, fmt.Errorf("service spec: name is required")
	}
	if len(apiSpec.Selector) == 0 {
		return nil, fmt.Errorf("service spec: selector is required")
	}
	if len(apiSpec.Ports) == 0 {
		return nil, fmt.Errorf("service spec: at least one port is required")
	}

	serviceType := corev1.ServiceTypeClusterIP
	if apiType := apiSpec.Type.Get(); apiType != nil {
		serviceType = corev1.ServiceType(*apiType)
	}

	ports := make([]corev1.ServicePort, 0, len(apiSpec.Ports))
	for i := range apiSpec.Ports {
		port, err := convertServicePort(&apiSpec.Ports[i])
		if err != nil {
			return nil, fmt.Errorf("service spec ports[%d]: %w", i, err)
		}
		ports = append(ports, port)
	}

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CoreV1,
			Kind:       constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   apiSpec.Name,
			Labels: cloneOptionalStringMap(apiSpec.Labels),
		},
		Spec: corev1.ServiceSpec{
			Type:     serviceType,
			Ports:    ports,
			Selector: apiSpec.Selector,
		},
	}, nil
}

func cloneOptionalStringMap(source *map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	return maps.Clone(*source)
}

func convertServicePort(apiPort *workloadsManagerApi.ServicePort) (corev1.ServicePort, error) {
	if apiPort.Port <= 0 {
		return corev1.ServicePort{}, fmt.Errorf("port must be greater than 0")
	}

	port := corev1.ServicePort{
		Port:     apiPort.Port,
		Protocol: corev1.ProtocolTCP,
	}
	if name := apiPort.Name.Get(); name != nil {
		port.Name = *name
	}
	if targetPort := apiPort.TargetPort.Get(); targetPort != nil {
		port.TargetPort = intstr.FromInt32(*targetPort)
	}
	if nodePort := apiPort.NodePort.Get(); nodePort != nil {
		port.NodePort = *nodePort
	}
	if protocol := apiPort.Protocol.Get(); protocol != nil {
		port.Protocol = corev1.Protocol(*protocol)
	}

	return port, nil
}

func convertAgentServiceCreate(apiCreate *workloadsManagerApi.AgentServiceCreate) (*corev1.Service, error) {
	if apiCreate == nil {
		return nil, fmt.Errorf("service create: value is required")
	}
	if apiCreate.Port <= 0 {
		return nil, fmt.Errorf("service create: port must be greater than 0")
	}
	if len(apiCreate.Selector) == 0 {
		return nil, fmt.Errorf("service create: selector is required")
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: constants.CoreV1,
			Kind:       constants.KindService,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port:     apiCreate.Port,
				Protocol: corev1.ProtocolTCP,
			}},
			Selector: apiCreate.Selector,
		},
	}, nil
}

func convertAuthorization(apiAuth *workloadsManagerApi.Workloadv2Authorization) *dromosv1alpha1.NetworkingAuthConfig {
	endpointAuth := convertEndpointAuthorization(apiAuth)
	if endpointAuth == nil {
		return nil
	}
	return &dromosv1alpha1.NetworkingAuthConfig{
		AuthorizationType: dromosv1alpha1.AuthorizationType(endpointAuth.AuthorizationType),
		AuthorizedUsers:   endpointAuth.AuthorizedUsers,
		AuthorizedGroups:  endpointAuth.AuthorizedGroups,
		ForceJsonResponse: endpointAuth.ForceJsonResponse,
	}
}

func convertEndpointAuthorization(apiAuth *workloadsManagerApi.Workloadv2Authorization) *endpoints.AuthConfig {
	if apiAuth == nil {
		return nil
	}
	auth := &endpoints.AuthConfig{}

	if apiAuth.AuthorizationType.IsSet() {
		authType := apiAuth.AuthorizationType.Get()
		auth.AuthorizationType = endpoints.AuthorizationType(*authType)
	}

	if len(apiAuth.AuthorizedUsers) > 0 {
		auth.AuthorizedUsers = strings.Join(apiAuth.AuthorizedUsers, ",")
	}

	if len(apiAuth.AuthorizedGroups) > 0 {
		auth.AuthorizedGroups = strings.Join(apiAuth.AuthorizedGroups, ",")
	}

	if apiAuth.ForceJsonResponse.IsSet() {
		auth.ForceJsonResponse = apiAuth.ForceJsonResponse.Get()
	}

	return auth
}

func convertExternalAccess(apiEA *workloadsManagerApi.ExternalAccess, useOpenShiftRoute bool) *dromosv1alpha1.NetworkingExternalAccessConfig {
	config := &dromosv1alpha1.NetworkingExternalAccessConfig{}
	endpointAccess := convertEndpointExternalAccess(apiEA, useOpenShiftRoute)
	if endpointAccess == nil {
		return config
	}
	config.Ingress = endpointAccess.Ingress
	config.Route = endpointAccess.Route
	config.HTTPRoute = endpointAccess.HTTPRoute
	return config
}

func convertEndpointExternalAccess(apiEA *workloadsManagerApi.ExternalAccess, useOpenShiftRoute bool) *endpoints.ExternalAccess {
	if apiEA == nil {
		return nil
	}
	config := &endpoints.ExternalAccess{}
	if apiEA.HasIngress() {
		ing := unmarshalIngress(apiEA.GetIngress())
		if ing != nil {
			config.Ingress = ing
		}
	}
	if apiEA.HasRoute() {
		route := unmarshalRoute(apiEA.GetRoute())
		if route != nil {
			config.Route = route
		}
	}
	if apiEA.HasHttpRoute() {
		httpRoute := unmarshalHTTPRoute(apiEA.GetHttpRoute())
		if httpRoute != nil {
			config.HTTPRoute = httpRoute
		}
	}
	if config.Ingress != nil && config.Route != nil {
		if useOpenShiftRoute {
			config.Ingress = nil
		} else {
			config.Route = nil
		}
	}
	if config.Ingress == nil && config.Route == nil && config.HTTPRoute == nil {
		return nil
	}
	return config
}

func unmarshalIngress(m map[string]interface{}) *networkingv1.Ingress {
	if len(m) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var ing networkingv1.Ingress
	if json.Unmarshal(b, &ing) != nil {
		return nil
	}
	return &ing
}

func unmarshalRoute(m map[string]interface{}) *openshiftv1.Route {
	if len(m) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var route openshiftv1.Route
	if json.Unmarshal(b, &route) != nil {
		return nil
	}
	return &route
}

func unmarshalHTTPRoute(m map[string]interface{}) *gatewayv1.HTTPRoute {
	if len(m) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var httpRoute gatewayv1.HTTPRoute
	if json.Unmarshal(b, &httpRoute) != nil {
		return nil
	}
	return &httpRoute
}
