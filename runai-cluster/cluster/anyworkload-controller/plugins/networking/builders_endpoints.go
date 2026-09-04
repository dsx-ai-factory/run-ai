/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"crypto/sha256"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getEndpointAuthProxyName(workloadServiceName string) string {
	name := workloadServiceName + AuthProxySuffix
	const (
		maxResourceNameLength = 63
		hashLength            = 10
	)
	if len(name) <= maxResourceNameLength {
		return name
	}

	// Preserve existing names whenever Kubernetes accepts them. For longer
	// derived names, retain the resource-type suffix and hash the complete
	// original name so reconciliation stays deterministic without collisions.
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(name)))[:hashLength]
	baseLength := maxResourceNameLength - len(AuthProxySuffix) - hashLength - 1
	return fmt.Sprintf("%s-%s%s", workloadServiceName[:baseLength], digest, AuthProxySuffix)
}

func getEndpointExternalHostAndPath(endpoint *endpoints.Endpoint) (host, path string) {
	if endpoint == nil || endpoint.ExternalAccess == nil {
		return "", "/"
	}
	ea := endpoint.ExternalAccess
	return resolveExternalHostAndPath(ea.Ingress, ea.Route, ea.HTTPRoute)
}

// buildEndpointWorkloadService preserves the spec name verbatim so producers retain control of cross-Endpoint backend resolution.
func (p *Plugin) buildEndpointWorkloadService(
	runner *dromosv1alpha1.WorkloadRunner,
	serviceCreate *corev1.Service,
) *corev1.Service {
	svc := serviceCreate.DeepCopy()
	svc.TypeMeta = metav1.TypeMeta{
		APIVersion: constants.CoreV1,
		Kind:       constants.KindService,
	}
	svc.Namespace = runner.Namespace
	if isMetricsService(svc) {
		if svc.Labels == nil {
			svc.Labels = map[string]string{}
		}
		// This label is controller-owned bookkeeping used only to scope metrics
		// Service pruning. All labels supplied by the composer remain intact.
		svc.Labels[LabelWorkloadRunner] = runner.Name
	}
	return svc
}

func (p *Plugin) buildEndpointAuthProxyDeployment(
	runner *dromosv1alpha1.WorkloadRunner,
	svc *endpoints.Service,
	name string,
	upstreamURL string,
	externalHost string,
	externalPath string,
) (*appsv1.Deployment, error) {
	container, err := p.buildEndpointOAuth2ProxyContainer(svc, upstreamURL, name, externalHost, externalPath)
	if err != nil {
		return nil, err
	}
	initContainers, err := p.buildEndpointAuthProxyInitContainers(svc, name)
	if err != nil {
		return nil, err
	}
	return composeAuthProxyDeployment(
		runner, name, container,
		initContainers,
		p.buildAuthProxyVolumes(),
	), nil
}

func (p *Plugin) buildEndpointAuthProxyInitContainers(
	svc *endpoints.Service,
	nick string,
) ([]corev1.Container, error) {
	if svc == nil || svc.Auth == nil || svc.Auth.AuthorizedUsers == "" {
		return nil, nil
	}
	initContainer, err := workloadbuilders.BuildUsernameLoaderInitContainerWithError(svc.Auth.AuthorizedUsers, nick)
	if err != nil {
		return nil, err
	}
	return []corev1.Container{*initContainer}, nil
}

func (p *Plugin) buildEndpointOAuth2ProxyContainer(
	svc *endpoints.Service,
	upstreamURL string,
	cookiePrefix string,
	externalHost string,
	externalPath string,
) (*corev1.Container, error) {
	var authorizedUsers, authorizedGroups *string
	forceJsonResponse := false
	if svc != nil && svc.Auth != nil {
		auth := svc.Auth
		if auth.AuthorizedUsers != "" {
			authorizedUsers = &auth.AuthorizedUsers
		}
		if auth.AuthorizedGroups != "" {
			authorizedGroups = &auth.AuthorizedGroups
		}
		if auth.ForceJsonResponse != nil {
			forceJsonResponse = *auth.ForceJsonResponse
		}
	}
	return buildOAuth2ProxyContainerFromAuth(
		upstreamURL, cookiePrefix, externalHost, externalPath,
		authorizedUsers, authorizedGroups, forceJsonResponse,
	)
}
