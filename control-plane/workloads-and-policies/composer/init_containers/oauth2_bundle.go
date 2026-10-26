/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	corev1 "k8s.io/api/core/v1"
)

// OAuth2Bundle is the full set of pod artifacts that implement oauth2 auth for
// one workload's ExposedUrlInfo. The caller decides how to install them:
// batch/v1 Jobs and Kubeflow put Proxies in InitContainers (native sidecars);
// Knative puts them in Containers (regular sidecars, one port-declaring).
type OAuth2Bundle struct {
	// UsernameLoader is the runai-auth-user-loader init container that
	// materializes the per-URL allowlist files on the oauth2 auth volume. Nil
	// when no item carries an authorizedUsers list.
	UsernameLoader *corev1.Container

	// Proxies is one oauth2-proxy container per surviving ExposedUrlInfoItem,
	// in the same order as BuildOAuth2ProxyContainers returns them.
	Proxies []corev1.Container

	// Volumes is the oauth2 auth EmptyDir volume plus the optional CA secret
	// volume, ready to be appended to podSpec.Volumes.
	Volumes []corev1.Volume
}

// BuildOAuth2Bundle assembles the oauth2 proxies, the username-loader, and the
// supporting volumes for a workload from its ExposedUrlInfo. Returns an empty
// (zero Proxies, nil loader, empty volumes) bundle when the info has no items
// or when no proxies survive the ExternalAuthUrlEnabled filter, so callers can
// unconditionally range over Proxies without a nil check.
func BuildOAuth2Bundle(info *composer.ExposedUrlInfo) (*OAuth2Bundle, *workloadsApi.Error) {
	proxies, err := BuildOAuth2ProxyContainers(info)
	if err != nil {
		return nil, err
	}
	if len(proxies) == 0 {
		return &OAuth2Bundle{}, nil
	}

	loader, loaderErr := BuildUsernameLoaderInitContainer(info)
	if loaderErr != nil {
		return nil, loaderErr
	}

	return &OAuth2Bundle{
		UsernameLoader: loader,
		Proxies:        proxies,
		Volumes:        BuildOAuth2ProxyVolumes(info),
	}, nil
}
