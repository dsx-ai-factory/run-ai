package storage

type Config struct {
	Enabled bool `mapstructure:"STORAGE_ENABLED"`
	// CleanupOnCompletionDefault is the default for whether the plugin deletes the PVCs can be override per runner
	CleanupOnCompletionDefault bool `mapstructure:"STORAGE_CLEANUP_ON_COMPLETION_DEFAULT"`
}
