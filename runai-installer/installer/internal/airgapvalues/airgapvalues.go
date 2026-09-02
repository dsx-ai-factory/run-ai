/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package airgapvalues builds the Helm values that point the Run:ai charts at a
// mirror registry. It is the one place that knows how each chart consumes an
// image reference, which the charts do not agree on: most take a
// {registry, repository, tag} triple, Bitnami subcharts additionally guard the
// registry they accept, some have no registry field at all, and a few images
// are created by the cluster operator rather than by Helm.
//
// Run:ai's own images need no entry: they are composed from the chart-wide
// registry the entry points set.
package airgapvalues

import "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/artifacts"

// Image is one mirrored image with its lock group and source tag.
type Image struct {
	Group        string
	GroupVersion string
	Name         string
	Tag          string
}

type imageKey struct {
	group string
	name  string
}

const (
	controlPlaneGroup = "runai-control-plane"
	clusterGroup      = "runai-cluster"
	kaiGroup          = "kai-scheduler"
	kartaGroup        = "karta"

	// ControlPlaneValuesFile is the generated control-plane values filename.
	ControlPlaneValuesFile = "control-plane-airgap-values.yaml"
	// ClusterValuesFile is the generated cluster values filename.
	ClusterValuesFile = "cluster-airgap-values.yaml"
)

// controlPlaneBindings maps a bundled image to where the control-plane chart
// reads its reference from.
var controlPlaneBindings = map[imageKey][]binding{
	{controlPlaneGroup, "postgresql"}:  {bitnamiImage{subchart: "postgresql"}},
	{controlPlaneGroup, "thanos"}:      {bitnamiImage{subchart: "thanos"}},
	{controlPlaneGroup, "nats"}:        {bitnamiImage{subchart: "nats"}},
	{controlPlaneGroup, "keycloak"}:    {fullReferenceImage{path: "keycloakx.image"}},
	{controlPlaneGroup, "traefik"}:     {registryImage{path: "traefik.image"}},
	{controlPlaneGroup, "postgres"}:    {registryImage{path: "postgresqlUserInit.image"}},
	{clusterGroup, "wait-ubi-minimal"}: {registryImage{path: "global.waitImage"}},
}

// clusterBindings maps a bundled image to where the cluster chart reads its
// reference from. The images the operator creates at runtime are reached
// through clusterConfig, which patches RunaiConfig.
var clusterBindings = map[imageKey][]binding{
	{kartaGroup, "operator"}: {fullReferenceImage{path: "karta.image"}},
	{clusterGroup, "valkey"}: {runaiConfigImage{path: "cluster-redis.image"}},
	{clusterGroup, "git-sync"}: {
		runaiConfigImage{path: "anyworkload-controller.gitSyncImage"},
	},
	{clusterGroup, "wait-ubi-minimal"}: {
		runaiConfigImage{path: "anyworkload-controller.runaiOauth2ProxyUsernameLoaderImage"},
	},
}

// ControlPlane builds the values that install the control-plane chart from the mirror.
func ControlPlane(registry string, images []Image) map[string]any {
	v := values{}
	v.set("global.installationType", "airgap")
	v.set("global.image.registry", artifacts.RegistryPrefix(registry, controlPlaneGroup))
	v.set("global.image.tag", groupVersion(images, controlPlaneGroup))
	// The cluster chart is installed from the bundle, not downloaded.
	v.set("clusterService.config.clusterInstallation.helmRegistryUrl", "")
	apply(v, controlPlaneBindings, registry, images)
	return v
}

// Cluster builds the values that install the cluster chart from the mirror.
func Cluster(registry string, images []Image) map[string]any {
	v := values{}
	v.set("global.image.registry", artifacts.RegistryPrefix(registry, clusterGroup))
	// The KAI subchart composes its own images; a top-level global.registry
	// would leak into it, so its own key is set instead.
	v.set("kai-scheduler.global.registry", artifacts.RegistryPrefix(registry, kaiGroup))
	apply(v, clusterBindings, registry, images)
	return v
}

// apply writes every bundled image that the chart has a binding for. An image
// with no binding needs none, and a binding with no image was not packed.
func apply(v values, bindings map[imageKey][]binding, registry string, images []Image) {
	for _, image := range images {
		for _, b := range bindings[imageKey{image.Group, image.Name}] {
			b.apply(v, artifacts.RegistryPrefix(registry, image.Group), image)
		}
	}
}

func groupVersion(images []Image, group string) string {
	for _, image := range images {
		if image.Group == group {
			return image.GroupVersion
		}
	}
	return ""
}
