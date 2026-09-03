/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package enrich

import "github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/enrich/enrichers"

type EnrichingConfigProvider interface {
	GetGlobalEnrichers() []string
	GetGlobalEnrichingConfig() *EnrichingConfig
}

type EnrichingConfig struct {
	KubeMetadataEnricherConfig *enrichers.KubeMetadataEnricherConfig `mapstructure:"kube_metadata" json:"kube_metadata_enricher_config"`
	StaticEnricherConfig       *enrichers.StaticEnricherConfig       `mapstructure:"static" json:"static_enricher_config"`
}
