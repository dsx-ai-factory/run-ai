/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

// BuildDistributedInferenceIntent maps the resolved native LWS serving topology.
// It does not establish trusted submission origin or inference runtime compatibility;
// a future submission caller must establish both before selecting this adapter.
// Only the public HTTP, root-hostname topology is representable in this version.
func BuildDistributedInferenceIntent(lws *lwsv1.LeaderWorkerSet, endpoint api.NetworkingEndpoint, defaults *Defaults) (*api.AgentInferenceEndpoint, error) {
	if defaults == nil || defaults.GatewayName == "" || len(validation.IsDNS1123Subdomain(defaults.GatewayName)) != 0 ||
		defaults.GatewayNamespace == "" || len(validation.IsDNS1123Label(defaults.GatewayNamespace)) != 0 || defaults.EPPProfile != "gie-v1.5" {
		return nil, fmt.Errorf("distributed inference requires valid administrator Gateway defaults and the gie-v1.5 EPP profile")
	}
	if lws == nil || lws.Name == "" || lws.Namespace == "" || lws.Spec.LeaderWorkerTemplate.LeaderTemplate == nil ||
		len(lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers) == 0 {
		return nil, fmt.Errorf("distributed inference requires a named LWS with an explicit leader container")
	}
	service, err := distributedServingService(lws, endpoint)
	if err != nil {
		return nil, err
	}
	hostname, err := distributedServingHostname(endpoint.ExternalAccess, service, defaults)
	if err != nil {
		return nil, err
	}
	return &api.AgentInferenceEndpoint{
		SchemaVersion: runv1.InferenceEndpointIntegrationV1,
		Spec: map[string]interface{}{
			"workload": map[string]interface{}{
				"selector": map[string]interface{}{"matchLabels": maps.Clone(*service.Selector)},
				"port":     *service.Ports[0].TargetPort,
			},
			"routing": map[string]interface{}{
				"gatewayRef": map[string]interface{}{"name": defaults.GatewayName, "namespace": defaults.GatewayNamespace},
				"hostnames":  []string{hostname},
			},
			"epp": map[string]interface{}{"profile": defaults.EPPProfile},
		},
	}, nil
}

func distributedServingService(lws *lwsv1.LeaderWorkerSet, endpoint api.NetworkingEndpoint) (*api.ServiceSpec, error) {
	if endpoint.Services == nil || len(*endpoint.Services) != 1 {
		return nil, fmt.Errorf("distributed inference requires exactly one created serving Service")
	}
	networkingService := (*endpoint.Services)[0]
	service := networkingService.Spec
	if service == nil || networkingService.Ref != nil || networkingService.Authorization != nil || service.Name == "" || len(service.Ports) != 1 {
		return nil, fmt.Errorf("distributed inference requires one public created Service with one serving port")
	}
	if service.Selector == nil || (*service.Selector)[lwsv1.SetNameLabelKey] != lws.Name || (*service.Selector)[lwsv1.WorkerIndexLabelKey] != "0" {
		return nil, fmt.Errorf("distributed inference serving selector must select only leaders of LWS %s", lws.Name)
	}
	leader := lws.Spec.LeaderWorkerTemplate.LeaderTemplate
	for key, value := range *service.Selector {
		if key == lwsv1.SetNameLabelKey || key == lwsv1.WorkerIndexLabelKey {
			continue
		}
		if actual, found := leader.Labels[key]; !found || actual != value {
			return nil, fmt.Errorf("distributed inference serving selector %q does not match the leader template", key)
		}
	}
	port := service.Ports[0]
	if port.Port < 1 || port.Port > 65535 || port.TargetPort == nil || *port.TargetPort < 1 || *port.TargetPort > 65535 ||
		(port.Protocol != nil && *port.Protocol != api.SERVICEPROTOCOL_TCP) {
		return nil, fmt.Errorf("distributed inference serving Service requires an explicit TCP container target port")
	}
	// Native DINF puts the application in the first leader container. In particular,
	// do not turn a Service targeting an authorization proxy into direct pod access.
	for _, declared := range leader.Spec.Containers[0].Ports {
		if declared.ContainerPort == *port.TargetPort && (declared.Protocol == "" || declared.Protocol == corev1.ProtocolTCP) {
			return service, nil
		}
	}
	return nil, fmt.Errorf("distributed inference target port must address the primary leader container directly; proxy routing is unsupported")
}

func distributedServingHostname(access *api.ExternalAccess, service *api.ServiceSpec, defaults *Defaults) (string, error) {
	if access == nil || access.HttpRoute == nil || access.Ingress != nil || access.Route != nil {
		return "", fmt.Errorf("distributed inference requires an external Gateway HTTPRoute without Ingress or OpenShift Route")
	}
	raw, err := json.Marshal(*access.HttpRoute)
	if err != nil {
		return "", fmt.Errorf("encode distributed inference HTTPRoute: %w", err)
	}
	var route gatewayv1.HTTPRoute
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&route); err != nil {
		return "", fmt.Errorf("decode distributed inference HTTPRoute: %w", err)
	}
	if route.APIVersion != gatewayv1.GroupVersion.String() || route.Kind != "HTTPRoute" || len(route.Annotations) != 0 || len(route.Spec.Hostnames) != 1 ||
		len(validation.IsDNS1123Subdomain(string(route.Spec.Hostnames[0]))) != 0 {
		return "", fmt.Errorf("distributed inference requires a v1 HTTPRoute with one concrete hostname and no annotation-based routing")
	}
	// Match the native composer's complete route shape so filters, authentication,
	// path prefixes and listener restrictions cannot be silently dropped by the CR.
	expected := gatewayv1.HTTPRouteSpec{
		CommonRouteSpec: gatewayv1.CommonRouteSpec{ParentRefs: []gatewayv1.ParentReference{{
			Name: gatewayv1.ObjectName(defaults.GatewayName), Namespace: ptr.To(gatewayv1.Namespace(defaults.GatewayNamespace)),
		}}},
		Hostnames: route.Spec.Hostnames,
		Rules: []gatewayv1.HTTPRouteRule{{
			Matches: []gatewayv1.HTTPRouteMatch{{Path: &gatewayv1.HTTPPathMatch{Type: ptr.To(gatewayv1.PathMatchPathPrefix), Value: ptr.To("/")}}},
			BackendRefs: []gatewayv1.HTTPBackendRef{{BackendRef: gatewayv1.BackendRef{BackendObjectReference: gatewayv1.BackendObjectReference{
				Name: gatewayv1.ObjectName(service.Name), Port: ptr.To(gatewayv1.PortNumber(service.Ports[0].Port)),
			}}}},
		}},
	}
	if !reflect.DeepEqual(route.Spec, expected) {
		return "", fmt.Errorf("distributed inference HTTPRoute must use the administrator Gateway and one unfiltered root path to its serving Service")
	}
	return string(route.Spec.Hostnames[0]), nil
}
