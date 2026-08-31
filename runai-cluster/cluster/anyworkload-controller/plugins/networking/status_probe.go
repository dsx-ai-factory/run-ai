/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"

	openshiftv1 "github.com/openshift/api/route/v1"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type endpointNetworkingProbe struct {
	auth     authProxyProbeResult
	external externalAccessProbeResult
}

type authProxyProbeResult struct {
	result           endpointPhaseResult
	ready            *bool
	conditionMessage string
	err              error
}

func disabledAuthProxyProbe() authProxyProbeResult {
	return authProxyProbeResult{
		result: endpointPhaseResult{
			phase:   dromosv1alpha1.PluginPhaseDisabled,
			reason:  AuthProxyReasonDisabled,
			message: "auth disabled",
		},
		conditionMessage: "Auth proxy is disabled",
	}
}

func (p *Plugin) probeAuthProxy(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	deploymentName string,
	k8sClient client.Client,
) authProxyProbeResult {
	deployment := &appsv1.Deployment{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: runner.Namespace}, deployment)
	switch {
	case apierrors.IsNotFound(err):
		ready := false
		return authProxyProbeResult{
			result: endpointPhaseResult{
				phase:   dromosv1alpha1.PluginPhaseCreating,
				reason:  AuthProxyReasonNotFound,
				message: fmt.Sprintf("auth proxy %s not found", deploymentName),
			},
			ready:            &ready,
			conditionMessage: fmt.Sprintf("Auth proxy deployment %s not found", deploymentName),
		}
	case err != nil:
		ready := false
		log.FromContext(ctx).Error(err, "failed to get auth proxy deployment", "name", deploymentName, "namespace", runner.Namespace)
		return authProxyProbeResult{
			result: endpointPhaseResult{
				phase:   dromosv1alpha1.PluginPhaseFailed,
				reason:  AuthProxyReasonGetFailed,
				message: fmt.Sprintf("failed to get auth proxy %s: %v", deploymentName, err),
			},
			ready:            &ready,
			conditionMessage: fmt.Sprintf("Failed to get auth proxy deployment: %v", err),
			err:              fmt.Errorf("failed to get auth proxy %s: %w", deploymentName, err),
		}
	}

	if deployment.Status.AvailableReplicas < deployment.Status.Replicas || deployment.Status.Replicas == 0 {
		ready := false
		return authProxyProbeResult{
			result: endpointPhaseResult{
				phase:  dromosv1alpha1.PluginPhaseInitializing,
				reason: AuthProxyReasonNotReady,
				message: fmt.Sprintf("auth proxy %s not ready (%d/%d)",
					deploymentName, deployment.Status.AvailableReplicas, deployment.Status.Replicas),
			},
			ready: &ready,
			conditionMessage: fmt.Sprintf("Auth proxy deployment %s is not ready (%d/%d replicas available)",
				deploymentName, deployment.Status.AvailableReplicas, deployment.Status.Replicas),
		}
	}

	ready := true
	return authProxyProbeResult{
		result: endpointPhaseResult{
			phase:   dromosv1alpha1.PluginPhaseReady,
			reason:  AuthProxyReadyConditionName,
			message: fmt.Sprintf("auth proxy %s ready", deploymentName),
		},
		ready:            &ready,
		conditionMessage: fmt.Sprintf("Auth proxy deployment %s is ready", deploymentName),
	}
}

func (r authProxyProbeResult) condition() metav1.Condition {
	return metav1.Condition{
		Type:    AuthProxyReadyConditionName,
		Status:  conditionStatusFromResult(r.result, metav1.ConditionFalse),
		Reason:  r.result.reason,
		Message: r.conditionMessage,
	}
}

type externalAccessProbeResult struct {
	result           endpointPhaseResult
	conditionStatus  metav1.ConditionStatus
	conditionMessage string
	status           *endpoints.ExternalAccessStatus
	serviceURLs      []endpoints.ServiceURLStatus
	err              error
}

type externalAccessSpecs struct {
	ingress   *networkingv1.Ingress
	route     *openshiftv1.Route
	httpRoute *gatewayv1.HTTPRoute
}

func disabledExternalAccessProbe() externalAccessProbeResult {
	return externalAccessProbeResult{
		result: endpointPhaseResult{
			phase:   dromosv1alpha1.PluginPhaseDisabled,
			reason:  ExternalAccessReasonNotFound,
			message: "external access disabled",
		},
		conditionStatus:  metav1.ConditionFalse,
		conditionMessage: "external access disabled",
	}
}

func (r externalAccessProbeResult) condition() metav1.Condition {
	return metav1.Condition{
		Type:    ExternalAccessReadyConditionName,
		Status:  r.conditionStatus,
		Reason:  r.result.reason,
		Message: r.conditionMessage,
	}
}

type externalAccessResourceProbe struct {
	result           endpointPhaseResult
	conditionStatus  metav1.ConditionStatus
	conditionMessage string
	status           endpoints.ExternalAccessResourceStatus
	serviceURLs      []endpoints.ServiceURLStatus
	err              error
}

func (p *Plugin) probeExternalAccess(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	k8sClient client.Client,
	specs externalAccessSpecs,
) externalAccessProbeResult {
	if specs.ingress == nil && specs.route == nil && specs.httpRoute == nil {
		return disabledExternalAccessProbe()
	}

	var probes []externalAccessResourceProbe
	if specs.ingress != nil {
		probes = append(probes, p.probeIngress(ctx, runner, getResourceName(specs.ingress.Name, runner), k8sClient))
	}
	if specs.route != nil {
		probes = append(probes, p.probeRoute(ctx, runner, getResourceName(specs.route.Name, runner), k8sClient))
	}
	if specs.httpRoute != nil {
		probes = append(probes, p.probeHTTPRoute(ctx, runner, getResourceName(specs.httpRoute.Name, runner), k8sClient))
	}

	return aggregateExternalAccessProbes(probes)
}

func (p *Plugin) probeIngress(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
) externalAccessResourceProbe {
	ingress := &networkingv1.Ingress{}
	resource, handled := probeExternalAccessResource(ctx, runner, name, k8sClient, ingress, endpoints.ExternalAccessKindIngress, "Ingress", "ingress")
	if handled {
		return resource
	}

	url := ingressURL(ingress)
	message := fmt.Sprintf("ingress %s exists", name)
	conditionMessage := fmt.Sprintf("Ingress %s exists", name)
	if url != "" {
		message = fmt.Sprintf("ingress %s exists, url: %s", name, url)
		conditionMessage = fmt.Sprintf("Ingress %s exists, url: %s", name, url)
	}
	resource.result = endpointPhaseResult{
		phase:   dromosv1alpha1.PluginPhaseReady,
		reason:  ExternalAccessReadyConditionName,
		message: message,
	}
	resource.conditionStatus = metav1.ConditionTrue
	resource.conditionMessage = conditionMessage
	resource.status.Phase = endpoints.PhaseReady
	resource.status.Reason = ExternalAccessReadyConditionName
	resource.status.Message = fmt.Sprintf("ingress %s exists", name)
	resource.status.Hosts = ingressHosts(ingress)
	resource.serviceURLs = serviceURLs(endpoints.ExternalAccessKindIngress, name, ingressURLs(ingress))
	return resource
}

func (p *Plugin) probeRoute(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
) externalAccessResourceProbe {
	route := &openshiftv1.Route{}
	resource, handled := probeExternalAccessResource(ctx, runner, name, k8sClient, route, endpoints.ExternalAccessKindRoute, "Route", "route")
	if handled {
		return resource
	}
	url := routeURL(route)
	message := fmt.Sprintf("route %s exists", name)
	conditionMessage := fmt.Sprintf("Route %s exists", name)
	if url != "" {
		message = fmt.Sprintf("route %s exists, url: %s", name, url)
		conditionMessage = fmt.Sprintf("Route %s exists, url: %s", name, url)
	}
	resource.result = endpointPhaseResult{
		phase:   dromosv1alpha1.PluginPhaseReady,
		reason:  ExternalAccessReadyConditionName,
		message: message,
	}
	resource.conditionStatus = metav1.ConditionTrue
	resource.conditionMessage = conditionMessage
	resource.status.Phase = endpoints.PhaseReady
	resource.status.Reason = ExternalAccessReadyConditionName
	resource.status.Message = fmt.Sprintf("route %s exists", name)
	resource.status.Hosts = routeHosts(route)
	resource.serviceURLs = serviceURLs(endpoints.ExternalAccessKindRoute, name, routeURLs(route))
	return resource
}

func (p *Plugin) probeHTTPRoute(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
) externalAccessResourceProbe {
	httpRoute := &gatewayv1.HTTPRoute{}
	resource, handled := probeExternalAccessResource(ctx, runner, name, k8sClient, httpRoute, endpoints.ExternalAccessKindHTTPRoute, "HTTPRoute", "httproute")
	if handled {
		return resource
	}

	resource.status.Hosts = httpRouteHosts(httpRoute)
	acceptance := httpRouteAcceptanceResult(httpRoute)
	resource.status.Accepted = acceptance.accepted
	resource.status.Phase = detailsPhaseFromPluginPhase(acceptance.phase)
	resource.status.Reason = acceptance.reason
	resource.status.Message = httpRouteAcceptanceMessage("httproute", name, acceptance)
	resource.result = endpointPhaseResult{
		phase:   acceptance.phase,
		reason:  acceptance.reason,
		message: resource.status.Message,
	}
	resource.conditionStatus = acceptance.status
	resource.conditionMessage = httpRouteAcceptanceMessage("HTTPRoute", name, acceptance)
	if acceptance.phase == dromosv1alpha1.PluginPhaseReady {
		if url := httpRouteURL(httpRoute); url != "" {
			resource.result.message = fmt.Sprintf("httproute %s exists, url: %s", name, url)
			resource.conditionMessage = fmt.Sprintf("HTTPRoute %s exists, url: %s", name, url)
		}
		resource.serviceURLs = serviceURLs(endpoints.ExternalAccessKindHTTPRoute, name, httpRouteURLs(httpRoute))
	}
	return resource
}

func probeExternalAccessResource(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	name string,
	k8sClient client.Client,
	obj client.Object,
	kind endpoints.ExternalAccessKind,
	displayName string,
	lowerName string,
) (externalAccessResourceProbe, bool) {
	resource := externalAccessResourceProbe{
		status: endpoints.ExternalAccessResourceStatus{
			Kind: kind,
			Name: name,
		},
	}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: runner.Namespace}, obj)
	switch {
	case apierrors.IsNotFound(err):
		resource.result = endpointPhaseResult{
			phase:   dromosv1alpha1.PluginPhaseCreating,
			reason:  ExternalAccessReasonNotFound,
			message: fmt.Sprintf("%s %s not found", lowerName, name),
		}
		resource.conditionStatus = metav1.ConditionFalse
		resource.conditionMessage = fmt.Sprintf("%s %s not found", displayName, name)
		resource.status.Phase = endpoints.PhaseCreating
		resource.status.Reason = ExternalAccessReasonNotFound
		resource.status.Message = fmt.Sprintf("%s %s not found", lowerName, name)
		return resource, true
	case err != nil:
		log.FromContext(ctx).Error(err, "failed to get "+lowerName, "name", name, "namespace", runner.Namespace)
		resource.result = endpointPhaseResult{
			phase:   dromosv1alpha1.PluginPhaseFailed,
			reason:  ExternalAccessReasonGetFailed,
			message: fmt.Sprintf("failed to get %s %s: %v", lowerName, name, err),
		}
		resource.conditionStatus = metav1.ConditionUnknown
		resource.conditionMessage = fmt.Sprintf("Failed to get %s: %v", lowerName, err)
		resource.status.Phase = endpoints.PhaseFailed
		resource.status.Reason = ExternalAccessReasonGetFailed
		resource.status.Message = fmt.Sprintf("failed to get %s %s: %v", lowerName, name, err)
		resource.err = fmt.Errorf("failed to get %s %s: %w", lowerName, name, err)
		return resource, true
	default:
		return resource, false
	}
}

func aggregateExternalAccessProbes(probes []externalAccessResourceProbe) externalAccessProbeResult {
	if len(probes) == 0 {
		return disabledExternalAccessProbe()
	}

	status := &endpoints.ExternalAccessStatus{}
	var serviceURLs []endpoints.ServiceURLStatus
	var accumErr error

	for _, probe := range probes {
		status.Resources = append(status.Resources, probe.status)
		serviceURLs = append(serviceURLs, probe.serviceURLs...)
		accumErr = errors.Join(accumErr, probe.err)
	}

	// Aggregate readiness preserves the legacy priority order used by condition checks:
	// Ingress > Route > HTTPRoute. Secondary resources remain visible in Resources.
	primary := probes[0]
	status.Phase = primary.status.Phase
	status.Reason = primary.status.Reason
	status.Message = primary.status.Message

	return externalAccessProbeResult{
		result:           primary.result,
		conditionStatus:  primary.conditionStatus,
		conditionMessage: primary.conditionMessage,
		status:           status,
		serviceURLs:      serviceURLs,
		err:              accumErr,
	}
}

func conditionStatusFromResult(result endpointPhaseResult, disabledStatus metav1.ConditionStatus) metav1.ConditionStatus {
	switch result.phase {
	case dromosv1alpha1.PluginPhaseReady:
		return metav1.ConditionTrue
	case dromosv1alpha1.PluginPhaseFailed, dromosv1alpha1.PluginPhaseUnknown:
		return metav1.ConditionUnknown
	case dromosv1alpha1.PluginPhaseDisabled:
		return disabledStatus
	default:
		return metav1.ConditionFalse
	}
}
