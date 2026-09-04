/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package route

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/enrich"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/enrich/enrichers"
	"github.com/spf13/viper"
)

func (brcp *BasicRouteConfigurationProvider) GetGlobalEnrichingConfig() *enrich.EnrichingConfig {
	return &brcp.config.Enrich
}

func (brcp *BasicRouteConfigurationProvider) GetGlobalEnrichers() []string {
	enrichers := []string{}
	enrichers = addIfNotNil(brcp.config.Enrich.KubeMetadataEnricherConfig, enrichers, "kube_metadata")
	enrichers = addIfNotNil(brcp.config.Enrich.StaticEnricherConfig, enrichers, "static")
	return enrichers
}

func setEnrichingDefaults(v *viper.Viper) {
	if config.ShouldRootTakeDefaults(v, "enrich.kube_metadata") {
		kubeMetadataDefaults(v)
	}
	if config.ShouldRootTakeDefaults(v, "enrich.static") {
		staticDefaults(v)
	}
}

func kubeMetadataDefaults(v *viper.Viper) {
	config.SetDefault(v, "enrich.kube_metadata", enrichers.KubeMetadataEnricherConfig{
		Labels: map[string]string{
			// TODO: remove this once we have a unified label for workload display name
			"release":                        "workloadDisplayName",
			"training.kubeflow.org/job-name": "workloadDisplayName",
			"serving.knative.dev/service":    "workloadDisplayName",

			"project": "project",
		},
		Annotations: map[string]string{
			"user": "user",
		},
	})
}

func staticDefaults(v *viper.Viper) {
	config.SetDefault(v, "enrich.static", enrichers.StaticEnricherConfig{})
}

func addIfNotNil(cfg interface{}, values []string, value string) []string {
	if config.IsConfigFieldEmpty(cfg) {
		return values
	}
	return append(values, value)
}
