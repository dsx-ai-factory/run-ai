package networking

type Config struct {
	Enabled bool `mapstructure:"enabled"`
}

func (c *Config) SetDefaults() {
	c.Enabled = true
}
