package types

type EnrichingConfig struct {
	KubeMetadataEnricherConfig *KubeMetadataEnricherConfig `mapstructure:"kube_metadata" json:"kube_metadata_enricher_config"`
	StaticEnricherConfig       *StaticEnricherConfig       `mapstructure:"static" json:"static_enricher_config"`
}

type StaticEnricherConfig map[string]string

type KubeMetadataEnricherConfig struct {
	// Maps kubernetes object labels to event properties
	Labels map[string]string `mapstructure:"labels" json:"labels"`
	// Maps kubernetes object annotations to event properties
	Annotations map[string]string `mapstructure:"annotations" json:"annotations"`
}
