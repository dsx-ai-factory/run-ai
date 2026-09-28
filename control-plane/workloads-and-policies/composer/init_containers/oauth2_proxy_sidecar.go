/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package init_containers

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	corev1 "k8s.io/api/core/v1"
)

// ServingPortItemNick is the nickname used for the synthetic ExposedUrlInfoItem
// that represents the authenticated serving-port proxy.
const ServingPortItemNick = "serving-port"

// ServingProxySpec bundles the caller-controlled inputs required by
// BuildOauth2ProxySidecar. Callers precompute the serving URL and cookie prefix;
// this helper only owns the item construction, optional exposed-URL merge, and
// the raw install into a *corev1.PodSpec.
type ServingProxySpec struct {
	// Required — the internal port the serving proxy forwards to (containerPort).
	ContainerPort int32
	// Required — the upstream URL the serving proxy is reachable at.
	ServingURL string
	// Required — cookie prefix / name for this workload's proxy.
	CookiePrefix string

	// Comma-separated allowlists. Empty means "no explicit allowlist for this axis".
	AuthorizedUsers  string
	AuthorizedGroups string

	// Optional — proxy → user-container upstream response timeout. Zero omits.
	UpstreamTimeoutSeconds int64

	// Required — used by BuildOAuth2Proxy* helpers for naming/labeling and
	// cluster-config lookups.
	WorkloadName string
	ProjectName  string
	Config       composer.ClusterStorageConfig

	// Optional. When non-nil, the caller's exposed-URL items are merged into
	// the same info map alongside the serving-port item; each merged item is
	// forced to EnableSidecarRestartPolicy=false and OmitPortDeclaration=true
	// (the Knative-compatible external-URL proxy shape). Used by inference to
	// fold auth-gated exposedUrls into the sidecar path. Distributed inference
	// passes nil.
	ExposedURLInfo *composer.ExposedUrlInfo

	// CustomizeServingProxy runs after the serving proxy has been built but
	// before it (and any external proxies / username-loader / volumes) is
	// appended to podSpec. Callers may mutate the passed-in servingProxy
	// (mutations survive because the copy that gets appended is the one being
	// mutated) and may mutate podSpec itself — including any pointers they
	// captured into podSpec.Containers before calling BuildOauth2ProxySidecar,
	// which would otherwise be invalidated by the append's slice reallocation.
	// Optional.
	CustomizeServingProxy func(podSpec *corev1.PodSpec, servingProxy *corev1.Container) *workloadsApi.Error
}

// BuildOauth2ProxySidecar builds the oauth2-proxy sidecar (and, if applicable,
// any external-URL proxies + username-loader + volumes) for a workload's
// authenticated serving port, and installs them into podSpec.
//
// This helper is shared by the inference (Knative) and distributed-inference
// (LWS) composers. It does not enforce a "public means skip" policy — callers
// are expected to bail out before invoking it when the port is public.
func BuildOauth2ProxySidecar(podSpec *corev1.PodSpec, spec ServingProxySpec) *workloadsApi.Error {
	if podSpec == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			"oauth2-proxy sidecar: podSpec is required")
	}
	if spec.ContainerPort <= 0 {
		return workloadsApi.NewError(http.StatusBadRequest,
			"servingPort.container must be greater than zero")
	}
	if spec.Config == nil {
		spec.Config = composer.DefaultClusterStorageConfig()
	}

	item := &composer.ExposedUrlInfoItem{
		Nick:                             ServingPortItemNick,
		ContainerPort:                    convutils.Ref(spec.ContainerPort),
		Url:                              spec.ServingURL,
		OAuthPort:                        convutils.Ref(composer.Oauth2ContainerDefaultPort),
		BuildWhenExternalAuthURLDisabled: true,
		OmitPortDeclaration:              true,
		ForceJSONResponse:                true,
		EnableSidecarRestartPolicy:       convutils.Ref(false),
		CookiePrefix:                     spec.CookiePrefix,
		UpstreamTimeoutSeconds:           spec.UpstreamTimeoutSeconds,
	}
	if spec.AuthorizedUsers != "" {
		item.AuthorizedUsers = convutils.Ref(spec.AuthorizedUsers)
	}
	if spec.AuthorizedGroups != "" {
		item.AuthorizedGroups = convutils.Ref(spec.AuthorizedGroups)
	}

	items := make(map[string]*composer.ExposedUrlInfoItem, 1)
	items[ServingPortItemNick] = item
	if spec.ExposedURLInfo != nil {
		for nick, exposedItem := range spec.ExposedURLInfo.Items {
			if exposedItem == nil {
				items[nick] = nil
				continue
			}
			// Force Knative-compatible shape for the external-URL proxies
			// that ride alongside the serving proxy in spec.containers.
			knativeItem := *exposedItem
			knativeItem.EnableSidecarRestartPolicy = convutils.Ref(false)
			knativeItem.OmitPortDeclaration = true
			items[nick] = &knativeItem
		}
	}
	info := &composer.ExposedUrlInfo{
		Items:        items,
		WorkloadName: spec.WorkloadName,
		ProjectName:  spec.ProjectName,
		Config:       spec.Config,
	}
	if spec.ExposedURLInfo != nil {
		info.ResourcePrefix = spec.ExposedURLInfo.ResourcePrefix
		info.ForceSuffixInUrls = spec.ExposedURLInfo.ForceSuffixInUrls
	}

	proxies, buildErr := BuildOAuth2ProxyContainers(info)
	if buildErr != nil {
		return buildErr
	}

	var servingProxy *corev1.Container
	var externalProxies []corev1.Container
	for i := range proxies {
		proxy := &proxies[i]
		if len(proxy.Ports) != 0 {
			if servingProxy != nil {
				return workloadsApi.NewError(http.StatusInternalServerError,
					"oauth2-proxy sidecar: multiple serving proxy containers were generated")
			}
			servingProxy = proxy
			continue
		}
		externalProxies = append(externalProxies, *proxy)
	}
	if servingProxy == nil {
		return workloadsApi.NewError(http.StatusInternalServerError,
			fmt.Sprintf("oauth2-proxy sidecar: serving proxy container was not generated from %d proxies", len(proxies)))
	}

	loader, loaderErr := BuildUsernameLoaderInitContainer(info)
	if loaderErr != nil {
		return loaderErr
	}

	// Run the caller's proxy customization BEFORE we touch podSpec so that
	// (a) mutations to servingProxy are captured by the append below, and
	// (b) any caller-held pointers into podSpec.Containers stay valid (the
	//     append that adds the proxies would otherwise reallocate the slice
	//     and strand those pointers on the old backing array).
	if spec.CustomizeServingProxy != nil {
		if err := spec.CustomizeServingProxy(podSpec, servingProxy); err != nil {
			return err
		}
	}

	if loader != nil {
		podSpec.InitContainers = append(podSpec.InitContainers, *loader)
	}
	podSpec.Containers = append(podSpec.Containers, *servingProxy)
	podSpec.Containers = append(podSpec.Containers, externalProxies...)
	podSpec.Volumes = append(podSpec.Volumes, BuildOAuth2ProxyVolumes(info)...)
	return nil
}
