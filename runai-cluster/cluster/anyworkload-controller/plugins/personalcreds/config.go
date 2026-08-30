package personalcreds

import commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"

type Config struct {
	Enabled        bool   `mapstructure:"enabled"`
	RunaiNamespace string `mapstructure:"runai_namespace"`
}

func (c *Config) SetDefaults() {
	c.Enabled = true
	c.RunaiNamespace = commonconstants.NamespaceRunAi
}
