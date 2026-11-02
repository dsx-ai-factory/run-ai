/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	composer "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

// BuildDistributedInferenceIntent returns the externalAccess.inference block for
// the resolved native LWS serving topology. The caller sets it on the same
// endpoint's externalAccess, next to httpRoute. The networking plugin derives the
// InferenceEndpoint workload selector and port, Gateway, hostnames and default
// backend from that endpoint's Service and httpRoute; this function checks that
// they describe the leaders of lws behind the administrator Gateway, so the block
// only carries what cannot be derived.
// It does not establish trusted submission origin or inference runtime compatibility;
// a future submission caller must establish both before selecting this adapter,
// and must set the block after validating the caller's networking, which rejects it.
// Only the public HTTP, root-hostname topology is representable in this version.
// protected is the serving port's access mode from the request (servingPort.authorizationType
// other than public). The composer protects distributed inference with an oauth2-proxy
// sidecar in the leader Pod and does not set authorization on the networking Service,
// so protection cannot be read from the endpoint.
// rewrites are the workload's modelRewrites; the key is omitted when there are none
// because the endpoint CRD has no nullable fields.
func BuildDistributedInferenceIntent(lws *lwsv1.LeaderWorkerSet, endpoint api.NetworkingEndpoint, protected bool, rewrites []api.ModelRewrite, defaults *Defaults) (*api.ExternalAccessInference, error) {
	if defaults == nil || defaults.GatewayName == "" || len(validation.IsDNS1123Subdomain(defaults.GatewayName)) != 0 ||
		defaults.GatewayNamespace == "" || len(validation.IsDNS1123Label(defaults.GatewayNamespace)) != 0 || defaults.EPPProfile != "gie-v1.5" {
		return nil, fmt.Errorf("distributed inference requires valid administrator Gateway defaults and the gie-v1.5 EPP profile")
	}
	if lws == nil || lws.Name == "" || lws.Namespace == "" || lws.Spec.LeaderWorkerTemplate.LeaderTemplate == nil ||
		len(lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers) == 0 {
		return nil, fmt.Errorf("distributed inference requires a named LWS with an explicit leader container")
	}
	service, err := distributedServingService(lws, endpoint, protected)
	if err != nil {
		return nil, err
	}
	if _, err := distributedServingHostname(endpoint.ExternalAccess, service, defaults); err != nil {
		return nil, err
	}
	inference := &api.ExternalAccessInference{Epp: map[string]interface{}{"profile": defaults.EPPProfile}}
	if len(rewrites) > 0 {
		rules := modelRewrites(rewrites)
		inference.Rewrites = &rules
	}
	return inference, nil
}

func modelRewrites(rewrites []api.ModelRewrite) []map[string]interface{} {
	rules := make([]map[string]interface{}, 0, len(rewrites))
	for _, rewrite := range rewrites {
		targets := make([]interface{}, 0, len(rewrite.Targets))
		for _, target := range rewrite.Targets {
			targets = append(targets, map[string]interface{}{"model": target.Model, "weight": target.Weight})
		}
		rules = append(rules, map[string]interface{}{"model": rewrite.Model, "targets": targets})
	}
	return rules
}

func distributedServingService(lws *lwsv1.LeaderWorkerSet, endpoint api.NetworkingEndpoint, protected bool) (*api.ServiceSpec, error) {
	if endpoint.Services == nil || len(*endpoint.Services) != 1 {
		return nil, fmt.Errorf("distributed inference requires exactly one created serving Service")
	}
	networkingService := (*endpoint.Services)[0]
	service := networkingService.Spec
	if service == nil || networkingService.Ref != nil || service.Name == "" || len(service.Ports) != 1 {
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
	if protected {
		// The auth proxy is a sidecar in the leader Pod. It is the only valid target for
		// protected serving; the model port behind it must never be exposed directly.
		if *port.TargetPort != composer.Oauth2ContainerDefaultPort {
			return nil, fmt.Errorf("protected distributed inference must target the auth proxy port %d", composer.Oauth2ContainerDefaultPort)
		}
		if !containersDeclarePort(leader.Spec.Containers, *port.TargetPort) {
			return nil, fmt.Errorf("protected distributed inference requires a leader container that declares the auth proxy port %d", *port.TargetPort)
		}
		return service, nil
	}
	if *port.TargetPort == composer.Oauth2ContainerDefaultPort {
		return nil, fmt.Errorf("public distributed inference must not target the auth proxy port %d", composer.Oauth2ContainerDefaultPort)
	}
	// Native DINF puts the application in the first leader container. In particular,
	// do not turn a Service targeting an authorization proxy into direct pod access.
	if containersDeclarePort(leader.Spec.Containers[:1], *port.TargetPort) {
		return service, nil
	}
	return nil, fmt.Errorf("distributed inference target port must address the primary leader container directly; proxy routing is unsupported")
}

func containersDeclarePort(containers []corev1.Container, port int32) bool {
	for _, container := range containers {
		for _, declared := range container.Ports {
			if declared.ContainerPort == port && (declared.Protocol == "" || declared.Protocol == corev1.ProtocolTCP) {
				return true
			}
		}
	}
	return false
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
