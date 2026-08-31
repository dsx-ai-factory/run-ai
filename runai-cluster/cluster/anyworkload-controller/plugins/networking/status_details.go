/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"

	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setNetworkingAdditionalInfo(status framework.PluginStatusSetter, details *endpoints.NetworkingStatus) error {
	raw, err := endpoints.NetworkingStatusToRawExtension(details)
	if err != nil {
		return err
	}
	status.SetAdditionalInfo(raw)
	return nil
}

func (p *Plugin) buildInstructionsNetworkingDetails(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	config *dromosv1alpha1.NetworkingConfig,
	k8sClient client.Client,
	authProbe authProxyProbeResult,
	externalProbe externalAccessProbeResult,
) (*endpoints.NetworkingStatus, error) {
	details := endpoints.NewNetworkingStatus(runner.Generation)
	endpointStatus := endpoints.EndpointStatus{}
	var accumErr error

	authEnabled := isAuthEnabled(config)
	authResult := authProbe.result

	if hasService(config) {
		info := getWorkloadServiceInfo(config, runner.Name)
		serviceStatus, err := p.buildServiceDetails(ctx, runner, info.Name, k8sClient)
		accumErr = errors.Join(accumErr, err)
		if authEnabled {
			serviceStatus.AuthProxyReady = authProbe.ready
		}
		endpointStatus.Services = append(endpointStatus.Services, serviceStatus)
	}

	if config != nil && config.ExternalAccess != nil && externalProbe.status != nil {
		endpointStatus.ExternalAccess = externalProbe.status
		addURLsToServices(&endpointStatus, externalProbe.serviceURLs)
	}

	endpointStatus.Phase = endpointDetailsPhase(endpointStatus, authResult)
	details.Endpoints = []endpoints.EndpointStatus{endpointStatus}
	return details, accumErr
}

func (p *Plugin) buildEndpointsNetworkingDetails(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	cfg *endpoints.Config,
	k8sClient client.Client,
	probes []endpointNetworkingProbe,
) (*endpoints.NetworkingStatus, error) {
	details := endpoints.NewNetworkingStatus(runner.Generation)
	details.Endpoints = make([]endpoints.EndpointStatus, 0, len(cfg.Endpoints))
	var accumErr error

	for i := range cfg.Endpoints {
		endpoint := &cfg.Endpoints[i]
		probe := endpointNetworkingProbe{
			auth:     disabledAuthProxyProbe(),
			external: disabledExternalAccessProbe(),
		}
		if i < len(probes) {
			probe = probes[i]
		}
		authResult := probe.auth.result

		endpointStatus := endpoints.EndpointStatus{}
		for j := range endpoint.Services {
			svc := &endpoint.Services[j]
			serviceStatus, err := p.buildServiceDetails(ctx, runner, endpointServiceName(svc), k8sClient)
			accumErr = errors.Join(accumErr, err)
			if svc.Auth != nil {
				serviceStatus.AuthProxyReady = probe.auth.ready
			}
			endpointStatus.Services = append(endpointStatus.Services, serviceStatus)
		}

		if endpoint.ExternalAccess != nil && probe.external.status != nil {
			endpointStatus.ExternalAccess = probe.external.status
			addURLsToServices(&endpointStatus, probe.external.serviceURLs)
		}

		endpointStatus.Phase = endpointDetailsPhase(endpointStatus, authResult)
		details.Endpoints = append(details.Endpoints, endpointStatus)
	}

	return details, accumErr
}
