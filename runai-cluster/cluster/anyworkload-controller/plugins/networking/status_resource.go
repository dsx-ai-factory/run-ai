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
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	AuthProxyReadyConditionName = "AuthProxyReady"

	AuthProxyReasonDisabled  = "Disabled"
	AuthProxyReasonNotFound  = "AuthProxyNotFound"
	AuthProxyReasonGetFailed = "AuthProxyGetFailed"
	AuthProxyReasonNotReady  = "AuthProxyNotReady"

	ExternalAccessReadyConditionName = "ExternalAccessReady"

	ExternalAccessReasonNotFound  = "ExternalAccessNotFound"
	ExternalAccessReasonGetFailed = "ExternalAccessGetFailed"

	HTTPRouteNotAcceptedReason = "HTTPRouteNotAccepted"
)

// Status dispatches on input shape; see README "Input shapes".
func (p *Plugin) Status(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, status framework.PluginStatusSetter, state *reconcile.State) error {
	logger := log.FromContext(ctx)
	if status == nil {
		logger.V(3).Info("nil status for plugin", "workloadRunner", runner.Name, "plugin", p.Name())
		return fmt.Errorf("nil status for plugin %s", p.Name())
	}
	if err := setNetworkingAdditionalInfo(status, nil); err != nil {
		return err
	}

	if state != nil {
		for _, iss := range state.PluginIssues(p) {
			if iss.IsCritical {
				status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
				return nil
			}
		}
	}

	shape, cfg, classifyErr := classifyInput(runner)
	if classifyErr != nil {
		status.SetPhase(dromosv1alpha1.PluginPhaseFailed)
		status.SetCondition(metav1.Condition{
			Type:    ExternalAccessReadyConditionName,
			Status:  metav1.ConditionUnknown,
			Reason:  "InvalidNetworkingInput",
			Message: classifyErr.Error(),
		})
		return nil //nolint:nilerr // user-input error surfaced on status; requeue cannot fix it.
	}
	switch shape {
	case inputLegacy:
		return p.statusLegacy(ctx, runner, k8sClient, status)
	case inputAnnotation:
		return p.statusEndpoints(ctx, runner, cfg, k8sClient, status)
	case inputNone:
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}
	return nil
}

func (p *Plugin) statusLegacy(ctx context.Context, runner *dromosv1alpha1.WorkloadRunner, k8sClient client.Client, status framework.PluginStatusSetter) error {
	config := extractNetworkingConfig(runner)

	if !shouldReportNetworkingStatus(config) {
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	authProbe := disabledAuthProxyProbe()
	if !isAuthEnabled(config) {
		status.SetCondition(authProbe.condition())
	} else {
		authProbe = p.probeAuthProxy(ctx, runner, getAuthProxyName(runner), k8sClient)
		status.SetCondition(authProbe.condition())
	}

	externalProbe := disabledExternalAccessProbe()
	if config == nil || config.ExternalAccess == nil {
		externalProbe = disabledExternalAccessProbe()
	} else {
		externalProbe = p.probeExternalAccess(ctx, runner, k8sClient, externalAccessSpecs{
			ingress:   config.ExternalAccess.Ingress,
			route:     config.ExternalAccess.Route,
			httpRoute: config.ExternalAccess.HTTPRoute,
		})
		status.SetCondition(externalProbe.condition())
	}

	details, detailsErr := p.buildInstructionsNetworkingDetails(ctx, runner, config, k8sClient, authProbe, externalProbe)
	detailsErr = errors.Join(detailsErr, setNetworkingAdditionalInfo(status, details))
	status.SetPhase(pluginPhaseFromNetworkingStatus(details))
	if joinErr := errors.Join(authProbe.err, externalProbe.err, detailsErr); joinErr != nil {
		return joinErr
	}
	return nil
}

func ingressURL(ingress *networkingv1.Ingress) string {
	if len(ingress.Spec.Rules) == 0 || ingress.Spec.Rules[0].Host == "" {
		return ""
	}
	host := ingress.Spec.Rules[0].Host
	// Default https: controllers commonly terminate TLS via wildcard certs without populating Spec.TLS.
	scheme := "https"
	path := "/"
	if ingress.Spec.Rules[0].HTTP != nil && len(ingress.Spec.Rules[0].HTTP.Paths) > 0 {
		path = ingress.Spec.Rules[0].HTTP.Paths[0].Path
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

func routeURL(route *openshiftv1.Route) string {
	if route.Spec.Host == "" {
		return ""
	}
	scheme := "http"
	if route.Spec.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/", scheme, route.Spec.Host)
}

func httpRouteURL(httpRoute *gatewayv1.HTTPRoute) string {
	if len(httpRoute.Spec.Hostnames) == 0 {
		return ""
	}
	return fmt.Sprintf("https://%s/", httpRoute.Spec.Hostnames[0])
}

func httpRouteAcceptedConditions(httpRoute *gatewayv1.HTTPRoute) []metav1.Condition {
	conditions := []metav1.Condition{}
	for _, parentStatus := range httpRoute.Status.Parents {
		for i := range parentStatus.Conditions {
			if parentStatus.Conditions[i].Type == string(gatewayv1.RouteConditionAccepted) {
				conditions = append(conditions, metav1.Condition{
					Status:  parentStatus.Conditions[i].Status,
					Message: parentStatus.Conditions[i].Message,
				})
			}
		}
	}
	return conditions
}

type httpRouteAcceptance struct {
	phase    dromosv1alpha1.PluginPhase
	status   metav1.ConditionStatus
	reason   string
	message  string
	accepted *bool
}

func httpRouteAcceptanceResult(httpRoute *gatewayv1.HTTPRoute) httpRouteAcceptance {
	acceptedConditions := httpRouteAcceptedConditions(httpRoute)
	if len(acceptedConditions) == 0 {
		return httpRouteAcceptance{
			phase:  dromosv1alpha1.PluginPhaseReady,
			status: metav1.ConditionTrue,
			reason: ExternalAccessReadyConditionName,
		}
	}

	var rejectedCondition metav1.Condition
	hasRejectedCondition := false
	for i := range acceptedConditions {
		switch acceptedConditions[i].Status {
		case metav1.ConditionTrue:
			return httpRouteAcceptance{
				phase:    dromosv1alpha1.PluginPhaseReady,
				status:   metav1.ConditionTrue,
				reason:   ExternalAccessReadyConditionName,
				accepted: ptr.To(true),
			}
		case metav1.ConditionFalse:
			if !hasRejectedCondition {
				rejectedCondition = acceptedConditions[i]
				hasRejectedCondition = true
			}
		}
	}

	if hasRejectedCondition {
		return httpRouteAcceptance{
			phase:    dromosv1alpha1.PluginPhaseInitializing,
			status:   metav1.ConditionFalse,
			reason:   HTTPRouteNotAcceptedReason,
			message:  rejectedCondition.Message,
			accepted: ptr.To(false),
		}
	}

	return httpRouteAcceptance{
		phase:  dromosv1alpha1.PluginPhaseReady,
		status: metav1.ConditionTrue,
		reason: ExternalAccessReadyConditionName,
	}
}

func httpRouteAcceptanceMessage(kind, name string, acceptance httpRouteAcceptance) string {
	switch acceptance.reason {
	case ExternalAccessReadyConditionName:
		return fmt.Sprintf("%s %s exists", kind, name)
	case HTTPRouteNotAcceptedReason:
		if acceptance.message != "" {
			return fmt.Sprintf("%s %s not accepted by gateway: %s", kind, name, acceptance.message)
		}
		return fmt.Sprintf("%s %s not accepted by gateway", kind, name)
	default:
		return fmt.Sprintf("%s %s exists", kind, name)
	}
}

// worstPhase returns the more severe of two plugin phases.
// Severity (ascending): Disabled < Ready < Creating < Initializing < Failed.
func worstPhase(a, b dromosv1alpha1.PluginPhase) dromosv1alpha1.PluginPhase {
	return pluginPhases[max(pluginPhaseRank(a), pluginPhaseRank(b))]
}

var pluginPhases = []dromosv1alpha1.PluginPhase{
	dromosv1alpha1.PluginPhaseDisabled,
	dromosv1alpha1.PluginPhaseReady,
	dromosv1alpha1.PluginPhaseCreating,
	dromosv1alpha1.PluginPhaseInitializing,
	dromosv1alpha1.PluginPhaseFailed,
}

func pluginPhaseRank(phase dromosv1alpha1.PluginPhase) int {
	for i, p := range pluginPhases {
		if p == phase {
			return i
		}
	}
	return len(pluginPhases) - 1 // unknown phases treated as worst
}
