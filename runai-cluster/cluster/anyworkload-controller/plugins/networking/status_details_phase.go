/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
)

func endpointDetailsPhase(endpointStatus endpoints.EndpointStatus, authResult endpointPhaseResult) endpoints.Phase {
	// Endpoint phase mirrors the plugin readiness contract: auth proxy plus external access.
	// ServiceStatus.Phase is still reported per service, but service existence has not driven plugin phase.
	phase := endpoints.PhaseDisabled
	if endpointStatus.ExternalAccess != nil {
		phase = worstDetailsPhase(phase, endpointStatus.ExternalAccess.Phase)
	}
	phase = worstDetailsPhase(phase, detailsPhaseFromPluginPhase(authResult.phase))
	return phase
}

func detailsPhaseFromPluginPhase(phase dromosv1alpha1.PluginPhase) endpoints.Phase {
	switch phase {
	case dromosv1alpha1.PluginPhaseReady:
		return endpoints.PhaseReady
	case dromosv1alpha1.PluginPhaseCreating:
		return endpoints.PhaseCreating
	case dromosv1alpha1.PluginPhasePending:
		return endpoints.PhasePending
	case dromosv1alpha1.PluginPhaseInitializing:
		return endpoints.PhaseInitializing
	case dromosv1alpha1.PluginPhaseFailed:
		return endpoints.PhaseFailed
	case dromosv1alpha1.PluginPhaseUnknown:
		return endpoints.PhaseUnknown
	default:
		return endpoints.PhaseDisabled
	}
}

func pluginPhaseFromNetworkingStatus(status *endpoints.NetworkingStatus) dromosv1alpha1.PluginPhase {
	if status == nil || len(status.Endpoints) == 0 {
		return dromosv1alpha1.PluginPhaseDisabled
	}
	phase := endpoints.PhaseDisabled
	for _, endpoint := range status.Endpoints {
		phase = worstDetailsPhase(phase, endpoint.Phase)
	}
	return pluginPhaseFromDetailsPhase(phase)
}

func pluginPhaseFromDetailsPhase(phase endpoints.Phase) dromosv1alpha1.PluginPhase {
	switch phase {
	case endpoints.PhaseReady:
		return dromosv1alpha1.PluginPhaseReady
	case endpoints.PhaseCreating:
		return dromosv1alpha1.PluginPhaseCreating
	case endpoints.PhasePending:
		return dromosv1alpha1.PluginPhasePending
	case endpoints.PhaseInitializing:
		return dromosv1alpha1.PluginPhaseInitializing
	case endpoints.PhaseFailed:
		return dromosv1alpha1.PluginPhaseFailed
	case endpoints.PhaseUnknown:
		return dromosv1alpha1.PluginPhaseUnknown
	default:
		return dromosv1alpha1.PluginPhaseDisabled
	}
}

func worstDetailsPhase(a, b endpoints.Phase) endpoints.Phase {
	if detailsPhaseRank(b) > detailsPhaseRank(a) {
		return b
	}
	return a
}

func detailsPhaseRank(phase endpoints.Phase) int {
	switch phase {
	case endpoints.PhaseDisabled:
		return 0
	case endpoints.PhaseReady:
		return 1
	case endpoints.PhaseCreating, endpoints.PhasePending:
		return 2
	case endpoints.PhaseInitializing:
		return 3
	case endpoints.PhaseFailed, endpoints.PhaseUnknown:
		return 4
	default:
		return 4
	}
}
