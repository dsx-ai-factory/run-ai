package enrich

import "github.com/run-ai/runai/control-plane/notifications-service/internal/enrich/enrichers"

type EnrichingConfigProvider interface {
	GetGlobalEnrichers() []string
	GetGlobalEnrichingConfig() *EnrichingConfig
}

type EnrichingConfig struct {
	KubeMetadataEnricherConfig *enrichers.KubeMetadataEnricherConfig `mapstructure:"kube_metadata" json:"kube_metadata_enricher_config"`
	StaticEnricherConfig       *enrichers.StaticEnricherConfig       `mapstructure:"static" json:"static_enricher_config"`
}
