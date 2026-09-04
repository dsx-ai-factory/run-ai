/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"
	"strings"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// endpointPhaseResult is one Endpoint's contribution to the folded AuthProxyReady / ExternalAccessReady condition.
type endpointPhaseResult struct {
	phase   dromosv1alpha1.PluginPhase
	reason  string
	message string
}

func (p *Plugin) statusEndpoints(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	cfg *endpoints.Config,
	k8sClient client.Client,
	status framework.PluginStatusSetter,
) error {
	if cfg == nil || len(cfg.Endpoints) == 0 {
		status.SetPhase(dromosv1alpha1.PluginPhaseDisabled)
		return nil
	}

	probes := make([]endpointNetworkingProbe, len(cfg.Endpoints))
	authResults := make([]endpointPhaseResult, len(cfg.Endpoints))
	accessResults := make([]endpointPhaseResult, len(cfg.Endpoints))
	var probeErr error

	for i := range cfg.Endpoints {
		endpoint := &cfg.Endpoints[i]
		probes[i] = p.probeEndpointNetworking(ctx, runner, endpoint, k8sClient)
		authResults[i] = probes[i].auth.result
		accessResults[i] = probes[i].external.result
		probeErr = errors.Join(probeErr, probes[i].auth.err, probes[i].external.err)
	}

	status.SetCondition(foldAuthProxyConditions(authResults))
	status.SetCondition(foldExternalAccessConditions(accessResults))

	details, detailsErr := p.buildEndpointsNetworkingDetails(ctx, runner, cfg, k8sClient, probes)
	detailsErr = errors.Join(detailsErr, setNetworkingAdditionalInfo(status, details))
	status.SetPhase(pluginPhaseFromNetworkingStatus(details))

	return errors.Join(probeErr, detailsErr)
}

func (p *Plugin) probeEndpointNetworking(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	endpoint *endpoints.Endpoint,
	k8sClient client.Client,
) endpointNetworkingProbe {
	authProbe := disabledAuthProxyProbe()
	svc := firstService(endpoint)
	if svc != nil && svc.Auth != nil {
		workloadServiceName := endpointServiceName(svc)
		if workloadServiceName == "" {
			authProbe = authProxyProbeResult{
				result: endpointPhaseResult{
					phase:   dromosv1alpha1.PluginPhaseFailed,
					reason:  AuthProxyReasonGetFailed,
					message: "endpoint service has no resolvable name",
				},
				conditionMessage: "endpoint service has no resolvable name",
			}
		} else {
			authProbe = p.probeAuthProxy(ctx, runner, getEndpointAuthProxyName(workloadServiceName), k8sClient)
		}
	}

	externalProbe := disabledExternalAccessProbe()
	if endpoint != nil && endpoint.ExternalAccess != nil {
		externalProbe = p.probeExternalAccess(ctx, runner, k8sClient, externalAccessSpecs{
			ingress:   endpoint.ExternalAccess.Ingress,
			route:     endpoint.ExternalAccess.Route,
			httpRoute: endpoint.ExternalAccess.HTTPRoute,
		})
	}

	return endpointNetworkingProbe{
		auth:     authProbe,
		external: externalProbe,
	}
}

// Reason values mirror the legacy single-Endpoint format so substring-grep consumers (cluster-sync) keep working.
func foldAuthProxyConditions(results []endpointPhaseResult) metav1.Condition {
	return foldResultsCondition(results, AuthProxyReadyConditionName, AuthProxyReasonDisabled)
}

func foldExternalAccessConditions(results []endpointPhaseResult) metav1.Condition {
	return foldResultsCondition(results, ExternalAccessReadyConditionName, ExternalAccessReasonNotFound)
}

func foldResultsCondition(results []endpointPhaseResult, conditionType, disabledReason string) metav1.Condition {
	allDisabled := true
	for _, r := range results {
		if r.phase != dromosv1alpha1.PluginPhaseDisabled {
			allDisabled = false
			break
		}
	}
	worst := worstPhaseFromResults(results)

	cond := metav1.Condition{Type: conditionType}
	switch {
	case allDisabled:
		cond.Status = metav1.ConditionFalse
		cond.Reason = disabledReason
		cond.Message = formatResultsMessage(results, "disabled across all endpoints")
	case worst == dromosv1alpha1.PluginPhaseReady:
		cond.Status = metav1.ConditionTrue
		cond.Reason = conditionType
		cond.Message = formatResultsMessage(results, "ready across all endpoints")
	default:
		cond.Status = metav1.ConditionFalse
		cond.Reason = firstNonReadyReason(results, conditionType)
		cond.Message = formatResultsMessage(results, fmt.Sprintf("aggregate phase: %s", worst))
	}
	return cond
}

func formatResultsMessage(results []endpointPhaseResult, fallback string) string {
	if len(results) == 1 {
		return results[0].message
	}
	parts := make([]string, 0, len(results))
	for i, r := range results {
		parts = append(parts, fmt.Sprintf("endpoint[%d]=%s", i, r.message))
	}
	if len(parts) == 0 {
		return fallback
	}
	return strings.Join(parts, "; ")
}

func firstNonReadyReason(results []endpointPhaseResult, defaultReason string) string {
	for _, r := range results {
		if r.phase != dromosv1alpha1.PluginPhaseReady && r.phase != dromosv1alpha1.PluginPhaseDisabled {
			return r.reason
		}
	}
	return defaultReason
}

func worstPhaseFromResults(results []endpointPhaseResult) dromosv1alpha1.PluginPhase {
	worst := dromosv1alpha1.PluginPhaseDisabled
	for _, r := range results {
		worst = worstPhase(worst, r.phase)
	}
	return worst
}
