package mnnvl

type Config struct {
	Enabled              bool `mapstructure:"ENABLE_MNNVL_SUPPORT"`
	ElasticComputeDomain bool `mapstructure:"ENABLE_ELASTIC_COMPUTE_DOMAIN"`
}
